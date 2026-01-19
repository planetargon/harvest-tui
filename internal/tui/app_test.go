package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/planetargon/argon-harvest-tui/internal/config"
	"github.com/planetargon/argon-harvest-tui/internal/harvest"
	"github.com/planetargon/argon-harvest-tui/internal/state"
)

func TestNewModel(t *testing.T) {
	t.Run("given valid config and dependencies when NewModel called then returns model with correct initial state", func(t *testing.T) {
		cfg := &config.Config{
			Harvest: config.HarvestConfig{
				AccountID:   "12345",
				AccessToken: "test-token",
			},
		}
		client := harvest.NewClient("12345", "test-token")
		appState := &state.State{
			Recents: []state.RecentEntry{},
		}

		model := NewModel(cfg, client, appState)

		// Test initial view state
		if model.currentView != ViewList {
			t.Errorf("expected currentView to be ViewList, got %v", model.currentView)
		}

		// Test dependencies
		if model.config != cfg {
			t.Error("expected config to be set correctly")
		}
		if model.harvestClient != client {
			t.Error("expected harvestClient to be set correctly")
		}
		if model.appState != appState {
			t.Error("expected appState to be set correctly")
		}

		// Test data initialization
		if model.timeEntries == nil {
			t.Error("expected timeEntries to be initialized")
		}
		if len(model.timeEntries) != 0 {
			t.Errorf("expected timeEntries to be empty, got %d entries", len(model.timeEntries))
		}
		if model.projectsWithTasks == nil {
			t.Error("expected projectsWithTasks to be initialized")
		}
		if len(model.projectsWithTasks) != 0 {
			t.Errorf("expected projectsWithTasks to be empty, got %d entries", len(model.projectsWithTasks))
		}

		// Test date initialization (should be today)
		now := time.Now()
		if model.currentDate.Year() != now.Year() || model.currentDate.Month() != now.Month() || model.currentDate.Day() != now.Day() {
			t.Errorf("expected currentDate to be today, got %v", model.currentDate)
		}

		// Test selection state
		if model.selectedEntryIndex != 0 {
			t.Errorf("expected selectedEntryIndex to be 0, got %d", model.selectedEntryIndex)
		}

		// Test new entry state
		if model.selectedProject != nil {
			t.Error("expected selectedProject to be nil")
		}
		if model.selectedTask != nil {
			t.Error("expected selectedTask to be nil")
		}
		if model.newEntryNotes != "" {
			t.Errorf("expected newEntryNotes to be empty, got '%s'", model.newEntryNotes)
		}
		if model.newEntryHours != "" {
			t.Errorf("expected newEntryHours to be empty, got '%s'", model.newEntryHours)
		}
		if model.newEntryBillable != true {
			t.Errorf("expected newEntryBillable to be true, got %t", model.newEntryBillable)
		}

		// Test edit state
		if model.editingEntry != nil {
			t.Error("expected editingEntry to be nil")
		}
		if model.editNotes != "" {
			t.Errorf("expected editNotes to be empty, got '%s'", model.editNotes)
		}
		if model.editHours != "" {
			t.Errorf("expected editHours to be empty, got '%s'", model.editHours)
		}
		if model.editBillable != true {
			t.Errorf("expected editBillable to be true, got %t", model.editBillable)
		}

		// Test UI state
		if model.loading != false {
			t.Errorf("expected loading to be false, got %t", model.loading)
		}
		if model.errorMessage != "" {
			t.Errorf("expected errorMessage to be empty, got '%s'", model.errorMessage)
		}
		if model.statusMessage != "" {
			t.Errorf("expected statusMessage to be empty, got '%s'", model.statusMessage)
		}
		if model.showSpinner != false {
			t.Errorf("expected showSpinner to be false, got %t", model.showSpinner)
		}

		// Test window dimensions
		if model.width != 80 {
			t.Errorf("expected width to be 80, got %d", model.width)
		}
		if model.height != 24 {
			t.Errorf("expected height to be 24, got %d", model.height)
		}
	})
}

