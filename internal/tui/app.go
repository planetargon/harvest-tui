package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/planetargon/argon-harvest-tui/internal/config"
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
		return m, nil

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

	default:
		return m, nil
	}
}

// View renders the current view.
func (m Model) View() string {
	switch m.currentView {
	case ViewList:
		return m.renderListView()
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
	default:
		return "Unknown view"
	}
}

// handleKeyPress processes key presses for the current view.
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global keybindings that work in all views
	switch msg.String() {
	case "ctrl+c", "esc":
		if m.currentView != ViewList {
			// Return to main list view
			m.currentView = ViewList
			m.clearEditState()
			return m, nil
		}
		// Exit application if already in list view
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
	m.editingEntry = nil
	m.editNotes = ""
	m.editHours = ""
	m.editBillable = true
}

// renderListView renders the main list view showing time entries for the current date.
func (m Model) renderListView() string {
	styles := DefaultStyles()

	// Header with title and current date
	header := styles.Header.Render(
		lipgloss.JoinHorizontal(lipgloss.Left,
			styles.Title.Render("Harvest TUI"),
			"  ",
			styles.Subtitle.Render(m.currentDate.Format("Monday, January 2, 2006")),
		),
	)

	// Handle loading state
	if m.loading {
		content := styles.Content.Render("Loading...")
		return lipgloss.JoinVertical(lipgloss.Left, header, content)
	}

	// Handle error state
	if m.errorMessage != "" {
		content := styles.Content.Render(
			styles.ErrorText.Render("Error: " + m.errorMessage),
		)
		return lipgloss.JoinVertical(lipgloss.Left, header, content)
	}

	// Handle empty state
	if len(m.timeEntries) == 0 {
		emptyState := styles.Content.Render(
			lipgloss.JoinVertical(lipgloss.Center,
				styles.MutedText.Render("No time entries for this date"),
				"",
				styles.SecondaryText.Render("Press 'n' to create a new entry"),
			),
		)

		// Add daily total (0:00 for empty day)
		footer := styles.Footer.Render(
			lipgloss.JoinHorizontal(lipgloss.Left,
				"Daily total: "+styles.PrimaryText.Render("0:00"),
			),
		)

		return lipgloss.JoinVertical(lipgloss.Left, header, emptyState, footer)
	}

	// Render time entries
	var entries []string
	for i, entry := range m.timeEntries {
		isSelected := i == m.selectedEntryIndex
		entryView := m.renderTimeEntry(entry, isSelected, styles)
		entries = append(entries, entryView)
	}

	// Calculate daily total
	totalHours := 0.0
	for _, entry := range m.timeEntries {
		totalHours += entry.Hours
	}

	content := styles.Content.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.JoinVertical(lipgloss.Left, entries...),
			"",
			styles.StatusBar.Render(
				"Daily total: "+styles.PrimaryText.Render(formatHoursSimple(totalHours)),
			),
		),
	)

	// Status message
	var statusLine string
	if m.statusMessage != "" {
		statusLine = styles.Footer.Render(styles.SuccessText.Render(m.statusMessage))
	}

	parts := []string{header, content}
	if statusLine != "" {
		parts = append(parts, statusLine)
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// renderTimeEntry renders a single time entry with appropriate styling.
func (m Model) renderTimeEntry(entry harvest.TimeEntry, isSelected bool, styles Styles) string {
	// Status indicator
	statusIcon, _ := styles.StatusIndicator(entry.IsRunning, entry.IsLocked)

	// Format hours
	hoursText := formatHoursSimple(entry.Hours)
	if entry.IsRunning {
		hoursText = styles.RunningIndicator.Render(hoursText)
	}

	// Build the entry content
	entryContent := lipgloss.JoinVertical(lipgloss.Left,
		// First line: Project info and hours
		lipgloss.JoinHorizontal(lipgloss.Left,
			statusIcon,
			styles.PrimaryText.Render(entry.Client.Name+" → "+entry.Project.Name+" → "+entry.Task.Name),
			lipgloss.NewStyle().Width(4).Align(lipgloss.Right).Render(""),
			hoursText,
		),
		// Second line: Notes (if any)
		func() string {
			if entry.Notes != "" {
				return "  " + styles.SecondaryText.Render(entry.Notes)
			}
			return ""
		}(),
	)

	// Apply appropriate styling based on entry state
	entryStyle := styles.TimeEntryStyle(isSelected, entry.IsRunning, entry.IsLocked)
	return entryStyle.Render(entryContent)
}

// formatHoursSimple formats hours as H:MM format.
func formatHoursSimple(hours float64) string {
	h := int(hours)
	m := int((hours - float64(h)) * 60)
	return fmt.Sprintf("%d:%02d", h, m)
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
	return i.client.Name + " → " + i.project.Name
}

func (i projectItem) Description() string {
	return "Project ID: " + fmt.Sprintf("%d", i.project.ID)
}

// taskItem represents a task in the selection list.

type taskItem struct {
	task harvest.Task
}

func (i taskItem) FilterValue() string {
	return i.task.Name
}

func (i taskItem) Title() string {
	return i.task.Name
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
					break
				}
			}
		}
	}

	// Add all projects sorted by client then project name
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
			items = append(items, projectItem{
				project: pwt.Project,
				client:  pwt.Project.Client,
			})
		}
	}

	m.projectList.SetItems(items)
	m.projectList.Title = "Select Project"
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
	return "Task select view - TODO: implement"
}

