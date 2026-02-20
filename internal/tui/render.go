package tui

import (
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/planetargon/argon-harvest-tui/internal/harvest"
)

// renderStyledListView renders the main list view with proper styling.
func (m Model) renderStyledListView() string {
	// Calculate dimensions
	width := 65 // Fixed width for consistent layout
	if m.width > 0 {
		width = min(m.width-2, 80) // Cap at 80 chars wide
	}

	// Format date navigation with Tokyo Night styling
	dateStr := m.currentDate.Format("Mon, Jan 2, 2006")
	dateNav := ArrowNavStyle.Render("◀ ") + DateStyle.Render(dateStr) + ArrowNavStyle.Render(" ▶")

	// Title bar with Tokyo Night styling
	titleText := "  " + TitleStyle.Render("🌾 Harvest Time Tracker")
	titleSuffix := dateNav + "  "
	spacerWidth := width - 2 - lipgloss.Width(titleText) - lipgloss.Width(titleSuffix)
	if spacerWidth < 1 {
		spacerWidth = 1
	}
	titleBar := titleText + strings.Repeat(" ", spacerWidth) + titleSuffix

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
		return m.wrapInStyledBox(strings.Join(content, "\n"), width)
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
		return m.wrapInStyledBox(strings.Join(content, "\n"), width)
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
		return m.wrapInStyledBox(strings.Join(content, "\n"), width)
	}

	// Render time entries with styles
	var entryLines []string
	for i, entry := range m.timeEntries {
		isSelected := i == m.selectedEntryIndex
		entryLines = append(entryLines, m.renderStyledTimeEntry(entry, isSelected, width-4))
	}

	// Build content
	contentLines := []string{titleBar, sectionHeader, divider}
	contentLines = append(contentLines, entryLines...)

	// Add status message with appropriate styling
	if m.statusMessage != "" {
		statusStyle := SuccessText
		msgLower := strings.ToLower(m.statusMessage)
		if strings.Contains(msgLower, "error") ||
			strings.Contains(msgLower, "failed") ||
			strings.Contains(msgLower, "cannot") {
			statusStyle = ErrorText
		} else if strings.Contains(msgLower, "locked") {
			statusStyle = WarningText
		}
		contentLines = append(contentLines, "", "  "+statusStyle.Render(m.statusMessage))
	}

	return m.wrapInStyledBox(strings.Join(contentLines, "\n"), width)
}

// wrapInStyledBox wraps content in a styled box border.
func (m Model) wrapInStyledBox(content string, width int) string {
	borderStyle := lipgloss.NewStyle().Foreground(borderColor)

	// Top border
	top := "┌" + strings.Repeat("─", width-2) + "┐"

	// Content lines with side borders
	lines := strings.Split(content, "\n")
	var boxedLines []string
	boxedLines = append(boxedLines, borderStyle.Render(top))

	for _, line := range lines {
		// Pad line to width
		lineWidth := lipgloss.Width(line)
		padded := line
		if lineWidth < width-2 {
			padded = line + strings.Repeat(" ", width-2-lineWidth)
		} else if lineWidth > width-2 {
			// Truncate if too long
			padded = truncateStyledLine(line, width-2)
		}
		boxedLines = append(boxedLines, borderStyle.Render("│")+padded+borderStyle.Render("│"))
	}

	// Footer with keybindings using Tokyo Night styling
	footerSeparator := borderStyle.Render("├" + strings.Repeat("─", width-2) + "┤")

	// Build styled keybindings
	keybindings := []string{
		RenderKeybinding("n", "new"),
		RenderKeybinding("e", "edit"),
		RenderKeybinding("s", "start/stop"),
		RenderKeybinding("d", "delete"),
		RenderKeybinding("?", "help"),
		RenderKeybinding("q", "quit"),
	}
	footerText := " " + strings.Join(keybindings, " ")
	footerWidth := lipgloss.Width(footerText)
	if footerWidth < width-2 {
		footerPadded := footerText + strings.Repeat(" ", width-2-footerWidth)
		footerText = footerPadded
	}

	// Bottom border
	bottom := borderStyle.Render("└" + strings.Repeat("─", width-2) + "┘")

	boxedLines = append(boxedLines, footerSeparator)
	boxedLines = append(boxedLines, borderStyle.Render("│")+footerText+borderStyle.Render("│"))
	boxedLines = append(boxedLines, bottom)

	return strings.Join(boxedLines, "\n")
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

	// Calculate padding for alignment
	pathWidth := lipgloss.Width(entryPath)
	durationWidth := lipgloss.Width(styledDuration)
	indicatorWidth := lipgloss.Width(indicator)
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
		notesText := RenderNotes(truncateString(entry.Notes, maxWidth-8))
		// Indent notes to align with entry content (3 chars from entry style + 2 indent)
		lines = append(lines, "     "+notesText)
	}

	return strings.Join(lines, "\n")
}

// truncateStyledLine truncates a styled line to fit within maxWidth
func truncateStyledLine(line string, maxWidth int) string {
	// This is a simple implementation that may need refinement for complex styled text
	if lipgloss.Width(line) <= maxWidth {
		return line
	}
	// For now, return the line as-is if it contains styling
	// A more sophisticated implementation would parse and preserve styles
	return line
}

// max returns the maximum of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
