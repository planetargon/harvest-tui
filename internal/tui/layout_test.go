package tui

import (
	"fmt"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/planetargon/harvest-tui/internal/harvest"
)

func populatedLayoutModel() Model {
	m := newTestModel()
	for i := 0; i < 30; i++ {
		m.timeEntries = append(m.timeEntries, harvest.TimeEntry{
			ID: i + 1, Hours: 1, Notes: fmt.Sprintf("entry-%02d", i),
			Client:  harvest.TimeEntryClient{Name: "Client"},
			Project: harvest.TimeEntryProject{Name: "Project"},
			Task:    harvest.TimeEntryTask{Name: "Task"},
		})
	}
	return m
}

func resizeLayoutModel(m Model, width, height int) Model {
	updated, _ := m.Update(tea.WindowSizeMsg{Width: width, Height: height})
	return updated.(Model)
}

func pressLayoutKey(m Model, key tea.KeyType) Model {
	updated, _ := m.Update(tea.KeyMsg{Type: key})
	return updated.(Model)
}

func TestResponsiveLayout(t *testing.T) {
	for _, view := range []ViewState{ViewLoading, ViewList, ViewHelp, ViewNewEntry, ViewEditEntry, ViewConfirmDelete, ViewSelectProject, ViewSelectTask, ViewNotesInput, ViewDurationInput, ViewBillableToggle} {
		for _, size := range [][2]int{{120, 40}, {64, 14}, {25, 8}, {2, 6}, {1, 1}, {0, 0}} {
			t.Run(fmt.Sprintf("given view %d when resized to %dx%d then output fits terminal", view, size[0], size[1]), func(t *testing.T) {
				m := populatedLayoutModel()
				m.currentView = view
				m = resizeLayoutModel(m, size[0], size[1])
				output := m.View()
				if size[0] == 0 || size[1] == 0 {
					if output != "" {
						t.Fatalf("expected empty output before terminal has dimensions")
					}
					return
				}
				lines := strings.Split(output, "\n")
				if len(lines) > size[1] {
					t.Fatalf("rendered %d rows in %d-row terminal", len(lines), size[1])
				}
				for i, line := range lines {
					if width := lipgloss.Width(line); width > size[0] {
						t.Errorf("line %d has width %d, exceeds %d", i, width, size[0])
					}
				}
				if size[0] >= 2 && size[1] >= m.minimumHeight() {
					if len(lines) != size[1] || lipgloss.Width(lines[0]) != size[0] {
						t.Errorf("expected shell to fill %dx%d terminal", size[0], size[1])
					}
				}
			})
		}
	}
}

