package appsettings

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestResolveDatabasePathPersistsInstallerSelection(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config", "config.json")
	defaultPath := filepath.Join(dir, "config", "omnicred.db")
	installedPath := filepath.Join(dir, "chosen", "omnicred.db")

	path, returnedConfigPath, err := resolveDatabasePath("", configPath, defaultPath, installedPath)
	if err != nil {
		t.Fatalf("resolve installer selection: %v", err)
	}
	if path != installedPath || returnedConfigPath != configPath {
		t.Fatalf("paths = %q, %q; want %q, %q", path, returnedConfigPath, installedPath, configPath)
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read saved config: %v", err)
	}
	var saved config
	if err := json.Unmarshal(content, &saved); err != nil {
		t.Fatalf("decode saved config: %v", err)
	}
	if saved.DatabasePath != installedPath {
		t.Fatalf("saved database path = %q; want %q", saved.DatabasePath, installedPath)
	}

	otherInstallerPath := filepath.Join(dir, "other", "omnicred.db")
	path, _, err = resolveDatabasePath("", configPath, defaultPath, otherInstallerPath)
	if err != nil {
		t.Fatalf("resolve existing config: %v", err)
	}
	if path != installedPath {
		t.Fatalf("existing config path = %q; want %q", path, installedPath)
	}
}