func TestModelInit(t *testing.T) {
	t.Run("given new model when Init called then returns fetch commands", func(t *testing.T) {
		cfg := &config.Config{
			Harvest: config.HarvestConfig{
				AccountID:   "12345",
				AccessToken: "test-token",
			},
		}
		client := harvest.NewClient("12345", "test-token")
		appState := &state.State{}

		model := NewModel(cfg, client, appState)
		cmd := model.Init()

		// Since we're returning tea.Batch(), we can't easily test the exact commands
		// without implementing the actual fetch commands, but we can verify it returns a command
		if cmd == nil {
			t.Error("expected Init to return a command, got nil")
		}
	})
}

func TestClearEditState(t *testing.T) {
	t.Run("given model with edit state when clearEditState called then resets all edit fields", func(t *testing.T) {
		cfg := &config.Config{
			Harvest: config.HarvestConfig{
				AccountID:   "12345",
				AccessToken: "test-token",
			},
		}
		client := harvest.NewClient("12345", "test-token")
		appState := &state.State{}

		model := NewModel(cfg, client, appState)

		// Set up some edit state
		model.selectedProject = &harvest.Project{ID: 1, Name: "Test Project"}
		model.selectedTask = &harvest.Task{ID: 2, Name: "Test Task"}
		model.newEntryNotes = "Some notes"
		model.newEntryHours = "2.5"
		model.newEntryBillable = false
		model.editingEntry = &harvest.TimeEntry{ID: 3}
		model.editNotes = "Edit notes"
		model.editHours = "1.5"
		model.editBillable = false

		// Clear edit state
		model.clearEditState()

		// Verify everything is cleared
		if model.selectedProject != nil {
			t.Error("expected selectedProject to be nil after clearEditState")
		}
		if model.selectedTask != nil {
			t.Error("expected selectedTask to be nil after clearEditState")
		}
		if model.newEntryNotes != "" {
			t.Errorf("expected newEntryNotes to be empty after clearEditState, got '%s'", model.newEntryNotes)
		}
		if model.newEntryHours != "" {
			t.Errorf("expected newEntryHours to be empty after clearEditState, got '%s'", model.newEntryHours)
		}
		if model.newEntryBillable != true {
			t.Errorf("expected newEntryBillable to be true after clearEditState, got %t", model.newEntryBillable)
		}
		if model.editingEntry != nil {
			t.Error("expected editingEntry to be nil after clearEditState")
		}
		if model.editNotes != "" {
			t.Errorf("expected editNotes to be empty after clearEditState, got '%s'", model.editNotes)
		}
		if model.editHours != "" {
			t.Errorf("expected editHours to be empty after clearEditState, got '%s'", model.editHours)
		}
		if model.editBillable != true {
			t.Errorf("expected editBillable to be true after clearEditState, got %t", model.editBillable)
		}
	})
}

func TestViewStateTransitions(t *testing.T) {
	t.Run("given model in different views when view state checked then returns correct state", func(t *testing.T) {
		cfg := &config.Config{
			Harvest: config.HarvestConfig{
				AccountID:   "12345",
				AccessToken: "test-token",
			},
		}
		client := harvest.NewClient("12345", "test-token")
		appState := &state.State{}

		model := NewModel(cfg, client, appState)

		// Test initial state
		if model.currentView != ViewList {
			t.Errorf("expected initial view to be ViewList, got %v", model.currentView)
		}

		// Test changing states
		model.currentView = ViewSelectProject
		if model.currentView != ViewSelectProject {
			t.Errorf("expected view to be ViewSelectProject, got %v", model.currentView)
		}

		model.currentView = ViewSelectTask
		if model.currentView != ViewSelectTask {
			t.Errorf("expected view to be ViewSelectTask, got %v", model.currentView)
		}

		model.currentView = ViewEditEntry
		if model.currentView != ViewEditEntry {
			t.Errorf("expected view to be ViewEditEntry, got %v", model.currentView)
		}

		model.currentView = ViewConfirmDelete
		if model.currentView != ViewConfirmDelete {
			t.Errorf("expected view to be ViewConfirmDelete, got %v", model.currentView)
		}

		model.currentView = ViewHelp
		if model.currentView != ViewHelp {
			t.Errorf("expected view to be ViewHelp, got %v", model.currentView)
		}
	})
}

