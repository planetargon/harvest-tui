package tui

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/planetargon/argon-harvest-tui/internal/config"
	"github.com/planetargon/argon-harvest-tui/internal/harvest"
	"github.com/planetargon/argon-harvest-tui/internal/state"
)

func TestDeleteEntry(t *testing.T) {
	cfg := &config.Config{Harvest: config.HarvestConfig{AccountID: "123456", AccessToken: "test-token"}}
	client := &harvest.Client{}
	appState := &state.State{}

	t.Run("given entry selected when delete key pressed then transitions to confirm delete view", func(t *testing.T) {
		model := NewModel(cfg, client, appState)
		model.currentView = ViewList
		model.timeEntries = []harvest.TimeEntry{
			{
				ID:        1,
				Hours:     1.5,
				Notes:     "Test entry",
				IsRunning: false,
				IsLocked:  false,
			},
		}
		model.selectedEntryIndex = 0

		// Press 'd' to delete
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")}
		updatedModel, _ := model.Update(msg)

		// Should transition to confirm delete view
		if updatedModel.(Model).currentView != ViewConfirmDelete {
			t.Errorf("expected view to be ViewConfirmDelete, got %v", updatedModel.(Model).currentView)
		}

		// Should set editingEntry (used for deletion)
		if updatedModel.(Model).editingEntry == nil {
			t.Error("expected editingEntry to be set for deletion")
		} else {
			if updatedModel.(Model).editingEntry.ID != 1 {
				t.Errorf("expected editingEntry.ID to be 1, got %d", updatedModel.(Model).editingEntry.ID)
			}
		}
	})

	t.Run("given locked entry when delete key pressed then shows error", func(t *testing.T) {
		model := NewModel(cfg, client, appState)
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

		// Press 'd' to delete
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")}
		updatedModel, _ := model.Update(msg)

		// Should not transition to confirm delete view
		if updatedModel.(Model).currentView != ViewList {
			t.Errorf("expected to remain in ViewList, got %v", updatedModel.(Model).currentView)
		}

		// Should show error message
		if updatedModel.(Model).statusMessage != "Cannot delete locked time entry." {
			t.Errorf("expected locked entry message, got '%s'", updatedModel.(Model).statusMessage)
		}
	})

	t.Run("given running entry when delete key pressed then shows error", func(t *testing.T) {
		model := NewModel(cfg, client, appState)
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

		// Press 'd' to delete
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")}
		updatedModel, _ := model.Update(msg)

		// Should not transition to confirm delete view
		if updatedModel.(Model).currentView != ViewList {
			t.Errorf("expected to remain in ViewList, got %v", updatedModel.(Model).currentView)
		}

		// Should show error message
		if updatedModel.(Model).statusMessage != "Cannot delete running time entry. Stop the timer first." {
			t.Errorf("expected running entry message, got '%s'", updatedModel.(Model).statusMessage)
		}
	})

	t.Run("given confirm delete view when 'y' pressed then deletes entry", func(t *testing.T) {
		model := NewModel(cfg, client, appState)
		model.currentView = ViewConfirmDelete
		model.editingEntry = &harvest.TimeEntry{
			ID:    1,
			Hours: 1.5,
			Notes: "Test entry",
		}

		// Press 'y' to confirm
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")}
		_, cmd := model.Update(msg)

		// Should return a command to delete the entry
		if cmd == nil {
			t.Error("expected command to delete entry, got nil")
		}
	})

	t.Run("given confirm delete view when 'n' pressed then cancels deletion", func(t *testing.T) {
		model := NewModel(cfg, client, appState)
		model.currentView = ViewConfirmDelete
		model.editingEntry = &harvest.TimeEntry{
			ID:    1,
			Hours: 1.5,
			Notes: "Test entry",
		}

		// Press 'n' to cancel
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")}
		updatedModel, cmd := model.Update(msg)

		// Should return to list view
		if updatedModel.(Model).currentView != ViewList {
			t.Errorf("expected view to be ViewList, got %v", updatedModel.(Model).currentView)
		}

		// Should not return a command
		if cmd != nil {
			t.Error("expected no command on cancel, got command")
		}

		// Should clear editingEntry
		if updatedModel.(Model).editingEntry != nil {
			t.Error("expected editingEntry to be cleared on cancel")
		}
	})

	t.Run("given confirm delete view when escape pressed then cancels deletion", func(t *testing.T) {
		model := NewModel(cfg, client, appState)
		model.currentView = ViewConfirmDelete
		model.editingEntry = &harvest.TimeEntry{
			ID:    1,
			Hours: 1.5,
			Notes: "Test entry",
		}

		// Press escape
		msg := tea.KeyMsg{Type: tea.KeyEscape}
		updatedModel, _ := model.Update(msg)

		// Should return to list view
		if updatedModel.(Model).currentView != ViewList {
			t.Errorf("expected view to be ViewList, got %v", updatedModel.(Model).currentView)
		}

		// Should clear editingEntry
		if updatedModel.(Model).editingEntry != nil {
			t.Error("expected editingEntry to be cleared on escape")
		}
	})

	t.Run("given entry deletion succeeds when message received then removes from list", func(t *testing.T) {
		model := NewModel(cfg, client, appState)
		model.timeEntries = []harvest.TimeEntry{
			{ID: 1, Hours: 1.5, Notes: "Entry 1"},
			{ID: 2, Hours: 2.0, Notes: "Entry 2"},
			{ID: 3, Hours: 0.5, Notes: "Entry 3"},
		}

		// Simulate successful deletion of entry 2
		msg := timeEntryDeletedMsg{entryID: 2}
		updatedModel, _ := model.Update(msg)

		// Should remove entry from list
		if len(updatedModel.(Model).timeEntries) != 2 {
			t.Errorf("expected 2 entries after deletion, got %d", len(updatedModel.(Model).timeEntries))
		}

		// Should not contain deleted entry
		for _, entry := range updatedModel.(Model).timeEntries {
			if entry.ID == 2 {
				t.Error("expected entry with ID 2 to be removed")
			}
		}

		// Should show success message
		if updatedModel.(Model).statusMessage != "Time entry deleted successfully" {
			t.Errorf("expected success message, got '%s'", updatedModel.(Model).statusMessage)
		}

		// Should return to list view
		if updatedModel.(Model).currentView != ViewList {
			t.Errorf("expected view to be ViewList, got %v", updatedModel.(Model).currentView)
		}
	})

	t.Run("given entry deletion fails when message received then shows error", func(t *testing.T) {
		model := NewModel(cfg, client, appState)

		// Simulate failed deletion
		msg := timeEntryDeletedMsg{err: fmt.Errorf("API error")}
		updatedModel, _ := model.Update(msg)

		// Should show error message
		if updatedModel.(Model).statusMessage == "" {
			t.Error("expected error message to be set")
		}
	})
}
