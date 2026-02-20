package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/planetargon/harvest-tui/internal/config"
	"github.com/planetargon/harvest-tui/internal/harvest"
	"github.com/planetargon/harvest-tui/internal/state"
)

func newTestModel() Model {
	cfg := &config.Config{Harvest: config.HarvestConfig{AccountID: "12345", AccessToken: "test-token"}}
	client := harvest.NewClient("12345", "test-token")
	appState := &state.State{Recents: []state.RecentEntry{}}
	return NewModel(cfg, client, appState, &harvest.User{FirstName: "Test", LastName: "User"})
}

func TestBoxBorderAlignment(t *testing.T) {
	t.Run("given rendered list view when box borders checked then all lines have equal width", func(t *testing.T) {
		model := newTestModel()
		model.timeEntries = []harvest.TimeEntry{
			{
				ID:      1,
				Hours:   1.5,
				Notes:   "Test entry",
				Client:  harvest.TimeEntryClient{ID: 1, Name: "Acme"},
				Project: harvest.TimeEntryProject{ID: 1, Name: "Web"},
				Task:    harvest.TimeEntryTask{ID: 1, Name: "Dev"},
			},
		}
		model.currentDate = time.Date(2025, 1, 19, 0, 0, 0, 0, time.UTC)

		output := model.View()
		lines := strings.Split(output, "\n")

		if len(lines) < 3 {
			t.Fatal("expected at least 3 lines in output")
		}

		// Top border sets the expected width
		topBorderWidth := lipgloss.Width(lines[0])

		for i, line := range lines {
			lineWidth := lipgloss.Width(line)
			if lineWidth != topBorderWidth {
				t.Errorf("line %d has width %d, expected %d (top border width)\nline: %q", i, lineWidth, topBorderWidth, line)
			}
		}
	})

	t.Run("given empty entries when box rendered then borders still align", func(t *testing.T) {
		model := newTestModel()
		model.timeEntries = []harvest.TimeEntry{}
		model.currentDate = time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)

		output := model.View()
		lines := strings.Split(output, "\n")

		topBorderWidth := lipgloss.Width(lines[0])
		for i, line := range lines {
			lineWidth := lipgloss.Width(line)
			if lineWidth != topBorderWidth {
				t.Errorf("line %d has width %d, expected %d\nline: %q", i, lineWidth, topBorderWidth, line)
			}
		}
	})

	t.Run("given loading state when box rendered then borders align", func(t *testing.T) {
		model := newTestModel()
		model.loading = true

		output := model.View()
		lines := strings.Split(output, "\n")

		topBorderWidth := lipgloss.Width(lines[0])
		for i, line := range lines {
			lineWidth := lipgloss.Width(line)
			if lineWidth != topBorderWidth {
				t.Errorf("line %d has width %d, expected %d\nline: %q", i, lineWidth, topBorderWidth, line)
			}
		}
	})

	t.Run("given error state when box rendered then borders align", func(t *testing.T) {
		model := newTestModel()
		model.errorMessage = "Something went wrong"

		output := model.View()
		lines := strings.Split(output, "\n")

		topBorderWidth := lipgloss.Width(lines[0])
		for i, line := range lines {
			lineWidth := lipgloss.Width(line)
			if lineWidth != topBorderWidth {
				t.Errorf("line %d has width %d, expected %d\nline: %q", i, lineWidth, topBorderWidth, line)
			}
		}
	})
}

