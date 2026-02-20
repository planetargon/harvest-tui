package tui

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/planetargon/harvest-tui/internal/config"
	"github.com/planetargon/harvest-tui/internal/harvest"
	"github.com/planetargon/harvest-tui/internal/state"
)

// ViewState represents the different views in the TUI application.
type ViewState int

const (
	// ViewLoading is the initial loading screen shown during startup.
	ViewLoading ViewState = iota
	// ViewList is the main list view showing time entries for the current date.
	ViewList
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
	currentUser        *harvest.User

	// New entry creation state
	selectedProject      *harvest.Project
	selectedTask         *harvest.Task
	newEntryNotes        string
	newEntryHours        string
	newEntryBillable     bool
	newEntryCurrentField int // 0=project, 1=task, 2=notes, 3=duration

	// Edit entry state
	editingEntry     *harvest.TimeEntry
	editTask         *harvest.Task
	editNotes        string
	editHours        string
	editBillable     bool
	editCurrentField int // 0=task, 1=notes, 2=duration
	pendingTaskEdit  bool

	// UI state
	loading           bool
	errorMessage      string
	statusMessage     string
	statusMessageTime time.Time // Track when the status message was set
	lastFetchTime     time.Time // Track last API fetch to avoid rate limiting
	spinner           spinner.Model
	timeEntriesLoaded bool
	projectsLoaded    bool

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
func NewModel(cfg *config.Config, client *harvest.Client, appState *state.State, user *harvest.User) Model {
	s := spinner.New()
	s.Spinner = spinner.MiniDot
	s.Style = lipgloss.NewStyle().Foreground(accentColor)

	return Model{
		currentView:        ViewLoading,
		config:             cfg,
		harvestClient:      client,
		appState:           appState,
		currentDate:        time.Now(),
		timeEntries:        []harvest.TimeEntry{},
		projectsWithTasks:  []harvest.ProjectWithTasks{},
		selectedEntryIndex: 0,
		currentUser:        user,
		newEntryNotes:      "",
		newEntryHours:      "",
		newEntryBillable:   true,
		editNotes:          "",
		editHours:          "",
		editBillable:       true,
		loading:            false,
		errorMessage:       "",
		statusMessage:      "",
		statusMessageTime:  time.Time{},
		spinner:            s,
		projectList:        newShellList(newProjectDelegate()),
		taskList:           newShellList(newTaskDelegate()),
		width:              80,
		height:             24,
	}
}

// Init initializes the model and returns initial commands.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
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
		m.setListSizes()
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case spinner.TickMsg:
		if m.currentView == ViewLoading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil

	case timeEntriesFetchedMsg:
		if msg.err != nil {
			m.errorMessage = "Failed to fetch time entries: " + msg.err.Error()
		} else {
			m.timeEntries = msg.entries
			m.errorMessage = ""
			m.lastFetchTime = time.Now()
		}
		m.loading = false
		m.timeEntriesLoaded = true

		// Transition from loading screen when both fetches complete
		if m.currentView == ViewLoading && m.projectsLoaded {
			m.currentView = ViewList
		}

		// If there's a running timer, continue ticking
		if m.hasRunningTimer() {
			return m, tickCmd()
		}
		return m, nil

	case tickMsg:
		// Clear status message after 3 seconds
		if m.statusMessage != "" && !m.statusMessageTime.IsZero() {
			if time.Since(m.statusMessageTime) > 3*time.Second {
				m.statusMessage = ""
				m.statusMessageTime = time.Time{}
			}
		}

		// Check if we have a running timer and it's time to refresh from API
		if m.hasRunningTimer() && m.currentView == ViewList && !m.loading {
			if time.Since(m.lastFetchTime) >= 25*time.Second {
				m.lastFetchTime = time.Now()
				return m, tea.Batch(
					fetchTimeEntriesCmd(m.harvestClient, m.currentDate),
					tickCmd(),
				)
			}
			return m, tickCmd()
		}
		// Continue ticking if we have a running timer or status message
		if m.hasRunningTimer() || m.statusMessage != "" {
			return m, tickCmd()
		}
		return m, nil

	case projectsWithTasksFetchedMsg:
		if msg.err != nil {
			m.errorMessage = "Failed to fetch projects: " + msg.err.Error()
			m.pendingTaskEdit = false
		} else {
			m.projectsWithTasks = msg.projectsWithTasks
			m.errorMessage = ""

			// If user requested task edit while projects were loading, open it now
			if m.pendingTaskEdit && m.editingEntry != nil && m.currentView == ViewEditEntry {
				m.pendingTaskEdit = false
				if !m.openTaskSelectionForEdit() {
					m.setStatusMessage("No tasks found for this project")
				}
			}
		}
		m.projectsLoaded = true

		// Transition from loading screen when both fetches complete
		if m.currentView == ViewLoading && m.timeEntriesLoaded {
			m.currentView = ViewList
		}

		return m, nil

	case timeEntryStartedMsg:
		if msg.err != nil {
			m.setStatusMessage("Failed to start timer: " + msg.err.Error())
		} else {
			// Update the entry in our local list
			for i, entry := range m.timeEntries {
				if entry.ID == msg.entry.ID {
					m.timeEntries[i] = *msg.entry
					break
				}
			}
			m.setStatusMessage("Timer started successfully")
			// Re-fetch entries so previously running timer shows as stopped
			m.lastFetchTime = time.Now()
			return m, tea.Batch(
				fetchTimeEntriesCmd(m.harvestClient, m.currentDate),
				tickCmd(),
			)
		}
		return m, nil

	case timeEntryStoppedMsg:
		if msg.err != nil {
			m.setStatusMessage("Failed to stop timer: " + msg.err.Error())
		} else {
			// Update the entry in our local list
			for i, entry := range m.timeEntries {
				if entry.ID == msg.entry.ID {
					m.timeEntries[i] = *msg.entry
					break
				}
			}
			m.setStatusMessage("Timer stopped successfully")
		}
		return m, nil

	case timeEntryCreatedMsg:
		if msg.err != nil {
			m.setStatusMessage("Failed to create entry: " + msg.err.Error())
		} else {
			// Add the new entry to our local list
			m.timeEntries = append([]harvest.TimeEntry{*msg.entry}, m.timeEntries...)
			m.setStatusMessage("Time entry created successfully")
			// Clear new entry state and return to main list
			m.clearEditState()
			m.currentView = ViewList
		}
		return m, nil

	case timeEntryUpdatedMsg:
		if msg.err != nil {
			m.setStatusMessage("Failed to update entry: " + msg.err.Error())
		} else {
			// Update the entry in our local list
			for i, entry := range m.timeEntries {
				if entry.ID == msg.entry.ID {
					m.timeEntries[i] = *msg.entry
					break
				}
			}
			m.setStatusMessage("Time entry updated successfully")
			// Clear edit state and return to main list
			m.clearEditState()
			m.currentView = ViewList
		}
		return m, nil

	case timeEntryDeletedMsg:
		if msg.err != nil {
			m.setStatusMessage("Failed to delete entry: " + msg.err.Error())
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

			m.setStatusMessage("Time entry deleted successfully")
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
	case ViewLoading:
		return m.renderLoadingView()
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
	previousView := m.currentView

	// Ctrl+C always quits, even during loading
	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}

	// Block all other input during loading
	if m.currentView == ViewLoading {
		return m, nil
	}

	// Global keybindings that work in all views
	switch msg.String() {
	case "?":
		if m.currentView == ViewHelp {
			m.currentView = ViewList
		} else {
			m.currentView = ViewHelp
		}
		m.clearStatusMessage()
		return m, nil
	}

	// View-specific keybindings
	var result tea.Model
	var cmd tea.Cmd

	switch m.currentView {
	case ViewList:
		result, cmd = m.handleListViewKeys(msg)
	case ViewSelectProject:
		result, cmd = m.handleProjectSelectKeys(msg)
	case ViewSelectTask:
		result, cmd = m.handleTaskSelectKeys(msg)
	case ViewNewEntry:
		result, cmd = m.handleNewEntryKeys(msg)
	case ViewEditEntry:
		result, cmd = m.handleEditViewKeys(msg)
	case ViewConfirmDelete:
		result, cmd = m.handleConfirmDeleteKeys(msg)
	case ViewHelp:
		result, cmd = m.handleHelpViewKeys(msg)
	case ViewNotesInput:
		result, cmd = m.handleNotesInputKeys(msg)
	case ViewDurationInput:
		result, cmd = m.handleDurationInputKeys(msg)
	case ViewBillableToggle:
		result, cmd = m.handleBillableToggleKeys(msg)
	default:
		return m, nil
	}

	// Clear status message when transitioning between views
	if resultModel, ok := result.(Model); ok && resultModel.currentView != previousView {
		resultModel.clearStatusMessage()
		return resultModel, cmd
	}

	return result, cmd
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
	m.editTask = nil
	m.editNotes = ""
	m.editHours = ""
	m.editBillable = true
	m.editNotesInput = nil
	m.editDurationInput = nil
	m.editCurrentField = 0
	m.pendingTaskEdit = false
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

// parseDuration parses a duration string in HH:MM format and returns hours as a float64.
func parseDuration(durationStr string) (float64, error) {
	durationStr = strings.TrimSpace(durationStr)
	if durationStr == "" {
		return 0, fmt.Errorf("duration cannot be empty")
	}

	parts := strings.Split(durationStr, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid duration format. Use HH:MM (e.g., 1:30)")
	}

	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("invalid duration format. Use HH:MM (e.g., 1:30)")
	}

	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("invalid duration format. Use HH:MM (e.g., 1:30)")
	}

	if hours < 0 || minutes < 0 || minutes >= 60 {
		return 0, fmt.Errorf("invalid duration format. Use HH:MM (e.g., 1:30)")
	}

	return float64(hours) + float64(minutes)/60.0, nil
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
	return ""
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
	return ""
}

