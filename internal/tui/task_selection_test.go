package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/planetargon/harvest-tui/internal/config"
	"github.com/planetargon/harvest-tui/internal/harvest"
	"github.com/planetargon/harvest-tui/internal/state"
)

func TestTaskSelectionTransition(t *testing.T) {
	cfg := &config.Config{Harvest: config.HarvestConfig{AccountID: "123456", AccessToken: "test-token"}}
	client := &harvest.Client{}
	appState := &state.State{}

	t.Run("given project with multiple tasks when project selected then transitions to task selection view", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})

		// Set up project with multiple tasks
		model.projectsWithTasks = []harvest.ProjectWithTasks{
			{
				Project: harvest.Project{
					ID:   1,
					Name: "Website Redesign",
					Client: harvest.ProjectClient{
						ID:   100,
						Name: "Acme Corp",
					},
				},
				Tasks: []harvest.Task{
					{ID: 1, Name: "Development"},
					{ID: 2, Name: "Design"},
					{ID: 3, Name: "Testing"},
				},
			},
		}

		// Start in project selection view
		model.currentView = ViewSelectProject
		model.updateProjectList()

		// Simulate selecting the project
		model.projectList.Select(0)

		// Simulate pressing enter
		msg := tea.KeyMsg{Type: tea.KeyEnter}
		updatedModel, _ := model.handleProjectSelectKeys(msg)

		// Should transition to task selection view
		if updatedModel.(Model).currentView != ViewSelectTask {
			t.Errorf("expected view to be ViewSelectTask, got %v", updatedModel.(Model).currentView)
		}

		// Should have selected the project
		if updatedModel.(Model).selectedProject == nil {
			t.Error("expected selectedProject to be set")
		} else if updatedModel.(Model).selectedProject.ID != 1 {
			t.Errorf("expected selectedProject.ID to be 1, got %d", updatedModel.(Model).selectedProject.ID)
		}
	})

	t.Run("given project with single task when project selected then skips task selection", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})

		// Set up project with single task
		model.projectsWithTasks = []harvest.ProjectWithTasks{
			{
				Project: harvest.Project{
					ID:   2,
					Name: "Mobile App",
					Client: harvest.ProjectClient{
						ID:   200,
						Name: "BigCorp Inc",
					},
				},
				Tasks: []harvest.Task{
					{ID: 10, Name: "Development"},
				},
			},
		}

		// Start in project selection view
		model.currentView = ViewSelectProject
		model.updateProjectList()

		// Simulate selecting the project
		model.projectList.Select(0)

		// Simulate pressing enter
		msg := tea.KeyMsg{Type: tea.KeyEnter}
		updatedModel, _ := model.handleProjectSelectKeys(msg)

		// Should go to notes input (with single task auto-selected)
		if updatedModel.(Model).currentView != ViewNotesInput {
			t.Errorf("expected view to be ViewNotesInput, got %v", updatedModel.(Model).currentView)
		}

		// Should have selected both project and task
		if updatedModel.(Model).selectedProject == nil {
			t.Error("expected selectedProject to be set")
		} else if updatedModel.(Model).selectedProject.ID != 2 {
			t.Errorf("expected selectedProject.ID to be 2, got %d", updatedModel.(Model).selectedProject.ID)
		}

		if updatedModel.(Model).selectedTask == nil {
			t.Error("expected selectedTask to be set")
		} else if updatedModel.(Model).selectedTask.ID != 10 {
			t.Errorf("expected selectedTask.ID to be 10, got %d", updatedModel.(Model).selectedTask.ID)
		}
	})

	t.Run("given divider item when enter pressed then moves to next item", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})

		// Set up state with recents to ensure divider is present
		model.appState.Recents = []state.RecentEntry{
			{ClientID: 100, ProjectID: 1, TaskID: 1},
		}

		model.projectsWithTasks = []harvest.ProjectWithTasks{
			{
				Project: harvest.Project{
					ID:   1,
					Name: "Website Redesign",
					Client: harvest.ProjectClient{
						ID:   100,
						Name: "Acme Corp",
					},
				},
				Tasks: []harvest.Task{{ID: 1, Name: "Development"}},
			},
			{
				Project: harvest.Project{
					ID:   2,
					Name: "Mobile App",
					Client: harvest.ProjectClient{
						ID:   200,
						Name: "BigCorp Inc",
					},
				},
				Tasks: []harvest.Task{{ID: 2, Name: "Testing"}},
			},
		}

		model.currentView = ViewSelectProject
		model.updateProjectList()

		// Select the divider item (item 1)
		model.projectList.Select(1)

		// Simulate pressing enter on divider
		msg := tea.KeyMsg{Type: tea.KeyEnter}
		updatedModel, _ := model.handleProjectSelectKeys(msg)

		// Should remain in project selection view
		if updatedModel.(Model).currentView != ViewSelectProject {
			t.Errorf("expected to remain in ViewSelectProject, got %v", updatedModel.(Model).currentView)
		}

		// Should not have selected any project
		if updatedModel.(Model).selectedProject != nil {
			t.Error("expected selectedProject to remain nil when divider selected")
		}

		// Cursor should move down (handled by CursorDown in implementation)
	})

	t.Run("given project with no tasks when project selected then shows error", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})

		// Set up project with no tasks (edge case)
		model.projectsWithTasks = []harvest.ProjectWithTasks{
			{
				Project: harvest.Project{
					ID:   3,
					Name: "Empty Project",
					Client: harvest.ProjectClient{
						ID:   300,
						Name: "Empty Corp",
					},
				},
				Tasks: []harvest.Task{}, // No tasks
			},
		}

		// Start in project selection view
		model.currentView = ViewSelectProject
		model.updateProjectList()

		// Simulate selecting the project
		model.projectList.Select(0)

		// Simulate pressing enter
		msg := tea.KeyMsg{Type: tea.KeyEnter}
		updatedModel, _ := model.handleProjectSelectKeys(msg)

		// Should remain in project selection view (no tasks to select)
		if updatedModel.(Model).currentView != ViewSelectProject {
			t.Errorf("expected to remain in ViewSelectProject when project has no tasks, got %v", updatedModel.(Model).currentView)
		}

		// Should not have selected the project (since we can't proceed)
		if updatedModel.(Model).selectedProject != nil {
			t.Error("expected selectedProject to remain nil when project has no tasks")
		}

		// Should show error message
		if updatedModel.(Model).statusMessage != "No tasks available for this project" {
			t.Errorf("expected status message about no tasks, got '%s'", updatedModel.(Model).statusMessage)
		}
	})

	t.Run("given recent combo selected when enter pressed then skips task selection", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})

		// Set up state with a recent that has all IDs (client, project, task)
		model.appState.Recents = []state.RecentEntry{
			{ClientID: 100, ProjectID: 1, TaskID: 5},
		}

		// Set up matching project with tasks
		model.projectsWithTasks = []harvest.ProjectWithTasks{
			{
				Project: harvest.Project{
					ID:   1,
					Name: "Website Redesign",
					Client: harvest.ProjectClient{
						ID:   100,
						Name: "Acme Corp",
					},
				},
				Tasks: []harvest.Task{
					{ID: 5, Name: "Development"},
					{ID: 6, Name: "Testing"},
					{ID: 7, Name: "Deployment"},
				},
			},
		}

		// Start in project selection view
		model.currentView = ViewSelectProject
		model.updateProjectList()

		// The first item should be the recent
		model.projectList.Select(0)

		// Simulate pressing enter on the recent
		msg := tea.KeyMsg{Type: tea.KeyEnter}
		updatedModel, _ := model.handleProjectSelectKeys(msg)

		// Should skip task selection and go to notes input
		if updatedModel.(Model).currentView != ViewNotesInput {
			t.Errorf("expected view to be ViewNotesInput when recent selected, got %v", updatedModel.(Model).currentView)
		}

		// Should have selected both project and task from the recent
		if updatedModel.(Model).selectedProject == nil {
			t.Error("expected selectedProject to be set from recent")
		} else if updatedModel.(Model).selectedProject.ID != 1 {
			t.Errorf("expected selectedProject.ID to be 1, got %d", updatedModel.(Model).selectedProject.ID)
		}

		if updatedModel.(Model).selectedTask == nil {
			t.Error("expected selectedTask to be set from recent")
		} else if updatedModel.(Model).selectedTask.ID != 5 {
			t.Errorf("expected selectedTask.ID to be 5 (from recent), got %d", updatedModel.(Model).selectedTask.ID)
		}
	})

	t.Run("given recent with non-existent task when selected then shows task selection", func(t *testing.T) {
		model := NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})

		// Set up state with a recent that has a task ID that no longer exists
		model.appState.Recents = []state.RecentEntry{
			{ClientID: 100, ProjectID: 1, TaskID: 999}, // Task 999 doesn't exist
		}

		// Set up matching project with different tasks
		model.projectsWithTasks = []harvest.ProjectWithTasks{
			{
				Project: harvest.Project{
					ID:   1,
					Name: "Website Redesign",
					Client: harvest.ProjectClient{
						ID:   100,
						Name: "Acme Corp",
					},
				},
				Tasks: []harvest.Task{
					{ID: 5, Name: "Development"},
					{ID: 6, Name: "Testing"},
				},
			},
		}

		// Start in project selection view
		model.currentView = ViewSelectProject
		model.updateProjectList()

		// The first item should be the recent
		model.projectList.Select(0)

		// Simulate pressing enter on the recent
		msg := tea.KeyMsg{Type: tea.KeyEnter}
		updatedModel, _ := model.handleProjectSelectKeys(msg)

		// Should transition to task selection since the task from recent doesn't exist
		if updatedModel.(Model).currentView != ViewSelectTask {
			t.Errorf("expected view to be ViewSelectTask when recent task doesn't exist, got %v", updatedModel.(Model).currentView)
		}

		// Should have selected the project but not the task
		if updatedModel.(Model).selectedProject == nil {
			t.Error("expected selectedProject to be set")
		} else if updatedModel.(Model).selectedProject.ID != 1 {
			t.Errorf("expected selectedProject.ID to be 1, got %d", updatedModel.(Model).selectedProject.ID)
		}

		if updatedModel.(Model).selectedTask != nil {
			t.Error("expected selectedTask to remain nil when recent task doesn't exist")
		}
	})
}