func TestDividerWidth(t *testing.T) {
	t.Run("given default width when divider rendered then matches box interior width", func(t *testing.T) {
		model := newTestModel()
		model.timeEntries = []harvest.TimeEntry{}
		model.currentDate = time.Date(2025, 1, 19, 0, 0, 0, 0, time.UTC)

		output := model.View()
		lines := strings.Split(output, "\n")

		// Find the divider line (contains ─ but not ┌┐┘└├┤)
		for _, line := range lines {
			stripped := strings.TrimSpace(line)
			if strings.Contains(stripped, "│") {
				inner := strings.TrimPrefix(stripped, "│")
				inner = strings.TrimSuffix(inner, "│")
				// Check if this is a divider-only line
				trimmedInner := strings.TrimSpace(inner)
				if len(trimmedInner) > 0 && !strings.ContainsAny(trimmedInner, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789") {
					// This line only has divider/special chars - verify it fits within the box
					// Just need to ensure it doesn't overflow
					innerWidth := lipgloss.Width(inner)
					if innerWidth > 80 {
						t.Errorf("divider inner width %d exceeds max box width", innerWidth)
					}
				}
			}
		}
	})

	t.Run("given RenderDividerWidth called with various widths then returns correct width", func(t *testing.T) {
		tests := []struct {
			width    int
			expected int
		}{
			{10, 10},
			{40, 40},
			{1, 1},
			{0, 1},  // Minimum 1
			{-5, 1}, // Minimum 1
		}

		for _, tc := range tests {
			result := RenderDividerWidth(tc.width)
			// The visual width should match (lipgloss.Width handles ANSI)
			if lipgloss.Width(result) != tc.expected {
				t.Errorf("RenderDividerWidth(%d) visual width = %d, expected %d", tc.width, lipgloss.Width(result), tc.expected)
			}
		}
	})
}

func TestWindowResizeHandling(t *testing.T) {
	t.Run("given model when window resize received then updates dimensions", func(t *testing.T) {
		model := newTestModel()

		msg := tea.WindowSizeMsg{Width: 120, Height: 40}
		updatedModel, cmd := model.Update(msg)
		m := updatedModel.(Model)

		if m.width != 120 {
			t.Errorf("expected width to be 120, got %d", m.width)
		}
		if m.height != 40 {
			t.Errorf("expected height to be 40, got %d", m.height)
		}
		if cmd != nil {
			t.Error("expected no command from window resize")
		}
	})

	t.Run("given small terminal when rendered then box width is capped appropriately", func(t *testing.T) {
		model := newTestModel()
		model.width = 50
		model.timeEntries = []harvest.TimeEntry{}
		model.currentDate = time.Date(2025, 1, 19, 0, 0, 0, 0, time.UTC)

		output := model.View()
		lines := strings.Split(output, "\n")

		topBorderWidth := lipgloss.Width(lines[0])
		// Should be capped at width-2 = 48
		if topBorderWidth != 48 {
			t.Errorf("expected box width to be 48 for terminal width 50, got %d", topBorderWidth)
		}
	})

	t.Run("given large terminal when rendered then box width is capped at 80", func(t *testing.T) {
		model := newTestModel()
		model.width = 200
		model.timeEntries = []harvest.TimeEntry{}
		model.currentDate = time.Date(2025, 1, 19, 0, 0, 0, 0, time.UTC)

		output := model.View()
		lines := strings.Split(output, "\n")

		topBorderWidth := lipgloss.Width(lines[0])
		if topBorderWidth != 80 {
			t.Errorf("expected box width to be capped at 80, got %d", topBorderWidth)
		}
	})
}

func TestFormatHoursSimple(t *testing.T) {
	tests := []struct {
		name     string
		hours    float64
		expected string
	}{
		{"given zero hours when formatted then returns 0:00", 0.0, "0:00"},
		{"given one hour when formatted then returns 1:00", 1.0, "1:00"},
		{"given half hour when formatted then returns 0:30", 0.5, "0:30"},
		{"given 1.5 hours when formatted then returns 1:30", 1.5, "1:30"},
		{"given quarter hour when formatted then returns 0:15", 0.25, "0:15"},
		{"given 8.75 hours when formatted then returns 8:45", 8.75, "8:45"},
		{"given 10+ hours when formatted then returns correct format", 12.5, "12:30"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := formatHoursSimple(tc.hours)
			if result != tc.expected {
				t.Errorf("formatHoursSimple(%f) = %q, expected %q", tc.hours, result, tc.expected)
			}
		})
	}
}

