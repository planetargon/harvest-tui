package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/planetargon/argon-harvest-tui/internal/config"
	"github.com/planetargon/argon-harvest-tui/internal/domain"
	"github.com/planetargon/argon-harvest-tui/internal/harvest"
	"github.com/planetargon/argon-harvest-tui/internal/state"
)

// ViewState represents the different views in the TUI application.
type ViewState int

const (
	// ViewList is the main list view showing time entries for the current date.
	ViewList ViewState = iota
	// ViewSelectProject is the project selection view when creating a new time entry.
	ViewSelectProject
	// ViewSelectTask is the task selection view when creating a new time entry.
	ViewSelectTask
	// ViewEditEntry is the view for editing an existing time entry.
	ViewEditEntry
	// ViewConfirmDelete is the confirmation view when deleting a time entry.
	ViewConfirmDelete
	// ViewHelp is the help overlay showing all keybindings.
	ViewHelp
	// ViewNotesInput is the view for entering notes for a new time entry.
	ViewNotesInput
	// ViewDurationInput is the view for entering duration for a new time entry.
	ViewDurationInput
	// ViewBillableToggle is the view for toggling billable status for a new time entry.
	ViewBillableToggle
)

// Model represents the state of the TUI application.
type Model struct {
	// Current view state
	currentView ViewState

	// Configuration and external dependencies
	config        *config.Config
	harvestClient *harvest.Client
	appState      *state.State

	// Data
	currentDate        time.Time
	timeEntries        []harvest.TimeEntry
	projectsWithTasks  []harvest.ProjectWithTasks
	selectedEntryIndex int

	// New entry creation state
	selectedProject  *harvest.Project
	selectedTask     *harvest.Task
	newEntryNotes    string
	newEntryHours    string
	newEntryBillable bool

	// Edit entry state
	editingEntry *harvest.TimeEntry
	editNotes    string
	editHours    string
	editBillable bool

	// UI state
	loading       bool
	errorMessage  string
	statusMessage string
	showSpinner   bool

	// List components for selection views
	projectList list.Model
	taskList    list.Model

	// Text input components
	notesInput        *textinput.Model
	durationInput     *textinput.Model
	editNotesInput    *textinput.Model
	editDurationInput *textinput.Model

	// Window dimensions
	width  int
	height int
}

// NewModel creates a new TUI model with the given configuration.
func NewModel(cfg *config.Config, client *harvest.Client, appState *state.State) Model {
	return Model{
		currentView:        ViewList,
		config:             cfg,
		harvestClient:      client,
		appState:           appState,
		currentDate:        time.Now(),
		timeEntries:        []harvest.TimeEntry{},
		projectsWithTasks:  []harvest.ProjectWithTasks{},
		selectedEntryIndex: 0,
		newEntryNotes:      "",
		newEntryHours:      "",
		newEntryBillable:   true,
		editNotes:          "",
		editHours:          "",
		editBillable:       true,
		loading:            false,
		errorMessage:       "",
		statusMessage:      "",
		showSpinner:        false,
		projectList:        list.New([]list.Item{}, newProjectDelegate(), 0, 0),
		taskList:           list.New([]list.Item{}, newTaskDelegate(), 0, 0),
		width:              80,
		height:             24,
	}
}

// Init initializes the model and returns initial commands.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		fetchTimeEntriesCmd(m.harvestClient, m.currentDate),
		fetchProjectsWithTasksCmd(m.harvestClient),
		tickCmd(), // Start the ticker for real-time updates
	)
}