// newProjectDelegate creates a new delegate for project list items.
func newProjectDelegate() list.DefaultDelegate {
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	return delegate
}

// newTaskDelegate creates a new delegate for task list items.
func newTaskDelegate() list.DefaultDelegate {
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	return delegate
}

// setListSizes updates the project and task list dimensions based on the shell width and window height.
func (m *Model) setListSizes() {
	contentW := m.shellWidth() - 4
	contentH := m.height - 7
	if contentH < 5 {
		contentH = 5
	}
	m.projectList.SetSize(contentW, contentH)
	m.taskList.SetSize(contentW, contentH)
}

// newShellList creates a list.Model with title and status bar disabled.
func newShellList(delegate list.DefaultDelegate) list.Model {
	l := list.New([]list.Item{}, delegate, 0, 0)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	return l
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
	allProjects := make([]harvest.ProjectWithTasks, len(m.projectsWithTasks))
	copy(allProjects, m.projectsWithTasks)

	sort.Slice(allProjects, func(i, j int) bool {
		if allProjects[i].Project.Client.Name != allProjects[j].Project.Client.Name {
			return allProjects[i].Project.Client.Name < allProjects[j].Project.Client.Name
		}
		return allProjects[i].Project.Name < allProjects[j].Project.Name
	})

	// Add all projects to items (including those in recents)
	for _, pwt := range allProjects {
		items = append(items, projectItem{
			project: pwt.Project,
			client:  pwt.Project.Client,
		})
	}

	m.projectList.SetItems(items)
}