func TestHasRunningTimer(t *testing.T) {
	t.Run("given model with no running entries when checked then returns false", func(t *testing.T) {
		model := newTestModel()
		model.timeEntries = []harvest.TimeEntry{
			{ID: 1, IsRunning: false},
			{ID: 2, IsRunning: false},
		}

		if model.hasRunningTimer() {
			t.Error("expected hasRunningTimer to be false when no entries are running")
		}
	})

	t.Run("given model with running entry when checked then returns true", func(t *testing.T) {
		model := newTestModel()
		model.timeEntries = []harvest.TimeEntry{
			{ID: 1, IsRunning: false},
			{ID: 2, IsRunning: true},
			{ID: 3, IsRunning: false},
		}

		if !model.hasRunningTimer() {
			t.Error("expected hasRunningTimer to be true when an entry is running")
		}
	})

	t.Run("given model with empty entries when checked then returns false", func(t *testing.T) {
		model := newTestModel()
		model.timeEntries = []harvest.TimeEntry{}

		if model.hasRunningTimer() {
			t.Error("expected hasRunningTimer to be false for empty entries")
		}
	})
}

func TestStatusMessageAutoClearing(t *testing.T) {
	t.Run("given status message when tick received after 3 seconds then message is cleared", func(t *testing.T) {
		model := newTestModel()
		model.setStatusMessage("Timer started successfully")
		// Set the timestamp to 4 seconds ago to simulate time passing
		model.statusMessageTime = time.Now().Add(-4 * time.Second)

		msg := tickMsg(time.Now())
		updatedModel, _ := model.Update(msg)
		m := updatedModel.(Model)

		if m.statusMessage != "" {
			t.Errorf("expected status message to be cleared after 3 seconds, got %q", m.statusMessage)
		}
	})

	t.Run("given status message when tick received before 3 seconds then message persists", func(t *testing.T) {
		model := newTestModel()
		model.setStatusMessage("Timer started successfully")
		// Set the timestamp to 1 second ago
		model.statusMessageTime = time.Now().Add(-1 * time.Second)

		msg := tickMsg(time.Now())
		updatedModel, _ := model.Update(msg)
		m := updatedModel.(Model)

		if m.statusMessage == "" {
			t.Error("expected status message to persist before 3 seconds")
		}
	})

	t.Run("given no status message when tick received then no change", func(t *testing.T) {
		model := newTestModel()
		model.statusMessage = ""

		msg := tickMsg(time.Now())
		updatedModel, _ := model.Update(msg)
		m := updatedModel.(Model)

		if m.statusMessage != "" {
			t.Errorf("expected no status message, got %q", m.statusMessage)
		}
	})
}

func TestStatusMessageStyling(t *testing.T) {
	t.Run("given success status message when rendered then appears in output", func(t *testing.T) {
		model := newTestModel()
		model.timeEntries = []harvest.TimeEntry{
			{
				ID: 1, Hours: 1.0,
				Client:  harvest.TimeEntryClient{ID: 1, Name: "C"},
				Project: harvest.TimeEntryProject{ID: 1, Name: "P"},
				Task:    harvest.TimeEntryTask{ID: 1, Name: "T"},
			},
		}
		model.setStatusMessage("Timer started successfully")
		model.currentDate = time.Date(2025, 1, 19, 0, 0, 0, 0, time.UTC)

		output := model.View()

		if !strings.Contains(output, "Timer started successfully") {
			t.Error("expected success status message to appear in output")
		}
	})

	t.Run("given error status message when rendered then appears in output", func(t *testing.T) {
		model := newTestModel()
		model.timeEntries = []harvest.TimeEntry{
			{
				ID: 1, Hours: 1.0,
				Client:  harvest.TimeEntryClient{ID: 1, Name: "C"},
				Project: harvest.TimeEntryProject{ID: 1, Name: "P"},
				Task:    harvest.TimeEntryTask{ID: 1, Name: "T"},
			},
		}
		model.setStatusMessage("Error: failed to connect")
		model.currentDate = time.Date(2025, 1, 19, 0, 0, 0, 0, time.UTC)

		output := model.View()

		if !strings.Contains(output, "Error: failed to connect") {
			t.Error("expected error status message to appear in output")
		}
	})
}