func TestMainListViewRendering(t *testing.T) {
	t.Run("given model with time entries when renderListView called then displays entries with correct formatting", func(t *testing.T) {
		cfg := &config.Config{
			Harvest: config.HarvestConfig{
				AccountID:   "12345",
				AccessToken: "test-token",
			},
		}
		client := harvest.NewClient("12345", "test-token")
		appState := &state.State{}

		model := NewModel(cfg, client, appState)

		// Add mock time entries
		model.timeEntries = []harvest.TimeEntry{
			{
				ID:         1,
				SpentDate:  "2025-01-19",
				Hours:      2.5,
				Notes:      "Development work",
				IsRunning:  false,
				IsLocked:   false,
				IsBillable: true,
				Client:     harvest.TimeEntryClient{ID: 1, Name: "Acme Corp"},
				Project:    harvest.TimeEntryProject{ID: 2, Name: "Website Redesign"},
				Task:       harvest.TimeEntryTask{ID: 3, Name: "Frontend Development"},
			},
			{
				ID:         2,
				SpentDate:  "2025-01-19",
				Hours:      1.0,
				Notes:      "Daily standup",
				IsRunning:  true,
				IsLocked:   false,
				IsBillable: false,
				Client:     harvest.TimeEntryClient{ID: 1, Name: "Acme Corp"},
				Project:    harvest.TimeEntryProject{ID: 4, Name: "Internal"},
				Task:       harvest.TimeEntryTask{ID: 5, Name: "Meetings"},
			},
			{
				ID:         3,
				SpentDate:  "2025-01-19",
				Hours:      3.0,
				Notes:      "Code review",
				IsRunning:  false,
				IsLocked:   true,
				IsBillable: true,
				Client:     harvest.TimeEntryClient{ID: 2, Name: "BigCorp"},
				Project:    harvest.TimeEntryProject{ID: 6, Name: "API Integration"},
				Task:       harvest.TimeEntryTask{ID: 7, Name: "Code Review"},
			},
		}

		// Set current date
		model.currentDate = time.Date(2025, 1, 19, 0, 0, 0, 0, time.UTC)

		output := model.View()

		// Check that the output contains expected content
		if !strings.Contains(output, "Harvest TUI") {
			t.Error("expected output to contain title 'Harvest TUI'")
		}

		if !strings.Contains(output, "2025-01-19") && !strings.Contains(output, "January") {
			t.Errorf("expected output to contain current date, got output: %s", output)
		}

		if !strings.Contains(output, "Development work") {
			t.Error("expected output to contain first entry notes")
		}

		if !strings.Contains(output, "Acme Corp") {
			t.Error("expected output to contain client name")
		}

		if !strings.Contains(output, "Website Redesign") {
			t.Error("expected output to contain project name")
		}

		if !strings.Contains(output, "Frontend Development") {
			t.Error("expected output to contain task name")
		}

		// Check for running indicator
		if !strings.Contains(output, "Daily standup") && !strings.Contains(output, "⏱️") {
			t.Error("expected running entry to have indicator or be marked as running")
		}

		// Check for locked indicator
		if !strings.Contains(output, "Code review") && !strings.Contains(output, "🔒") {
			t.Error("expected locked entry to have indicator or be marked as locked")
		}
	})

	t.Run("given model with no time entries when renderListView called then displays empty state", func(t *testing.T) {
		cfg := &config.Config{
			Harvest: config.HarvestConfig{
				AccountID:   "12345",
				AccessToken: "test-token",
			},
		}
		client := harvest.NewClient("12345", "test-token")
		appState := &state.State{}

		model := NewModel(cfg, client, appState)
		model.currentDate = time.Date(2025, 1, 19, 0, 0, 0, 0, time.UTC)

		output := model.View()

		// Check for empty state message
		if !strings.Contains(output, "No time entries") && !strings.Contains(output, "Press 'n'") {
			t.Error("expected output to show empty state guidance")
		}

		if !strings.Contains(output, "2025-01-19") && !strings.Contains(output, "January") {
			t.Errorf("expected output to contain current date even when empty, got output: %s", output)
		}
	})

	t.Run("given model with loading state when renderListView called then displays spinner", func(t *testing.T) {
		cfg := &config.Config{
			Harvest: config.HarvestConfig{
				AccountID:   "12345",
				AccessToken: "test-token",
			},
		}
		client := harvest.NewClient("12345", "test-token")
		appState := &state.State{}

		model := NewModel(cfg, client, appState)
		model.loading = true

		output := model.View()

		// Check for loading indication
		if !strings.Contains(output, "Loading") && !strings.Contains(output, "...") {
			t.Error("expected output to show loading state")
		}
	})

	t.Run("given model with error when renderListView called then displays error message", func(t *testing.T) {
		cfg := &config.Config{
			Harvest: config.HarvestConfig{
				AccountID:   "12345",
				AccessToken: "test-token",
			},
		}
		client := harvest.NewClient("12345", "test-token")
		appState := &state.State{}

		model := NewModel(cfg, client, appState)
		model.errorMessage = "Failed to fetch data"

		output := model.View()

		// Check for error message
		if !strings.Contains(output, "Failed to fetch data") && !strings.Contains(output, "Error") {
			t.Error("expected output to show error message")
		}
	})
}

