package tui

import (
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/planetargon/argon-harvest-tui/internal/config"
	"github.com/planetargon/argon-harvest-tui/internal/harvest"
	"github.com/planetargon/argon-harvest-tui/internal/state"
)

func TestNotesInputView(t *testing.T) {
	cfg := &config.Config{Harvest: config.HarvestConfig{AccountID: "123456", AccessToken: "test-token"}}
	client := &harvest.Client{}
	appState := &state.State{}

	t.Run("given task selected when transitioning to notes input then shows notes input view", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})

		// Set up selected project and task
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

		// Transition to notes input view
		notesInput := textinput.New()
		notesInput.Focus()
		model.notesInput = &notesInput
		model.currentView = ViewNotesInput

		// View should be notes input
		if model.currentView != ViewNotesInput {
			t.Errorf("expected view to be ViewNotesInput, got %v", model.currentView)
		}

		// Should have notes input field initialized
		if model.notesInput == nil {
			t.Error("expected notesInput to be initialized")
		}
	})

	t.Run("given notes input view when typing then updates input field", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})
		model.selectedProject = &harvest.Project{ID: 1, Name: "Test Project"}
		model.selectedTask = &harvest.Task{ID: 1, Name: "Test Task"}

		// Initialize notes input
		notesInput := textinput.New()
		model.notesInput = &notesInput
		model.notesInput.Focus()
		model.currentView = ViewNotesInput

		// Type some text
		msgs := []tea.Msg{
			tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Working on feature")},
		}

		for _, msg := range msgs {
			updatedModel, _ := model.Update(msg)
			model = updatedModel.(Model)
		}

		// Check that text was entered
		if model.notesInput.Value() != "Working on feature" {
			t.Errorf("expected notes to be 'Working on feature', got '%s'", model.notesInput.Value())
		}
	})

	t.Run("given notes input view when enter pressed then transitions to duration input", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})
		model.selectedProject = &harvest.Project{ID: 1, Name: "Test Project"}
		model.selectedTask = &harvest.Task{ID: 1, Name: "Test Task"}
		notesInput := textinput.New()
		model.notesInput = &notesInput
		model.notesInput.SetValue("Test notes")
		model.currentView = ViewNotesInput

		// Press enter
		msg := tea.KeyMsg{Type: tea.KeyEnter}
		updatedModel, _ := model.Update(msg)

		// Should transition to duration input
		if updatedModel.(Model).currentView != ViewDurationInput {
			t.Errorf("expected view to be ViewDurationInput, got %v", updatedModel.(Model).currentView)
		}

		// Notes should be stored
		if updatedModel.(Model).newEntryNotes != "Test notes" {
			t.Errorf("expected notes to be stored as 'Test notes', got '%s'", updatedModel.(Model).newEntryNotes)
		}
	})

	t.Run("given notes input view when escape pressed then returns to main list", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})
		model.selectedProject = &harvest.Project{ID: 1, Name: "Test Project"}
		model.selectedTask = &harvest.Task{ID: 1, Name: "Test Task"}
		notesInput := textinput.New()
		model.notesInput = &notesInput
		model.currentView = ViewNotesInput

		// Press escape
		msg := tea.KeyMsg{Type: tea.KeyEscape}
		updatedModel, _ := model.Update(msg)

		// Should return to main list
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
	})

	t.Run("given duration input view when valid duration entered then stores hours", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})
		model.selectedProject = &harvest.Project{ID: 1, Name: "Test Project"}
		model.selectedTask = &harvest.Task{ID: 1, Name: "Test Task"}
		durationInput := textinput.New()
		model.durationInput = &durationInput
		model.durationInput.SetValue("1:30")
		model.currentView = ViewDurationInput

		// Press enter
		msg := tea.KeyMsg{Type: tea.KeyEnter}
		updatedModel, _ := model.Update(msg)

		// Should transition to billable toggle
		if updatedModel.(Model).currentView != ViewBillableToggle {
			t.Errorf("expected view to be ViewBillableToggle, got %v", updatedModel.(Model).currentView)
		}

		// Duration should be stored
		if updatedModel.(Model).newEntryHours != "1:30" {
			t.Errorf("expected duration to be '1:30', got '%s'", updatedModel.(Model).newEntryHours)
		}
	})

	t.Run("given duration input view when invalid duration entered then shows error", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})
		model.selectedProject = &harvest.Project{ID: 1, Name: "Test Project"}
		model.selectedTask = &harvest.Task{ID: 1, Name: "Test Task"}
		durationInput := textinput.New()
		model.durationInput = &durationInput
		model.durationInput.SetValue("invalid")
		model.currentView = ViewDurationInput

		// Press enter
		msg := tea.KeyMsg{Type: tea.KeyEnter}
		updatedModel, _ := model.Update(msg)

		// Should remain in duration input
		if updatedModel.(Model).currentView != ViewDurationInput {
			t.Errorf("expected to remain in ViewDurationInput, got %v", updatedModel.(Model).currentView)
		}

		// Should show error message
		if updatedModel.(Model).statusMessage == "" {
			t.Error("expected error message for invalid duration")
		}
	})

	t.Run("given duration input view when escape pressed then returns to notes input", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})
		model.selectedProject = &harvest.Project{ID: 1, Name: "Test Project"}
		model.selectedTask = &harvest.Task{ID: 1, Name: "Test Task"}
		durationInput := textinput.New()
		model.durationInput = &durationInput
		model.currentView = ViewDurationInput

		// Press escape
		msg := tea.KeyMsg{Type: tea.KeyEscape}
		updatedModel, _ := model.Update(msg)

		// Should return to notes input
		if updatedModel.(Model).currentView != ViewNotesInput {
			t.Errorf("expected view to be ViewNotesInput, got %v", updatedModel.(Model).currentView)
		}
	})
}
