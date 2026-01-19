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
	// ViewNewEntry is the unified new entry form view.
	ViewNewEntry
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
	selectedProject      *harvest.Project
	selectedTask         *harvest.Task
	newEntryNotes        string
	newEntryHours        string
	newEntryBillable     bool
	newEntryCurrentField int // 0=project, 1=task, 2=notes, 3=duration, 4=billable

	// Edit entry state
	editingEntry     *harvest.TimeEntry
	editNotes        string
	editHours        string
	editBillable     bool
	editCurrentField int // 0=notes, 1=duration, 2=billable

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
	case ViewNewEntry:
		return m.renderNewEntryModal()
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
	case ViewNewEntry:
		return m.handleNewEntryKeys(msg)
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
	m.editNotesInput = nil
	m.editDurationInput = nil
	m.editCurrentField = 0
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
		projectInfo := fmt.Sprintf("%s → %s → %s",
			m.editingEntry.Client.Name,
			m.editingEntry.Project.Name,
			m.editingEntry.Task.Name,
		)
		info = styles.SecondaryText.Render(projectInfo)
	}

	// Build field views
	var fieldViews []string

	// Notes field
	notesLabel := "Notes:"
	if m.editCurrentField == 0 {
		notesLabel = styles.HighlightText.Render("▶ Notes:")
	} else {
		notesLabel = styles.SecondaryText.Render("  Notes:")
	}
	var notesView string
	if m.editNotesInput != nil {
		notesView = m.editNotesInput.View()
	} else {
		notesView = m.editNotes
	}
	fieldViews = append(fieldViews, lipgloss.JoinHorizontal(lipgloss.Left, notesLabel, " ", notesView))

	// Duration field
	durationLabel := "Duration:"
	if m.editCurrentField == 1 {
		durationLabel = styles.HighlightText.Render("▶ Duration:")
	} else {
		durationLabel = styles.SecondaryText.Render("  Duration:")
	}
	var durationView string
	if m.editDurationInput != nil {
		durationView = m.editDurationInput.View()
	} else {
		durationView = m.editHours
	}
	fieldViews = append(fieldViews, "", lipgloss.JoinHorizontal(lipgloss.Left, durationLabel, " ", durationView))

	// Billable field
	billableLabel := "Billable:"
	if m.editCurrentField == 2 {
		billableLabel = styles.HighlightText.Render("▶ Billable:")
	} else {
		billableLabel = styles.SecondaryText.Render("  Billable:")
	}
	billableStatus := "[ ] Non-billable"
	if m.editBillable {
		billableStatus = "[x] Billable"
	}
	fieldViews = append(fieldViews, "", lipgloss.JoinHorizontal(lipgloss.Left, billableLabel, " ", billableStatus))

	fields := lipgloss.JoinVertical(lipgloss.Left, fieldViews...)

	// Instructions
	instructions := styles.SecondaryText.Render("Tab/Shift+Tab to navigate • Space/B to toggle billable • Enter to save • Esc to cancel")

	// Status message if any
	statusMsg := ""
	if m.statusMessage != "" {
		statusMsg = styles.ErrorText.Render(m.statusMessage)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		info,
		"",
		fields,
		"",
		instructions,
		statusMsg,
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
			m.currentView = ViewNewEntry
			m.clearEditState()
			// Initialize the new entry form
			m.newEntryCurrentField = 0
			m.newEntryNotes = ""
			m.newEntryHours = "0:00"
			m.newEntryBillable = true
			m.selectedProject = nil
			m.selectedTask = nil

			// Initialize text inputs for new entry
			notesInput := textinput.New()
			notesInput.Placeholder = "Enter notes (optional)"
			notesInput.Width = 50
			m.notesInput = &notesInput

			durationInput := textinput.New()
			durationInput.SetValue("0:00")
			durationInput.Placeholder = "Enter duration (e.g., 1:30)"
			durationInput.Width = 20
			m.durationInput = &durationInput

			m.updateProjectList()
			m.projectList.SetSize(m.width-10, min(len(m.projectList.Items()), 10))
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
			m.editCurrentField = 0

			// Initialize text inputs for editing
			notesInput := textinput.New()
			notesInput.SetValue(selectedEntry.Notes)
			notesInput.Focus()
			notesInput.Placeholder = "Enter notes (optional)"
			notesInput.Width = 50
			m.editNotesInput = &notesInput

			durationInput := textinput.New()
			durationInput.SetValue(formatHoursSimple(selectedEntry.Hours))
			durationInput.Placeholder = "Enter duration (e.g., 1:30)"
			durationInput.Width = 20
			m.editDurationInput = &durationInput

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
		// Check if we're coming from new entry form
		if m.newEntryCurrentField >= 0 && m.newEntryCurrentField <= 4 {
			// Return to new entry form
			m.currentView = ViewNewEntry
			return m, nil
		}
		// Otherwise, cancel project selection and return to main list
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
	var cmd tea.Cmd

	switch msg.String() {
	case "esc":
		// Return to main list, clearing all edit and entry state
		m.currentView = ViewList
		m.clearEditState()
		return m, nil

	case "tab":
		// Move to next field
		m.editCurrentField = (m.editCurrentField + 1) % 3
		// Update focus based on current field
		if m.editCurrentField == 0 && m.editNotesInput != nil {
			m.editNotesInput.Focus()
			if m.editDurationInput != nil {
				m.editDurationInput.Blur()
			}
		} else if m.editCurrentField == 1 && m.editDurationInput != nil {
			m.editDurationInput.Focus()
			if m.editNotesInput != nil {
				m.editNotesInput.Blur()
			}
		} else if m.editCurrentField == 2 {
			// Billable field - blur both inputs
			if m.editNotesInput != nil {
				m.editNotesInput.Blur()
			}
			if m.editDurationInput != nil {
				m.editDurationInput.Blur()
			}
		}
		return m, nil

	case "shift+tab":
		// Move to previous field
		m.editCurrentField = (m.editCurrentField - 1 + 3) % 3
		// Update focus based on current field
		if m.editCurrentField == 0 && m.editNotesInput != nil {
			m.editNotesInput.Focus()
			if m.editDurationInput != nil {
				m.editDurationInput.Blur()
			}
		} else if m.editCurrentField == 1 && m.editDurationInput != nil {
			m.editDurationInput.Focus()
			if m.editNotesInput != nil {
				m.editNotesInput.Blur()
			}
		} else if m.editCurrentField == 2 {
			// Billable field - blur both inputs
			if m.editNotesInput != nil {
				m.editNotesInput.Blur()
			}
			if m.editDurationInput != nil {
				m.editDurationInput.Blur()
			}
		}
		return m, nil

	case " ", "b":
		// Toggle billable if on billable field
		if m.editCurrentField == 2 {
			m.editBillable = !m.editBillable
		}
		return m, nil

	case "enter":
		// Save changes
		return m, m.updateTimeEntry()

	default:
		// Pass to the appropriate input field if it's focused
		if m.editCurrentField == 0 && m.editNotesInput != nil {
			*m.editNotesInput, cmd = m.editNotesInput.Update(msg)
			m.editNotes = m.editNotesInput.Value()
		} else if m.editCurrentField == 1 && m.editDurationInput != nil {
			*m.editDurationInput, cmd = m.editDurationInput.Update(msg)
			m.editHours = m.editDurationInput.Value()
		}
	}

	return m, cmd
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

// updateTimeEntry updates an existing time entry and returns a command
func (m Model) updateTimeEntry() tea.Cmd {
	if m.editingEntry == nil {
		return nil
	}

	// Validate duration
	hours, err := domain.ParseDuration(m.editHours)
	if err != nil {
		// Return an error message
		return func() tea.Msg {
			return timeEntryUpdatedMsg{err: fmt.Errorf("Invalid duration format. Use HH:MM (e.g., 1:30)")}
		}
	}

	request := harvest.UpdateTimeEntryRequest{
		Hours:      &hours,
		Notes:      &m.editNotes,
		IsBillable: &m.editBillable,
	}

	entryID := m.editingEntry.ID

	return func() tea.Msg {
		entry, err := m.harvestClient.UpdateTimeEntry(entryID, request)
		if err != nil {
			return timeEntryUpdatedMsg{err: err}
		}

		return timeEntryUpdatedMsg{entry: entry}
	}
}

// renderNewEntryModal renders the new entry form as a modal overlay
func (m Model) renderNewEntryModal() string {
	// First render the background (list view)
	background := m.renderStyledListView()

	// Modal dimensions
	modalWidth := 60
	modalHeight := 20

	// Center the modal
	startX := (m.width - modalWidth) / 2
	startY := (m.height - modalHeight) / 2

	// Create modal content
	modalContent := m.renderNewEntryForm(modalWidth)

	// Create overlay effect by rendering the modal on top
	// For simplicity, we'll just return the modal form
	// In a more sophisticated implementation, we'd overlay it on the background
	return m.renderModalOverlay(background, modalContent, startX, startY, modalWidth, modalHeight)
}

// renderModalOverlay overlays modal content on background
func (m Model) renderModalOverlay(background, modal string, x, y, width, height int) string {
	// For now, just return the modal with a dimmed indication
	// A full implementation would blend the two views
	return modal
}

// renderNewEntryForm renders the new entry form content
func (m Model) renderNewEntryForm(width int) string {
	styles := DefaultStyles()

	// Header
	header := styles.Header.Render(
		lipgloss.JoinHorizontal(lipgloss.Left,
			styles.Title.Render("New Time Entry"),
		),
	)

	// Build field views
	var fieldViews []string

	// Project field
	projectLabel := "Project:"
	if m.newEntryCurrentField == 0 {
		projectLabel = styles.HighlightText.Render("▶ Project:")
	} else {
		projectLabel = styles.SecondaryText.Render("  Project:")
	}
	projectValue := "(none selected)"
	if m.selectedProject != nil {
		projectValue = fmt.Sprintf("%s → %s", m.selectedProject.Client.Name, m.selectedProject.Name)
	}
	fieldViews = append(fieldViews, lipgloss.JoinHorizontal(lipgloss.Left, projectLabel, " ", projectValue))

	// Task field
	taskLabel := "Task:"
	if m.newEntryCurrentField == 1 {
		taskLabel = styles.HighlightText.Render("▶ Task:")
	} else {
		taskLabel = styles.SecondaryText.Render("  Task:")
	}
	taskValue := "(none selected)"
	if m.selectedTask != nil {
		taskValue = m.selectedTask.Name
	}
	fieldViews = append(fieldViews, "", lipgloss.JoinHorizontal(lipgloss.Left, taskLabel, " ", taskValue))

	// Notes field
	notesLabel := "Notes:"
	if m.newEntryCurrentField == 2 {
		notesLabel = styles.HighlightText.Render("▶ Notes:")
	} else {
		notesLabel = styles.SecondaryText.Render("  Notes:")
	}
	var notesView string
	if m.notesInput != nil && m.newEntryCurrentField == 2 {
		m.notesInput.Focus()
		notesView = m.notesInput.View()
	} else if m.notesInput != nil {
		m.notesInput.Blur()
		notesView = m.notesInput.View()
	} else {
		notesView = m.newEntryNotes
	}
	fieldViews = append(fieldViews, "", lipgloss.JoinHorizontal(lipgloss.Left, notesLabel, " ", notesView))

	// Duration field
	durationLabel := "Duration:"
	if m.newEntryCurrentField == 3 {
		durationLabel = styles.HighlightText.Render("▶ Duration:")
	} else {
		durationLabel = styles.SecondaryText.Render("  Duration:")
	}
	var durationView string
	if m.durationInput != nil && m.newEntryCurrentField == 3 {
		m.durationInput.Focus()
		durationView = m.durationInput.View()
	} else if m.durationInput != nil {
		m.durationInput.Blur()
		durationView = m.durationInput.View()
	} else {
		durationView = m.newEntryHours
	}
	fieldViews = append(fieldViews, "", lipgloss.JoinHorizontal(lipgloss.Left, durationLabel, " ", durationView))

	// Billable field
	billableLabel := "Billable:"
	if m.newEntryCurrentField == 4 {
		billableLabel = styles.HighlightText.Render("▶ Billable:")
	} else {
		billableLabel = styles.SecondaryText.Render("  Billable:")
	}
	billableStatus := "[ ] Non-billable"
	if m.newEntryBillable {
		billableStatus = "[x] Billable"
	}
	fieldViews = append(fieldViews, "", lipgloss.JoinHorizontal(lipgloss.Left, billableLabel, " ", billableStatus))

	fields := lipgloss.JoinVertical(lipgloss.Left, fieldViews...)

	// Instructions
	instructions := styles.SecondaryText.Render("Tab/Shift+Tab to navigate • Enter on project/task to select • Space/B for billable • Ctrl+S to save • Esc to cancel")

	// Status message if any
	statusMsg := ""
	if m.statusMessage != "" {
		statusMsg = styles.ErrorText.Render(m.statusMessage)
	}

	// Create modal box
	content := lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		fields,
		"",
		instructions,
		statusMsg,
	)

	// Style the modal with border and background
	modal := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(AccentColor).
		Padding(1, 2).
		Width(width - 4).
		Background(lipgloss.AdaptiveColor{Light: "#FAFAFA", Dark: "#1A1A1A"}).
		Render(content)

	return modal
}

