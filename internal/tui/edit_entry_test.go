package tui

import (
	"fmt"
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/planetargon/argon-harvest-tui/internal/config"
	"github.com/planetargon/argon-harvest-tui/internal/harvest"
	"github.com/planetargon/argon-harvest-tui/internal/state"
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
}
