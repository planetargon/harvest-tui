package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const SetupInstructionsURL = "https://github.com/planetargon/harvest-tui?tab=readme-ov-file#getting-harvest-api-credentials"

type Config struct {
	Harvest HarvestConfig `toml:"harvest"`
}

type HarvestConfig struct {
	AccountID   string `toml:"account_id"`
	AccessToken string `toml:"access_token"`
}

func Load() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, fmt.Errorf("could not determine config path: %w", err)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("could not load config file. Create %s with your Harvest credentials.\n\nTo get started, set up your Harvest API credentials:\n%s", configPath, SetupInstructionsURL)
	}

	var config Config
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		return nil, fmt.Errorf("could not parse config file: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &config, nil
}

func (c *Config) Validate() error {
	if c.Harvest.AccountID == "" {
		return fmt.Errorf("account_id is required.\n\nTo get started, set up your Harvest API credentials:\n%s", SetupInstructionsURL)
	}
	if c.Harvest.AccessToken == "" {
		return fmt.Errorf("access_token is required.\n\nTo get started, set up your Harvest API credentials:\n%s", SetupInstructionsURL)
	}
	return nil
}

func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", "harvest-tui", "config.toml"), nil
}
