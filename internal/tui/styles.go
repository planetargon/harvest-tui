package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Tokyo Night Color Palette
var (
	// Base colors
	bgColor     = lipgloss.Color("#1a1b26")
	cardBg      = lipgloss.Color("#24283b")
	selectedBg  = lipgloss.Color("#2f3549")
	borderColor = lipgloss.Color("#3b4261")

	// Accent
	accentColor = lipgloss.Color("#ff9e64") // warm orange

	// Semantic colors
	clientColor  = lipgloss.Color("#bb9af7") // purple
	projectColor = lipgloss.Color("#7aa2f7") // blue
	taskColor    = lipgloss.Color("#c0caf5") // light text

	greenColor  = lipgloss.Color("#9ece6a") // success/running
	yellowColor = lipgloss.Color("#e0af68") // warning/locked
	pinkColor   = lipgloss.Color("#f7768e") // error/delete

	// Text hierarchy
	primaryText = lipgloss.Color("#c0caf5")
	mutedText   = lipgloss.Color("#565f89")
	dimText     = lipgloss.Color("#3b4261")
)

// Tokyo Night Component Styles
var (
	// Base text styles
	BaseText = lipgloss.NewStyle().
			Foreground(primaryText)

	SecondaryText = lipgloss.NewStyle().
			Foreground(mutedText)

	MutedText = lipgloss.NewStyle().
			Foreground(mutedText)

	AccentText = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true)

	SuccessText = lipgloss.NewStyle().
			Foreground(greenColor)

	WarningText = lipgloss.NewStyle().
			Foreground(yellowColor)

	ErrorText = lipgloss.NewStyle().
			Foreground(pinkColor).
			Bold(true)

	// Entry styles
	SelectedEntry = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accentColor).
			Background(selectedBg).
			Padding(0, 1).
			MarginBottom(1)

	UnselectedEntry = lipgloss.NewStyle().
			PaddingLeft(3).
			MarginBottom(1)

	// Entry path styles
	ClientStyle  = lipgloss.NewStyle().Foreground(clientColor).Bold(true)
	ArrowStyle   = lipgloss.NewStyle().Foreground(dimText)
	ProjectStyle = lipgloss.NewStyle().Foreground(projectColor)
	TaskStyle    = lipgloss.NewStyle().Foreground(taskColor)

	// Duration styles
	DurationStyle = lipgloss.NewStyle().
			Foreground(primaryText).
			Width(6).
			Align(lipgloss.Right)

	RunningDurationStyle = DurationStyle.
				Foreground(accentColor).
				Bold(true)

	// Indicators
	RunningDot = lipgloss.NewStyle().
			Foreground(greenColor)

	LockedIcon = lipgloss.NewStyle().
			Foreground(yellowColor)

	LockedEntryStyle = lipgloss.NewStyle().
				Foreground(mutedText)

	// Notes style
	NotesStyle = lipgloss.NewStyle().
			Foreground(mutedText).
			Italic(true).
			PaddingLeft(2)

	// Header styles
	TitleStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true).
			PaddingRight(2)

	DateStyle = lipgloss.NewStyle().
			Foreground(primaryText).
			Bold(true)

	ArrowNavStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true)

	// Section header
	SectionHeaderStyle = lipgloss.NewStyle().
				Foreground(mutedText).
				MarginBottom(1).
				MarginLeft(1)

	// Summary bar styles
	SummaryBar = lipgloss.NewStyle().
			Background(cardBg).
			Padding(0, 2).
			MarginTop(1).
			MarginBottom(1)

	TotalLabel = lipgloss.NewStyle().
			Foreground(mutedText)

	TotalValue = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true)

	RunningBadge = lipgloss.NewStyle().
			Background(selectedBg).
			Foreground(mutedText).
			Padding(0, 1)

	// Keybinding styles
	KeyStyle = lipgloss.NewStyle().
			Background(cardBg).
			Foreground(accentColor).
			Bold(true).
			Padding(0, 1)

	KeyDescStyle = lipgloss.NewStyle().
			Foreground(mutedText).
			MarginRight(2)

	// Divider
	DividerStyle = lipgloss.NewStyle().
			Foreground(borderColor)

	// Empty state
	EmptyState = lipgloss.NewStyle().
			Foreground(mutedText).
			Italic(true).
			Padding(2, 0).
			Align(lipgloss.Center)

	// Delete confirmation
	ConfirmBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(pinkColor).
			Padding(1, 2).
			Width(50)

	ConfirmTitle = lipgloss.NewStyle().
			Foreground(pinkColor).
			Bold(true)

	// Help overlay
	HelpBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(accentColor).
		Padding(1, 2).
		Width(40)

	HelpTitleStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true).
			MarginBottom(1)

	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Width(12)

	HelpDescStyle = lipgloss.NewStyle().
			Foreground(primaryText)

	// Input styles
	FocusedInput = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(accentColor)

	UnfocusedInput = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(borderColor)
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
	HighlightText lipgloss.Style

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

