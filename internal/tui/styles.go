package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Tokyo Night Color Palette
var (
	// Base colors
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
			Border(lipgloss.Border{Left: "▎"}).
			BorderForeground(accentColor).
			BorderTop(false).
			BorderRight(false).
			BorderBottom(false).
			BorderLeft(true).
			Background(selectedBg).
			PaddingLeft(2).
			Padding(1, 0, 1, 2).
			MarginBottom(0)

	UnselectedEntry = lipgloss.NewStyle().
			PaddingLeft(3).
			Padding(1, 0, 1, 3).
			MarginBottom(0)

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
	TotalLabel = lipgloss.NewStyle().
			Foreground(mutedText)

	TotalValue = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true)

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
)

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

// RenderDivider renders a horizontal divider with a default width.
func RenderDivider() string {
	return RenderDividerWidth(41)
}

// RenderDividerWidth renders a horizontal divider with the specified width.
func RenderDividerWidth(width int) string {
	if width < 1 {
		width = 1
	}
	return DividerStyle.Render(strings.Repeat("─", width))
}

// RenderKeybinding renders a key binding with styled key and description.
func RenderKeybinding(key, desc string) string {
	return KeyStyle.Render(key) + " " + KeyDescStyle.Render(desc)
}

// RenderEmptyState renders the empty state message.
func RenderEmptyState() string {
	return EmptyState.Render("No entries yet. Press 'n' to start tracking.")
}
