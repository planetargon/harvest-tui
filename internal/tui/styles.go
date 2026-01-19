package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Color palette - Harvest branded colors with adaptive support
var (
	// Primary Colors
	AccentColor = lipgloss.AdaptiveColor{Light: "#E85D00", Dark: "#FF6600"} // Harvest orange
	MutedColor  = lipgloss.AdaptiveColor{Light: "#666666", Dark: "#666666"}
	SuccessColor = lipgloss.AdaptiveColor{Light: "#16A34A", Dark: "#22C55E"}
	WarningColor = lipgloss.AdaptiveColor{Light: "#CA8A04", Dark: "#EAB308"}
	ErrorColor   = lipgloss.AdaptiveColor{Light: "#DC2626", Dark: "#EF4444"}

	// Text Colors
	PrimaryTextColor   = lipgloss.AdaptiveColor{Light: "#18181B", Dark: "#FAFAFA"}
	SecondaryTextColor = lipgloss.AdaptiveColor{Light: "#52525B", Dark: "#A1A1AA"}
	InverseTextColor   = lipgloss.AdaptiveColor{Light: "#FAFAFA", Dark: "#18181B"}

	// Background Colors
	SelectedBgColor = lipgloss.AdaptiveColor{Light: "#E4E4E7", Dark: "#27272A"}
	BorderColor     = lipgloss.AdaptiveColor{Light: "#D4D4D8", Dark: "#3F3F46"}
)

// Base styles - composable building blocks
var (
	BaseText = lipgloss.NewStyle().
		Foreground(PrimaryTextColor)

	SecondaryText = lipgloss.NewStyle().
		Foreground(SecondaryTextColor)

	MutedText = lipgloss.NewStyle().
		Foreground(MutedColor)

	AccentText = lipgloss.NewStyle().
		Foreground(AccentColor).
		Bold(true)

	SuccessText = lipgloss.NewStyle().
		Foreground(SuccessColor)

	WarningText = lipgloss.NewStyle().
		Foreground(WarningColor)

	ErrorText = lipgloss.NewStyle().
		Foreground(ErrorColor).
		Bold(true)

	// Selected item with accent left border
	SelectedItem = lipgloss.NewStyle().
		Background(SelectedBgColor).
		BorderLeft(true).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(AccentColor).
		PaddingLeft(1)

	// Running timer indicator
	RunningIndicator = lipgloss.NewStyle().
		Foreground(AccentColor).
		Bold(true)

	// Locked indicator
	LockedIndicator = lipgloss.NewStyle().
		Foreground(WarningColor)

	// Header styles
	AppTitle = lipgloss.NewStyle().
		Foreground(AccentColor).
		Bold(true)

	DateHeader = lipgloss.NewStyle().
		Bold(true)

	// Section headers
	SectionHeader = lipgloss.NewStyle().
		Foreground(MutedColor).
		Italic(true)

	// Dialog styles
	DialogBorder = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(WarningColor).
		Padding(1, 2)

	DialogTitle = lipgloss.NewStyle().
		Foreground(WarningColor).
		Bold(true)

	// Input styles
	FocusedInput = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(AccentColor)

	UnfocusedInput = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(BorderColor)
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

// DefaultStyles returns the default styling configuration using the Harvest color palette.
func DefaultStyles() Styles {
	return Styles{
		// Base styles
		App:     BaseText,
		Header:  BaseText.Padding(0, 1),
		Footer:  MutedText.Padding(0, 1),
		Content: BaseText.Padding(0, 1),

		// Navigation styles
		Title:    AppTitle,
		Subtitle: SecondaryText,

		// List styles
		ListItem: BaseText,
		SelectedItem: BaseText.
			Inherit(SelectedItem),
		RunningItem: BaseText.
			Foreground(AccentColor),
		LockedItem: MutedText,

		// Text styles
		PrimaryText:   BaseText,
		SecondaryText: SecondaryText,
		MutedText:     MutedText,
		ErrorText:     ErrorText,
		SuccessText:   SuccessText,
		WarningText:   WarningText,

		// Component styles
		Card: BaseText.
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(BorderColor),

		SelectedCard: BaseText.
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(AccentColor),

		Input: UnfocusedInput.
			Padding(0, 1),

		Button: BaseText.
			Padding(0, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(BorderColor),

		ActiveButton: BaseText.
			Padding(0, 2).
			Background(AccentColor).
			Foreground(InverseTextColor).
			Bold(true),

		// Status indicators
		StatusBar:        MutedText.Align(lipgloss.Center),
		RunningIndicator: RunningIndicator,
		LockedIndicator:  LockedIndicator,

		// Help styles
		HelpKey:   AccentText,
		HelpValue: SecondaryText,
		HelpTitle: BaseText.Bold(true),

		// Dialog styles
		Dialog:       DialogBorder,
		DialogTitle:  DialogTitle,
		DialogButton: MutedText.Padding(0, 2),

		// Spinner and loading
		Spinner: AccentText,
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
	return fmt.Sprintf("%d:%02d", h, m)
}