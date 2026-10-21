package tui

import (
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/planetargon/harvest-tui/internal/harvest"
)

// shellWidth returns the terminal width, allowing room for both side borders.
func (m Model) shellWidth() int {
	return max(2, m.width)
}

// renderTitleBar renders the title bar with date navigation.
func (m Model) renderTitleBar() string {
	width := m.shellWidth()

	dateStr := m.currentDate.Format("Mon, Jan 2, 2006")
	dateNav := ArrowNavStyle.Render("◀ ") + DateStyle.Render(dateStr) + ArrowNavStyle.Render(" ▶")

	titleText := "  " + TitleStyle.Render("🌾 Harvest Time Tracker")
	titleSuffix := dateNav + "  "
	spacerWidth := width - 2 - lipgloss.Width(titleText) - lipgloss.Width(titleSuffix)
	if spacerWidth < 1 {
		spacerWidth = 1
	}
	return titleText + strings.Repeat(" ", spacerWidth) + titleSuffix
}

// renderLoadingView renders the loading screen shown during startup.
func (m Model) renderLoadingView() shellView {
	titleBar := m.renderTitleBar()

	loadingMsg := "  " + m.spinner.View() + " " + AccentText.Render("Harvesting your data...")

	contentLines := []string{
		titleBar,
		"",
		loadingMsg,
		"",
	}

	content := strings.Join(contentLines, "\n")

	footerKeys := []string{
		RenderKeybinding("ctrl+c", "quit"),
	}

	return shellView{content: content, footer: footerKeys}
}

// renderStyledListView renders the main list view with proper styling.
func (m Model) renderStyledListView() shellView {
	// Calculate dimensions
	width := m.shellWidth()

	titleBar := m.renderTitleBar()

	// Calculate daily total with accent color (add elapsed time for running entry)
	totalHours := 0.0
	for _, entry := range m.timeEntries {
		if entry.IsRunning && !m.lastFetchTime.IsZero() {
			totalHours += entry.Hours + time.Since(m.lastFetchTime).Hours()
		} else {
			totalHours += entry.Hours
		}
	}
	totalStr := formatHoursSimple(totalHours)

	// Section header with Tokyo Night styling
	// Check if currentDate is today
	now := time.Now()
	isToday := m.currentDate.Year() == now.Year() &&
		m.currentDate.Month() == now.Month() &&
		m.currentDate.Day() == now.Day()

	headerText := "Today's Entries"
	if !isToday {
		headerText = m.currentDate.Format("Monday's Entries")
	}
	entriesText := SectionHeaderStyle.Render(headerText)
	totalLabelText := TotalLabel.Render("Total: ")
	totalValue := TotalValue.Render(totalStr)
	paddingWidth := width - lipgloss.Width(entriesText) - lipgloss.Width(totalLabelText) - lipgloss.Width(totalValue) - 4
	if paddingWidth < 1 {
		paddingWidth = 1
	}
	sectionHeader := "  " + entriesText + strings.Repeat(" ", paddingWidth) + totalLabelText + totalValue + "  "

	// Divider with Tokyo Night styling
	divider := "  " + RenderDividerWidth(width-4)

	// Handle loading state
	if m.loading {
		content := []string{
			titleBar,
			sectionHeader,
			divider,
			"    " + MutedText.Render("Loading..."),
			"",
		}
		return m.listShell(strings.Join(content, "\n"))
	}

	// Handle error state
	if m.errorMessage != "" {
		content := []string{
			titleBar,
			sectionHeader,
			divider,
			"    " + ErrorText.Render("Error: "+m.errorMessage),
			"",
		}
		return m.listShell(strings.Join(content, "\n"))
	}

	// Handle empty state
	if len(m.timeEntries) == 0 {
		content := []string{
			titleBar,
			sectionHeader,
			divider,
			"",
			"    " + RenderEmptyState(),
			"",
		}
		return m.listShell(strings.Join(content, "\n"))
	}

	// Record the selected entry's line range so navigation can keep it visible.
	contentLines := []string{titleBar, sectionHeader, divider}
	view := shellView{footer: listViewFooterKeys()}
	line := strings.Count(strings.Join(contentLines, "\n"), "\n")
	for i, entry := range m.timeEntries {
		rendered := m.renderStyledTimeEntry(entry, i == m.selectedEntryIndex, width-4)
		height := lipgloss.Height(rendered)
		if i == m.selectedEntryIndex {
			// Skip the card's decorative top padding when space is tight.
			view.focusStart, view.focusEnd = line+1, line+height
		}
		contentLines = append(contentLines, rendered)
		line += height
	}
	view.content = strings.Join(contentLines, "\n")
	return view
}