// updateTaskList updates the task list with tasks from the selected project.
func (m *Model) updateTaskList(tasks []harvest.Task) {
	var items []list.Item
	for _, task := range tasks {
		items = append(items, taskItem{task: task})
	}
	m.taskList.SetItems(items)
}

func (m Model) renderProjectSelectView() string {
	width := m.shellWidth()

	titleBar := m.renderTitleBar()

	// Breadcrumb header
	breadcrumb := "  " + AccentText.Render("New Time Entry") + ArrowStyle.Render(" → ") + MutedText.Render("Step 1: Choose Project")

	divider := "  " + RenderDividerWidth(width-4)

	// Render the list
	listView := m.projectList.View()

	content := strings.Join([]string{titleBar, breadcrumb, divider, "", listView}, "\n")

	footerKeys := []string{
		RenderKeybinding("↑↓", "navigate"),
		RenderKeybinding("/", "filter"),
		RenderKeybinding("enter", "select"),
		RenderKeybinding("esc", "back"),
	}

	return m.buildShellBox(content, width, footerKeys)
}

func (m Model) renderTaskSelectView() string {
	width := m.shellWidth()

	titleBar := m.renderTitleBar()

	// Breadcrumb header
	var breadcrumb string
	if m.editingEntry != nil {
		breadcrumb = "  " + AccentText.Render("Edit Time Entry") + ArrowStyle.Render(" → ") + MutedText.Render("Change Task")
	} else {
		breadcrumb = "  " + AccentText.Render("New Time Entry") + ArrowStyle.Render(" → ") + MutedText.Render("Step 2: Choose Task")
	}

	// Show selected project
	projectInfo := ""
	if m.selectedProject != nil {
		projectInfo = "  " + MutedText.Render(fmt.Sprintf("Project: %s → %s",
			m.selectedProject.Client.Name,
			m.selectedProject.Name))
	}

	divider := "  " + RenderDividerWidth(width-4)

	// Render the list
	listView := m.taskList.View()

	content := strings.Join([]string{titleBar, breadcrumb, projectInfo, divider, "", listView}, "\n")

	footerKeys := []string{
		RenderKeybinding("↑↓", "navigate"),
		RenderKeybinding("/", "filter"),
		RenderKeybinding("enter", "select"),
		RenderKeybinding("esc", "back"),
	}

	return m.buildShellBox(content, width, footerKeys)
}