func TestHelpViewRendering(t *testing.T) {
	t.Run("given help view when rendered then displays all keybinding categories", func(t *testing.T) {
		model := newTestModel()
		model.currentView = ViewHelp

		output := model.View()

		expectedContent := []string{
			"Help",
			"Navigation",
			"Move up",
			"Move down",
			"Previous day",
			"Next day",
			"Jump to today",
			"Time Entry Actions",
			"New entry",
			"Edit entry",
			"Delete entry",
			"Start/stop timer",
			"General",
			"Toggle this help",
			"Quit/Go back",
			"Force quit",
		}

		for _, expected := range expectedContent {
			if !strings.Contains(output, expected) {
				t.Errorf("expected help output to contain %q", expected)
			}
		}
	})

	t.Run("given help view when escape pressed then returns to list", func(t *testing.T) {
		model := newTestModel()
		model.currentView = ViewHelp

		msg := tea.KeyMsg{Type: tea.KeyEscape}
		updatedModel, _ := model.Update(msg)
		m := updatedModel.(Model)

		if m.currentView != ViewList {
			t.Errorf("expected view to return to ViewList, got %v", m.currentView)
		}
	})

	t.Run("given help view when q pressed then returns to list", func(t *testing.T) {
		model := newTestModel()
		model.currentView = ViewHelp

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}
		updatedModel, _ := model.Update(msg)
		m := updatedModel.(Model)

		if m.currentView != ViewList {
			t.Errorf("expected view to return to ViewList on q, got %v", m.currentView)
		}
	})
}

func TestConfirmDeleteViewRendering(t *testing.T) {
	t.Run("given confirm delete view with entry when rendered then shows entry details", func(t *testing.T) {
		model := newTestModel()
		model.currentView = ViewConfirmDelete
		model.editingEntry = &harvest.TimeEntry{
			ID:    1,
			Hours: 2.5,
			Notes: "Important work",
		}

		output := model.View()

		if !strings.Contains(output, "Confirm Delete") {
			t.Error("expected output to contain 'Confirm Delete'")
		}
		if !strings.Contains(output, "Important work") {
			t.Error("expected output to contain entry notes")
		}
		if !strings.Contains(output, "2:30") {
			t.Error("expected output to contain formatted duration")
		}
		if !strings.Contains(output, "y") || !strings.Contains(output, "confirm") {
			t.Error("expected output to contain confirmation instructions")
		}
	})
}

func TestEditFormTabNavigation(t *testing.T) {
	t.Run("given edit view on task field when tab pressed then moves to notes field", func(t *testing.T) {
		model := newTestModel()
		model.currentView = ViewEditEntry
		model.editingEntry = &harvest.TimeEntry{ID: 1, Hours: 1.5, Notes: "Test"}
		model.editCurrentField = 0

		msg := tea.KeyMsg{Type: tea.KeyTab}
		updatedModel, _ := model.Update(msg)
		m := updatedModel.(Model)

		if m.editCurrentField != 1 {
			t.Errorf("expected editCurrentField to be 1 after tab, got %d", m.editCurrentField)
		}
	})

	t.Run("given edit view on duration field when tab pressed then wraps to task field", func(t *testing.T) {
		model := newTestModel()
		model.currentView = ViewEditEntry
		model.editingEntry = &harvest.TimeEntry{ID: 1, Hours: 1.5, Notes: "Test"}
		model.editCurrentField = 2

		msg := tea.KeyMsg{Type: tea.KeyTab}
		updatedModel, _ := model.Update(msg)
		m := updatedModel.(Model)

		if m.editCurrentField != 0 {
			t.Errorf("expected editCurrentField to wrap to 0 after tab, got %d", m.editCurrentField)
		}
	})

	t.Run("given edit view on task field when shift+tab pressed then wraps to duration field", func(t *testing.T) {
		model := newTestModel()
		model.currentView = ViewEditEntry
		model.editingEntry = &harvest.TimeEntry{ID: 1, Hours: 1.5, Notes: "Test"}
		model.editCurrentField = 0

		msg := tea.KeyMsg{Type: tea.KeyShiftTab}
		updatedModel, _ := model.Update(msg)
		m := updatedModel.(Model)

		if m.editCurrentField != 2 {
			t.Errorf("expected editCurrentField to wrap to 2 on shift+tab, got %d", m.editCurrentField)
		}
	})
}

