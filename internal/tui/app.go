package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
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
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

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

// Placeholder methods for view rendering - these will be implemented in subsequent steps
func (m Model) renderListView() string {
	return "List view - TODO: implement"
}

func (m Model) renderProjectSelectView() string {
	return "Project select view - TODO: implement"
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

// Placeholder methods for key handling - these will be implemented in subsequent steps
func (m Model) handleListViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m Model) handleProjectSelectKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, nil
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

// Commands for fetching data - these will be implemented in subsequent steps
func fetchTimeEntriesCmd(client *harvest.Client, date time.Time) tea.Cmd {
	return nil
}

func fetchProjectsWithTasksCmd(client *harvest.Client) tea.Cmd {
	return nil
}
