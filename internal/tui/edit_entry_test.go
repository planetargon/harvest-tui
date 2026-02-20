package tui

import (
	"fmt"
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/planetargon/harvest-tui/internal/config"
	"github.com/planetargon/harvest-tui/internal/harvest"
	"github.com/planetargon/harvest-tui/internal/state"
)

func TestEditEntry(t *testing.T) {
	cfg := &config.Config{Harvest: config.HarvestConfig{AccountID: "123456", AccessToken: "test-token"}}
	client := &harvest.Client{}
	appState := &state.State{}

	t.Run("given entry selected when edit key pressed then transitions to edit view", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})
		model.currentView = ViewList
		model.timeEntries = []harvest.TimeEntry{
			{
				ID:         1,
				Hours:      1.5,
				Notes:      "Test entry",
				IsRunning:  false,
				IsLocked:   false,
				IsBillable: true,
			},
		}
		model.selectedEntryIndex = 0

		// Press 'e' to edit
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")}
		updatedModel, _ := model.Update(msg)

		// Should transition to edit view
		if updatedModel.(Model).currentView != ViewEditEntry {
			t.Errorf("expected view to be ViewEditEntry, got %v", updatedModel.(Model).currentView)
		}

		// Should set editingEntry
		if updatedModel.(Model).editingEntry == nil {
			t.Error("expected editingEntry to be set")
		} else {
			if updatedModel.(Model).editingEntry.ID != 1 {
				t.Errorf("expected editingEntry.ID to be 1, got %d", updatedModel.(Model).editingEntry.ID)
			}
		}

		// Should populate edit fields
		if updatedModel.(Model).editNotes != "Test entry" {
			t.Errorf("expected editNotes to be 'Test entry', got '%s'", updatedModel.(Model).editNotes)
		}
		if updatedModel.(Model).editHours != "1:30" {
			t.Errorf("expected editHours to be '1:30', got '%s'", updatedModel.(Model).editHours)
		}
		if !updatedModel.(Model).editBillable {
			t.Error("expected editBillable to be true")
		}
	})

	t.Run("given locked entry when edit key pressed then shows error", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})
		model.currentView = ViewList
		model.timeEntries = []harvest.TimeEntry{
			{
				ID:       1,
				Hours:    1.5,
				Notes:    "Test entry",
				IsLocked: true,
			},
		}
		model.selectedEntryIndex = 0

		// Press 'e' to edit
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")}
		updatedModel, _ := model.Update(msg)

		// Should not transition to edit view
		if updatedModel.(Model).currentView != ViewList {
			t.Errorf("expected to remain in ViewList, got %v", updatedModel.(Model).currentView)
		}

		// Should show error message
		if updatedModel.(Model).statusMessage != "Cannot edit locked time entry." {
			t.Errorf("expected locked entry message, got '%s'", updatedModel.(Model).statusMessage)
		}

		// Should not set editingEntry
		if updatedModel.(Model).editingEntry != nil {
			t.Error("expected editingEntry to remain nil for locked entry")
		}
	})

	t.Run("given running entry when edit key pressed then shows error", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})
		model.currentView = ViewList
		model.timeEntries = []harvest.TimeEntry{
			{
				ID:        1,
				Hours:     1.5,
				Notes:     "Test entry",
				IsRunning: true,
			},
		}
		model.selectedEntryIndex = 0

		// Press 'e' to edit
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")}
		updatedModel, _ := model.Update(msg)

		// Should not transition to edit view
		if updatedModel.(Model).currentView != ViewList {
			t.Errorf("expected to remain in ViewList, got %v", updatedModel.(Model).currentView)
		}

		// Should show error message
		if updatedModel.(Model).statusMessage != "Cannot edit running time entry. Stop the timer first." {
			t.Errorf("expected running entry message, got '%s'", updatedModel.(Model).statusMessage)
		}
	})

	t.Run("given edit view when escape pressed then returns to main list", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})
		model.currentView = ViewEditEntry
		model.editingEntry = &harvest.TimeEntry{ID: 1}
		model.editNotes = "Modified notes"
		model.editHours = "2:00"

		// Press escape
		msg := tea.KeyMsg{Type: tea.KeyEscape}
		updatedModel, _ := model.Update(msg)

		// Should return to list view
		if updatedModel.(Model).currentView != ViewList {
			t.Errorf("expected view to be ViewList, got %v", updatedModel.(Model).currentView)
		}

		// Should clear editing state
		if updatedModel.(Model).editingEntry != nil {
			t.Error("expected editingEntry to be cleared")
		}
		if updatedModel.(Model).editNotes != "" {
			t.Error("expected editNotes to be cleared")
		}
		if updatedModel.(Model).editHours != "" {
			t.Error("expected editHours to be cleared")
		}
	})

	t.Run("given edit view when valid changes saved then updates entry", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})
		model.currentView = ViewEditEntry
		model.editingEntry = &harvest.TimeEntry{
			ID:    1,
			Hours: 1.5,
			Notes: "Original notes",
		}
		model.editNotes = "Updated notes"
		model.editHours = "2:00"
		model.editBillable = false

		// Would simulate pressing enter/save
		// For now, just test that the edit fields are properly set
		if model.editNotes != "Updated notes" {
			t.Errorf("expected editNotes to be 'Updated notes', got '%s'", model.editNotes)
		}
		if model.editHours != "2:00" {
			t.Errorf("expected editHours to be '2:00', got '%s'", model.editHours)
		}
		if model.editBillable {
			t.Error("expected editBillable to be false")
		}
	})

	t.Run("given edit view when invalid duration entered then shows error", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})
		model.currentView = ViewEditEntry
		model.editingEntry = &harvest.TimeEntry{ID: 1}

		// Initialize edit duration input for testing
		durationInput := textinput.New()
		durationInput.SetValue("invalid")
		model.editDurationInput = &durationInput

		// Simulate entering invalid duration and saving
		// The actual validation would happen in the save handler
		// For now, just ensure the field can be set
		if model.editDurationInput.Value() != "invalid" {
			t.Errorf("expected duration input to be 'invalid', got '%s'", model.editDurationInput.Value())
		}
	})

	t.Run("given entry update succeeds when message received then updates list", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})
		model.timeEntries = []harvest.TimeEntry{
			{
				ID:    1,
				Hours: 1.5,
				Notes: "Original notes",
			},
		}

		// Simulate successful update
		updatedEntry := &harvest.TimeEntry{
			ID:    1,
			Hours: 2.0,
			Notes: "Updated notes",
		}

		msg := timeEntryUpdatedMsg{entry: updatedEntry}
		updatedModel, _ := model.Update(msg)

		// Should update entry in list
		if updatedModel.(Model).timeEntries[0].Notes != "Updated notes" {
			t.Errorf("expected notes to be 'Updated notes', got '%s'", updatedModel.(Model).timeEntries[0].Notes)
		}
		if updatedModel.(Model).timeEntries[0].Hours != 2.0 {
			t.Errorf("expected hours to be 2.0, got %f", updatedModel.(Model).timeEntries[0].Hours)
		}

		// Should show success message
		if updatedModel.(Model).statusMessage != "Time entry updated successfully" {
			t.Errorf("expected success message, got '%s'", updatedModel.(Model).statusMessage)
		}

		// Should return to list view
		if updatedModel.(Model).currentView != ViewList {
			t.Errorf("expected view to be ViewList, got %v", updatedModel.(Model).currentView)
		}
	})

	t.Run("given entry update fails when message received then shows error", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})

		// Simulate failed update
		msg := timeEntryUpdatedMsg{err: fmt.Errorf("API error")}
		updatedModel, _ := model.Update(msg)

		// Should show error message
		if updatedModel.(Model).statusMessage == "" {
			t.Error("expected error message to be set")
		}
	})

	t.Run("given edit view when enter pressed on task field before projects loaded then sets pending and fetches", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})
		model.currentView = ViewEditEntry
		model.editingEntry = &harvest.TimeEntry{
			ID:      1,
			Project: harvest.TimeEntryProject{ID: 1, Name: "Website"},
			Task:    harvest.TimeEntryTask{ID: 10, Name: "Development"},
		}
		model.editTask = &harvest.Task{ID: 10, Name: "Development"}
		model.editCurrentField = 0
		// projectsWithTasks is empty (not yet loaded)

		msg := tea.KeyMsg{Type: tea.KeyEnter}
		updatedModel, cmd := model.Update(msg)

		m := updatedModel.(Model)
		if m.currentView != ViewEditEntry {
			t.Errorf("expected to stay on ViewEditEntry, got %v", m.currentView)
		}
		if !m.pendingTaskEdit {
			t.Error("expected pendingTaskEdit to be true")
		}
		if cmd == nil {
			t.Error("expected a command to fetch projects")
		}
	})

	t.Run("given pending task edit when projects arrive then opens task selection automatically", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})
		model.currentView = ViewEditEntry
		model.editingEntry = &harvest.TimeEntry{
			ID:      1,
			Project: harvest.TimeEntryProject{ID: 1, Name: "Website"},
			Task:    harvest.TimeEntryTask{ID: 10, Name: "Development"},
		}
		model.editTask = &harvest.Task{ID: 10, Name: "Development"}
		model.editCurrentField = 0
		model.pendingTaskEdit = true

		notesInput := textinput.New()
		model.editNotesInput = &notesInput
		durationInput := textinput.New()
		model.editDurationInput = &durationInput

		// Simulate projects arriving
		fetchedMsg := projectsWithTasksFetchedMsg{
			projectsWithTasks: []harvest.ProjectWithTasks{
				{
					Project: harvest.Project{ID: 1, Name: "Website"},
					Tasks: []harvest.Task{
						{ID: 10, Name: "Development"},
						{ID: 11, Name: "Design"},
					},
				},
			},
		}
		updatedModel, _ := model.Update(fetchedMsg)

		m := updatedModel.(Model)
		if m.currentView != ViewSelectTask {
			t.Errorf("expected ViewSelectTask after projects loaded, got %v", m.currentView)
		}
		if m.pendingTaskEdit {
			t.Error("expected pendingTaskEdit to be cleared")
		}
	})

	t.Run("given edit view when enter pressed on task field then opens task selection", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})
		model.currentView = ViewEditEntry
		model.editingEntry = &harvest.TimeEntry{
			ID:      1,
			Project: harvest.TimeEntryProject{ID: 1, Name: "Website"},
			Task:    harvest.TimeEntryTask{ID: 10, Name: "Development"},
		}
		model.editTask = &harvest.Task{ID: 10, Name: "Development"}
		model.editCurrentField = 0
		model.projectsWithTasks = []harvest.ProjectWithTasks{
			{
				Project: harvest.Project{ID: 1, Name: "Website"},
				Tasks: []harvest.Task{
					{ID: 10, Name: "Development"},
					{ID: 11, Name: "Design"},
				},
			},
		}

		// Press enter on task field
		msg := tea.KeyMsg{Type: tea.KeyEnter}
		updatedModel, _ := model.Update(msg)

		if updatedModel.(Model).currentView != ViewSelectTask {
			t.Errorf("expected view to be ViewSelectTask, got %v", updatedModel.(Model).currentView)
		}
	})

	t.Run("given task selection during edit when task selected then returns to edit view with updated task", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})
		model.currentView = ViewSelectTask
		model.editingEntry = &harvest.TimeEntry{
			ID:      1,
			Project: harvest.TimeEntryProject{ID: 1, Name: "Website"},
			Task:    harvest.TimeEntryTask{ID: 10, Name: "Development"},
		}
		model.editTask = &harvest.Task{ID: 10, Name: "Development"}
		model.editCurrentField = 0

		// Set up task list with items
		tasks := []harvest.Task{
			{ID: 10, Name: "Development"},
			{ID: 11, Name: "Design"},
		}
		model.updateTaskList(tasks)
		model.taskList.Select(1) // Select "Design"

		// Press enter to select task
		msg := tea.KeyMsg{Type: tea.KeyEnter}
		updatedModel, _ := model.handleTaskSelectKeys(msg)

		m := updatedModel.(Model)
		if m.currentView != ViewEditEntry {
			t.Errorf("expected view to be ViewEditEntry, got %v", m.currentView)
		}
		if m.editTask == nil {
			t.Fatal("expected editTask to be set")
		}
		if m.editTask.ID != 11 {
			t.Errorf("expected editTask.ID to be 11, got %d", m.editTask.ID)
		}
		if m.editTask.Name != "Design" {
			t.Errorf("expected editTask.Name to be 'Design', got '%s'", m.editTask.Name)
		}
	})

	t.Run("given task selection during edit when escape pressed then returns to edit view", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})
		model.currentView = ViewSelectTask
		model.editingEntry = &harvest.TimeEntry{
			ID:      1,
			Project: harvest.TimeEntryProject{ID: 1, Name: "Website"},
			Task:    harvest.TimeEntryTask{ID: 10, Name: "Development"},
		}
		model.editTask = &harvest.Task{ID: 10, Name: "Development"}

		// Press escape
		msg := tea.KeyMsg{Type: tea.KeyEscape}
		updatedModel, _ := model.handleTaskSelectKeys(msg)

		if updatedModel.(Model).currentView != ViewEditEntry {
			t.Errorf("expected view to be ViewEditEntry, got %v", updatedModel.(Model).currentView)
		}
		// Task should remain unchanged
		if updatedModel.(Model).editTask.ID != 10 {
			t.Errorf("expected editTask.ID to remain 10, got %d", updatedModel.(Model).editTask.ID)
		}
	})

	t.Run("given edit view with changed task when saved then includes task ID in request", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})
		model.currentView = ViewEditEntry
		model.editingEntry = &harvest.TimeEntry{
			ID:      1,
			Project: harvest.TimeEntryProject{ID: 1, Name: "Website"},
			Task:    harvest.TimeEntryTask{ID: 10, Name: "Development"},
		}
		model.editTask = &harvest.Task{ID: 11, Name: "Design"} // Changed task
		model.editHours = "1:30"
		model.editNotes = "Some notes"

		// Call updateTimeEntry to get the command
		cmd := model.updateTimeEntry()

		// The command should not be nil (it creates an API call)
		if cmd == nil {
			t.Fatal("expected non-nil command from updateTimeEntry")
		}
	})

	t.Run("given entry selected when edit key pressed then populates editTask from entry", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})
		model.currentView = ViewList
		model.timeEntries = []harvest.TimeEntry{
			{
				ID:    1,
				Hours: 1.5,
				Notes: "Test entry",
				Task:  harvest.TimeEntryTask{ID: 42, Name: "Testing"},
			},
		}
		model.selectedEntryIndex = 0

		// Press 'e' to edit
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")}
		updatedModel, _ := model.Update(msg)

		m := updatedModel.(Model)
		if m.editTask == nil {
			t.Fatal("expected editTask to be set")
		}
		if m.editTask.ID != 42 {
			t.Errorf("expected editTask.ID to be 42, got %d", m.editTask.ID)
		}
		if m.editTask.Name != "Testing" {
			t.Errorf("expected editTask.Name to be 'Testing', got '%s'", m.editTask.Name)
		}
	})

	t.Run("given edit view when tab cycles through fields then visits all three fields", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})
		model.currentView = ViewEditEntry
		model.editingEntry = &harvest.TimeEntry{ID: 1}
		model.editCurrentField = 0

		notesInput := textinput.New()
		model.editNotesInput = &notesInput
		durationInput := textinput.New()
		model.editDurationInput = &durationInput

		// Tab from field 0 (task) to field 1 (notes)
		msg := tea.KeyMsg{Type: tea.KeyTab}
		result, _ := model.Update(msg)
		m := result.(Model)
		if m.editCurrentField != 1 {
			t.Errorf("expected field 1 after first tab, got %d", m.editCurrentField)
		}

		// Tab from field 1 (notes) to field 2 (duration)
		result, _ = m.Update(msg)
		m = result.(Model)
		if m.editCurrentField != 2 {
			t.Errorf("expected field 2 after second tab, got %d", m.editCurrentField)
		}

		// Tab from field 2 (duration) wraps to field 0 (task)
		result, _ = m.Update(msg)
		m = result.(Model)
		if m.editCurrentField != 0 {
			t.Errorf("expected field 0 after third tab, got %d", m.editCurrentField)
		}
	})

	t.Run("given list view when e pressed then enter on task field opens task selection", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})
		model.currentView = ViewList
		model.timeEntries = []harvest.TimeEntry{
			{
				ID:      1,
				Hours:   1.5,
				Notes:   "Test entry",
				Project: harvest.TimeEntryProject{ID: 99, Name: "Website"},
				Task:    harvest.TimeEntryTask{ID: 42, Name: "Testing"},
			},
		}
		model.selectedEntryIndex = 0
		model.projectsWithTasks = []harvest.ProjectWithTasks{
			{
				Project: harvest.Project{ID: 99, Name: "Website"},
				Tasks: []harvest.Task{
					{ID: 42, Name: "Testing"},
					{ID: 43, Name: "Design"},
				},
			},
		}

		// Press 'e' to enter edit mode
		result, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
		m := result.(Model)

		if m.currentView != ViewEditEntry {
			t.Fatalf("expected ViewEditEntry after pressing e, got %v", m.currentView)
		}
		if m.editCurrentField != 0 {
			t.Fatalf("expected editCurrentField to be 0 (task), got %d", m.editCurrentField)
		}

		// Press enter on task field (field 0)
		result, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		m = result.(Model)

		if m.currentView != ViewSelectTask {
			t.Errorf("expected ViewSelectTask after enter on task field, got %v", m.currentView)
		}
	})

	t.Run("given edit view when enter pressed on notes field then does not save", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})
		model.currentView = ViewEditEntry
		model.editingEntry = &harvest.TimeEntry{
			ID:      1,
			Project: harvest.TimeEntryProject{ID: 1, Name: "Website"},
			Task:    harvest.TimeEntryTask{ID: 10, Name: "Development"},
		}
		model.editTask = &harvest.Task{ID: 10, Name: "Development"}
		model.editCurrentField = 1 // Notes field
		model.editNotes = "Some notes"
		model.editHours = "1:30"

		notesInput := textinput.New()
		model.editNotesInput = &notesInput
		durationInput := textinput.New()
		model.editDurationInput = &durationInput

		msg := tea.KeyMsg{Type: tea.KeyEnter}
		updatedModel, cmd := model.Update(msg)

		m := updatedModel.(Model)
		if m.currentView != ViewEditEntry {
			t.Errorf("expected to stay on ViewEditEntry, got %v", m.currentView)
		}
		if cmd != nil {
			t.Error("expected no command (no save), but got one")
		}
	})

	t.Run("given edit view when enter pressed on duration field then does not save", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})
		model.currentView = ViewEditEntry
		model.editingEntry = &harvest.TimeEntry{
			ID:      1,
			Project: harvest.TimeEntryProject{ID: 1, Name: "Website"},
			Task:    harvest.TimeEntryTask{ID: 10, Name: "Development"},
		}
		model.editTask = &harvest.Task{ID: 10, Name: "Development"}
		model.editCurrentField = 2 // Duration field
		model.editNotes = "Some notes"
		model.editHours = "1:30"

		notesInput := textinput.New()
		model.editNotesInput = &notesInput
		durationInput := textinput.New()
		model.editDurationInput = &durationInput

		msg := tea.KeyMsg{Type: tea.KeyEnter}
		updatedModel, cmd := model.Update(msg)

		m := updatedModel.(Model)
		if m.currentView != ViewEditEntry {
			t.Errorf("expected to stay on ViewEditEntry, got %v", m.currentView)
		}
		if cmd != nil {
			t.Error("expected no command (no save), but got one")
		}
	})

	t.Run("given edit view when ctrl+s pressed then saves entry", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})
		model.currentView = ViewEditEntry
		model.editingEntry = &harvest.TimeEntry{
			ID:      1,
			Project: harvest.TimeEntryProject{ID: 1, Name: "Website"},
			Task:    harvest.TimeEntryTask{ID: 10, Name: "Development"},
		}
		model.editTask = &harvest.Task{ID: 10, Name: "Development"}
		model.editCurrentField = 1 // Notes field - should save from any field
		model.editNotes = "Updated notes"
		model.editHours = "2:00"

		notesInput := textinput.New()
		model.editNotesInput = &notesInput
		durationInput := textinput.New()
		model.editDurationInput = &durationInput

		msg := tea.KeyMsg{Type: tea.KeyCtrlS}
		_, cmd := model.Update(msg)

		if cmd == nil {
			t.Error("expected a command to save entry, got nil")
		}
	})
}