func TestHandleListViewKeys(t *testing.T) {
	cfg := &config.Config{
		Harvest: config.HarvestConfig{
			AccountID:   "12345",
			AccessToken: "test-token",
		},
	}
	client := harvest.NewClient("12345", "test-token")
	appState := &state.State{}

	t.Run("given model with time entries when up key pressed then moves selection up", func(t *testing.T) {
		model := NewModel(cfg, client, appState)
		model.timeEntries = []harvest.TimeEntry{
			{ID: 1, Notes: "First entry"},
			{ID: 2, Notes: "Second entry"},
			{ID: 3, Notes: "Third entry"},
		}
		model.selectedEntryIndex = 2

		msg := tea.KeyMsg{Type: tea.KeyUp}
		newModel, _ := model.handleListViewKeys(msg)
		m := newModel.(Model)

		if m.selectedEntryIndex != 1 {
			t.Errorf("expected selectedEntryIndex to be 1, got %d", m.selectedEntryIndex)
		}
	})

	t.Run("given model at top of list when up key pressed then selection stays at top", func(t *testing.T) {
		model := NewModel(cfg, client, appState)
		model.timeEntries = []harvest.TimeEntry{
			{ID: 1, Notes: "First entry"},
			{ID: 2, Notes: "Second entry"},
		}
		model.selectedEntryIndex = 0

		msg := tea.KeyMsg{Type: tea.KeyUp}
		newModel, _ := model.handleListViewKeys(msg)
		m := newModel.(Model)

		if m.selectedEntryIndex != 0 {
			t.Errorf("expected selectedEntryIndex to stay at 0, got %d", m.selectedEntryIndex)
		}
	})

	t.Run("given model with time entries when down key pressed then moves selection down", func(t *testing.T) {
		model := NewModel(cfg, client, appState)
		model.timeEntries = []harvest.TimeEntry{
			{ID: 1, Notes: "First entry"},
			{ID: 2, Notes: "Second entry"},
			{ID: 3, Notes: "Third entry"},
		}
		model.selectedEntryIndex = 0

		msg := tea.KeyMsg{Type: tea.KeyDown}
		newModel, _ := model.handleListViewKeys(msg)
		m := newModel.(Model)

		if m.selectedEntryIndex != 1 {
			t.Errorf("expected selectedEntryIndex to be 1, got %d", m.selectedEntryIndex)
		}
	})

	t.Run("given model at bottom of list when down key pressed then selection stays at bottom", func(t *testing.T) {
		model := NewModel(cfg, client, appState)
		model.timeEntries = []harvest.TimeEntry{
			{ID: 1, Notes: "First entry"},
			{ID: 2, Notes: "Second entry"},
		}
		model.selectedEntryIndex = 1

		msg := tea.KeyMsg{Type: tea.KeyDown}
		newModel, _ := model.handleListViewKeys(msg)
		m := newModel.(Model)

		if m.selectedEntryIndex != 1 {
			t.Errorf("expected selectedEntryIndex to stay at 1, got %d", m.selectedEntryIndex)
		}
	})

	t.Run("given empty time entries when navigation keys pressed then selection stays at 0", func(t *testing.T) {
		model := NewModel(cfg, client, appState)
		model.timeEntries = []harvest.TimeEntry{}
		model.selectedEntryIndex = 0

		// Test up key
		msg := tea.KeyMsg{Type: tea.KeyUp}
		newModel, _ := model.handleListViewKeys(msg)
		m := newModel.(Model)

		if m.selectedEntryIndex != 0 {
			t.Errorf("expected selectedEntryIndex to stay at 0 with empty entries, got %d", m.selectedEntryIndex)
		}

		// Test down key
		msg = tea.KeyMsg{Type: tea.KeyDown}
		newModel, _ = model.handleListViewKeys(msg)
		m = newModel.(Model)

		if m.selectedEntryIndex != 0 {
			t.Errorf("expected selectedEntryIndex to stay at 0 with empty entries, got %d", m.selectedEntryIndex)
		}
	})
}

