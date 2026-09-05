package appsettings

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"omnicred/internal/credential"
	"omnicred/internal/sqlite"
)

func TestMigrateStorageCopiesDatabaseAndSavesConfig(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	currentPath := filepath.Join(dir, "current.db")
	db, err := sqlite.Open(ctx, currentPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()
	now := time.Now().UTC()
	if _, err := sqlite.NewStore(db).Create(ctx, credential.Credential{
		Provider: "github", Account: "user@example.com", Password: "secret", CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatalf("seed credential: %v", err)
	}

	configPath := filepath.Join(dir, "config", "config.json")
	service := NewService(db, currentPath, configPath, "0.1.0", "127.0.0.1:8787", "https://github.com/mibgb65-cloud/OmniCred", false, nil)
	target := filepath.Join(dir, "moved", "vault.db")
	result, err := service.MigrateStorage(ctx, StorageInput{DatabasePath: target})
	if err != nil || !result.RestartRequired || result.DatabasePath != target {
		t.Fatalf("MigrateStorage() = %#v, %v", result, err)
	}

	copied, err := sqlite.Open(ctx, target)
	if err != nil {
		t.Fatalf("open copied database: %v", err)
	}
	defer copied.Close()
	items, err := sqlite.NewStore(copied).List(ctx, credential.Filter{})
	if err != nil || len(items) != 1 || items[0].Account != "user@example.com" {
		t.Fatalf("copied credentials = %#v, %v", items, err)
	}
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	var saved config
	if err := json.Unmarshal(content, &saved); err != nil || saved.DatabasePath != target {
		t.Fatalf("saved config = %#v, %v", saved, err)
	}
	if _, err := service.MigrateStorage(ctx, StorageInput{DatabasePath: target}); !errors.As(err, new(*ValidationError)) {
		t.Fatalf("existing target error = %v", err)
	}
}
