package tui

import (
	"strings"

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
	titleBar := lipgloss.JoinHorizontal(
		lipgloss.Left,
		"  "+TitleStyle.Render("🌾 Harvest Time Tracker"),
		lipgloss.NewStyle().Width(width-30-lipgloss.Width(dateNav)).Render(""),
		dateNav,
		"  ",
	)

	// Calculate daily total with accent color
	totalHours := 0.0
	for _, entry := range m.timeEntries {
		totalHours += entry.Hours
	}
	totalStr := formatHoursSimple(totalHours)

	// Section header with Tokyo Night styling
	entriesText := SectionHeaderStyle.Render("Today's Entries")
	totalLabelText := TotalLabel.Render("Total: ")
	totalValue := TotalValue.Render(totalStr)
	paddingWidth := width - lipgloss.Width(entriesText) - lipgloss.Width(totalLabelText) - lipgloss.Width(totalValue) - 4
	if paddingWidth < 1 {
		paddingWidth = 1
	}
	sectionHeader := "  " + entriesText + strings.Repeat(" ", paddingWidth) + totalLabelText + totalValue + "  "

	// Divider with Tokyo Night styling
	divider := "  " + RenderDivider()

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
	footerText := "  " + strings.Join(keybindings, "  ")
	footerPadded := footerText + strings.Repeat(" ", max(0, width-2-lipgloss.Width(footerText)))

	// Bottom border
	bottom := borderStyle.Render("└" + strings.Repeat("─", width-2) + "┘")

	boxedLines = append(boxedLines, footerSeparator)
	boxedLines = append(boxedLines, borderStyle.Render("│")+footerPadded+borderStyle.Render("│"))
	boxedLines = append(boxedLines, bottom)

	return strings.Join(boxedLines, "\n")
}

// renderStyledTimeEntry renders a single time entry with Tokyo Night styling.
func (m Model) renderStyledTimeEntry(entry harvest.TimeEntry, isSelected bool, maxWidth int) string {
	var lines []string

	// Build the entry path with Tokyo Night colors
	clientName := truncateString(entry.Client.Name, 20)
	projectName := truncateString(entry.Project.Name, 25)
	taskName := truncateString(entry.Task.Name, 20)
	entryPath := RenderEntryPath(clientName, projectName, taskName)

	// Format duration with Tokyo Night styling
	styledDuration := ""
	if entry.IsRunning {
		styledDuration = RunningDurationStyle.Render(formatHoursSimple(entry.Hours))
	} else if entry.IsLocked {
		styledDuration = DurationStyle.Copy().Foreground(mutedText).Render(formatHoursSimple(entry.Hours))
	} else {
		styledDuration = DurationStyle.Render(formatHoursSimple(entry.Hours))
	}

	// Add running indicator
	indicator := ""
	if entry.IsRunning {
		indicator = " " + RunningDot.Render("●")
	} else if entry.IsLocked {
		indicator = " " + LockedIcon.Render("🔒")
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
		// Selected entry with rounded border card
		entryContent := entryPath + strings.Repeat(" ", padding) + styledDuration + indicator
		entryLine = SelectedEntry.Render(entryContent)
	} else {
		// Unselected entry with left padding to align with border
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
		if isSelected {
			// Add padding to align with card content
			lines = append(lines, "   "+notesText)
		} else {
			// Add padding to align with unselected entries
			lines = append(lines, "      "+notesText)
		}
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