func (m Model) renderEditView() string {
	width := m.shellWidth()

	titleBar := m.renderTitleBar()

	// Breadcrumb
	breadcrumb := "  " + AccentText.Render("Edit Time Entry")

	// Entry info (client → project breadcrumb, task is now an editable field)
	info := ""
	if m.editingEntry != nil {
		info = "  " + MutedText.Render(fmt.Sprintf("%s → %s",
			m.editingEntry.Client.Name,
			m.editingEntry.Project.Name))
	}

	divider := "  " + RenderDividerWidth(width-4)

	// Build field views
	taskLabel := fieldLabel("Task:", m.editCurrentField == 0)
	taskName := ""
	if m.editTask != nil {
		taskName = m.editTask.Name
	}
	taskView := taskName
	if m.editCurrentField == 0 {
		taskView = taskName + MutedText.Render("  (press enter to change)")
	}

	notesLabel := fieldLabel("Notes:", m.editCurrentField == 1)
	var notesView string
	if m.editNotesInput != nil {
		notesView = m.editNotesInput.View()
	} else {
		notesView = m.editNotes
	}

	durationLabel := fieldLabel("Duration:", m.editCurrentField == 2)
	var durationView string
	if m.editDurationInput != nil {
		durationView = m.editDurationInput.View()
	} else {
		durationView = m.editHours
	}

	// Status message if any
	statusLine := m.renderStatusLine()

	contentLines := []string{
		titleBar,
		breadcrumb,
		info,
		divider,
		"",
		"  " + taskLabel + " " + taskView,
		"",
		"  " + notesLabel + " " + notesView,
		"",
		"  " + durationLabel + " " + durationView,
	}
	if statusLine != "" {
		contentLines = append(contentLines, "", statusLine)
	}

	content := strings.Join(contentLines, "\n")

	footerKeys := []string{
		RenderKeybinding("tab", "next field"),
	}
	if m.editCurrentField == 0 {
		footerKeys = append(footerKeys, RenderKeybinding("enter", "select"))
	}
	footerKeys = append(footerKeys,
		RenderKeybinding("ctrl+s", "save"),
		RenderKeybinding("esc", "cancel"),
	)

	return m.buildShellBox(content, width, footerKeys)
}

// fieldLabel renders a field label with ▶ indicator if active.
func fieldLabel(label string, active bool) string {
	if active {
		return AccentText.Render("▶ " + label)
	}
	return MutedText.Render("  " + label)
}

func (m Model) renderConfirmDeleteView() string {
	width := m.shellWidth()

	titleBar := m.renderTitleBar()

	breadcrumb := "  " + ErrorText.Render("Confirm Delete")

	divider := "  " + RenderDividerWidth(width-4)

	// Entry details
	var detailLines []string
	detailLines = append(detailLines, "  "+AccentText.Render("Are you sure you want to delete this entry?"))
	if m.editingEntry != nil {
		detailLines = append(detailLines, "")
		if m.editingEntry.Notes != "" {
			detailLines = append(detailLines, "  "+MutedText.Render("Notes: "+m.editingEntry.Notes))
		}
		detailLines = append(detailLines, "  "+MutedText.Render("Duration: "+formatHoursSimple(m.editingEntry.Hours)))
	}

	contentLines := []string{titleBar, breadcrumb, divider, ""}
	contentLines = append(contentLines, detailLines...)

	content := strings.Join(contentLines, "\n")

	footerKeys := []string{
		RenderKeybinding("y", "confirm"),
		RenderKeybinding("n", "cancel"),
		RenderKeybinding("esc", "cancel"),
	}

	return m.buildShellBox(content, width, footerKeys)
}