// Update handles messages and updates the model state.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.projectList.SetSize(msg.Width, msg.Height-8)
		m.taskList.SetSize(msg.Width, msg.Height-8)
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case timeEntriesFetchedMsg:
		if msg.err != nil {
			m.errorMessage = "Failed to fetch time entries: " + msg.err.Error()
		} else {
			m.timeEntries = msg.entries
			m.errorMessage = ""
		}
		m.loading = false
		
		// If there's a running timer, continue ticking
		if m.hasRunningTimer() {
			return m, tickCmd()
		}
		return m, nil
	
	case tickMsg:
		// Check if we have a running timer
		if m.hasRunningTimer() && m.currentView == ViewList && !m.loading {
			// Fetch updated time entries to get the latest hours
			return m, tea.Batch(
				fetchTimeEntriesCmd(m.harvestClient, m.currentDate),
				tickCmd(), // Continue ticking
			)
		}
		// If no running timer or not in list view, just continue ticking
		return m, tickCmd()

	case projectsWithTasksFetchedMsg:
		if msg.err != nil {
			m.errorMessage = "Failed to fetch projects: " + msg.err.Error()
		} else {
			m.projectsWithTasks = msg.projectsWithTasks
			m.errorMessage = ""
		}
		return m, nil

	case timeEntryStartedMsg:
		if msg.err != nil {
			m.statusMessage = "Failed to start timer: " + msg.err.Error()
		} else {
			// Update the entry in our local list
			for i, entry := range m.timeEntries {
				if entry.ID == msg.entry.ID {
					m.timeEntries[i] = *msg.entry
					break
				}
			}
			m.statusMessage = "Timer started successfully"
			// Start ticking for real-time updates
			return m, tickCmd()
		}
		return m, nil

	case timeEntryStoppedMsg:
		if msg.err != nil {
			m.statusMessage = "Failed to stop timer: " + msg.err.Error()
		} else {
			// Update the entry in our local list
			for i, entry := range m.timeEntries {
				if entry.ID == msg.entry.ID {
					m.timeEntries[i] = *msg.entry
					break
				}
			}
			m.statusMessage = "Timer stopped successfully"
		}
		return m, nil

	case timeEntryCreatedMsg:
		if msg.err != nil {
			m.statusMessage = "Failed to create entry: " + msg.err.Error()
		} else {
			// Add the new entry to our local list
			m.timeEntries = append([]harvest.TimeEntry{*msg.entry}, m.timeEntries...)
			m.statusMessage = "Time entry created successfully"
			// Clear new entry state and return to main list
			m.clearEditState()
			m.currentView = ViewList
		}
		return m, nil

	case timeEntryUpdatedMsg:
		if msg.err != nil {
			m.statusMessage = "Failed to update entry: " + msg.err.Error()
		} else {
			// Update the entry in our local list
			for i, entry := range m.timeEntries {
				if entry.ID == msg.entry.ID {
					m.timeEntries[i] = *msg.entry
					break
				}
			}
			m.statusMessage = "Time entry updated successfully"
			// Clear edit state and return to main list
			m.clearEditState()
			m.currentView = ViewList
		}
		return m, nil

	case timeEntryDeletedMsg:
		if msg.err != nil {
			m.statusMessage = "Failed to delete entry: " + msg.err.Error()
		} else {
			// Remove the entry from our local list
			newEntries := []harvest.TimeEntry{}
			for _, entry := range m.timeEntries {
				if entry.ID != msg.entryID {
					newEntries = append(newEntries, entry)
				}
			}
			m.timeEntries = newEntries

			// Adjust selected index if necessary
			if m.selectedEntryIndex >= len(m.timeEntries) && m.selectedEntryIndex > 0 {
				m.selectedEntryIndex--
			}

			m.statusMessage = "Time entry deleted successfully"
			// Clear edit state and return to main list
			m.clearEditState()
			m.currentView = ViewList
		}
		return m, nil

	default:
		return m, nil
	}
}

// View renders the current view.
func (m Model) View() string {
	switch m.currentView {
	case ViewList:
		return m.renderStyledListView()
	case ViewSelectProject:
		return m.renderProjectSelectView()
	case ViewSelectTask:
		return m.renderTaskSelectView()
	case ViewEditEntry:
		return m.renderEditView()
	case ViewConfirmDelete:
		return m.renderConfirmDeleteView()
	case ViewHelp:
		return m.renderHelpView()
	case ViewNotesInput:
		return m.renderNotesInputView()
	case ViewDurationInput:
		return m.renderDurationInputView()
	case ViewBillableToggle:
		return m.renderBillableToggleView()
	default:
		return "Unknown view"
	}
}

// handleKeyPress processes key presses for the current view.
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global keybindings that work in all views
	switch msg.String() {
	case "ctrl+c":
		// Ctrl+C always quits
		return m, tea.Quit
	case "?":
		if m.currentView == ViewHelp {
			m.currentView = ViewList
		} else {
			m.currentView = ViewHelp
		}
		return m, nil
	}

	// View-specific keybindings
	switch m.currentView {
	case ViewList:
		return m.handleListViewKeys(msg)
	case ViewSelectProject:
		return m.handleProjectSelectKeys(msg)
	case ViewSelectTask:
		return m.handleTaskSelectKeys(msg)
	case ViewEditEntry:
		return m.handleEditViewKeys(msg)
	case ViewConfirmDelete:
		return m.handleConfirmDeleteKeys(msg)
	case ViewHelp:
		return m.handleHelpViewKeys(msg)
	case ViewNotesInput:
		return m.handleNotesInputKeys(msg)
	case ViewDurationInput:
		return m.handleDurationInputKeys(msg)
	case ViewBillableToggle:
		return m.handleBillableToggleKeys(msg)
	}

	return m, nil
}

// clearEditState resets the editing and new entry state.
func (m *Model) clearEditState() {
	m.selectedProject = nil
	m.selectedTask = nil
	m.newEntryNotes = ""
	m.newEntryHours = ""
	m.newEntryBillable = true
	m.notesInput = nil
	m.durationInput = nil
	m.editingEntry = nil
	m.editNotes = ""
	m.editHours = ""
	m.editBillable = true
}

