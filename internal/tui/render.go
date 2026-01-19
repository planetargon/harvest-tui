package tui

import (
	"fmt"
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

	// Format date navigation with styled arrows
	dateStr := m.currentDate.Format("Mon, Jan 2, 2006")
	dateNav := MutedText.Render("◀ ") + DateHeader.Render(dateStr) + MutedText.Render(" ▶")

	// Title bar with styled app title
	titleBar := lipgloss.JoinHorizontal(
		lipgloss.Left,
		"  "+AppTitle.Render("Harvest Time Tracker"),
		lipgloss.NewStyle().Width(width-26-lipgloss.Width(dateNav)).Render(""),
		dateNav,
		"  ",
	)

	// Calculate daily total with accent color
	totalHours := 0.0
	for _, entry := range m.timeEntries {
		totalHours += entry.Hours
	}
	totalStr := formatHoursSimple(totalHours)

	// Section header with styled total
	entriesText := "  Entries"
	totalLabel := "Total: "
	totalValue := AccentText.Render(totalStr)
	paddingWidth := width - len(entriesText) - len(totalLabel) - lipgloss.Width(totalValue) - 2
	if paddingWidth < 1 {
		paddingWidth = 1
	}
	sectionHeader := entriesText + strings.Repeat(" ", paddingWidth) + totalLabel + totalValue + "  "

	// Divider
	divider := "  " + strings.Repeat("─", width-4)

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
			"    " + MutedText.Render("No time entries for this date"),
			"    " + SecondaryText.Render("Press 'n' to create a new entry"),
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
	borderStyle := lipgloss.NewStyle().Foreground(BorderColor)

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

	// Footer with keybindings
	footerSeparator := borderStyle.Render("├" + strings.Repeat("─", width-2) + "┤")
	footerText := "n new  e edit  s start/stop  d delete  ? help  q quit"
	footer := MutedText.Render("  " + footerText)
	footerPadded := footer + strings.Repeat(" ", max(0, width-2-lipgloss.Width(footer)))

	// Bottom border
	bottom := borderStyle.Render("└" + strings.Repeat("─", width-2) + "┘")

	boxedLines = append(boxedLines, footerSeparator)
	boxedLines = append(boxedLines, borderStyle.Render("│")+footerPadded+borderStyle.Render("│"))
	boxedLines = append(boxedLines, bottom)

	return strings.Join(boxedLines, "\n")
}

// renderStyledTimeEntry renders a single time entry with proper styling.
func (m Model) renderStyledTimeEntry(entry harvest.TimeEntry, isSelected bool, maxWidth int) string {
	var lines []string

	// Determine text style based on state
	var lineStyle lipgloss.Style
	if entry.IsLocked {
		lineStyle = MutedText
	} else if isSelected {
		lineStyle = BaseText.Background(SelectedBgColor)
	} else {
		lineStyle = BaseText
	}

	// Status indicator
	statusIcon := "  "
	if isSelected {
		// Add selection indicator with accent color
		statusIcon = AccentText.Render("▶ ")
	}

	// Timer/lock icons
	timerIcon := ""
	if entry.IsRunning {
		timerIcon = " ⏱️"
	} else if entry.IsLocked {
		timerIcon = " 🔒"
	}

	// Format hours with appropriate style
	hoursText := formatHoursSimple(entry.Hours)
	var styledHours string
	if entry.IsRunning {
		styledHours = RunningIndicator.Render(hoursText)
	} else if entry.IsLocked {
		styledHours = MutedText.Render(hoursText)
	} else if isSelected {
		styledHours = lineStyle.Render(hoursText)
	} else {
		styledHours = BaseText.Render(hoursText)
	}

	// Build the entry line
	clientName := truncateString(entry.Client.Name, 20)
	projectName := truncateString(entry.Project.Name, 25)
	taskName := truncateString(entry.Task.Name, 20)

	entryText := fmt.Sprintf("%s → %s → %s", clientName, projectName, taskName)

	// Calculate padding for hours alignment
	textWidth := len(statusIcon) + len(entryText)
	hoursWidth := lipgloss.Width(styledHours)
	iconWidth := len(timerIcon)
	padding := maxWidth - textWidth - hoursWidth - iconWidth - 2
	if padding < 1 {
		padding = 1
	}

	// Build the complete first line
	firstLine := statusIcon + lineStyle.Render(entryText+strings.Repeat(" ", padding)) + styledHours + timerIcon

	// Add selection border if selected
	if isSelected {
		// Create left border for selection
		firstLine = lipgloss.NewStyle().
			BorderLeft(true).
			BorderStyle(lipgloss.ThickBorder()).
			BorderForeground(AccentColor).
			PaddingLeft(0).
			Render(firstLine)
	}

	lines = append(lines, firstLine)

	// Second line: Notes (if any)
	if entry.Notes != "" {
		notesStyle := SecondaryText
		if entry.IsLocked {
			notesStyle = MutedText
		}
		notesText := fmt.Sprintf("    \"%s\"", truncateString(entry.Notes, maxWidth-8))
		lines = append(lines, notesStyle.Render(notesText))
	}

	// Add blank line after entry
	lines = append(lines, "")

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