func (m Model) renderHelpView() string {
	width := m.shellWidth()

	titleBar := m.renderTitleBar()

	breadcrumb := "  " + AccentText.Render("Help")

	divider := "  " + RenderDividerWidth(width-4)

	// Keybindings organized by category
	contentLines := []string{
		titleBar,
		breadcrumb,
		divider,
		"",
		"  " + AccentText.Render("Navigation"),
		"    ↑/k       Move up",
		"    ↓/j       Move down",
		"    ←/h       Previous day",
		"    →/l       Next day",
		"    t         Jump to today",
		"",
		"  " + AccentText.Render("Time Entry Actions"),
		"    n         New entry",
		"    e         Edit entry",
		"    d         Delete entry",
		"    s         Start/stop timer",
		"",
		"  " + AccentText.Render("General"),
		"    ?         Toggle this help",
		"    q/Esc     Quit/Go back",
		"    Ctrl+C    Force quit",
	}

	content := strings.Join(contentLines, "\n")

	footerKeys := []string{
		RenderKeybinding("?", "close"),
		RenderKeybinding("esc", "back"),
	}

	return m.buildShellBox(content, width, footerKeys)
}

func (m Model) renderNotesInputView() string {
	width := m.shellWidth()

	titleBar := m.renderTitleBar()

	breadcrumb := "  " + AccentText.Render("New Time Entry") + ArrowStyle.Render(" → ") + MutedText.Render("Step 3: Enter Notes")

	// Selected project and task info
	info := ""
	if m.selectedProject != nil && m.selectedTask != nil {
		info = "  " + MutedText.Render(fmt.Sprintf("%s → %s → %s",
			m.selectedProject.Client.Name,
			m.selectedProject.Name,
			m.selectedTask.Name))
	}

	divider := "  " + RenderDividerWidth(width-4)

	// Input field
	inputView := ""
	if m.notesInput != nil {
		inputView = "  " + m.notesInput.View()
	}

	content := strings.Join([]string{titleBar, breadcrumb, info, divider, "", inputView}, "\n")

	footerKeys := []string{
		RenderKeybinding("enter", "continue"),
		RenderKeybinding("esc", "cancel"),
	}

	return m.buildShellBox(content, width, footerKeys)
}

func (m Model) renderDurationInputView() string {
	width := m.shellWidth()

	titleBar := m.renderTitleBar()

	breadcrumb := "  " + AccentText.Render("New Time Entry") + ArrowStyle.Render(" → ") + MutedText.Render("Step 4: Enter Duration")

	// Selected info
	info := ""
	if m.selectedProject != nil && m.selectedTask != nil {
		info = "  " + MutedText.Render(fmt.Sprintf("%s → %s → %s",
			m.selectedProject.Client.Name,
			m.selectedProject.Name,
			m.selectedTask.Name))
	}

	// Notes info
	notesInfo := ""
	if m.newEntryNotes != "" {
		notesInfo = "  " + MutedText.Render("Notes: "+m.newEntryNotes)
	}

	divider := "  " + RenderDividerWidth(width-4)

	// Input field
	inputView := ""
	if m.durationInput != nil {
		inputView = "  " + m.durationInput.View()
	}

	// Status message
	statusLine := m.renderStatusLine()

	contentLines := []string{titleBar, breadcrumb, info, notesInfo, divider, "", inputView}
	if statusLine != "" {
		contentLines = append(contentLines, "", statusLine)
	}

	content := strings.Join(contentLines, "\n")

	footerKeys := []string{
		RenderKeybinding("enter", "continue"),
		RenderKeybinding("esc", "back"),
	}

	return m.buildShellBox(content, width, footerKeys)
}

