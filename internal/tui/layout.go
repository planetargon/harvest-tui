package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// shellView separates view content from the shared terminal layout. The first
// content line is the fixed title; focus is a line range in the remaining body.
type shellView struct {
	content    string
	footer     []string
	focusStart int
	focusEnd   int
}

// focusLastLine follows the active control in the single-control wizard views.
func (s shellView) focusLastLine() shellView {
	s.focusEnd = strings.Count(s.content, "\n")
	s.focusStart = max(0, s.focusEnd-1)
	return s
}

// Update synchronizes the scrollable body after applying application events.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.viewport = m.bodyViewport(m.currentShell())
	updated, cmd := m.update(msg)
	next := updated.(Model)

	viewChanged := next.currentView != m.currentView || !next.currentDate.Equal(m.currentDate)
	_, resized := msg.(tea.WindowSizeMsg)
	statusChanged := (next.statusMessage != "") != (m.statusMessage != "")
	if viewChanged || resized || statusChanged {
		next.setListSizes()
	}
	if viewChanged {
		next.viewport.GotoTop()
	}
	next.setInputSizes()
	shell := next.currentShell()
	next.viewport = next.bodyViewport(shell)
	_, entriesRefreshed := msg.(timeEntriesFetchedMsg)
	if viewChanged || resized || entriesRefreshed || next.viewport.Height != m.viewport.Height || next.selectedEntryIndex != m.selectedEntryIndex ||
		next.newEntryCurrentField != m.newEntryCurrentField || next.editCurrentField != m.editCurrentField {
		next.revealFocus(shell)
	}
	return next, cmd
}

func (m Model) bodyViewport(shell shellView) viewport.Model {
	v := m.viewport
	v.Width = max(1, m.shellWidth()-2)
	v.Height = m.bodyHeight()
	lines := strings.Split(shell.content, "\n")
	body := lines[1:]
	for i, line := range body {
		body[i] = truncateStyledLine(line, v.Width)
	}
	v.SetContent(strings.Join(body, "\n"))
	v.SetYOffset(v.YOffset) // Clamp after a resize or content removal.
	return v
}

func (m Model) bodyHeight() int {
	// Top border, title, footer separator, footer and bottom border stay fixed.
	height := max(1, m.height-5)
	if m.statusMessage != "" && m.height > 6 {
		height--
	}
	return height
}

func (m *Model) revealFocus(shell shellView) {
	if shell.focusEnd <= shell.focusStart {
		return
	}
	if shell.focusStart < m.viewport.YOffset || shell.focusEnd-shell.focusStart > m.viewport.Height {
		m.viewport.SetYOffset(shell.focusStart)
	} else if shell.focusEnd > m.viewport.YOffset+m.viewport.Height {
		m.viewport.SetYOffset(shell.focusEnd - m.viewport.Height)
	}
}

// View fills the terminal while keeping controls visible above and below the body.
func (m Model) View() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}
	if m.width < 2 || m.height < m.minimumHeight() {
		message := "Terminal too small"
		if m.height < m.minimumHeight() {
			message += fmt.Sprintf(": need at least %d rows", m.minimumHeight())
		}
		return truncateStyledLine(message, m.width)
	}
	shell := m.currentShell()
	width := m.shellWidth()
	inner := width - 2
	border := lipgloss.NewStyle().Foreground(borderColor)
	boxLine := func(line string) string {
		line = truncateStyledLine(line, inner)
		return border.Render("│") + line + strings.Repeat(" ", max(0, inner-lipgloss.Width(line))) + border.Render("│")
	}
	lines := []string{
		border.Render("┌" + strings.Repeat("─", inner) + "┐"),
		boxLine(strings.SplitN(shell.content, "\n", 2)[0]),
	}
	v := m.bodyViewport(shell)
	for _, line := range strings.Split(v.View(), "\n") {
		lines = append(lines, boxLine(line))
	}
	if m.statusMessage != "" && m.height > 6 {
		lines = append(lines, boxLine(m.renderStatusLine()))
	}
	lines = append(lines,
		border.Render("├"+strings.Repeat("─", inner)+"┤"),
		boxLine(" "+strings.Join(shell.footer, " ")),
		border.Render("└"+strings.Repeat("─", inner)+"┘"),
	)
	return strings.Join(lines, "\n")
}

func (m Model) minimumHeight() int {
	height := 6
	switch m.currentView {
	case ViewSelectProject:
		height = 12 // Shell, picker header, filter prompt, item and pagination.
	case ViewSelectTask:
		height = 13 // Task picker also shows the selected project.
	default:
		return 6
	}
	if m.statusMessage != "" {
		height++
	}
	return height
}

// handleScroll leaves day navigation and picker pagination with their owners.
// Entry scrolling moves selection too, so actions always target a visible entry.
func (m *Model) handleScroll(msg tea.Msg) bool {
	if m.currentView == ViewLoading || m.currentView == ViewSelectProject || m.currentView == ViewSelectTask {
		return false
	}
	direction, page, edge := 0, false, false
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "pgdown":
			direction, page = 1, true
		case "pgup":
			direction, page = -1, true
		case "home", "end":
			if m.currentView != ViewList && m.currentView != ViewHelp {
				return false // Preserve cursor movement in text inputs.
			}
			direction, edge = -1, true
			if msg.String() == "end" {
				direction = 1
			}
		case "up", "k":
			if m.currentView != ViewHelp {
				return false
			}
			direction = -1
		case "down", "j":
			if m.currentView != ViewHelp {
				return false
			}
			direction = 1
		default:
			return false
		}
	case tea.MouseMsg:
		if msg.Action != tea.MouseActionPress {
			return false
		}
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			direction = -1
		case tea.MouseButtonWheelDown:
			direction = 1
		default:
			return false
		}
	default:
		return false
	}
	if m.currentView == ViewList {
		if edge {
			m.selectedEntryIndex = 0
			if direction > 0 {
				m.selectedEntryIndex = max(0, len(m.timeEntries)-1)
			}
		} else {
			lines := 0
			for len(m.timeEntries) > 0 {
				index := m.selectedEntryIndex
				lines += lipgloss.Height(m.renderStyledTimeEntry(m.timeEntries[index], true, m.shellWidth()-4))
				m.selectedEntryIndex = min(max(0, index+direction), len(m.timeEntries)-1)
				if !page || lines >= m.viewport.Height || index == m.selectedEntryIndex {
					break
				}
			}
		}
		m.clearStatusMessage()
		m.revealFocus(m.currentShell())
		return true
	}
	switch {
	case edge && direction > 0:
		m.viewport.GotoBottom()
	case edge:
		m.viewport.GotoTop()
	case page:
		m.viewport.SetYOffset(m.viewport.YOffset + direction*m.viewport.Height)
	default:
		m.viewport.SetYOffset(m.viewport.YOffset + direction)
	}
	return true
}

func (m *Model) setInputSizes() {
	width := max(1, m.shellWidth()-19)
	for _, input := range []*textinput.Model{m.notesInput, m.editNotesInput} {
		resizeInput(input, width)
	}
	for _, input := range []*textinput.Model{m.durationInput, m.editDurationInput} {
		resizeInput(input, min(20, width))
	}
}

func resizeInput(input *textinput.Model, width int) {
	if input == nil || input.Width == width {
		return
	}
	position := input.Position()
	input.Width = width
	// Recompute Bubbles' horizontal offsets at the new width, then restore focus.
	input.CursorEnd()
	input.SetCursor(position)
}
