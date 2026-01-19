package tui

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/planetargon/argon-harvest-tui/internal/config"
	"github.com/planetargon/argon-harvest-tui/internal/harvest"
	"github.com/planetargon/argon-harvest-tui/internal/state"
)

func TestTimerControl(t *testing.T) {
	cfg := &config.Config{Harvest: config.HarvestConfig{AccountID: "123456", AccessToken: "test-token"}}
	client := &harvest.Client{}
	appState := &state.State{}

	t.Run("given stopped entry when start key pressed then starts timer", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})
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

		// Press 's' to start timer
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")}
		_, cmd := model.Update(msg)

		// Should return a command to start the timer
		if cmd == nil {
			t.Error("expected command to start timer, got nil")
		}
	})

	t.Run("given running entry when stop key pressed then stops timer", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})
		model.currentView = ViewList
		model.timeEntries = []harvest.TimeEntry{
			{
				ID:        1,
				Hours:     1.5,
				Notes:     "Test entry",
				IsRunning: true,
				IsLocked:  false,
			},
		}
		model.selectedEntryIndex = 0

		// Press 's' to stop timer (toggle)
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")}
		_, cmd := model.Update(msg)

		// Should return a command to stop the timer
		if cmd == nil {
			t.Error("expected command to stop timer, got nil")
		}
	})

	t.Run("given locked entry when start key pressed then shows error", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})
		model.currentView = ViewList
		model.timeEntries = []harvest.TimeEntry{
			{
				ID:        1,
				Hours:     1.5,
				Notes:     "Test entry",
				IsRunning: false,
				IsLocked:  true,
			},
		}
		model.selectedEntryIndex = 0

		// Press 's' to start timer
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")}
		updatedModel, cmd := model.Update(msg)

		// Should not return a command
		if cmd != nil {
			t.Error("expected no command for locked entry, got command")
		}

		// Should show error message
		if updatedModel.(Model).statusMessage != "Cannot start locked time entry." {
			t.Errorf("expected locked entry message, got '%s'", updatedModel.(Model).statusMessage)
		}
	})

	t.Run("given locked entry when stop key pressed then shows error", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})
		model.currentView = ViewList
		model.timeEntries = []harvest.TimeEntry{
			{
				ID:        1,
				Hours:     1.5,
				Notes:     "Test entry",
				IsRunning: true,
				IsLocked:  true,
			},
		}
		model.selectedEntryIndex = 0

		// Press 's' to stop timer (toggle)
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")}
		updatedModel, cmd := model.Update(msg)

		// Should not return a command
		if cmd != nil {
			t.Error("expected no command for locked entry, got command")
		}

		// Should show error message
		if updatedModel.(Model).statusMessage != "Cannot stop locked time entry." {
			t.Errorf("expected locked entry message, got '%s'", updatedModel.(Model).statusMessage)
		}
	})

	t.Run("given already running entry when start key pressed then stops timer", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})
		model.currentView = ViewList
		model.timeEntries = []harvest.TimeEntry{
			{
				ID:        1,
				Hours:     1.5,
				Notes:     "Test entry",
				IsRunning: true,
				IsLocked:  false,
			},
		}
		model.selectedEntryIndex = 0

		// Press 's' to stop timer (toggle)
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")}
		_, cmd := model.Update(msg)

		// Should return a command to stop the timer
		if cmd == nil {
			t.Error("expected command to stop timer, got nil")
		}
	})

	t.Run("given timer start succeeds when message received then updates entry", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})
		model.timeEntries = []harvest.TimeEntry{
			{
				ID:        1,
				Hours:     1.5,
				Notes:     "Test entry",
				IsRunning: false,
			},
		}

		// Simulate successful timer start
		updatedEntry := &harvest.TimeEntry{
			ID:        1,
			Hours:     1.5,
			Notes:     "Test entry",
			IsRunning: true,
		}

		msg := timeEntryStartedMsg{entry: updatedEntry}
		updatedModel, _ := model.Update(msg)

		// Should update entry in list
		if !updatedModel.(Model).timeEntries[0].IsRunning {
			t.Error("expected entry to be running after timer start")
		}

		// Should show success message
		if updatedModel.(Model).statusMessage != "Timer started successfully" {
			t.Errorf("expected success message, got '%s'", updatedModel.(Model).statusMessage)
		}
	})

	t.Run("given timer stop succeeds when message received then updates entry", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})
		model.timeEntries = []harvest.TimeEntry{
			{
				ID:        1,
				Hours:     1.5,
				Notes:     "Test entry",
				IsRunning: true,
			},
		}

		// Simulate successful timer stop
		updatedEntry := &harvest.TimeEntry{
			ID:        1,
			Hours:     2.0, // Timer added 0.5 hours
			Notes:     "Test entry",
			IsRunning: false,
		}

		msg := timeEntryStoppedMsg{entry: updatedEntry}
		updatedModel, _ := model.Update(msg)

		// Should update entry in list
		if updatedModel.(Model).timeEntries[0].IsRunning {
			t.Error("expected entry to be stopped after timer stop")
		}

		// Should update hours
		if updatedModel.(Model).timeEntries[0].Hours != 2.0 {
			t.Errorf("expected hours to be 2.0, got %f", updatedModel.(Model).timeEntries[0].Hours)
		}

		// Should show success message
		if updatedModel.(Model).statusMessage != "Timer stopped successfully" {
			t.Errorf("expected success message, got '%s'", updatedModel.(Model).statusMessage)
		}
	})

	t.Run("given timer start fails when message received then shows error", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})

		// Simulate failed timer start
		msg := timeEntryStartedMsg{err: fmt.Errorf("API error")}
		updatedModel, _ := model.Update(msg)

		// Should show error message
		if updatedModel.(Model).statusMessage == "" {
			t.Error("expected error message to be set")
		}
	})

	t.Run("given timer stop fails when message received then shows error", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})

		// Simulate failed timer stop
		msg := timeEntryStoppedMsg{err: fmt.Errorf("API error")}
		updatedModel, _ := model.Update(msg)

		// Should show error message
		if updatedModel.(Model).statusMessage == "" {
			t.Error("expected error message to be set")
		}
	})
}