func (m Model) renderBillableToggleView() string {
	width := m.shellWidth()

	titleBar := m.renderTitleBar()

	breadcrumb := "  " + AccentText.Render("New Time Entry") + ArrowStyle.Render(" → ") + MutedText.Render("Step 5: Billable Status")

	// Selected info
	info := ""
	if m.selectedProject != nil && m.selectedTask != nil {
		info = "  " + MutedText.Render(fmt.Sprintf("%s → %s → %s",
			m.selectedProject.Client.Name,
			m.selectedProject.Name,
			m.selectedTask.Name))
	}

	// Entry details
	var detailLines []string
	if m.newEntryNotes != "" {
		detailLines = append(detailLines, "  "+MutedText.Render("Notes: "+m.newEntryNotes))
	}
	if m.newEntryHours != "" {
		detailLines = append(detailLines, "  "+MutedText.Render("Duration: "+m.newEntryHours))
	}

	divider := "  " + RenderDividerWidth(width-4)

	// Billable toggle
	billableStatus := "  [ ] Non-billable"
	if m.newEntryBillable {
		billableStatus = "  [x] Billable"
	}

	contentLines := []string{titleBar, breadcrumb, info}
	contentLines = append(contentLines, detailLines...)
	contentLines = append(contentLines, divider, "", billableStatus)

	content := strings.Join(contentLines, "\n")

	footerKeys := []string{
		RenderKeybinding("space", "toggle"),
		RenderKeybinding("enter", "create"),
		RenderKeybinding("esc", "back"),
	}

	return m.buildShellBox(content, width, footerKeys)
}