func TestDateNavigation(t *testing.T) {
	cfg := &config.Config{
		Harvest: config.HarvestConfig{
			AccountID:   "12345",
			AccessToken: "test-token",
		},
	}
	client := harvest.NewClient("12345", "test-token")
	appState := &state.State{}

	t.Run("given model with current date when left arrow pressed then navigates to previous day", func(t *testing.T) {
		model := NewModel(cfg, client, appState)
		startDate := time.Date(2025, 1, 19, 0, 0, 0, 0, time.UTC)
		model.currentDate = startDate
		model.selectedEntryIndex = 5

		msg := tea.KeyMsg{Type: tea.KeyLeft}
		newModel, cmd := model.handleListViewKeys(msg)
		m := newModel.(Model)

		expectedDate := startDate.AddDate(0, 0, -1)
		if m.currentDate != expectedDate {
			t.Errorf("expected currentDate to be %v, got %v", expectedDate, m.currentDate)
		}

		if m.selectedEntryIndex != 0 {
			t.Errorf("expected selectedEntryIndex to reset to 0, got %d", m.selectedEntryIndex)
		}

		if !m.loading {
			t.Error("expected loading to be true after date navigation")
		}

		if cmd == nil {
			t.Error("expected fetch command to be returned")
		}
	})

	t.Run("given model with current date when right arrow pressed then navigates to next day", func(t *testing.T) {
		model := NewModel(cfg, client, appState)
		startDate := time.Date(2025, 1, 19, 0, 0, 0, 0, time.UTC)
		model.currentDate = startDate
		model.selectedEntryIndex = 3

		msg := tea.KeyMsg{Type: tea.KeyRight}
		newModel, cmd := model.handleListViewKeys(msg)
		m := newModel.(Model)

		expectedDate := startDate.AddDate(0, 0, 1)
		if m.currentDate != expectedDate {
			t.Errorf("expected currentDate to be %v, got %v", expectedDate, m.currentDate)
		}

		if m.selectedEntryIndex != 0 {
			t.Errorf("expected selectedEntryIndex to reset to 0, got %d", m.selectedEntryIndex)
		}

		if !m.loading {
			t.Error("expected loading to be true after date navigation")
		}

		if cmd == nil {
			t.Error("expected fetch command to be returned")
		}
	})

	t.Run("given model with any date when 't' pressed then navigates to today", func(t *testing.T) {
		model := NewModel(cfg, client, appState)
		pastDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		model.currentDate = pastDate
		model.selectedEntryIndex = 2

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}}
		newModel, cmd := model.handleListViewKeys(msg)
		m := newModel.(Model)

		now := time.Now()
		if m.currentDate.Year() != now.Year() || m.currentDate.Month() != now.Month() || m.currentDate.Day() != now.Day() {
			t.Errorf("expected currentDate to be today, got %v", m.currentDate)
		}

		if m.selectedEntryIndex != 0 {
			t.Errorf("expected selectedEntryIndex to reset to 0, got %d", m.selectedEntryIndex)
		}

		if !m.loading {
			t.Error("expected loading to be true after jumping to today")
		}

		if cmd == nil {
			t.Error("expected fetch command to be returned")
		}
	})
}