func TestNewEntryFormTabNavigation(t *testing.T) {
	t.Run("given new entry form when tab pressed through fields then cycles correctly", func(t *testing.T) {
		model := newTestModel()
		model.currentView = ViewNewEntry
		model.newEntryCurrentField = 0

		expectedFields := []int{1, 2, 3, 0} // Tab cycles: 0->1->2->3->0
		for _, expected := range expectedFields {
			msg := tea.KeyMsg{Type: tea.KeyTab}
			updatedModel, _ := model.Update(msg)
			model = updatedModel.(Model)

			if model.newEntryCurrentField != expected {
				t.Errorf("expected newEntryCurrentField to be %d after tab, got %d", expected, model.newEntryCurrentField)
			}
		}
	})

	t.Run("given new entry form on project field when enter pressed then opens project selection", func(t *testing.T) {
		model := newTestModel()
		model.currentView = ViewNewEntry
		model.newEntryCurrentField = 0
		model.projectsWithTasks = []harvest.ProjectWithTasks{
			{
				Project: harvest.Project{ID: 1, Name: "Test", Client: harvest.ProjectClient{ID: 1, Name: "Client"}},
				Tasks:   []harvest.Task{{ID: 1, Name: "Task"}},
			},
		}

		msg := tea.KeyMsg{Type: tea.KeyEnter}
		updatedModel, _ := model.Update(msg)
		m := updatedModel.(Model)

		if m.currentView != ViewSelectProject {
			t.Errorf("expected view to be ViewSelectProject, got %v", m.currentView)
		}
	})
}

func TestBillableToggleInteraction(t *testing.T) {
	t.Run("given billable toggle view when space pressed then toggles billable", func(t *testing.T) {
		model := newTestModel()
		model.currentView = ViewBillableToggle
		model.newEntryBillable = true

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")}
		updatedModel, _ := model.Update(msg)
		m := updatedModel.(Model)

		if m.newEntryBillable {
			t.Error("expected billable to be toggled to false")
		}

		// Toggle again
		updatedModel, _ = m.Update(msg)
		m = updatedModel.(Model)

		if !m.newEntryBillable {
			t.Error("expected billable to be toggled back to true")
		}
	})

	t.Run("given billable toggle view when 'b' pressed then toggles billable", func(t *testing.T) {
		model := newTestModel()
		model.currentView = ViewBillableToggle
		model.newEntryBillable = true

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")}
		updatedModel, _ := model.Update(msg)
		m := updatedModel.(Model)

		if m.newEntryBillable {
			t.Error("expected billable to be toggled to false on 'b' press")
		}
	})

	t.Run("given billable toggle view when tab pressed then toggles billable", func(t *testing.T) {
		model := newTestModel()
		model.currentView = ViewBillableToggle
		model.newEntryBillable = false

		msg := tea.KeyMsg{Type: tea.KeyTab}
		updatedModel, _ := model.Update(msg)
		m := updatedModel.(Model)

		if !m.newEntryBillable {
			t.Error("expected billable to be toggled to true on tab press")
		}
	})

	t.Run("given billable toggle view when escape pressed then goes back to duration input", func(t *testing.T) {
		model := newTestModel()
		model.currentView = ViewBillableToggle

		msg := tea.KeyMsg{Type: tea.KeyEscape}
		updatedModel, _ := model.Update(msg)
		m := updatedModel.(Model)

		if m.currentView != ViewDurationInput {
			t.Errorf("expected view to be ViewDurationInput, got %v", m.currentView)
		}
	})
}

