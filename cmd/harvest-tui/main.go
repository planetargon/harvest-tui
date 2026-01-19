package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/planetargon/argon-harvest-tui/internal/config"
)

type model struct {
	config *config.Config
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	return fmt.Sprintf("Harvest TUI\nAccount: %s\n\nPress 'q' to quit.\n", m.config.Harvest.AccountID)
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	m := model{config: cfg}
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