func TestNewEntryAction(t *testing.T) {
	cfg := &config.Config{
		Harvest: config.HarvestConfig{
			AccountID:   "12345",
			AccessToken: "test-token",
		},
	}
	client := harvest.NewClient("12345", "test-token")
	appState := &state.State{}

	t.Run("given model with projects when 'n' pressed then transitions to project selection", func(t *testing.T) {
		model := NewModel(cfg, client, appState)
		model.projectsWithTasks = []harvest.ProjectWithTasks{
			{
				Project: harvest.Project{ID: 1, Name: "Test Project"},
				Tasks:   []harvest.Task{{ID: 1, Name: "Test Task"}},
			},
		}

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
		newModel, _ := model.handleListViewKeys(msg)
		m := newModel.(Model)

		if m.currentView != ViewSelectProject {
			t.Errorf("expected currentView to be ViewSelectProject, got %v", m.currentView)
		}

		// Verify edit state is cleared
		if m.selectedProject != nil || m.selectedTask != nil || m.newEntryNotes != "" || m.newEntryHours != "" {
			t.Error("expected edit state to be cleared when starting new entry")
		}
	})

	t.Run("given model with no projects when 'n' pressed then shows error message", func(t *testing.T) {
		model := NewModel(cfg, client, appState)
		model.projectsWithTasks = []harvest.ProjectWithTasks{}

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
		newModel, _ := model.handleListViewKeys(msg)
		m := newModel.(Model)

		if m.currentView != ViewList {
			t.Errorf("expected currentView to stay ViewList, got %v", m.currentView)
		}

		if m.statusMessage == "" {
			t.Error("expected status message to be set when no projects available")
		}

		if !strings.Contains(m.statusMessage, "No projects available") {
			t.Errorf("expected status message about no projects, got '%s'", m.statusMessage)
		}
	})
}

func TestEditEntryAction(t *testing.T) {
	cfg := &config.Config{
		Harvest: config.HarvestConfig{
			AccountID:   "12345",
			AccessToken: "test-token",
		},
	}
	client := harvest.NewClient("12345", "test-token")
	appState := &state.State{}

	t.Run("given model with unlocked entry when 'e' pressed then transitions to edit view", func(t *testing.T) {
		model := NewModel(cfg, client, appState)
		model.timeEntries = []harvest.TimeEntry{
			{
				ID:         1,
				Notes:      "Test notes",
				Hours:      2.5,
				IsBillable: true,
				IsLocked:   false,
			},
		}
		model.selectedEntryIndex = 0

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}
		newModel, _ := model.handleListViewKeys(msg)
		m := newModel.(Model)

		if m.currentView != ViewEditEntry {
			t.Errorf("expected currentView to be ViewEditEntry, got %v", m.currentView)
		}

		if m.editingEntry == nil {
			t.Error("expected editingEntry to be set")
		}

		if m.editNotes != "Test notes" {
			t.Errorf("expected editNotes to be 'Test notes', got '%s'", m.editNotes)
		}

		if m.editHours != "2:30" {
			t.Errorf("expected editHours to be '2:30', got '%s'", m.editHours)
		}

		if m.editBillable != true {
			t.Error("expected editBillable to be true")
		}
	})

	t.Run("given model with locked entry when 'e' pressed then shows error message", func(t *testing.T) {
		model := NewModel(cfg, client, appState)
		model.timeEntries = []harvest.TimeEntry{
			{
				ID:       1,
				IsLocked: true,
			},
		}
		model.selectedEntryIndex = 0

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}
		newModel, _ := model.handleListViewKeys(msg)
		m := newModel.(Model)

		if m.currentView != ViewList {
			t.Errorf("expected currentView to stay ViewList, got %v", m.currentView)
		}

		if !strings.Contains(m.statusMessage, "Cannot edit locked") {
			t.Errorf("expected status message about locked entry, got '%s'", m.statusMessage)
		}
	})

	t.Run("given model with no entries when 'e' pressed then nothing happens", func(t *testing.T) {
		model := NewModel(cfg, client, appState)
		model.timeEntries = []harvest.TimeEntry{}

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}
		newModel, _ := model.handleListViewKeys(msg)
		m := newModel.(Model)

		if m.currentView != ViewList {
			t.Errorf("expected currentView to stay ViewList, got %v", m.currentView)
		}
	})
}