// renderListView renders the main list view showing time entries for the current date.
func (m Model) renderListView() string {
	// Calculate dimensions
	width := 65 // Fixed width for consistent layout
	if m.width > 0 {
		width = min(m.width-2, 80) // Cap at 80 chars wide
	}

	// Format date navigation
	dateStr := m.currentDate.Format("Mon, Jan 2, 2006")
	dateNav := fmt.Sprintf("◀ %s ▶", dateStr)

	// Title bar with date navigation
	titleBar := lipgloss.JoinHorizontal(
		lipgloss.Left,
		"  Harvest Time Tracker",
		lipgloss.NewStyle().Width(width-26-len(dateNav)).Render(""),
		dateNav,
		"  ",
	)

	// Calculate daily total
	totalHours := 0.0
	for _, entry := range m.timeEntries {
		totalHours += entry.Hours
	}
	totalStr := formatHoursSimple(totalHours)

	// Section header
	entriesText := "  Entries"
	totalText := "Total: " + totalStr + "  "
	paddingWidth := width - len(entriesText) - len(totalText)
	if paddingWidth < 1 {
		paddingWidth = 1
	}
	sectionHeader := entriesText + strings.Repeat(" ", paddingWidth) + totalText

	// Divider
	divider := "  " + strings.Repeat("─", width-4)

	// Handle loading state
	if m.loading {
		content := []string{
			titleBar,
			sectionHeader,
			divider,
			"    Loading...",
			"",
		}
		return m.wrapInBox(strings.Join(content, "\n"), width)
	}

	// Handle error state
	if m.errorMessage != "" {
		content := []string{
			titleBar,
			sectionHeader,
			divider,
			"    Error: " + m.errorMessage,
			"",
		}
		return m.wrapInBox(strings.Join(content, "\n"), width)
	}

	// Handle empty state
	if len(m.timeEntries) == 0 {
		content := []string{
			titleBar,
			sectionHeader,
			divider,
			"",
			"    No time entries for this date",
			"    Press 'n' to create a new entry",
			"",
		}
		return m.wrapInBox(strings.Join(content, "\n"), width)
	}

	// Render time entries
	var entryLines []string
	for i, entry := range m.timeEntries {
		isSelected := i == m.selectedEntryIndex
		entryLines = append(entryLines, m.renderCompactTimeEntry(entry, isSelected, width-4))
	}

	// Build content
	contentLines := []string{titleBar, sectionHeader, divider}
	contentLines = append(contentLines, entryLines...)

	// Add status message if present
	if m.statusMessage != "" {
		contentLines = append(contentLines, "", "  "+m.statusMessage)
	}

	return m.wrapInBox(strings.Join(contentLines, "\n"), width)
}

// wrapInBox wraps content in a box border with footer.
func (m Model) wrapInBox(content string, width int) string {
	// Top border
	top := "┌" + strings.Repeat("─", width-2) + "┐"

	// Content lines with side borders
	lines := strings.Split(content, "\n")
	var boxedLines []string
	boxedLines = append(boxedLines, top)

	for _, line := range lines {
		// Pad line to width
		padded := line
		if len(line) < width-2 {
			padded = line + strings.Repeat(" ", width-2-len(line))
		} else if len(line) > width-2 {
			padded = line[:width-2]
		}
		boxedLines = append(boxedLines, "│"+padded+"│")
	}

	// Footer with keybindings
	footerSeparator := "├" + strings.Repeat("─", width-2) + "┤"
	footer := "  n new  e edit  s start/stop  d delete  ? help  q quit"
	if len(footer) > width-2 {
		footer = footer[:width-2]
	}
	footerPadded := footer + strings.Repeat(" ", width-2-len(footer))

	// Bottom border
	bottom := "└" + strings.Repeat("─", width-2) + "┘"

	boxedLines = append(boxedLines, footerSeparator)
	boxedLines = append(boxedLines, "│"+footerPadded+"│")
	boxedLines = append(boxedLines, bottom)

	return strings.Join(boxedLines, "\n")
}

// renderCompactTimeEntry renders a single time entry in compact format.
func (m Model) renderCompactTimeEntry(entry harvest.TimeEntry, isSelected bool, maxWidth int) string {
	var lines []string

	// Status indicators
	statusIcon := "  "
	if isSelected {
		statusIcon = "▶ "
	}

	timerIcon := ""
	if entry.IsRunning {
		timerIcon = " ⏱️"
	}
	if entry.IsLocked {
		timerIcon = " 🔒"
	}

	// Format hours
	hoursText := formatHoursSimple(entry.Hours)

	// First line: Client → Project → Task + hours
	clientName := truncateString(entry.Client.Name, 20)
	projectName := truncateString(entry.Project.Name, 25)
	taskName := truncateString(entry.Task.Name, 20)

	firstLine := fmt.Sprintf("%s%s → %s → %s",
		statusIcon,
		clientName,
		projectName,
		taskName,
	)

	// Calculate padding for hours alignment
	padding := maxWidth - len(firstLine) - len(hoursText) - len(timerIcon) - 2
	if padding < 1 {
		padding = 1
	}

	firstLine = firstLine + strings.Repeat(" ", padding) + hoursText + timerIcon
	lines = append(lines, firstLine)

	// Second line: Notes (if any)
	if entry.Notes != "" {
		notesLine := "    \"" + truncateString(entry.Notes, maxWidth-8) + "\""
		lines = append(lines, notesLine)
	}

	// Add blank line after entry (except for loading/error states)
	lines = append(lines, "")

	return strings.Join(lines, "\n")
}

// formatHoursSimple formats hours as H:MM format.
func formatHoursSimple(hours float64) string {
	h := int(hours)
	m := int((hours - float64(h)) * 60)
	return fmt.Sprintf("%d:%02d", h, m)
}