func TestSelectedEntryIndexAdjustmentOnDelete(t *testing.T) {
	t.Run("given last entry selected when entry deleted then index adjusts down", func(t *testing.T) {
		model := newTestModel()
		model.timeEntries = []harvest.TimeEntry{
			{ID: 1, Hours: 1.0, Notes: "Entry 1"},
			{ID: 2, Hours: 2.0, Notes: "Entry 2"},
			{ID: 3, Hours: 3.0, Notes: "Entry 3"},
		}
		model.selectedEntryIndex = 2 // Last entry

		// Delete the last entry
		msg := timeEntryDeletedMsg{entryID: 3}
		updatedModel, _ := model.Update(msg)
		m := updatedModel.(Model)

		if m.selectedEntryIndex != 1 {
			t.Errorf("expected selectedEntryIndex to adjust to 1, got %d", m.selectedEntryIndex)
		}
		if len(m.timeEntries) != 2 {
			t.Errorf("expected 2 entries, got %d", len(m.timeEntries))
		}
	})

	t.Run("given first entry selected when non-selected entry deleted then index unchanged", func(t *testing.T) {
		model := newTestModel()
		model.timeEntries = []harvest.TimeEntry{
			{ID: 1, Hours: 1.0, Notes: "Entry 1"},
			{ID: 2, Hours: 2.0, Notes: "Entry 2"},
			{ID: 3, Hours: 3.0, Notes: "Entry 3"},
		}
		model.selectedEntryIndex = 0

		// Delete second entry
		msg := timeEntryDeletedMsg{entryID: 2}
		updatedModel, _ := model.Update(msg)
		m := updatedModel.(Model)

		if m.selectedEntryIndex != 0 {
			t.Errorf("expected selectedEntryIndex to remain 0, got %d", m.selectedEntryIndex)
		}
	})
}

func TestEntryRenderingWithIndicators(t *testing.T) {
	t.Run("given running entry when rendered then shows running indicator", func(t *testing.T) {
		model := newTestModel()
		model.timeEntries = []harvest.TimeEntry{
			{
				ID:        1,
				Hours:     1.5,
				IsRunning: true,
				Client:    harvest.TimeEntryClient{ID: 1, Name: "Acme"},
				Project:   harvest.TimeEntryProject{ID: 1, Name: "Web"},
				Task:      harvest.TimeEntryTask{ID: 1, Name: "Dev"},
			},
		}
		model.currentDate = time.Date(2025, 1, 19, 0, 0, 0, 0, time.UTC)

		output := model.View()

		if !strings.Contains(output, "●") {
			t.Error("expected running entry to show ● indicator")
		}
	})

	t.Run("given locked entry when rendered then shows lock indicator", func(t *testing.T) {
		model := newTestModel()
		model.timeEntries = []harvest.TimeEntry{
			{
				ID:       1,
				Hours:    1.5,
				IsLocked: true,
				Client:   harvest.TimeEntryClient{ID: 1, Name: "Acme"},
				Project:  harvest.TimeEntryProject{ID: 1, Name: "Web"},
				Task:     harvest.TimeEntryTask{ID: 1, Name: "Dev"},
			},
		}
		model.currentDate = time.Date(2025, 1, 19, 0, 0, 0, 0, time.UTC)

		output := model.View()

		if !strings.Contains(output, "🔒") {
			t.Error("expected locked entry to show 🔒 indicator")
		}
	})
}

func TestDateHeaderTextForToday(t *testing.T) {
	t.Run("given current date is today when rendered then shows Today's Entries", func(t *testing.T) {
		model := newTestModel()
		model.currentDate = time.Now()
		model.timeEntries = []harvest.TimeEntry{}

		output := model.View()

		if !strings.Contains(output, "Today's Entries") {
			t.Error("expected header to show 'Today's Entries' for current date")
		}
	})

	t.Run("given current date is not today when rendered then shows weekday name", func(t *testing.T) {
		model := newTestModel()
		// Use a date that is definitely not today
		model.currentDate = time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC) // Wednesday
		model.timeEntries = []harvest.TimeEntry{}

		output := model.View()

		if !strings.Contains(output, "Wednesday's Entries") {
			t.Errorf("expected header to show 'Wednesday's Entries', got output:\n%s", output)
		}
	})
}