func TestDeleteEntryAction(t *testing.T) {
	cfg := &config.Config{
		Harvest: config.HarvestConfig{
			AccountID:   "12345",
			AccessToken: "test-token",
		},
	}
	client := harvest.NewClient("12345", "test-token")
	appState := &state.State{}

	t.Run("given model with unlocked stopped entry when 'd' pressed then transitions to confirm delete", func(t *testing.T) {
		model := NewModel(cfg, client, appState)
		model.timeEntries = []harvest.TimeEntry{
			{
				ID:        1,
				IsLocked:  false,
				IsRunning: false,
			},
		}
		model.selectedEntryIndex = 0

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
		newModel, _ := model.handleListViewKeys(msg)
		m := newModel.(Model)

		if m.currentView != ViewConfirmDelete {
			t.Errorf("expected currentView to be ViewConfirmDelete, got %v", m.currentView)
		}

		if m.editingEntry == nil {
			t.Error("expected editingEntry to be set for delete confirmation")
		}
	})

	t.Run("given model with locked entry when 'd' pressed then shows error message", func(t *testing.T) {
		model := NewModel(cfg, client, appState)
		model.timeEntries = []harvest.TimeEntry{
			{
				ID:        1,
				IsLocked:  true,
				IsRunning: false,
			},
		}
		model.selectedEntryIndex = 0

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
		newModel, _ := model.handleListViewKeys(msg)
		m := newModel.(Model)

		if m.currentView != ViewList {
			t.Errorf("expected currentView to stay ViewList, got %v", m.currentView)
		}

		if !strings.Contains(m.statusMessage, "Cannot delete locked") {
			t.Errorf("expected status message about locked entry, got '%s'", m.statusMessage)
		}
	})

	t.Run("given model with running entry when 'd' pressed then shows error message", func(t *testing.T) {
		model := NewModel(cfg, client, appState)
		model.timeEntries = []harvest.TimeEntry{
			{
				ID:        1,
				IsLocked:  false,
				IsRunning: true,
			},
		}
		model.selectedEntryIndex = 0

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
		newModel, _ := model.handleListViewKeys(msg)
		m := newModel.(Model)

		if m.currentView != ViewList {
			t.Errorf("expected currentView to stay ViewList, got %v", m.currentView)
		}

		if !strings.Contains(m.statusMessage, "Cannot delete running") {
			t.Errorf("expected status message about running entry, got '%s'", m.statusMessage)
		}
	})
}

