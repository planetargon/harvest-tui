package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfig(t *testing.T) {
	t.Run("given a valid config file when loaded then returns config with correct values", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "config.toml")

		configContent := `[harvest]
account_id = "12345"
access_token = "abc123def456"
`
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		if err != nil {
			t.Fatal(err)
		}

		originalHome := os.Getenv("HOME")
		t.Cleanup(func() { os.Setenv("HOME", originalHome) })

		os.Setenv("HOME", tempDir)

		configDir := filepath.Join(tempDir, ".config", "harvest-tui")
		err = os.MkdirAll(configDir, 0755)
		if err != nil {
			t.Fatal(err)
		}

		finalConfigPath := filepath.Join(configDir, "config.toml")
		err = os.WriteFile(finalConfigPath, []byte(configContent), 0644)
		if err != nil {
			t.Fatal(err)
		}

		config, err := Load()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if config.Harvest.AccountID != "12345" {
			t.Errorf("expected account_id 12345, got %s", config.Harvest.AccountID)
		}

		if config.Harvest.AccessToken != "abc123def456" {
			t.Errorf("expected access_token abc123def456, got %s", config.Harvest.AccessToken)
		}
	})

	t.Run("given a config with missing account_id when validated then returns error", func(t *testing.T) {
		config := &Config{
			Harvest: HarvestConfig{
				AccountID:   "",
				AccessToken: "abc123def456",
			},
		}

		err := config.Validate()
		if err == nil {
			t.Fatal("expected error for missing account_id")
		}

		if err.Error() != "account_id is required" {
			t.Errorf("expected 'account_id is required', got %s", err.Error())
		}
	})

	t.Run("given a config with missing access_token when validated then returns error", func(t *testing.T) {
		config := &Config{
			Harvest: HarvestConfig{
				AccountID:   "12345",
				AccessToken: "",
			},
		}

		err := config.Validate()
		if err == nil {
			t.Fatal("expected error for missing access_token")
		}

		if err.Error() != "access_token is required" {
			t.Errorf("expected 'access_token is required', got %s", err.Error())
		}
	})

	t.Run("given a valid config when validated then returns no error", func(t *testing.T) {
		config := &Config{
			Harvest: HarvestConfig{
				AccountID:   "12345",
				AccessToken: "abc123def456",
			},
		}

		err := config.Validate()
		if err != nil {
			t.Errorf("expected no error for valid config, got %v", err)
		}
	})

	t.Run("given missing config file when loaded then returns helpful error message", func(t *testing.T) {
		tempDir := t.TempDir()
		originalHome := os.Getenv("HOME")
		t.Cleanup(func() { os.Setenv("HOME", originalHome) })

		os.Setenv("HOME", tempDir)

		_, err := Load()
		if err == nil {
			t.Fatal("expected error for missing config file")
		}

		expectedPath := filepath.Join(tempDir, ".config", "harvest-tui", "config.toml")
		expectedMsg := "could not load config file. Create " + expectedPath + " with your Harvest credentials"
		if err.Error() != expectedMsg {
			t.Errorf("expected '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("given malformed config file when loaded then returns parse error", func(t *testing.T) {
		tempDir := t.TempDir()
		originalHome := os.Getenv("HOME")
		t.Cleanup(func() { os.Setenv("HOME", originalHome) })

		os.Setenv("HOME", tempDir)

		configDir := filepath.Join(tempDir, ".config", "harvest-tui")
		err := os.MkdirAll(configDir, 0755)
		if err != nil {
			t.Fatal(err)
		}

		configPath := filepath.Join(configDir, "config.toml")
		malformedContent := `[harvest
account_id = "12345"
access_token = "abc123"
`
		err = os.WriteFile(configPath, []byte(malformedContent), 0644)
		if err != nil {
			t.Fatal(err)
		}

		_, err = Load()
		if err == nil {
			t.Fatal("expected error for malformed config file")
		}

		if err.Error()[:28] != "could not parse config file:" {
			t.Errorf("expected parse error, got '%s'", err.Error())
		}
	})

	t.Run("given config file missing harvest section when loaded then returns validation error", func(t *testing.T) {
		tempDir := t.TempDir()
		originalHome := os.Getenv("HOME")
		t.Cleanup(func() { os.Setenv("HOME", originalHome) })

		os.Setenv("HOME", tempDir)

		configDir := filepath.Join(tempDir, ".config", "harvest-tui")
		err := os.MkdirAll(configDir, 0755)
		if err != nil {
			t.Fatal(err)
		}

		configPath := filepath.Join(configDir, "config.toml")
		emptyContent := `[other]
setting = "value"
`
		err = os.WriteFile(configPath, []byte(emptyContent), 0644)
		if err != nil {
			t.Fatal(err)
		}

		_, err = Load()
		if err == nil {
			t.Fatal("expected error for config missing harvest fields")
		}

		if err.Error() != "invalid config: account_id is required" {
			t.Errorf("expected account_id required error, got '%s'", err.Error())
		}
	})
}