func TestCtrlCQuit(t *testing.T) {
	t.Run("given any view when ctrl+c pressed then quits", func(t *testing.T) {
		views := []ViewState{ViewList, ViewEditEntry, ViewHelp, ViewConfirmDelete, ViewSelectProject, ViewSelectTask}

		for _, view := range views {
			model := newTestModel()
			model.currentView = view

			msg := tea.KeyMsg{Type: tea.KeyCtrlC}
			_, cmd := model.Update(msg)

			if cmd == nil {
				t.Errorf("expected quit command from ctrl+c in view %v", view)
			}
		}
	})
}

func TestNavigationClearsStatusMessage(t *testing.T) {
	t.Run("given status message when up key pressed then status message cleared", func(t *testing.T) {
		model := newTestModel()
		model.timeEntries = []harvest.TimeEntry{
			{ID: 1, Notes: "First"},
			{ID: 2, Notes: "Second"},
		}
		model.selectedEntryIndex = 1
		model.statusMessage = "Some status"

		msg := tea.KeyMsg{Type: tea.KeyUp}
		updatedModel, _ := model.handleListViewKeys(msg)
		m := updatedModel.(Model)

		if m.statusMessage != "" {
			t.Error("expected status message to be cleared on navigation")
		}
	})

	t.Run("given status message when date navigation occurs then status message cleared", func(t *testing.T) {
		model := newTestModel()
		model.currentDate = time.Date(2025, 1, 19, 0, 0, 0, 0, time.UTC)
		model.statusMessage = "Old message"
		model.statusMessageTime = time.Now()

		msg := tea.KeyMsg{Type: tea.KeyLeft}
		updatedModel, _ := model.handleListViewKeys(msg)
		m := updatedModel.(Model)

		if m.statusMessage != "" {
			t.Error("expected status message to be cleared on date navigation")
		}
	})
}

func TestRenderKeybinding(t *testing.T) {
	t.Run("given key and description when rendered then contains both", func(t *testing.T) {
		result := RenderKeybinding("n", "new")

		if !strings.Contains(result, "n") {
			t.Error("expected keybinding to contain key 'n'")
		}
		if !strings.Contains(result, "new") {
			t.Error("expected keybinding to contain description 'new'")
		}
	})
}

func TestRenderEntryPath(t *testing.T) {
	t.Run("given client project task when rendered then shows arrow-separated path", func(t *testing.T) {
		result := RenderEntryPath("Acme", "Website", "Dev")

		if !strings.Contains(result, "Acme") {
			t.Error("expected path to contain client name")
		}
		if !strings.Contains(result, "Website") {
			t.Error("expected path to contain project name")
		}
		if !strings.Contains(result, "Dev") {
			t.Error("expected path to contain task name")
		}
		if !strings.Contains(result, "→") {
			t.Error("expected path to contain arrow separators")
		}
	})
}

func TestRenderNotes(t *testing.T) {
	t.Run("given notes text when rendered then wraps in quotes", func(t *testing.T) {
		result := RenderNotes("Working on feature")

		if !strings.Contains(result, `"Working on feature"`) {
			t.Errorf("expected notes to be wrapped in quotes, got: %s", result)
		}
	})

	t.Run("given empty notes when rendered then returns empty string", func(t *testing.T) {
		result := RenderNotes("")

		if result != "" {
			t.Errorf("expected empty string for empty notes, got: %s", result)
		}
	})
}

func TestRenderEmptyState(t *testing.T) {
	t.Run("given empty state when rendered then contains guidance text", func(t *testing.T) {
		result := RenderEmptyState()

		if !strings.Contains(result, "No entries yet") {
			t.Error("expected empty state to contain 'No entries yet'")
		}
		if !strings.Contains(result, "'n'") {
			t.Error("expected empty state to mention 'n' key for creating entries")
		}
	})
}