// truncateString truncates a string to the given max length, adding "..." if truncated.
func truncateString(s string, maxLen int) string {
	if maxLen <= 3 || len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// min returns the minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// projectItem represents a project in the selection list.
type projectItem struct {
	project harvest.Project
	client  harvest.ProjectClient
}

func (i projectItem) FilterValue() string {
	return i.project.Name + " " + i.client.Name
}

func (i projectItem) Title() string {
	clientName := truncateString(i.client.Name, 25)
	projectName := truncateString(i.project.Name, 35)
	return clientName + " → " + projectName
}

func (i projectItem) Description() string {
	return "Project ID: " + fmt.Sprintf("%d", i.project.ID)
}

// dividerItem represents a divider in the selection list.
type dividerItem struct{}

func (i dividerItem) FilterValue() string {
	return ""
}

func (i dividerItem) Title() string {
	return "─────────────────────────────────────"
}

func (i dividerItem) Description() string {
	return ""
}

// taskItem represents a task in the selection list.

type taskItem struct {
	task harvest.Task
}

func (i taskItem) FilterValue() string {
	return i.task.Name
}

func (i taskItem) Title() string {
	return truncateString(i.task.Name, 50)
}

func (i taskItem) Description() string {
	return "Task ID: " + fmt.Sprintf("%d", i.task.ID)
}

// newProjectDelegate creates a new delegate for project list items.
func newProjectDelegate() list.DefaultDelegate {
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true
	return delegate
}

// newTaskDelegate creates a new delegate for task list items.
func newTaskDelegate() list.DefaultDelegate {
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true
	return delegate
}

// updateProjectList updates the project list with current projects and recents.
func (m *Model) updateProjectList() {
	var items []list.Item
	recentsAdded := 0

	// Add recents section first
	if len(m.appState.Recents) > 0 {
		for _, recent := range m.appState.Recents {
			// Find the matching project and client
			for _, pwt := range m.projectsWithTasks {
				if pwt.Project.ID == recent.ProjectID && pwt.Project.Client.ID == recent.ClientID {
					items = append(items, projectItem{
						project: pwt.Project,
						client:  pwt.Project.Client,
					})
					recentsAdded++
					break
				}
			}
		}

		// Add divider after recents only if we actually added any
		if recentsAdded > 0 {
			items = append(items, dividerItem{})
		}
	}

	// Add all projects sorted by client then project name
	var nonRecentProjects []harvest.ProjectWithTasks
	for _, pwt := range m.projectsWithTasks {
		// Skip if already in recents
		isRecent := false
		for _, recent := range m.appState.Recents {
			if pwt.Project.ID == recent.ProjectID && pwt.Project.Client.ID == recent.ClientID {
				isRecent = true
				break
			}
		}
		if !isRecent {
			nonRecentProjects = append(nonRecentProjects, pwt)
		}
	}

	// Sort non-recent projects by client name then project name
	sort.Slice(nonRecentProjects, func(i, j int) bool {
		if nonRecentProjects[i].Project.Client.Name != nonRecentProjects[j].Project.Client.Name {
			return nonRecentProjects[i].Project.Client.Name < nonRecentProjects[j].Project.Client.Name
		}
		return nonRecentProjects[i].Project.Name < nonRecentProjects[j].Project.Name
	})

	// Add sorted non-recent projects to items
	for _, pwt := range nonRecentProjects {
		items = append(items, projectItem{
			project: pwt.Project,
			client:  pwt.Project.Client,
		})
	}

	m.projectList.SetItems(items)
	m.projectList.Title = "Select Project"
}

// updateTaskList updates the task list with tasks from the selected project.
func (m *Model) updateTaskList(tasks []harvest.Task) {
	var items []list.Item
	for _, task := range tasks {
		items = append(items, taskItem{task: task})
	}
	m.taskList.SetItems(items)
	m.taskList.Title = "Select Task"
}

func (m Model) renderProjectSelectView() string {
	styles := DefaultStyles()

	// Header
	header := styles.Header.Render(
		lipgloss.JoinHorizontal(lipgloss.Left,
			styles.Title.Render("New Time Entry"),
			"  ",
			styles.Subtitle.Render("Step 1: Choose Project"),
		),
	)

	// Instructions
	instructions := styles.SecondaryText.Render("Press Enter to select • Esc to cancel • / to filter")

	// Render the list
	listView := m.projectList.View()

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		instructions,
		"",
		listView,
	)
}

func (m Model) renderTaskSelectView() string {
	styles := DefaultStyles()

	// Header
	header := styles.Header.Render(
		lipgloss.JoinHorizontal(lipgloss.Left,
			styles.Title.Render("New Time Entry"),
			"  ",
			styles.Subtitle.Render("Step 2: Choose Task"),
		),
	)

	// Show selected project
	projectInfo := ""
	if m.selectedProject != nil {
		projectInfo = styles.SecondaryText.Render(
			fmt.Sprintf("Project: %s → %s",
				m.selectedProject.Client.Name,
				m.selectedProject.Name))
	}

	// Instructions
	instructions := styles.SecondaryText.Render("Press Enter to select • Esc to go back • / to filter")

	// Render the list
	listView := m.taskList.View()

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		projectInfo,
		"",
		instructions,
		"",
		listView,
	)
}