// handleNewEntryKeys handles key presses in the new entry form
func (m Model) handleNewEntryKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "esc":
		// Cancel and return to main list
		m.currentView = ViewList
		m.clearEditState()
		return m, nil

	case "tab":
		// Move to next field
		m.newEntryCurrentField = (m.newEntryCurrentField + 1) % 5
		// Update focus for text inputs
		if m.newEntryCurrentField == 2 && m.notesInput != nil {
			m.notesInput.Focus()
			if m.durationInput != nil {
				m.durationInput.Blur()
			}
		} else if m.newEntryCurrentField == 3 && m.durationInput != nil {
			m.durationInput.Focus()
			if m.notesInput != nil {
				m.notesInput.Blur()
			}
		} else {
			// Not on a text field, blur both
			if m.notesInput != nil {
				m.notesInput.Blur()
			}
			if m.durationInput != nil {
				m.durationInput.Blur()
			}
		}
		return m, nil

	case "shift+tab":
		// Move to previous field
		m.newEntryCurrentField = (m.newEntryCurrentField - 1 + 5) % 5
		// Update focus for text inputs
		if m.newEntryCurrentField == 2 && m.notesInput != nil {
			m.notesInput.Focus()
			if m.durationInput != nil {
				m.durationInput.Blur()
			}
		} else if m.newEntryCurrentField == 3 && m.durationInput != nil {
			m.durationInput.Focus()
			if m.notesInput != nil {
				m.notesInput.Blur()
			}
		} else {
			// Not on a text field, blur both
			if m.notesInput != nil {
				m.notesInput.Blur()
			}
			if m.durationInput != nil {
				m.durationInput.Blur()
			}
		}
		return m, nil

	case "enter":
		// Handle enter based on current field
		switch m.newEntryCurrentField {
		case 0: // Project field
			// Open project selection
			m.currentView = ViewSelectProject
			m.updateProjectList()
			return m, nil
		case 1: // Task field
			if m.selectedProject != nil {
				// Find tasks for selected project
				for _, pwt := range m.projectsWithTasks {
					if pwt.Project.ID == m.selectedProject.ID {
						if len(pwt.Tasks) > 0 {
							m.currentView = ViewSelectTask
							m.updateTaskList(pwt.Tasks)
						}
						break
					}
				}
			}
			return m, nil
		default:
			// On other fields, treat enter as save
			return m, m.createTimeEntry()
		}

	case "ctrl+s":
		// Save entry
		// Validate required fields
		if m.selectedProject == nil || m.selectedTask == nil {
			m.statusMessage = "Please select a project and task"
			return m, nil
		}

		// Store current values from inputs
		if m.notesInput != nil {
			m.newEntryNotes = m.notesInput.Value()
		}
		if m.durationInput != nil {
			m.newEntryHours = m.durationInput.Value()
		}

		// Validate duration
		if _, err := domain.ParseDuration(m.newEntryHours); err != nil {
			m.statusMessage = "Invalid duration format. Use HH:MM (e.g., 1:30)"
			return m, nil
		}

		return m, m.createTimeEntry()

	case " ", "b":
		// Toggle billable if on billable field
		if m.newEntryCurrentField == 4 {
			m.newEntryBillable = !m.newEntryBillable
		}
		return m, nil

	default:
		// Pass to text inputs if focused
		if m.newEntryCurrentField == 2 && m.notesInput != nil {
			*m.notesInput, cmd = m.notesInput.Update(msg)
			m.newEntryNotes = m.notesInput.Value()
		} else if m.newEntryCurrentField == 3 && m.durationInput != nil {
			*m.durationInput, cmd = m.durationInput.Update(msg)
			m.newEntryHours = m.durationInput.Value()
		}
	}

	return m, cmd
}
