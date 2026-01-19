package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/planetargon/argon-harvest-tui/internal/config"
	"github.com/planetargon/argon-harvest-tui/internal/harvest"
	"github.com/planetargon/argon-harvest-tui/internal/state"
	"github.com/planetargon/argon-harvest-tui/internal/tui"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		fmt.Printf("Configuration error: %v\n", err)
		os.Exit(1)
	}

	// Load application state
	appState, err := state.Load()
	if err != nil {
		fmt.Printf("Error loading state: %v\n", err)
		os.Exit(1)
	}

	// Initialize Harvest client
	harvestClient := harvest.NewClient(cfg.Harvest.AccountID, cfg.Harvest.AccessToken)

	// Validate authentication before starting TUI
	user, err := harvestClient.ValidateAuth()
	if err != nil {
		fmt.Printf("Authentication failed: %v\n", err)
		fmt.Println("Please check your Harvest credentials in ~/.config/harvest-tui/config.toml")
		os.Exit(1)
	}

	fmt.Printf("Welcome, %s!\n", user.FirstName+" "+user.LastName)
	fmt.Printf("Starting Harvest TUI...\n")

	// Initialize TUI model
	model := tui.NewModel(cfg, harvestClient, appState)

	// Create and run the program
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Save state on exit
	if err := appState.Save(); err != nil {
		fmt.Printf("Warning: Could not save state: %v\n", err)
	}
}
