package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette
const (
	// Primary colors
	primaryColor    = "#00D4AA" // Harvest green
	accentColor     = "#FF6B35" // Orange accent
	backgroundColor = "#1A1A1A" // Dark background
	surfaceColor    = "#2D2D2D" // Surface/card background

	// Text colors
	primaryText   = "#FFFFFF" // White text
	secondaryText = "#A0A0A0" // Gray text
	mutedText     = "#606060" // Muted gray text

	// Status colors
	successColor = "#00FF88" // Success green
	errorColor   = "#FF4757" // Error red
	warningColor = "#FFA502" // Warning orange
	runningColor = "#3742FA" // Running timer blue
	lockedColor  = "#747D8C" // Locked entry gray
)

// Styles defines all the styling for the TUI components.
type Styles struct {
	// Base styles
	App     lipgloss.Style
	Header  lipgloss.Style
	Footer  lipgloss.Style
	Content lipgloss.Style

	// Navigation styles
	Title    lipgloss.Style
	Subtitle lipgloss.Style

	// List styles
	ListItem     lipgloss.Style
	SelectedItem lipgloss.Style
	RunningItem  lipgloss.Style
	LockedItem   lipgloss.Style

	// Text styles
	PrimaryText   lipgloss.Style
	SecondaryText lipgloss.Style
	MutedText     lipgloss.Style
	ErrorText     lipgloss.Style
	SuccessText   lipgloss.Style
	WarningText   lipgloss.Style

	// Component styles
	Card         lipgloss.Style
	SelectedCard lipgloss.Style
	Input        lipgloss.Style
	Button       lipgloss.Style
	ActiveButton lipgloss.Style

	// Status indicators
	StatusBar        lipgloss.Style
	RunningIndicator lipgloss.Style
	LockedIndicator  lipgloss.Style

	// Help styles
	HelpKey   lipgloss.Style
	HelpValue lipgloss.Style
	HelpTitle lipgloss.Style

	// Dialog styles
	Dialog       lipgloss.Style
	DialogTitle  lipgloss.Style
	DialogButton lipgloss.Style

	// Spinner and loading
	Spinner lipgloss.Style
}

