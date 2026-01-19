package tui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines the keybindings for the application.
type KeyMap struct {
	// Global navigation
	Quit key.Binding
	Help key.Binding
	Back key.Binding

	// List view navigation
	Up      key.Binding
	Down    key.Binding
	PrevDay key.Binding
	NextDay key.Binding
	Today   key.Binding

	// Time entry actions
	New    key.Binding
	Edit   key.Binding
	Delete key.Binding
	Start  key.Binding
	Stop   key.Binding

	// Selection and confirmation
	Select  key.Binding
	Confirm key.Binding
	Cancel  key.Binding

	// Text input
	Submit key.Binding
	Clear  key.Binding
}

// DefaultKeyMap returns the default keybindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		// Global navigation
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back/cancel"),
		),

		// List view navigation
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "move down"),
		),
		PrevDay: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "previous day"),
		),
		NextDay: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "next day"),
		),
		Today: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "jump to today"),
		),

		// Time entry actions
		New: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new entry"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit entry"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete entry"),
		),
		Start: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "start/restart timer"),
		),
		Stop: key.NewBinding(
			key.WithKeys("S"),
			key.WithHelp("S", "stop timer"),
		),

		// Selection and confirmation
		Select: key.NewBinding(
			key.WithKeys("enter", " "),
			key.WithHelp("enter/space", "select"),
		),
		Confirm: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "confirm"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "cancel"),
		),

		// Text input
		Submit: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "submit"),
		),
		Clear: key.NewBinding(
			key.WithKeys("ctrl+u"),
			key.WithHelp("ctrl+u", "clear"),
		),
	}
}

// ShortHelp returns the short help for the current context.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

// FullHelp returns the full help for all keybindings.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		// First column: Navigation
		{k.Up, k.Down, k.PrevDay, k.NextDay, k.Today},
		// Second column: Actions
		{k.New, k.Edit, k.Delete, k.Start, k.Stop},
		// Third column: General
		{k.Select, k.Help, k.Back, k.Quit},
	}
}

// ListViewHelp returns help specific to the list view.
func (k KeyMap) ListViewHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.PrevDay, k.NextDay, k.Today},
		{k.New, k.Edit, k.Delete, k.Start, k.Stop},
		{k.Help, k.Quit},
	}
}

// SelectionViewHelp returns help for selection views (project/task selection).
func (k KeyMap) SelectionViewHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Select},
		{k.Back, k.Help, k.Quit},
	}
}

// EditViewHelp returns help for the edit view.
func (k KeyMap) EditViewHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Submit, k.Clear},
		{k.Back, k.Help, k.Quit},
	}
}

// ConfirmViewHelp returns help for confirmation dialogs.
func (k KeyMap) ConfirmViewHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Confirm, k.Cancel},
		{k.Help, k.Quit},
	}
}