func TestEntryScrolling(t *testing.T) {
	t.Run("given many entries when selection moves down then selected entry and footer stay visible", func(t *testing.T) {
		m := resizeLayoutModel(populatedLayoutModel(), 80, 14)
		for i := 0; i < 29; i++ {
			m = pressLayoutKey(m, tea.KeyDown)
		}
		output := ansi.Strip(m.View())
		if !strings.Contains(output, "entry-29") || strings.Contains(output, "entry-00") {
			t.Fatal("expected last selected entry to scroll into view")
		}
		if !strings.Contains(output, "Harvest Time Tracker") || !strings.Contains(output, "start/stop") {
			t.Fatal("expected title and footer to remain visible")
		}
		for i := 0; i < 29; i++ {
			m = pressLayoutKey(m, tea.KeyUp)
		}
		if !strings.Contains(m.View(), "entry-00") {
			t.Fatal("expected first entry to scroll back into view")
		}
	})

	t.Run("given entries when page keys and mouse wheel used then navigation scrolls within same day", func(t *testing.T) {
		m := resizeLayoutModel(populatedLayoutModel(), 80, 14)
		date := m.currentDate
		m = pressLayoutKey(m, tea.KeyPgDown)
		if m.selectedEntryIndex <= 1 || !strings.Contains(m.View(), fmt.Sprintf("entry-%02d", m.selectedEntryIndex)) {
			t.Fatal("expected page down to reveal a later selected entry")
		}
		before := m.selectedEntryIndex
		updated, _ := m.Update(tea.MouseMsg{Button: tea.MouseButtonWheelDown, Action: tea.MouseActionPress})
		m = updated.(Model)
		if m.selectedEntryIndex <= before || !m.currentDate.Equal(date) {
			t.Fatal("expected wheel to move entry selection without changing day")
		}
		m = pressLayoutKey(m, tea.KeyEnd)
		if m.selectedEntryIndex != 29 || !strings.Contains(m.View(), "entry-29") {
			t.Fatal("expected end to reveal last entry")
		}
		m = pressLayoutKey(m, tea.KeyHome)
		if m.selectedEntryIndex != 0 || !strings.Contains(m.View(), "entry-00") {
			t.Fatal("expected home to reveal first entry")
		}
	})

	t.Run("given scrolled entries when day changes then date moves and scroll resets", func(t *testing.T) {
		m := resizeLayoutModel(populatedLayoutModel(), 80, 14)
		m.currentDate = time.Date(2026, 10, 19, 12, 0, 0, 0, time.Local)
		m = pressLayoutKey(m, tea.KeyEnd)
		updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyLeft})
		m = updated.(Model)
		if cmd == nil || m.currentDate.Day() != 18 || m.selectedEntryIndex != 0 {
			t.Fatal("expected previous day to load with first entry selected")
		}
		updated, _ = m.Update(timeEntriesFetchedMsg{entries: m.timeEntries})
		m = updated.(Model)
		if !strings.Contains(m.View(), "entry-00") || strings.Contains(m.View(), "entry-29") {
			t.Fatal("expected new day's list to start at top")
		}
		m = pressLayoutKey(m, tea.KeyRight)
		if m.currentDate.Day() != 19 {
			t.Fatal("expected right arrow to advance one day")
		}
		updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
		m = updated.(Model)
		if cmd == nil || m.currentDate.Format("2006-01-02") != time.Now().Format("2006-01-02") {
			t.Fatal("expected t to reload today's entries")
		}
	})

	t.Run("given scrolled entries when terminal grows then viewport clamps and selection remains visible", func(t *testing.T) {
		m := resizeLayoutModel(populatedLayoutModel(), 64, 12)
		m = pressLayoutKey(m, tea.KeyEnd)
		m = resizeLayoutModel(m, 140, 60)
		if !strings.Contains(m.View(), "entry-29") || lipgloss.Height(m.View()) != 60 {
			t.Fatal("expected selected entry to remain visible after resize")
		}
		updated, _ := m.Update(timeEntriesFetchedMsg{entries: m.timeEntries[:1]})
		m = updated.(Model)
		if !strings.Contains(m.View(), "entry-00") {
			t.Fatal("expected viewport to clamp when entries disappear")
		}
	})
}