// handleListViewKeys handles key presses in the main list view.
func (m Model) handleListViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keys := DefaultKeyMap()

	// Check for quit first
	switch msg.String() {
	case "q":
		// Show farewell message
		if m.currentUser != nil {
			fullName := m.currentUser.FirstName + " " + m.currentUser.LastName
			return m, tea.Sequence(
				tea.Println(fmt.Sprintf("\nSee you next time, %s!", fullName)),
				tea.Quit,
			)
		}
		return m, tea.Quit
	}

	switch {
	case key.Matches(msg, keys.Up):
		if len(m.timeEntries) > 0 && m.selectedEntryIndex > 0 {
			m.selectedEntryIndex--
		}
		// Clear any status messages on navigation
		if m.statusMessage != "" {
			m.statusMessage = ""
			m.statusMessageTime = time.Time{}
		}
		return m, nil

	case key.Matches(msg, keys.Down):
		if len(m.timeEntries) > 0 && m.selectedEntryIndex < len(m.timeEntries)-1 {
			m.selectedEntryIndex++
		}
		// Clear any status messages on navigation
		if m.statusMessage != "" {
			m.statusMessage = ""
			m.statusMessageTime = time.Time{}
		}
		return m, nil

	case key.Matches(msg, keys.PrevDay):
		m.currentDate = m.currentDate.AddDate(0, 0, -1)
		m.selectedEntryIndex = 0
		m.loading = true
		m.clearStatusMessage()
		return m, fetchTimeEntriesCmd(m.harvestClient, m.currentDate)

	case key.Matches(msg, keys.NextDay):
		m.currentDate = m.currentDate.AddDate(0, 0, 1)
		m.selectedEntryIndex = 0
		m.loading = true
		m.clearStatusMessage()
		return m, fetchTimeEntriesCmd(m.harvestClient, m.currentDate)

	case key.Matches(msg, keys.Today):
		m.currentDate = time.Now()
		m.selectedEntryIndex = 0
		m.loading = true
		m.clearStatusMessage()
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
			m.setListSizes()
			return m, nil
		} else {
			m.setStatusMessage("No projects available. Please check your Harvest configuration.")
			return m, nil
		}

	case key.Matches(msg, keys.Edit):
		if len(m.timeEntries) > 0 && m.selectedEntryIndex < len(m.timeEntries) {
			selectedEntry := m.timeEntries[m.selectedEntryIndex]
			if selectedEntry.IsLocked {
				m.setStatusMessage("Cannot edit locked time entry.")
				return m, nil
			}
			if selectedEntry.IsRunning {
				m.setStatusMessage("Cannot edit running time entry. Stop the timer first.")
				return m, nil
			}
			m.editingEntry = &selectedEntry
			m.editTask = &harvest.Task{ID: selectedEntry.Task.ID, Name: selectedEntry.Task.Name}
			m.editNotes = selectedEntry.Notes
			m.editHours = formatHoursSimple(selectedEntry.Hours)
			m.editBillable = selectedEntry.IsBillable
			m.editCurrentField = 0

			// Initialize text inputs for editing
			notesInput := textinput.New()
			notesInput.SetValue(selectedEntry.Notes)
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
				m.setStatusMessage("Cannot delete locked time entry.")
				return m, nil
			}
			if selectedEntry.IsRunning {
				m.setStatusMessage("Cannot delete running time entry. Stop the timer first.")
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
					m.setStatusMessage("Cannot stop locked time entry.")
				} else {
					m.setStatusMessage("Cannot start locked time entry.")
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
		// If the list is filtering or has a filter applied, let the list handle esc
		if m.projectList.FilterState() != list.Unfiltered {
			break
		}
		// Check if we're coming from new entry form
		if m.newEntryCurrentField >= 0 && m.newEntryCurrentField <= 3 {
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

				// Find tasks for this project
				for _, pwt := range m.projectsWithTasks {
					if pwt.Project.ID == item.project.ID {
						if len(pwt.Tasks) == 0 {
							// No tasks available for this project
							m.setStatusMessage("No tasks available for this project")
							m.selectedProject = nil
							return m, nil
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
		// If the list is filtering or has a filter applied, let the list handle esc
		if m.taskList.FilterState() != list.Unfiltered {
			break
		}
		if m.editingEntry != nil {
			// Return to edit view when editing
			m.currentView = ViewEditEntry
			m.selectedProject = nil
			m.updateEditFieldFocus()
			return m, nil
		}
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
				if m.editingEntry != nil {
					// Set editTask and return to edit view
					m.editTask = &harvest.Task{ID: item.task.ID, Name: item.task.Name}
					m.selectedProject = nil
					m.currentView = ViewEditEntry
					m.updateEditFieldFocus()
					return m, nil
				}
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
		m.updateEditFieldFocus()
		return m, nil

	case "shift+tab":
		// Move to previous field
		m.editCurrentField = (m.editCurrentField - 1 + 3) % 3
		m.updateEditFieldFocus()
		return m, nil

	case "enter":
		if m.editCurrentField == 0 {
			// Open task selection for the current project
			if m.editingEntry != nil {
				if len(m.projectsWithTasks) == 0 {
					m.pendingTaskEdit = true
					m.setStatusMessage("Loading tasks...")
					return m, fetchProjectsWithTasksCmd(m.harvestClient)
				}
				if !m.openTaskSelectionForEdit() {
					m.setStatusMessage("No tasks found for this project")
				}
			}
		}
		return m, nil

	case "ctrl+s":
		// Save changes
		return m, m.updateTimeEntry()

	default:
		// Pass to the appropriate input field if it's focused
		if m.editCurrentField == 1 && m.editNotesInput != nil {
			*m.editNotesInput, cmd = m.editNotesInput.Update(msg)
			m.editNotes = m.editNotesInput.Value()
		} else if m.editCurrentField == 2 && m.editDurationInput != nil {
			*m.editDurationInput, cmd = m.editDurationInput.Update(msg)
			m.editHours = m.editDurationInput.Value()
		}
	}

	return m, cmd
}

// updateEditFieldFocus updates text input focus based on the current edit field.
func (m *Model) updateEditFieldFocus() {
	if m.editNotesInput != nil {
		if m.editCurrentField == 1 {
			m.editNotesInput.Focus()
		} else {
			m.editNotesInput.Blur()
		}
	}
	if m.editDurationInput != nil {
		if m.editCurrentField == 2 {
			m.editDurationInput.Focus()
		} else {
			m.editDurationInput.Blur()
		}
	}
}

// openTaskSelectionForEdit finds the editing entry's project tasks and switches to task selection.
// Returns true if the transition succeeded, false if the project was not found.
func (m *Model) openTaskSelectionForEdit() bool {
	for _, pwt := range m.projectsWithTasks {
		if pwt.Project.ID == m.editingEntry.Project.ID {
			m.selectedProject = &pwt.Project
			m.updateTaskList(pwt.Tasks)
			if m.editNotesInput != nil {
				m.editNotesInput.Blur()
			}
			if m.editDurationInput != nil {
				m.editDurationInput.Blur()
			}
			m.currentView = ViewSelectTask
			return true
		}
	}
	return false
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
			if _, err := parseDuration(duration); err != nil {
				m.setStatusMessage("Invalid duration format. Use HH:MM (e.g., 1:30)")
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
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// setStatusMessage sets a status message with a timestamp
func (m *Model) setStatusMessage(msg string) {
	m.statusMessage = msg
	m.statusMessageTime = time.Now()
}

// clearStatusMessage clears the status message
func (m *Model) clearStatusMessage() {
	m.statusMessage = ""
	m.statusMessageTime = time.Time{}
}

// renderStatusLine returns the status message styled based on its content.
// Success messages render green, errors red, warnings yellow.
func (m Model) renderStatusLine() string {
	if m.statusMessage == "" {
		return ""
	}
	style := SuccessText
	msgLower := strings.ToLower(m.statusMessage)
	if strings.Contains(msgLower, "error") ||
		strings.Contains(msgLower, "failed") ||
		strings.Contains(msgLower, "cannot") ||
		strings.Contains(msgLower, "no tasks") ||
		strings.Contains(msgLower, "invalid") {
		style = ErrorText
	} else if strings.Contains(msgLower, "locked") ||
		strings.Contains(msgLower, "loading") {
		style = WarningText
	}
	return "  " + style.Render(m.statusMessage)
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
	hours, err := parseDuration(m.newEntryHours)
	if err != nil {
		return nil
	}

	request := harvest.CreateTimeEntryRequest{
		ProjectID: m.selectedProject.ID,
		TaskID:    m.selectedTask.ID,
		SpentDate: m.currentDate.Format("2006-01-02"),
		Hours:     hours,
		Notes:     m.newEntryNotes,
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
	hours, err := parseDuration(m.editHours)
	if err != nil {
		// Return an error message
		return func() tea.Msg {
			return timeEntryUpdatedMsg{err: fmt.Errorf("Invalid duration format. Use HH:MM (e.g., 1:30)")}
		}
	}

	request := harvest.UpdateTimeEntryRequest{
		Hours: &hours,
		Notes: &m.editNotes,
	}

	// Include TaskID if the task was changed
	if m.editTask != nil && m.editTask.ID != m.editingEntry.Task.ID {
		request.TaskID = &m.editTask.ID
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

// renderNewEntryModal renders the new entry form inside the shell box.
func (m Model) renderNewEntryModal() string {
	width := m.shellWidth()

	titleBar := m.renderTitleBar()

	breadcrumb := "  " + AccentText.Render("New Time Entry")

	divider := "  " + RenderDividerWidth(width-4)

	// Project field
	projectValue := "(none selected)"
	if m.selectedProject != nil {
		projectValue = fmt.Sprintf("%s → %s", m.selectedProject.Client.Name, m.selectedProject.Name)
	}

	// Task field
	taskValue := "(none selected)"
	if m.selectedTask != nil {
		taskValue = m.selectedTask.Name
	}

	// Notes field
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

	// Duration field
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

	// Status message
	statusLine := m.renderStatusLine()

	contentLines := []string{
		titleBar,
		breadcrumb,
		divider,
		"",
		"  " + fieldLabel("Project:", m.newEntryCurrentField == 0) + " " + projectValue,
		"",
		"  " + fieldLabel("Task:", m.newEntryCurrentField == 1) + " " + taskValue,
		"",
		"  " + fieldLabel("Notes:", m.newEntryCurrentField == 2) + " " + notesView,
		"",
		"  " + fieldLabel("Duration:", m.newEntryCurrentField == 3) + " " + durationView,
	}
	if statusLine != "" {
		contentLines = append(contentLines, "", statusLine)
	}

	content := strings.Join(contentLines, "\n")

	footerKeys := []string{
		RenderKeybinding("tab", "next"),
	}
	if m.newEntryCurrentField <= 1 {
		footerKeys = append(footerKeys, RenderKeybinding("enter", "select"))
	}
	footerKeys = append(footerKeys,
		RenderKeybinding("ctrl+s", "save"),
		RenderKeybinding("esc", "cancel"),
	)

	return m.buildShellBox(content, width, footerKeys)
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
		m.newEntryCurrentField = (m.newEntryCurrentField + 1) % 4
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
		m.newEntryCurrentField = (m.newEntryCurrentField - 1 + 4) % 4
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
		}
		return m, nil

	case "ctrl+s":
		// Save entry
		// Validate required fields
		if m.selectedProject == nil || m.selectedTask == nil {
			m.setStatusMessage("Please select a project and task")
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
		if _, err := parseDuration(m.newEntryHours); err != nil {
			m.setStatusMessage("Invalid duration format. Use HH:MM (e.g., 1:30)")
			return m, nil
		}

		return m, m.createTimeEntry()

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