func (m Model) renderEditView() string {
	styles := DefaultStyles()

	// Header
	header := styles.Header.Render(
		lipgloss.JoinHorizontal(lipgloss.Left,
			styles.Title.Render("Edit Time Entry"),
		),
	)

	// Entry info
	var info string
	if m.editingEntry != nil {
		info = styles.SecondaryText.Render(fmt.Sprintf(
			"Editing entry from %s",
			m.currentDate.Format("Monday, January 2, 2006"),
		))
	}

	// Instructions
	instructions := styles.SecondaryText.Render("Tab to navigate fields • Enter to save • Esc to cancel")

	// Fields (placeholder for now)
	fields := lipgloss.JoinVertical(lipgloss.Left,
		"Notes: "+m.editNotes,
		"Duration: "+m.editHours,
		fmt.Sprintf("Billable: %v", m.editBillable),
	)

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		info,
		"",
		fields,
		"",
		instructions,
	)
}

func (m Model) renderConfirmDeleteView() string {
	styles := DefaultStyles()

	// Header
	header := styles.Header.Render(
		lipgloss.JoinHorizontal(lipgloss.Left,
			styles.Title.Render("Confirm Delete"),
		),
	)

	// Entry details
	var details string
	if m.editingEntry != nil {
		details = lipgloss.JoinVertical(lipgloss.Left,
			styles.Title.Render("Are you sure you want to delete this entry?"),
			"",
			styles.SecondaryText.Render(fmt.Sprintf("Notes: %s", m.editingEntry.Notes)),
			styles.SecondaryText.Render(fmt.Sprintf("Duration: %s", formatHoursSimple(m.editingEntry.Hours))),
		)
	}

	// Instructions
	instructions := styles.WarningText.Render("Press Y to confirm deletion • N or Esc to cancel")

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		details,
		"",
		instructions,
	)
}

func (m Model) renderHelpView() string {
	styles := DefaultStyles()

	// Header
	header := styles.Header.Render(
		lipgloss.JoinHorizontal(lipgloss.Left,
			styles.Title.Render("Help"),
		),
	)

	// Keybindings organized by category
	navigation := lipgloss.JoinVertical(lipgloss.Left,
		styles.Subtitle.Render("Navigation"),
		"  ↑/k       Move up",
		"  ↓/j       Move down",
		"  ←/h       Previous day",
		"  →/l       Next day",
		"  t         Jump to today",
	)

	entryActions := lipgloss.JoinVertical(lipgloss.Left,
		styles.Subtitle.Render("Time Entry Actions"),
		"  n         New entry",
		"  e         Edit entry",
		"  d         Delete entry",
		"  s         Start/stop timer",
	)

	general := lipgloss.JoinVertical(lipgloss.Left,
		styles.Subtitle.Render("General"),
		"  ?         Toggle this help",
		"  q/Esc     Quit/Go back",
		"  Ctrl+C    Force quit",
	)

	// Instructions
	instructions := styles.SecondaryText.Render("Press ? or Esc to close help")

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		navigation,
		"",
		entryActions,
		"",
		general,
		"",
		instructions,
	)
}

func (m Model) renderNotesInputView() string {
	styles := DefaultStyles()

	// Header
	header := styles.Header.Render(
		lipgloss.JoinHorizontal(lipgloss.Left,
			styles.Title.Render("New Time Entry"),
			"  ",
			styles.Subtitle.Render("Step 3: Enter Notes"),
		),
	)

	// Selected project and task info
	var info string
	if m.selectedProject != nil && m.selectedTask != nil {
		info = styles.SecondaryText.Render(fmt.Sprintf(
			"%s → %s → %s",
			m.selectedProject.Client.Name,
			m.selectedProject.Name,
			m.selectedTask.Name,
		))
	}

	// Instructions
	instructions := styles.SecondaryText.Render("Press Enter to continue • Esc to cancel")

	// Input field
	var inputView string
	if m.notesInput != nil {
		inputView = m.notesInput.View()
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		info,
		"",
		inputView,
		"",
		instructions,
	)
}

func (m Model) renderDurationInputView() string {
	styles := DefaultStyles()

	// Header
	header := styles.Header.Render(
		lipgloss.JoinHorizontal(lipgloss.Left,
			styles.Title.Render("New Time Entry"),
			"  ",
			styles.Subtitle.Render("Step 4: Enter Duration"),
		),
	)

	// Selected info
	var info string
	if m.selectedProject != nil && m.selectedTask != nil {
		info = styles.SecondaryText.Render(fmt.Sprintf(
			"%s → %s → %s",
			m.selectedProject.Client.Name,
			m.selectedProject.Name,
			m.selectedTask.Name,
		))
	}

	// Notes info
	notesInfo := ""
	if m.newEntryNotes != "" {
		notesInfo = styles.SecondaryText.Render(fmt.Sprintf("Notes: %s", m.newEntryNotes))
	}

	// Instructions
	instructions := styles.SecondaryText.Render("Enter duration in HH:MM format • Press Enter to continue • Esc to go back")

	// Input field
	var inputView string
	if m.durationInput != nil {
		inputView = m.durationInput.View()
	}

	// Status message if any
	statusMsg := ""
	if m.statusMessage != "" {
		statusMsg = styles.ErrorText.Render(m.statusMessage)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		info,
		notesInfo,
		"",
		inputView,
		"",
		instructions,
		statusMsg,
	)
}