func TestScrollLayoutDetails(t *testing.T) {
	t.Run("given scrolled entries when an action fails then status stays visible", func(t *testing.T) {
		m := resizeLayoutModel(populatedLayoutModel(), 80, 14)
		m = pressLayoutKey(m, tea.KeyEnd)
		updated, _ := m.Update(timeEntryStartedMsg{err: fmt.Errorf("permission denied")})
		m = updated.(Model)
		output := m.View()
		if !strings.Contains(output, "permission denied") || lipgloss.Height(output) != 14 {
			t.Fatal("expected status above footer without overflowing terminal")
		}
	})

	t.Run("given long styled paths when terminal is narrow then durations remain visible", func(t *testing.T) {
		m := populatedLayoutModel()
		m.timeEntries = m.timeEntries[:1]
		m.timeEntries[0].Client.Name = strings.Repeat("Client", 10)
		m.timeEntries[0].Project.Name = strings.Repeat("Project", 10)
		m.timeEntries[0].Task.Name = strings.Repeat("Task", 10)
		m = resizeLayoutModel(m, 40, 14)
		if !strings.Contains(m.View(), "1:00") {
			t.Fatal("expected duration to remain visible beside truncated path")
		}
	})

	t.Run("given short terminal when form focus advances then focused field scrolls into view", func(t *testing.T) {
		m := resizeLayoutModel(newTestModel(), 60, 10)
		m.currentView = ViewNewEntry
		for i := 0; i < 3; i++ {
			m = pressLayoutKey(m, tea.KeyTab)
		}
		if !strings.Contains(m.View(), "Duration:") {
			t.Fatal("expected focused duration field to scroll into view")
		}
	})

	t.Run("given long project and task lists when navigating then native pagination keeps selection visible", func(t *testing.T) {
		for _, view := range []ViewState{ViewSelectProject, ViewSelectTask} {
			m := populatedLayoutModel()
			var tasks []harvest.Task
			for i := 0; i < 30; i++ {
				m.projectsWithTasks = append(m.projectsWithTasks, harvest.ProjectWithTasks{
					Project: harvest.Project{ID: i + 1, Name: fmt.Sprintf("project-%02d", i)},
				})
				tasks = append(tasks, harvest.Task{ID: i + 1, Name: fmt.Sprintf("task-%02d", i)})
			}
			m.currentView = view
			m.updateProjectList()
			m.updateTaskList(tasks)
			height := 12
			if view == ViewSelectTask {
				height = 13
			}
			m = resizeLayoutModel(m, 80, height)
			for i := 0; i < 29; i++ {
				m = pressLayoutKey(m, tea.KeyDown)
			}
			expected := "project-29"
			if view == ViewSelectTask {
				expected = "task-29"
			}
			if !strings.Contains(m.View(), expected) {
				t.Fatalf("expected %s to remain visible", expected)
			}
		}
	})

	t.Run("given styled unicode text when clipped then display width stays bounded", func(t *testing.T) {
		line := "\x1b[31m日本語🌾 café\x1b[0m"
		for width := 0; width < 15; width++ {
			if result := truncateStyledLine(line, width); lipgloss.Width(result) > width {
				t.Errorf("clipped text exceeds %d display columns: %q", width, result)
			}
		}
	})

	t.Run("given unicode names when entry paths shorten then output remains valid unicode", func(t *testing.T) {
		m := populatedLayoutModel()
		m.timeEntries[0].Client.Name = strings.Repeat("日本語🌾", 20)
		m.timeEntries[0].Project.Name = strings.Repeat("café", 20)
		m = resizeLayoutModel(m, 60, 14)
		if !utf8.ValidString(m.View()) {
			t.Fatal("expected name truncation to preserve valid UTF-8")
		}
	})
}

func TestHelpScrolling(t *testing.T) {
	t.Run("given short terminal when help scrolls then hidden instructions become reachable", func(t *testing.T) {
		m := resizeLayoutModel(newTestModel(), 80, 12)
		m.currentView = ViewHelp
		if strings.Contains(m.View(), "Force quit") {
			t.Fatal("expected help to overflow short viewport")
		}
		m = pressLayoutKey(m, tea.KeyEnd)
		if !strings.Contains(m.View(), "Force quit") || !strings.Contains(m.View(), "back") {
			t.Fatal("expected help bottom and footer to remain reachable")
		}
		m = pressLayoutKey(m, tea.KeyHome)
		if !strings.Contains(m.View(), "Navigation") {
			t.Fatal("expected home to return help to top")
		}
	})
}

func TestInputResize(t *testing.T) {
	for _, position := range []int{70, 147} {
		t.Run(fmt.Sprintf("given input cursor at %d when terminal shrinks then cursor stays visible", position), func(t *testing.T) {
			m := newTestModel()
			m.currentView = ViewNewEntry
			m.newEntryCurrentField = 2
			input := textinput.New()
			value := strings.Repeat("A", 70) + "MID" + strings.Repeat("Z", 70) + "TAIL"
			input.SetValue(value)
			input.SetCursor(position)
			m.notesInput = &input
			m = resizeLayoutModel(m, 120, 18)
			m = resizeLayoutModel(m, 40, 12)
			if m.notesInput.Position() != position || m.notesInput.Value() != value {
				t.Fatal("expected resize to preserve cursor position and input value")
			}
			if lipgloss.Width(m.notesInput.View()) > m.notesInput.Width+lipgloss.Width(m.notesInput.Prompt)+1 {
				t.Fatal("expected input to fit its resized horizontal viewport")
			}
			expected := "MID"
			if position == len(value) {
				expected = "TAIL"
			}
			if !strings.Contains(m.View(), expected) {
				t.Fatalf("expected text next to cursor %q to remain visible", expected)
			}
		})
	}
}