func (m Model) renderEditView() string {
	return "Edit view - TODO: implement"
}

func (m Model) renderConfirmDeleteView() string {
	return "Confirm delete view - TODO: implement"
}

func (m Model) renderHelpView() string {
	return "Help view - TODO: implement"
}

// handleListViewKeys handles key presses in the main list view.
func (m Model) handleListViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keys := DefaultKeyMap()

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

	case key.Matches(msg, keys.Start):
		if len(m.timeEntries) > 0 && m.selectedEntryIndex < len(m.timeEntries) {
			selectedEntry := m.timeEntries[m.selectedEntryIndex]
			if selectedEntry.IsLocked {
				m.statusMessage = "Cannot start locked time entry."
				return m, nil
			}
			if selectedEntry.IsRunning {
				m.statusMessage = "Timer is already running for this entry."
				return m, nil
			}
			return m, restartTimeEntryCmd(m.harvestClient, selectedEntry.ID)
		}
		return m, nil

	case key.Matches(msg, keys.Stop):
		if len(m.timeEntries) > 0 && m.selectedEntryIndex < len(m.timeEntries) {
			selectedEntry := m.timeEntries[m.selectedEntryIndex]
			if selectedEntry.IsLocked {
				m.statusMessage = "Cannot stop locked time entry."
				return m, nil
			}
			if !selectedEntry.IsRunning {
				m.statusMessage = "Timer is not running for this entry."
				return m, nil
			}
			return m, stopTimeEntryCmd(m.harvestClient, selectedEntry.ID)
		}
		return m, nil
	}

	return m, nil
}

func (m Model) handleProjectSelectKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "enter":
		// Get the selected project
		selected := m.projectList.SelectedItem()
		if selected != nil {
			if item, ok := selected.(projectItem); ok {
				m.selectedProject = &item.project
				// Find tasks for this project
				for _, pwt := range m.projectsWithTasks {
					if pwt.Project.ID == item.project.ID {
						if len(pwt.Tasks) == 1 {
							// Only one task, skip task selection
							m.selectedTask = &pwt.Tasks[0]
							// TODO: Move to notes input view when implemented
							m.statusMessage = "Selected: " + item.client.Name + " → " + item.project.Name + " → " + pwt.Tasks[0].Name
							m.currentView = ViewList
						} else {
							// Multiple tasks, show task selection
							m.currentView = ViewSelectTask
							// TODO: Update task list
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
	return m, nil
}

func (m Model) handleEditViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m Model) handleConfirmDeleteKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m Model) handleHelpViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, nil
}

// Messages for handling async operations
type timeEntriesFetchedMsg struct {
	entries []harvest.TimeEntry
	err     error
}

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