func (m Model) renderBillableToggleView() string {
	styles := DefaultStyles()

	// Header
	header := styles.Header.Render(
		lipgloss.JoinHorizontal(lipgloss.Left,
			styles.Title.Render("New Time Entry"),
			"  ",
			styles.Subtitle.Render("Step 5: Billable Status"),
		),
	)

	// Selected info
	var info string
	if m.selectedProject != nil && m.selectedTask != nil {
		info = styles.SecondaryText.Render(fmt.Sprintf(
			"%s → %s → %s",
			m.selectedProject.Client.Name,
			m.selectedProject.Name,
			m.selectedTask.Name,
		))
	}

	// Entry details
	details := []string{}
	if m.newEntryNotes != "" {
		details = append(details, fmt.Sprintf("Notes: %s", m.newEntryNotes))
	}
	if m.newEntryHours != "" {
		details = append(details, fmt.Sprintf("Duration: %s", m.newEntryHours))
	}
	detailsView := styles.SecondaryText.Render(lipgloss.JoinVertical(lipgloss.Left, details...))

	// Billable toggle
	billableStatus := "[ ] Non-billable"
	if m.newEntryBillable {
		billableStatus = "[x] Billable"
	}
	billableView := styles.SecondaryText.Render(billableStatus)

	// Instructions
	instructions := styles.SecondaryText.Render("Press Space/Tab/B to toggle • Enter to create entry • Esc to go back")

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		info,
		"",
		detailsView,
		"",
		billableView,
		"",
		instructions,
	)
}

// handleListViewKeys handles key presses in the main list view.
func (m Model) handleListViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keys := DefaultKeyMap()

	// Check for escape/quit first
	switch msg.String() {
	case "esc", "q":
		return m, tea.Quit
	}

	switch {
	case key.Matches(msg, keys.Up):
		if len(m.timeEntries) > 0 && m.selectedEntryIndex > 0 {
			m.selectedEntryIndex--
		}
		return m, nil

	case key.Matches(msg, keys.Down):
		if len(m.timeEntries) > 0 && m.selectedEntryIndex < len(m.timeEntries)-1 {
			m.selectedEntryIndex++
		}
		return m, nil

	case key.Matches(msg, keys.PrevDay):
		m.currentDate = m.currentDate.AddDate(0, 0, -1)
		m.selectedEntryIndex = 0
		m.loading = true
		return m, fetchTimeEntriesCmd(m.harvestClient, m.currentDate)

	case key.Matches(msg, keys.NextDay):
		m.currentDate = m.currentDate.AddDate(0, 0, 1)
		m.selectedEntryIndex = 0
		m.loading = true
		return m, fetchTimeEntriesCmd(m.harvestClient, m.currentDate)

	case key.Matches(msg, keys.Today):
		m.currentDate = time.Now()
		m.selectedEntryIndex = 0
		m.loading = true
		return m, fetchTimeEntriesCmd(m.harvestClient, m.currentDate)

	case key.Matches(msg, keys.New):
		if len(m.projectsWithTasks) > 0 {
			m.currentView = ViewSelectProject
			m.clearEditState()
			m.updateProjectList()
			m.projectList.SetSize(m.width, m.height-8) // Account for header and instructions
			return m, nil
		} else {
			m.statusMessage = "No projects available. Please check your Harvest configuration."
			return m, nil
		}

	case key.Matches(msg, keys.Edit):
		if len(m.timeEntries) > 0 && m.selectedEntryIndex < len(m.timeEntries) {
			selectedEntry := m.timeEntries[m.selectedEntryIndex]
			if selectedEntry.IsLocked {
				m.statusMessage = "Cannot edit locked time entry."
				return m, nil
			}
			if selectedEntry.IsRunning {
				m.statusMessage = "Cannot edit running time entry. Stop the timer first."
				return m, nil
			}
			m.editingEntry = &selectedEntry
			m.editNotes = selectedEntry.Notes
			m.editHours = formatHoursSimple(selectedEntry.Hours)
			m.editBillable = selectedEntry.IsBillable
			m.currentView = ViewEditEntry
			return m, nil
		}
		return m, nil

	case key.Matches(msg, keys.Delete):
		if len(m.timeEntries) > 0 && m.selectedEntryIndex < len(m.timeEntries) {
			selectedEntry := m.timeEntries[m.selectedEntryIndex]
			if selectedEntry.IsLocked {
				m.statusMessage = "Cannot delete locked time entry."
				return m, nil
			}
			if selectedEntry.IsRunning {
				m.statusMessage = "Cannot delete running time entry. Stop the timer first."
				return m, nil
			}
			m.editingEntry = &selectedEntry
			m.currentView = ViewConfirmDelete
			return m, nil
		}
		return m, nil

	case key.Matches(msg, keys.StartStop):
		if len(m.timeEntries) > 0 && m.selectedEntryIndex < len(m.timeEntries) {
			selectedEntry := m.timeEntries[m.selectedEntryIndex]
			if selectedEntry.IsLocked {
				if selectedEntry.IsRunning {
					m.statusMessage = "Cannot stop locked time entry."
				} else {
					m.statusMessage = "Cannot start locked time entry."
				}
				return m, nil
			}
			// Toggle: if running, stop it; if stopped, start it
			if selectedEntry.IsRunning {
				return m, stopTimeEntryCmd(m.harvestClient, selectedEntry.ID)
			} else {
				return m, restartTimeEntryCmd(m.harvestClient, selectedEntry.ID)
			}
		}
		return m, nil
	}

	return m, nil
}