func TestTimeEntryFetchedMessage(t *testing.T) {
	t.Run("given successful fetch when message received then updates entries and clears loading", func(t *testing.T) {
		model := newTestModel()
		model.loading = true
		model.errorMessage = "old error"

		entries := []harvest.TimeEntry{
			{ID: 1, Hours: 1.5, Notes: "Entry 1"},
			{ID: 2, Hours: 2.0, Notes: "Entry 2"},
		}

		msg := timeEntriesFetchedMsg{entries: entries}
		updatedModel, _ := model.Update(msg)
		m := updatedModel.(Model)

		if m.loading {
			t.Error("expected loading to be false after fetch")
		}
		if m.errorMessage != "" {
			t.Errorf("expected error message to be cleared, got %q", m.errorMessage)
		}
		if len(m.timeEntries) != 2 {
			t.Errorf("expected 2 entries, got %d", len(m.timeEntries))
		}
	})

	t.Run("given failed fetch when message received then sets error message", func(t *testing.T) {
		model := newTestModel()
		model.loading = true

		msg := timeEntriesFetchedMsg{err: errForTest("network error")}
		updatedModel, _ := model.Update(msg)
		m := updatedModel.(Model)

		if m.loading {
			t.Error("expected loading to be false after failed fetch")
		}
		if m.errorMessage == "" {
			t.Error("expected error message to be set on fetch failure")
		}
	})
}

// errForTest is a simple error type for tests
type errForTest string

func (e errForTest) Error() string { return string(e) }

func TestProjectsWithTasksFetchedMessage(t *testing.T) {
	t.Run("given successful projects fetch when message received then updates projects", func(t *testing.T) {
		model := newTestModel()

		projects := []harvest.ProjectWithTasks{
			{
				Project: harvest.Project{ID: 1, Name: "Proj1"},
				Tasks:   []harvest.Task{{ID: 1, Name: "Task1"}},
			},
		}

		msg := projectsWithTasksFetchedMsg{projectsWithTasks: projects}
		updatedModel, _ := model.Update(msg)
		m := updatedModel.(Model)

		if len(m.projectsWithTasks) != 1 {
			t.Errorf("expected 1 project, got %d", len(m.projectsWithTasks))
		}
	})

	t.Run("given failed projects fetch when message received then sets error", func(t *testing.T) {
		model := newTestModel()

		msg := projectsWithTasksFetchedMsg{err: errForTest("API error")}
		updatedModel, _ := model.Update(msg)
		m := updatedModel.(Model)

		if m.errorMessage == "" {
			t.Error("expected error message to be set on projects fetch failure")
		}
	})
}

func TestNewEntryCreatedPrepended(t *testing.T) {
	t.Run("given existing entries when new entry created then prepended to list", func(t *testing.T) {
		model := newTestModel()
		model.timeEntries = []harvest.TimeEntry{
			{ID: 1, Hours: 1.0, Notes: "Old entry"},
		}

		newEntry := &harvest.TimeEntry{
			ID:    99,
			Hours: 0.5,
			Notes: "New entry",
		}

		msg := timeEntryCreatedMsg{entry: newEntry}
		updatedModel, _ := model.Update(msg)
		m := updatedModel.(Model)

		if len(m.timeEntries) != 2 {
			t.Fatalf("expected 2 entries, got %d", len(m.timeEntries))
		}
		if m.timeEntries[0].ID != 99 {
			t.Errorf("expected new entry (ID=99) to be first, got ID=%d", m.timeEntries[0].ID)
		}
		if m.timeEntries[1].ID != 1 {
			t.Errorf("expected old entry (ID=1) to be second, got ID=%d", m.timeEntries[1].ID)
		}
	})
}

func TestFooterKeybindings(t *testing.T) {
	t.Run("given list view when rendered then footer shows all keybindings", func(t *testing.T) {
		model := newTestModel()
		model.timeEntries = []harvest.TimeEntry{}
		model.currentDate = time.Date(2025, 1, 19, 0, 0, 0, 0, time.UTC)

		output := model.View()

		expectedKeys := []string{"n", "new", "e", "edit", "s", "start/stop", "d", "delete", "?", "help", "q", "quit"}
		for _, key := range expectedKeys {
			if !strings.Contains(output, key) {
				t.Errorf("expected footer to contain %q", key)
			}
		}
	})
}