func TestWizardFocusScrolling(t *testing.T) {
	for _, view := range []ViewState{ViewNotesInput, ViewDurationInput, ViewBillableToggle} {
		t.Run(fmt.Sprintf("given wizard view %d when terminal shrinks then active control stays visible", view), func(t *testing.T) {
			m := newTestModel()
			m.currentView = view
			notes := textinput.New()
			notes.SetValue("NOTES_SENTINEL")
			m.notesInput = &notes
			duration := textinput.New()
			duration.SetValue("1:23")
			m.durationInput = &duration
			m.newEntryNotes = "NOTES_SENTINEL"
			m.newEntryHours = "1:23"
			m = resizeLayoutModel(m, 60, 8)
			expected := "NOTES_SENTINEL"
			if view == ViewDurationInput {
				expected = "1:23"
			} else if view == ViewBillableToggle {
				expected = "[x] Billable"
			}
			if !strings.Contains(m.View(), expected) {
				t.Fatalf("expected active control %q to stay visible", expected)
			}
		})
	}
}

func TestRefreshSelectionVisibility(t *testing.T) {
	t.Run("given selected entry when refreshed preceding entries grow then selection stays visible", func(t *testing.T) {
		m := populatedLayoutModel()
		for i := range m.timeEntries {
			m.timeEntries[i].Notes = ""
		}
		m = resizeLayoutModel(m, 80, 14)
		for i := 0; i < 10; i++ {
			m = pressLayoutKey(m, tea.KeyDown)
		}
		entries := append([]harvest.TimeEntry(nil), m.timeEntries...)
		for i := range entries {
			entries[i].Notes = fmt.Sprintf("refreshed-%02d", i)
		}
		updated, _ := m.Update(timeEntriesFetchedMsg{entries: entries})
		m = updated.(Model)
		if m.selectedEntryIndex != 10 || !strings.Contains(m.View(), "refreshed-10") {
			t.Fatal("expected unchanged selected entry to stay visible after cards grow")
		}
	})
}

func TestDurationCursorWidth(t *testing.T) {
	t.Run("given narrow form when duration cursor is at end then input row fits without clipping cursor", func(t *testing.T) {
		m := newTestModel()
		m.currentView = ViewNewEntry
		m.newEntryCurrentField = 3
		duration := textinput.New()
		duration.SetValue("12:34")
		duration.CursorEnd()
		m.durationInput = &duration
		m = resizeLayoutModel(m, 23, 20)
		lines := strings.Split(m.currentShell().content, "\n")
		if lipgloss.Width(lines[len(lines)-1]) > m.shellWidth()-2 {
			t.Fatal("expected duration field and cursor to fit inside shell borders")
		}
	})
}

func TestShortPickerNotice(t *testing.T) {
	for _, view := range []ViewState{ViewSelectProject, ViewSelectTask} {
		t.Run(fmt.Sprintf("given picker %d when terminal cannot show items then display resize guidance", view), func(t *testing.T) {
			m := newTestModel()
			m.currentView = view
			height := 11
			if view == ViewSelectTask {
				height = 12
			}
			m = resizeLayoutModel(m, 80, height)
			if !strings.Contains(m.View(), "Terminal too small") {
				t.Fatal("expected resize guidance instead of a picker with invisible items")
			}
		})
	}
}

func TestPickerFilterHeight(t *testing.T) {
	for _, view := range []ViewState{ViewSelectProject, ViewSelectTask} {
		for _, status := range []string{"", "Task unavailable"} {
			t.Run(fmt.Sprintf("given picker %d with status %q when filtering at minimum height then results and pagination fit", view, status), func(t *testing.T) {
				m := newTestModel()
				m.currentView = view
				m.projectsWithTasks = []harvest.ProjectWithTasks{{Project: harvest.Project{ID: 1, Name: "project-visible"}}}
				m.updateProjectList()
				m.updateTaskList([]harvest.Task{{ID: 1, Name: "task-visible"}})
				height := 12
				expected := "project-visible"
				if view == ViewSelectTask {
					height, expected = 13, "task-visible"
				}
				if status != "" {
					height++
				}
				m = resizeLayoutModel(m, 80, height)
				updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
				m = updated.(Model)
				if status != "" {
					updated, _ = m.Update(timeEntryStartedMsg{err: fmt.Errorf("%s", status)})
					m = updated.(Model)
				}
				shell := m.currentShell()
				if strings.Count(shell.content, "\n") > m.bodyViewport(shell).Height {
					t.Fatal("expected complete filtered picker to fit available body height")
				}
				if !strings.Contains(m.View(), expected) {
					t.Fatal("expected matching item to stay visible while filtering")
				}
			})
		}
	}
}