func TestStartStopTimerActions(t *testing.T) {
	cfg := &config.Config{
		Harvest: config.HarvestConfig{
			AccountID:   "12345",
			AccessToken: "test-token",
		},
	}
	client := harvest.NewClient("12345", "test-token")
	appState := &state.State{}

	t.Run("given model with stopped unlocked entry when 's' pressed then starts timer", func(t *testing.T) {
		model := NewModel(cfg, client, appState)
		model.timeEntries = []harvest.TimeEntry{
			{
				ID:        1,
				IsLocked:  false,
				IsRunning: false,
			},
		}
		model.selectedEntryIndex = 0

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}
		newModel, cmd := model.handleListViewKeys(msg)
		m := newModel.(Model)

		if cmd == nil {
			t.Error("expected start timer command to be returned")
		}

		// Model should remain in ViewList
		if m.currentView != ViewList {
			t.Errorf("expected currentView to stay ViewList, got %v", m.currentView)
		}
	})

	t.Run("given model with running entry when 's' pressed then shows already running message", func(t *testing.T) {
		model := NewModel(cfg, client, appState)
		model.timeEntries = []harvest.TimeEntry{
			{
				ID:        1,
				IsLocked:  false,
				IsRunning: true,
			},
		}
		model.selectedEntryIndex = 0

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}
		newModel, _ := model.handleListViewKeys(msg)
		m := newModel.(Model)

		if !strings.Contains(m.statusMessage, "already running") {
			t.Errorf("expected status message about already running, got '%s'", m.statusMessage)
		}
	})

	t.Run("given model with locked entry when 's' pressed then shows locked message", func(t *testing.T) {
		model := NewModel(cfg, client, appState)
		model.timeEntries = []harvest.TimeEntry{
			{
				ID:        1,
				IsLocked:  true,
				IsRunning: false,
			},
		}
		model.selectedEntryIndex = 0

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}
		newModel, _ := model.handleListViewKeys(msg)
		m := newModel.(Model)

		if !strings.Contains(m.statusMessage, "Cannot start locked") {
			t.Errorf("expected status message about locked entry, got '%s'", m.statusMessage)
		}
	})

	t.Run("given model with running unlocked entry when 'S' pressed then stops timer", func(t *testing.T) {
		model := NewModel(cfg, client, appState)
		model.timeEntries = []harvest.TimeEntry{
			{
				ID:        1,
				IsLocked:  false,
				IsRunning: true,
			},
		}
		model.selectedEntryIndex = 0

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'S'}}
		newModel, cmd := model.handleListViewKeys(msg)
		m := newModel.(Model)

		if cmd == nil {
			t.Error("expected stop timer command to be returned")
		}

		// Model should remain in ViewList
		if m.currentView != ViewList {
			t.Errorf("expected currentView to stay ViewList, got %v", m.currentView)
		}
	})

	t.Run("given model with stopped entry when 'S' pressed then shows not running message", func(t *testing.T) {
		model := NewModel(cfg, client, appState)
		model.timeEntries = []harvest.TimeEntry{
			{
				ID:        1,
				IsLocked:  false,
				IsRunning: false,
			},
		}
		model.selectedEntryIndex = 0

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'S'}}
		newModel, _ := model.handleListViewKeys(msg)
		m := newModel.(Model)

		if !strings.Contains(m.statusMessage, "not running") {
			t.Errorf("expected status message about not running, got '%s'", m.statusMessage)
		}
	})
}

func TestGlobalKeyHandling(t *testing.T) {
	cfg := &config.Config{
		Harvest: config.HarvestConfig{
			AccountID:   "12345",
			AccessToken: "test-token",
		},
	}
	client := harvest.NewClient("12345", "test-token")
	appState := &state.State{}

	t.Run("given model in list view when escape pressed then quits application", func(t *testing.T) {
		model := NewModel(cfg, client, appState)
		model.currentView = ViewList

		msg := tea.KeyMsg{Type: tea.KeyEsc}
		_, cmd := model.Update(msg)

		// Should return quit command
		if cmd == nil {
			t.Error("expected quit command to be returned")
		}
	})

	t.Run("given model in edit view when escape pressed then returns to list view", func(t *testing.T) {
		model := NewModel(cfg, client, appState)
		model.currentView = ViewEditEntry
		model.selectedProject = &harvest.Project{ID: 1}
		model.newEntryNotes = "test"

		msg := tea.KeyMsg{Type: tea.KeyEsc}
		newModel, _ := model.Update(msg)
		m := newModel.(Model)

		if m.currentView != ViewList {
			t.Errorf("expected currentView to be ViewList, got %v", m.currentView)
		}

		// Edit state should be cleared
		if m.selectedProject != nil || m.newEntryNotes != "" {
			t.Error("expected edit state to be cleared on escape")
		}
	})

	t.Run("given model in list view when '?' pressed then shows help", func(t *testing.T) {
		model := NewModel(cfg, client, appState)
		model.currentView = ViewList

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
		newModel, _ := model.Update(msg)
		m := newModel.(Model)

		if m.currentView != ViewHelp {
			t.Errorf("expected currentView to be ViewHelp, got %v", m.currentView)
		}
	})

	t.Run("given model in help view when '?' pressed then returns to list", func(t *testing.T) {
		model := NewModel(cfg, client, appState)
		model.currentView = ViewHelp

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
		newModel, _ := model.Update(msg)
		m := newModel.(Model)

		if m.currentView != ViewList {
			t.Errorf("expected currentView to be ViewList, got %v", m.currentView)
		}
	})
}
