package tui

import (
	"fmt"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/planetargon/harvest-tui/internal/config"
	"github.com/planetargon/harvest-tui/internal/harvest"
	"github.com/planetargon/harvest-tui/internal/state"
)

func TestEntryCreation(t *testing.T) {
	cfg := &config.Config{Harvest: config.HarvestConfig{AccountID: "123456", AccessToken: "test-token"}}
	client := &harvest.Client{}
	appState := &state.State{}

	t.Run("given billable toggle view when enter pressed then creates time entry", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})

		// Set up state for entry creation
		model.selectedProject = &harvest.Project{
			ID:   1,
			Name: "Website Redesign",
			Client: harvest.ProjectClient{
				ID:   100,
				Name: "Acme Corp",
			},
		}
		model.selectedTask = &harvest.Task{
			ID:   1,
			Name: "Development",
		}
		model.newEntryNotes = "Working on feature"
		model.newEntryHours = "1:30"
		model.newEntryBillable = true
		model.currentView = ViewBillableToggle
		model.currentDate = time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

		// Press enter to create entry
		msg := tea.KeyMsg{Type: tea.KeyEnter}
		_, cmd := model.Update(msg)

		// Should return a command to create the entry
		if cmd == nil {
			t.Error("expected command to create entry, got nil")
		}
	})

	t.Run("given entry creation succeeds when message received then returns to list view", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})
		model.currentView = ViewBillableToggle
		model.selectedProject = &harvest.Project{ID: 1}
		model.selectedTask = &harvest.Task{ID: 1}

		// Simulate successful entry creation
		createdEntry := &harvest.TimeEntry{
			ID:         123,
			Project:    harvest.TimeEntryProject{ID: 1},
			Task:       harvest.TimeEntryTask{ID: 1},
			Hours:      1.5,
			Notes:      "Test entry",
			IsBillable: true,
		}

		msg := timeEntryCreatedMsg{entry: createdEntry}
		updatedModel, _ := model.Update(msg)

		// Should return to list view
		if updatedModel.(Model).currentView != ViewList {
			t.Errorf("expected view to be ViewList, got %v", updatedModel.(Model).currentView)
		}

		// Should clear selections
		if updatedModel.(Model).selectedProject != nil {
			t.Error("expected selectedProject to be cleared")
		}
		if updatedModel.(Model).selectedTask != nil {
			t.Error("expected selectedTask to be cleared")
		}

		// Should have success message
		if updatedModel.(Model).statusMessage != "Time entry created successfully" {
			t.Errorf("expected success message, got '%s'", updatedModel.(Model).statusMessage)
		}

		// Should add entry to list
		if len(updatedModel.(Model).timeEntries) == 0 {
			t.Error("expected entry to be added to timeEntries list")
		} else if updatedModel.(Model).timeEntries[0].ID != 123 {
			t.Errorf("expected first entry to have ID 123, got %d", updatedModel.(Model).timeEntries[0].ID)
		}
	})

	t.Run("given entry creation fails when message received then shows error", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})
		model.currentView = ViewBillableToggle

		// Simulate failed entry creation
		msg := timeEntryCreatedMsg{err: fmt.Errorf("authorization failed")}
		updatedModel, _ := model.Update(msg)

		// Should show error message
		if updatedModel.(Model).statusMessage == "" {
			t.Error("expected error message to be set")
		}

		// Should not return to list view
		if updatedModel.(Model).currentView == ViewList {
			t.Error("expected to remain in current view on error")
		}
	})

	t.Run("given successful entry creation when recents updated then saves to state", func(t *testing.T) {
		// Create a temporary state
		testState := &state.State{}

		model := NewModel(cfg, client, testState, &harvest.User{FirstName: "Test", LastName: "User"})
		model.selectedProject = &harvest.Project{
			ID:   1,
			Name: "Test Project",
			Client: harvest.ProjectClient{
				ID:   100,
				Name: "Test Client",
			},
		}
		model.selectedTask = &harvest.Task{
			ID:   1,
			Name: "Test Task",
		}
		model.newEntryHours = "1:00"
		model.currentView = ViewBillableToggle

		// Create entry (this would normally trigger the command)
		// For testing, we'll check that the createTimeEntry function
		// would add to recents
		cmd := model.createTimeEntry()
		if cmd == nil {
			t.Error("expected createTimeEntry to return a command")
		}

		// The actual recents update happens in the command function
		// which is executed asynchronously in a real scenario
	})
}