func (m Model) handleProjectSelectKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "esc":
		// Cancel project selection and return to main list
		m.currentView = ViewList
		m.selectedProject = nil
		m.selectedTask = nil
		return m, nil
	case "enter":
		// Get the selected project
		selected := m.projectList.SelectedItem()
		if selected != nil {
			// Skip divider items
			if _, ok := selected.(dividerItem); ok {
				// Move to next item
				m.projectList.CursorDown()
				return m, nil
			}

			if item, ok := selected.(projectItem); ok {
				m.selectedProject = &item.project

				// Check if this is a recent entry with a task already selected
				var recentTaskID *int
				selectedIndex := m.projectList.Index()
				if selectedIndex < len(m.appState.Recents) {
					// This is a recent entry (recents appear first in the list)
					for _, recent := range m.appState.Recents {
						if recent.ProjectID == item.project.ID && recent.ClientID == item.client.ID {
							recentTaskID = &recent.TaskID
							break
						}
					}
				}

				// Find tasks for this project
				for _, pwt := range m.projectsWithTasks {
					if pwt.Project.ID == item.project.ID {
						if len(pwt.Tasks) == 0 {
							// No tasks available for this project
							m.statusMessage = "No tasks available for this project"
							m.selectedProject = nil
							return m, nil
						}

						// If this is a recent with a task, try to use it
						if recentTaskID != nil {
							for i := range pwt.Tasks {
								if pwt.Tasks[i].ID == *recentTaskID {
									// Found the task from the recent
									m.selectedTask = &pwt.Tasks[i]
									// Initialize notes input
									notesInput := textinput.New()
									notesInput.Focus()
									notesInput.Placeholder = "Enter notes (optional)"
									notesInput.Width = 50
									m.notesInput = &notesInput
									m.currentView = ViewNotesInput
									return m, nil
								}
							}
							// Task from recent not found, fall through to show task selection
						}

						if len(pwt.Tasks) == 1 {
							// Only one task, skip task selection
							m.selectedTask = &pwt.Tasks[0]
							// Initialize notes input
							notesInput := textinput.New()
							notesInput.Focus()
							notesInput.Placeholder = "Enter notes (optional)"
							notesInput.Width = 50
							m.notesInput = &notesInput
							m.currentView = ViewNotesInput
						} else {
							// Multiple tasks, show task selection
							m.currentView = ViewSelectTask
							m.updateTaskList(pwt.Tasks)
						}
						break
					}
				}
			}
		}
		return m, nil
	}

	// Handle list navigation
	m.projectList, cmd = m.projectList.Update(msg)
	return m, cmd
}

func (m Model) handleTaskSelectKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "esc":
		// Go back to project selection
		m.currentView = ViewSelectProject
		m.selectedProject = nil
		m.selectedTask = nil
		return m, nil
	case "enter":
		// Get the selected task
		selected := m.taskList.SelectedItem()
		if selected != nil {
			if item, ok := selected.(taskItem); ok {
				m.selectedTask = &item.task
				// Initialize notes input
				notesInput := textinput.New()
				notesInput.Focus()
				notesInput.Placeholder = "Enter notes (optional)"
				notesInput.Width = 50
				m.notesInput = &notesInput
				m.currentView = ViewNotesInput
			}
		}
		return m, nil
	}

	// Handle list navigation
	m.taskList, cmd = m.taskList.Update(msg)
	return m, cmd
}

func (m Model) handleEditViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Return to main list, clearing all edit and entry state
		m.currentView = ViewList
		m.clearEditState()
		return m, nil
	}
	return m, nil
}

func (m Model) handleConfirmDeleteKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "n":
		// Cancel deletion and return to main list
		m.currentView = ViewList
		m.editingEntry = nil
		return m, nil
	case "y":
		// Confirm deletion
		if m.editingEntry != nil {
			return m, deleteTimeEntryCmd(m.harvestClient, m.editingEntry.ID)
		}
		return m, nil
	}
	return m, nil
}

func (m Model) handleHelpViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q", "?":
		// Return to main list
		m.currentView = ViewList
		return m, nil
	}
	return m, nil
}

func (m Model) handleNotesInputKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "esc":
		// Cancel and return to main list
		m.currentView = ViewList
		m.selectedProject = nil
		m.selectedTask = nil
		m.notesInput = nil
		return m, nil
	case "enter":
		// Store notes and move to duration input
		if m.notesInput != nil {
			m.newEntryNotes = m.notesInput.Value()
		}
		// Initialize duration input
		durationInput := textinput.New()
		durationInput.Focus()
		durationInput.Placeholder = "Enter duration (e.g., 1:30)"
		durationInput.Width = 20
		m.durationInput = &durationInput
		m.currentView = ViewDurationInput
		return m, nil
	}

	// Pass other messages to the text input
	if m.notesInput != nil {
		*m.notesInput, cmd = m.notesInput.Update(msg)
	}
	return m, cmd
}

