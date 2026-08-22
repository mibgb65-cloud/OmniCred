package appsettings

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type config struct {
	DatabasePath string `json:"database_path"`
}

func ResolveDatabasePath(override string) (string, string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "data"
	} else {
		configDir = filepath.Join(configDir, "OmniCred")
	}
	configPath := filepath.Join(configDir, "config.json")
	defaultPath := filepath.Join(configDir, "omnicred.db")
	if strings.TrimSpace(override) != "" {
		path, err := filepath.Abs(strings.TrimSpace(override))
		return path, configPath, err
	}

	content, err := os.ReadFile(configPath)
	if errors.Is(err, os.ErrNotExist) {
		return defaultPath, configPath, nil
	}
	if err != nil {
		return "", "", fmt.Errorf("read app config: %w", err)
	}
	var saved config
	if err := json.Unmarshal(content, &saved); err != nil {
		return "", "", fmt.Errorf("decode app config: %w", err)
	}
	if strings.TrimSpace(saved.DatabasePath) == "" {
		return defaultPath, configPath, nil
	}
	path, err := filepath.Abs(saved.DatabasePath)
	if err != nil {
		return "", "", fmt.Errorf("resolve configured database path: %w", err)
	}
	return path, configPath, nil
}

func saveDatabasePath(configPath, databasePath string) error {
	if err := os.MkdirAll(filepath.Dir(configPath), 0o700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	content, err := json.MarshalIndent(config{DatabasePath: databasePath}, "", "  ")
	if err != nil {
		return fmt.Errorf("encode app config: %w", err)
	}
	if err := os.WriteFile(configPath, append(content, '\n'), 0o600); err != nil {
		return fmt.Errorf("write app config: %w", err)
	}
	return nil
}
