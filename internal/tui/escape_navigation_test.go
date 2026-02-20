package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/planetargon/harvest-tui/internal/config"
	"github.com/planetargon/harvest-tui/internal/harvest"
	"github.com/planetargon/harvest-tui/internal/state"
)

func TestEscapeReturnsToMainList(t *testing.T) {
	cfg := &config.Config{Harvest: config.HarvestConfig{AccountID: "123456", AccessToken: "test-token"}}
	client := &harvest.Client{}
	appState := &state.State{}

	t.Run("given project selection view when escape pressed then returns to main list", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})

		// Start in project selection view (not from new entry form)
		model.currentView = ViewSelectProject
		model.newEntryCurrentField = -1 // Not from new entry form
		model.projectsWithTasks = []harvest.ProjectWithTasks{
			{
				Project: harvest.Project{
					ID:   1,
					Name: "Test Project",
					Client: harvest.ProjectClient{
						ID:   100,
						Name: "Test Client",
					},
				},
				Tasks: []harvest.Task{{ID: 1, Name: "Development"}},
			},
		}
		model.updateProjectList()

		// Simulate pressing escape
		msg := tea.KeyMsg{Type: tea.KeyEscape}
		updatedModel, cmd := model.Update(msg)

		// Should return to main list view
		if updatedModel.(Model).currentView != ViewList {
			t.Errorf("expected view to be ViewList, got %v", updatedModel.(Model).currentView)
		}

		// Should clear selected project
		if updatedModel.(Model).selectedProject != nil {
			t.Error("expected selectedProject to be cleared on escape")
		}

		// Should not quit
		if cmd != nil {
			if _, isQuit := cmd().(tea.QuitMsg); isQuit {
				t.Error("expected not to quit on escape from project selection")
			}
		}
	})

	t.Run("given task selection view when escape pressed then returns to project selection", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})

		// Set up project with tasks
		selectedProject := &harvest.Project{
			ID:   1,
			Name: "Test Project",
			Client: harvest.ProjectClient{
				ID:   100,
				Name: "Test Client",
			},
		}
		model.selectedProject = selectedProject
		model.projectsWithTasks = []harvest.ProjectWithTasks{
			{
				Project: *selectedProject,
				Tasks: []harvest.Task{
					{ID: 1, Name: "Development"},
					{ID: 2, Name: "Testing"},
				},
			},
		}

		// Start in task selection view
		model.currentView = ViewSelectTask
		model.updateTaskList(model.projectsWithTasks[0].Tasks)

		// Simulate pressing escape
		msg := tea.KeyMsg{Type: tea.KeyEscape}
		updatedModel, _ := model.Update(msg)

		// Should return to project selection view
		if updatedModel.(Model).currentView != ViewSelectProject {
			t.Errorf("expected view to be ViewSelectProject, got %v", updatedModel.(Model).currentView)
		}

		// Should clear selected project
		if updatedModel.(Model).selectedProject != nil {
			t.Error("expected selectedProject to be cleared on escape from task selection")
		}

		// Should clear selected task
		if updatedModel.(Model).selectedTask != nil {
			t.Error("expected selectedTask to be cleared on escape")
		}
	})

	// Note: Notes input view will be tested when implemented in Phase 16

	t.Run("given main list view when q pressed then quits application", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})

		// Start in main list view
		model.currentView = ViewList

		// Simulate pressing q
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}
		_, cmd := model.Update(msg)

		// Should quit the application
		if cmd == nil {
			t.Error("expected quit command, got nil")
		}
	})

	t.Run("given help view when escape pressed then returns to main list", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})

		// Start in help view
		model.currentView = ViewHelp

		// Simulate pressing escape
		msg := tea.KeyMsg{Type: tea.KeyEscape}
		updatedModel, cmd := model.Update(msg)

		// Should return to main list view
		if updatedModel.(Model).currentView != ViewList {
			t.Errorf("expected view to be ViewList, got %v", updatedModel.(Model).currentView)
		}

		// Should not quit
		if cmd != nil {
			if _, isQuit := cmd().(tea.QuitMsg); isQuit {
				t.Error("expected not to quit on escape from help view")
			}
		}
	})

	t.Run("given edit entry view when escape pressed then returns to main list", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})

		// Start in edit entry view
		model.currentView = ViewEditEntry

		// Simulate pressing escape
		msg := tea.KeyMsg{Type: tea.KeyEscape}
		updatedModel, cmd := model.Update(msg)

		// Should return to main list view
		if updatedModel.(Model).currentView != ViewList {
			t.Errorf("expected view to be ViewList, got %v", updatedModel.(Model).currentView)
		}

		// Should not quit
		if cmd != nil {
			if _, isQuit := cmd().(tea.QuitMsg); isQuit {
				t.Error("expected not to quit on escape from edit view")
			}
		}
	})

	t.Run("given confirm delete view when escape pressed then returns to main list", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})

		// Start in confirm delete view
		model.currentView = ViewConfirmDelete

		// Simulate pressing escape
		msg := tea.KeyMsg{Type: tea.KeyEscape}
		updatedModel, cmd := model.Update(msg)

		// Should return to main list view
		if updatedModel.(Model).currentView != ViewList {
			t.Errorf("expected view to be ViewList, got %v", updatedModel.(Model).currentView)
		}

		// Should not quit
		if cmd != nil {
			if _, isQuit := cmd().(tea.QuitMsg); isQuit {
				t.Error("expected not to quit on escape from delete confirmation")
			}
		}
	})

	t.Run("given multiple escape presses then returns to main list then q quits", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})

		// Start in task selection view (not from new entry form)
		model.currentView = ViewSelectTask
		model.newEntryCurrentField = -1 // Not from new entry form
		model.selectedProject = &harvest.Project{
			ID:   1,
			Name: "Test Project",
		}

		// First escape - should go to project selection
		msg := tea.KeyMsg{Type: tea.KeyEscape}
		updatedModel, _ := model.Update(msg)

		if updatedModel.(Model).currentView != ViewSelectProject {
			t.Errorf("expected first escape to go to ViewSelectProject, got %v", updatedModel.(Model).currentView)
		}

		// Second escape - should go to main list
		updatedModel, cmd := updatedModel.Update(msg)

		if updatedModel.(Model).currentView != ViewList {
			t.Errorf("expected second escape to go to ViewList, got %v", updatedModel.(Model).currentView)
		}
		if cmd != nil {
			if _, isQuit := cmd().(tea.QuitMsg); isQuit {
				t.Error("expected not to quit on second escape")
			}
		}

		// Third press - try 'q' to quit
		quitMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}
		_, cmd = updatedModel.Update(quitMsg)

		if cmd == nil {
			t.Error("expected quit command on 'q' press, got nil")
		}
	})
}