func (m Model) handleDurationInputKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "esc":
		// Go back to notes input
		m.currentView = ViewNotesInput
		m.durationInput = nil
		return m, nil
	case "enter":
		// Validate and store duration
		if m.durationInput != nil {
			duration := m.durationInput.Value()
			if duration == "" {
				duration = "0:00"
			}
			// Validate duration format
			if _, err := domain.ParseDuration(duration); err != nil {
				m.statusMessage = "Invalid duration format. Use HH:MM (e.g., 1:30)"
				return m, nil
			}
			m.newEntryHours = duration
			m.newEntryBillable = true // Default to billable
			m.currentView = ViewBillableToggle
		}
		return m, nil
	}

	// Pass other messages to the text input
	if m.durationInput != nil {
		*m.durationInput, cmd = m.durationInput.Update(msg)
	}
	return m, cmd
}

func (m Model) handleBillableToggleKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Go back to duration input
		m.currentView = ViewDurationInput
		return m, nil
	case "tab", " ", "b":
		// Toggle billable status
		m.newEntryBillable = !m.newEntryBillable
		return m, nil
	case "enter":
		// Create the time entry
		return m, m.createTimeEntry()
	}
	return m, nil
}

// Messages for handling async operations
type timeEntriesFetchedMsg struct {
	entries []harvest.TimeEntry
	err     error
}

// tickMsg is sent periodically to update running timers
type tickMsg time.Time

type projectsWithTasksFetchedMsg struct {
	projectsWithTasks []harvest.ProjectWithTasks
	err               error
}

type timeEntryStartedMsg struct {
	entry *harvest.TimeEntry
	err   error
}

type timeEntryStoppedMsg struct {
	entry *harvest.TimeEntry
	err   error
}

type timeEntryCreatedMsg struct {
	entry *harvest.TimeEntry
	err   error
}

type timeEntryUpdatedMsg struct {
	entry *harvest.TimeEntry
	err   error
}

type timeEntryDeletedMsg struct {
	entryID int
	err     error
}

// Commands for fetching data
func fetchTimeEntriesCmd(client *harvest.Client, date time.Time) tea.Cmd {
	return func() tea.Msg {
		dateStr := date.Format("2006-01-02")
		entries, err := client.FetchTimeEntries(dateStr)
		return timeEntriesFetchedMsg{entries: entries, err: err}
	}
}

func fetchProjectsWithTasksCmd(client *harvest.Client) tea.Cmd {
	return func() tea.Msg {
		// Fetch projects and task assignments, then aggregate them
		projects, err := client.FetchProjects()
		if err != nil {
			return projectsWithTasksFetchedMsg{err: err}
		}

		taskAssignments, err := client.FetchTaskAssignments()
		if err != nil {
			return projectsWithTasksFetchedMsg{err: err}
		}

		projectsWithTasks := harvest.AggregateProjectsWithTasks(projects, taskAssignments)
		return projectsWithTasksFetchedMsg{projectsWithTasks: projectsWithTasks}
	}
}

func restartTimeEntryCmd(client *harvest.Client, entryID int) tea.Cmd {
	return func() tea.Msg {
		entry, err := client.RestartTimeEntry(entryID)
		return timeEntryStartedMsg{entry: entry, err: err}
	}
}

func stopTimeEntryCmd(client *harvest.Client, entryID int) tea.Cmd {
	return func() tea.Msg {
		entry, err := client.StopTimeEntry(entryID)
		return timeEntryStoppedMsg{entry: entry, err: err}
	}
}

func deleteTimeEntryCmd(client *harvest.Client, entryID int) tea.Cmd {
	return func() tea.Msg {
		err := client.DeleteTimeEntry(entryID)
		return timeEntryDeletedMsg{entryID: entryID, err: err}
	}
}

// tickCmd returns a command that sends a tick message after a delay
func tickCmd() tea.Cmd {
	return tea.Tick(30*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// hasRunningTimer checks if any time entry has a running timer
func (m Model) hasRunningTimer() bool {
	for _, entry := range m.timeEntries {
		if entry.IsRunning {
			return true
		}
	}
	return false
}

// createTimeEntry creates a new time entry and returns a command
func (m Model) createTimeEntry() tea.Cmd {
	if m.selectedProject == nil || m.selectedTask == nil {
		return nil
	}

	// Parse duration
	hours, err := domain.ParseDuration(m.newEntryHours)
	if err != nil {
		return nil
	}

	request := harvest.CreateTimeEntryRequest{
		ProjectID:  m.selectedProject.ID,
		TaskID:     m.selectedTask.ID,
		SpentDate:  m.currentDate.Format("2006-01-02"),
		Hours:      hours,
		Notes:      m.newEntryNotes,
		IsBillable: &m.newEntryBillable,
	}

	return func() tea.Msg {
		entry, err := m.harvestClient.CreateTimeEntry(request)
		if err != nil {
			return timeEntryCreatedMsg{err: err}
		}

		// Update recents
		m.appState.AddRecent(
			m.selectedProject.Client.ID,
			m.selectedProject.ID,
			m.selectedTask.ID,
		)
		if saveErr := m.appState.Save(); saveErr != nil {
			// Log error but don't fail the operation
			fmt.Printf("Failed to save recents: %v\n", saveErr)
		}

		return timeEntryCreatedMsg{entry: entry}
	}
}