// DefaultStyles returns the default styling configuration.
func DefaultStyles() Styles {
	return Styles{
		// Base styles
		App: lipgloss.NewStyle().
			Background(lipgloss.Color(backgroundColor)).
			Foreground(lipgloss.Color(primaryText)),

		Header: lipgloss.NewStyle().
			Background(lipgloss.Color(surfaceColor)).
			Foreground(lipgloss.Color(primaryText)).
			Padding(1, 2).
			Bold(true),

		Footer: lipgloss.NewStyle().
			Background(lipgloss.Color(surfaceColor)).
			Foreground(lipgloss.Color(secondaryText)).
			Padding(0, 2),

		Content: lipgloss.NewStyle().
			Padding(1, 2),

		// Navigation styles
		Title: lipgloss.NewStyle().
			Foreground(lipgloss.Color(primaryColor)).
			Bold(true).
			MarginBottom(1),

		Subtitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(secondaryText)).
			MarginBottom(1),

		// List styles
		ListItem: lipgloss.NewStyle().
			Padding(0, 2).
			MarginBottom(1).
			Background(lipgloss.Color(surfaceColor)).
			Foreground(lipgloss.Color(primaryText)),

		SelectedItem: lipgloss.NewStyle().
			Padding(0, 2).
			MarginBottom(1).
			Background(lipgloss.Color(primaryColor)).
			Foreground(lipgloss.Color(backgroundColor)).
			Bold(true),

		RunningItem: lipgloss.NewStyle().
			Padding(0, 2).
			MarginBottom(1).
			Background(lipgloss.Color(runningColor)).
			Foreground(lipgloss.Color(primaryText)).
			Bold(true),

		LockedItem: lipgloss.NewStyle().
			Padding(0, 2).
			MarginBottom(1).
			Background(lipgloss.Color(lockedColor)).
			Foreground(lipgloss.Color(primaryText)),

		// Text styles
		PrimaryText: lipgloss.NewStyle().
			Foreground(lipgloss.Color(primaryText)),

		SecondaryText: lipgloss.NewStyle().
			Foreground(lipgloss.Color(secondaryText)),

		MutedText: lipgloss.NewStyle().
			Foreground(lipgloss.Color(mutedText)),

		ErrorText: lipgloss.NewStyle().
			Foreground(lipgloss.Color(errorColor)).
			Bold(true),

		SuccessText: lipgloss.NewStyle().
			Foreground(lipgloss.Color(successColor)).
			Bold(true),

		WarningText: lipgloss.NewStyle().
			Foreground(lipgloss.Color(warningColor)).
			Bold(true),

		// Component styles
		Card: lipgloss.NewStyle().
			Background(lipgloss.Color(surfaceColor)).
			Padding(1, 2).
			Margin(0, 0, 1, 0).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(mutedText)),

		SelectedCard: lipgloss.NewStyle().
			Background(lipgloss.Color(surfaceColor)).
			Padding(1, 2).
			Margin(0, 0, 1, 0).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(primaryColor)).
			BorderStyle(lipgloss.ThickBorder()),

		Input: lipgloss.NewStyle().
			Padding(0, 1).
			Background(lipgloss.Color(surfaceColor)).
			Foreground(lipgloss.Color(primaryText)).
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(mutedText)),

		Button: lipgloss.NewStyle().
			Padding(0, 2).
			Background(lipgloss.Color(surfaceColor)).
			Foreground(lipgloss.Color(primaryText)).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(mutedText)),

		ActiveButton: lipgloss.NewStyle().
			Padding(0, 2).
			Background(lipgloss.Color(primaryColor)).
			Foreground(lipgloss.Color(backgroundColor)).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(primaryColor)).
			Bold(true),

		// Status indicators
		StatusBar: lipgloss.NewStyle().
			Background(lipgloss.Color(surfaceColor)).
			Foreground(lipgloss.Color(secondaryText)).
			Padding(0, 1).
			Align(lipgloss.Center),

		RunningIndicator: lipgloss.NewStyle().
			Foreground(lipgloss.Color(runningColor)).
			Bold(true),

		LockedIndicator: lipgloss.NewStyle().
			Foreground(lipgloss.Color(lockedColor)).
			Bold(true),

		// Help styles
		HelpKey: lipgloss.NewStyle().
			Foreground(lipgloss.Color(primaryColor)).
			Bold(true),

		HelpValue: lipgloss.NewStyle().
			Foreground(lipgloss.Color(secondaryText)),

		HelpTitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(primaryText)).
			Bold(true).
			MarginBottom(1),

		// Dialog styles
		Dialog: lipgloss.NewStyle().
			Background(lipgloss.Color(surfaceColor)).
			Padding(2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(primaryColor)),

		DialogTitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(primaryColor)).
			Bold(true).
			MarginBottom(1),

		DialogButton: lipgloss.NewStyle().
			Padding(0, 2).
			Margin(0, 1).
			Background(lipgloss.Color(surfaceColor)).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(mutedText)),

		// Spinner and loading
		Spinner: lipgloss.NewStyle().
			Foreground(lipgloss.Color(primaryColor)),
	}
}

// TimeEntryStyle returns the appropriate style for a time entry based on its state.
func (s Styles) TimeEntryStyle(isSelected, isRunning, isLocked bool) lipgloss.Style {
	if isSelected {
		return s.SelectedItem
	}
	if isRunning {
		return s.RunningItem
	}
	if isLocked {
		return s.LockedItem
	}
	return s.ListItem
}

// StatusIndicator returns the appropriate status indicator text and style.
func (s Styles) StatusIndicator(isRunning, isLocked bool) (string, lipgloss.Style) {
	if isRunning {
		return "⏱️ ", s.RunningIndicator
	}
	if isLocked {
		return "🔒 ", s.LockedIndicator
	}
	return "", s.PrimaryText
}

// FormatDuration formats a duration with appropriate styling.
func (s Styles) FormatDuration(hours float64, isRunning bool) string {
	formatted := formatHours(hours)
	if isRunning {
		return s.RunningIndicator.Render(formatted)
	}
	return s.PrimaryText.Render(formatted)
}

// formatHours formats hours as HH:MM string.
func formatHours(hours float64) string {
	h := int(hours)
	m := int((hours - float64(h)) * 60)
	return lipgloss.NewStyle().Bold(true).Render(
		lipgloss.JoinHorizontal(lipgloss.Left,
			lipgloss.NewStyle().Render(string(rune('0'+h/10))),
			string(rune('0'+h%10)),
			":",
			string(rune('0'+m/10)),
			string(rune('0'+m%10)),
		))
}