// listViewFooterKeys returns the standard footer keybindings for the list view.
func listViewFooterKeys() []string {
	return []string{
		RenderKeybinding("n", "new"),
		RenderKeybinding("e", "edit"),
		RenderKeybinding("s", "start/stop"),
		RenderKeybinding("d", "delete"),
		RenderKeybinding("?", "help"),
		RenderKeybinding("q", "quit"),
	}
}

// listShell adds the standard controls to the entry list.
func (m Model) listShell(content string) shellView {
	return shellView{content: content, footer: listViewFooterKeys()}
}

// renderStyledTimeEntry renders a single time entry with Tokyo Night styling.
func (m Model) renderStyledTimeEntry(entry harvest.TimeEntry, isSelected bool, maxWidth int) string {
	var lines []string

	// Build the entry path
	clientName := truncateString(entry.Client.Name, 20)
	projectName := truncateString(entry.Project.Name, 25)
	taskName := truncateString(entry.Task.Name, 20)

	// For running entries, add elapsed time since last fetch for a live display
	displayHours := entry.Hours
	if entry.IsRunning && !m.lastFetchTime.IsZero() {
		elapsed := time.Since(m.lastFetchTime).Hours()
		displayHours += elapsed
	}

	// Build styled components with optional selected background
	var entryPath, styledDuration, indicator string
	if isSelected {
		bg := selectedBg
		entryPath = ClientStyle.Background(bg).Render(clientName) +
			ArrowStyle.Background(bg).Render(" → ") +
			ProjectStyle.Background(bg).Render(projectName) +
			ArrowStyle.Background(bg).Render(" → ") +
			TaskStyle.Background(bg).Render(taskName)

		if entry.IsRunning {
			styledDuration = RunningDurationStyle.Background(bg).Render(formatHoursSimple(displayHours))
		} else {
			styledDuration = DurationStyle.Background(bg).Render(formatHoursSimple(displayHours))
		}

		if entry.IsRunning {
			indicator = " " + RunningDot.Background(bg).Render("●")
		} else if entry.IsLocked {
			indicator = " " + LockedIcon.Background(bg).Render("🔒")
		}
	} else {
		entryPath = RenderEntryPath(clientName, projectName, taskName)

		if entry.IsRunning {
			styledDuration = RunningDurationStyle.Render(formatHoursSimple(displayHours))
		} else if entry.IsLocked {
			styledDuration = DurationStyle.Copy().Foreground(mutedText).Render(formatHoursSimple(displayHours))
		} else {
			styledDuration = DurationStyle.Render(formatHoursSimple(displayHours))
		}

		if entry.IsRunning {
			indicator = " " + RunningDot.Render("●")
		} else if entry.IsLocked {
			indicator = " " + LockedIcon.Render("🔒")
		}
	}

	// Reserve the duration and indicator before fitting the path to narrow terminals.
	durationWidth := lipgloss.Width(styledDuration)
	indicatorWidth := lipgloss.Width(indicator)
	entryPath = truncateStyledLine(entryPath, maxWidth-durationWidth-indicatorWidth-5)
	pathWidth := lipgloss.Width(entryPath)
	padding := maxWidth - pathWidth - durationWidth - indicatorWidth - 4
	if padding < 1 {
		padding = 1
	}

	// Build the entry line
	var entryLine string
	if isSelected {
		// Selected entry with accent bar and full-width background
		bgSpacer := lipgloss.NewStyle().Background(selectedBg).Render(strings.Repeat(" ", padding))
		entryContent := entryPath + bgSpacer + styledDuration + indicator
		entryLine = SelectedEntry.Width(maxWidth).Render(entryContent)
	} else {
		// Unselected entry with left padding
		entryContent := entryPath + strings.Repeat(" ", padding) + styledDuration + indicator
		if entry.IsLocked {
			entryContent = LockedEntryStyle.Render(entryContent)
		}
		entryLine = UnselectedEntry.Render(entryContent)
	}

	lines = append(lines, entryLine)

	// Notes line with Tokyo Night styling
	if entry.Notes != "" {
		notesText := RenderNotes(truncateStyledLine(entry.Notes, maxWidth-9))
		// Indent notes to align with entry content (3 chars from entry style + 2 indent)
		lines = append(lines, "     "+notesText)
	}

	return strings.Join(lines, "\n")
}

// truncateStyledLine truncates a styled line to fit within maxWidth
func truncateStyledLine(line string, maxWidth int) string {
	return ansi.Truncate(line, max(0, maxWidth), "…")
}

// max returns the maximum of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
