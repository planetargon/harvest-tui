package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/planetargon/harvest-tui/internal/config"
	"github.com/planetargon/harvest-tui/internal/harvest"
	"github.com/planetargon/harvest-tui/internal/state"
)

func newLoadingModel() Model {
	cfg := &config.Config{Harvest: config.HarvestConfig{AccountID: "12345", AccessToken: "test-token"}}
	client := harvest.NewClient("12345", "test-token")
	appState := &state.State{Recents: []state.RecentEntry{}}
	return NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})
}

func TestLoadingView(t *testing.T) {
	t.Run("given new model when created then starts in ViewLoading", func(t *testing.T) {
		model := newLoadingModel()

		if model.currentView != ViewLoading {
			t.Errorf("expected currentView to be ViewLoading, got %v", model.currentView)
		}
	})

	t.Run("given loading view when only time entries arrive then stays in ViewLoading", func(t *testing.T) {
		model := newLoadingModel()

		msg := timeEntriesFetchedMsg{
			entries: []harvest.TimeEntry{{ID: 1, Notes: "Test"}},
		}
		newModel, _ := model.Update(msg)
		m := newModel.(Model)

		if m.currentView != ViewLoading {
			t.Errorf("expected currentView to remain ViewLoading, got %v", m.currentView)
		}
		if !m.timeEntriesLoaded {
			t.Error("expected timeEntriesLoaded to be true")
		}
		if m.projectsLoaded {
			t.Error("expected projectsLoaded to be false")
		}
	})

	t.Run("given loading view when only projects arrive then stays in ViewLoading", func(t *testing.T) {
		model := newLoadingModel()

		msg := projectsWithTasksFetchedMsg{
			projectsWithTasks: []harvest.ProjectWithTasks{
				{
					Project: harvest.Project{ID: 1, Name: "Test Project"},
					Tasks:   []harvest.Task{{ID: 1, Name: "Test Task"}},
				},
			},
		}
		newModel, _ := model.Update(msg)
		m := newModel.(Model)

		if m.currentView != ViewLoading {
			t.Errorf("expected currentView to remain ViewLoading, got %v", m.currentView)
		}
		if m.timeEntriesLoaded {
			t.Error("expected timeEntriesLoaded to be false")
		}
		if !m.projectsLoaded {
			t.Error("expected projectsLoaded to be true")
		}
	})

	t.Run("given loading view when both fetches complete then transitions to ViewList", func(t *testing.T) {
		model := newLoadingModel()

		// First: time entries arrive
		msg1 := timeEntriesFetchedMsg{
			entries: []harvest.TimeEntry{{ID: 1, Notes: "Test"}},
		}
		newModel, _ := model.Update(msg1)
		m := newModel.(Model)

		// Second: projects arrive
		msg2 := projectsWithTasksFetchedMsg{
			projectsWithTasks: []harvest.ProjectWithTasks{
				{
					Project: harvest.Project{ID: 1, Name: "Test Project"},
					Tasks:   []harvest.Task{{ID: 1, Name: "Test Task"}},
				},
			},
		}
		newModel, _ = m.Update(msg2)
		m = newModel.(Model)

		if m.currentView != ViewList {
			t.Errorf("expected currentView to be ViewList, got %v", m.currentView)
		}
	})

	t.Run("given loading view when both fetches complete in reverse order then transitions to ViewList", func(t *testing.T) {
		model := newLoadingModel()

		// First: projects arrive
		msg1 := projectsWithTasksFetchedMsg{
			projectsWithTasks: []harvest.ProjectWithTasks{},
		}
		newModel, _ := model.Update(msg1)
		m := newModel.(Model)

		// Second: time entries arrive
		msg2 := timeEntriesFetchedMsg{
			entries: []harvest.TimeEntry{},
		}
		newModel, _ = m.Update(msg2)
		m = newModel.(Model)

		if m.currentView != ViewList {
			t.Errorf("expected currentView to be ViewList, got %v", m.currentView)
		}
	})

	t.Run("given loading view when fetch errors occur then still transitions to ViewList", func(t *testing.T) {
		model := newLoadingModel()

		// Both fetches return errors
		msg1 := timeEntriesFetchedMsg{
			err: fmt.Errorf("network error"),
		}
		newModel, _ := model.Update(msg1)
		m := newModel.(Model)

		msg2 := projectsWithTasksFetchedMsg{
			err: fmt.Errorf("network error"),
		}
		newModel, _ = m.Update(msg2)
		m = newModel.(Model)

		if m.currentView != ViewList {
			t.Errorf("expected currentView to be ViewList even with errors, got %v", m.currentView)
		}
	})

	t.Run("given loading view when key pressed then does nothing", func(t *testing.T) {
		model := newLoadingModel()

		// Press 'n' key during loading
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
		newModel, _ := model.Update(msg)
		m := newModel.(Model)

		if m.currentView != ViewLoading {
			t.Errorf("expected currentView to remain ViewLoading, got %v", m.currentView)
		}
	})

	t.Run("given loading view when ctrl+c pressed then quits", func(t *testing.T) {
		model := newLoadingModel()

		msg := tea.KeyMsg{Type: tea.KeyCtrlC}
		_, cmd := model.Update(msg)

		if cmd == nil {
			t.Error("expected quit command to be returned")
		}
	})

	t.Run("given loading view when rendered then contains harvest-themed message", func(t *testing.T) {
		model := newLoadingModel()

		output := model.View()

		if !strings.Contains(output, "Harvesting your data") {
			t.Error("expected output to contain 'Harvesting your data'")
		}

		if !strings.Contains(output, "Harvest Time Tracker") {
			t.Error("expected output to contain title 'Harvest Time Tracker'")
		}

		if !strings.Contains(output, "ctrl+c") {
			t.Error("expected output to contain ctrl+c keybinding")
		}
	})
}