// DefaultStyles returns the default styling configuration using the Tokyo Night theme.
func DefaultStyles() Styles {
	return Styles{
		// Base styles
		App:     BaseText,
		Header:  BaseText.Padding(0, 1),
		Footer:  MutedText.Padding(0, 1),
		Content: BaseText.Padding(0, 1),

		// Navigation styles
		Title:    TitleStyle,
		Subtitle: SecondaryText,

		// List styles
		ListItem: BaseText,
		SelectedItem: BaseText.
			Inherit(SelectedEntry),
		RunningItem: BaseText.
			Foreground(accentColor),
		LockedItem: LockedEntryStyle,

		// Text styles
		PrimaryText:   BaseText,
		SecondaryText: SecondaryText,
		MutedText:     MutedText,
		ErrorText:     ErrorText,
		SuccessText:   SuccessText,
		WarningText:   WarningText,
		HighlightText: AccentText,

		// Component styles
		Card: BaseText.
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Background(cardBg),

		SelectedCard: SelectedEntry,

		Input: UnfocusedInput.
			Padding(0, 1),

		Button: BaseText.
			Padding(0, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Background(cardBg),

		ActiveButton: BaseText.
			Padding(0, 2).
			Background(accentColor).
			Foreground(primaryText).
			Bold(true),

		// Status indicators
		StatusBar:        SummaryBar,
		RunningIndicator: RunningDot,
		LockedIndicator:  LockedIcon,

		// Help styles
		HelpKey:   HelpKeyStyle,
		HelpValue: HelpDescStyle,
		HelpTitle: HelpTitleStyle,

		// Dialog styles
		Dialog:       ConfirmBox,
		DialogTitle:  ConfirmTitle,
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
		return RunningDot.Render("●") + " ", s.RunningIndicator
	}
	if isLocked {
		return LockedIcon.Render("🔒") + " ", s.LockedIndicator
	}
	return "", s.PrimaryText
}

// FormatDuration formats a duration with appropriate styling.
func (s Styles) FormatDuration(hours float64, isRunning bool) string {
	formatted := formatHours(hours)
	if isRunning {
		return RunningDurationStyle.Render(formatted)
	}
	return DurationStyle.Render(formatted)
}

// formatHours formats hours as HH:MM string.
func formatHours(hours float64) string {
	h := int(hours)
	m := int((hours - float64(h)) * 60)
	return fmt.Sprintf("%d:%02d", h, m)
}

// RenderEntryPath renders the client → project → task path with proper styling.
func RenderEntryPath(client, project, task string) string {
	return ClientStyle.Render(client) +
		ArrowStyle.Render(" → ") +
		ProjectStyle.Render(project) +
		ArrowStyle.Render(" → ") +
		TaskStyle.Render(task)
}

// RenderNotes renders notes with proper styling and quotes.
func RenderNotes(notes string) string {
	if notes == "" {
		return ""
	}
	return NotesStyle.Render(`"` + notes + `"`)
}

// RenderDivider renders a horizontal divider.
func RenderDivider() string {
	return DividerStyle.Render("─────────────────────────────────────────")
}

// RenderKeybinding renders a key binding with styled key and description.
func RenderKeybinding(key, desc string) string {
	return KeyStyle.Render(key) + " " + KeyDescStyle.Render(desc)
}

// RenderEmptyState renders the empty state message.
func RenderEmptyState() string {
	return EmptyState.Render("No entries yet. Press 'n' to start tracking.")
}
