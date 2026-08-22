package appsettings

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
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
	service := NewService(db, currentPath, configPath, "0.1.0", "127.0.0.1:8787", "https://github.com/mibgb65-cloud/OmniCred", false)
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

func TestCheckUpdateHandlesReleaseAndEmptyRepository(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/empty" {
			response.WriteHeader(http.StatusNotFound)
			return
		}
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{"tag_name":"v0.2.0","html_url":"https://example.com/release","published_at":"2026-08-22T00:00:00Z"}`))
	}))
	defer server.Close()
	service := NewService(nil, "current.db", "config.json", "0.1.0", "127.0.0.1:8787", "https://github.com/mibgb65-cloud/OmniCred", false)
	service.client = server.Client()
	service.releasesEndpoint = server.URL + "/latest"

	update, err := service.CheckUpdate(context.Background())
	if err != nil || !update.UpdateAvailable || update.LatestVersion != "v0.2.0" {
		t.Fatalf("CheckUpdate() = %#v, %v", update, err)
	}
	service.releasesEndpoint = server.URL + "/empty"
	empty, err := service.CheckUpdate(context.Background())
	if err != nil || empty.Status != "no_releases" || empty.UpdateAvailable {
		t.Fatalf("CheckUpdate(empty) = %#v, %v", empty, err)
	}
}
