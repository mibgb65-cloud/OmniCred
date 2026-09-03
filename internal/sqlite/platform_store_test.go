package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"omnicred/internal/credential"
	"omnicred/internal/platform"
)

func TestPlatformStoreLifecycleAndCredentialRename(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "omnicred.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()
	store := NewStore(db)
	now := time.Date(2026, 8, 22, 4, 0, 0, 0, time.UTC)

	created, err := store.Create(ctx, credential.Credential{
		Provider: "github", Account: "user@example.com", Password: "secret", CreatedAt: now, UpdatedAt: now,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	items, err := store.ListPlatforms(ctx)
	if err != nil {
		t.Fatalf("ListPlatforms() error = %v", err)
	}
	github := findPlatform(items, "github")
	if github == nil || github.CredentialCount != 1 {
		t.Fatalf("github platform = %#v", github)
	}

	renamed, err := store.UpdatePlatform(ctx, github.ID, "github work", now.Add(time.Hour))
	if err != nil || renamed.CredentialCount != 1 {
		t.Fatalf("UpdatePlatform() = %#v, %v", renamed, err)
	}
	got, err := store.Get(ctx, created.ID)
	if err != nil || got.Provider != "github work" {
		t.Fatalf("credential after platform rename = %#v, %v", got, err)
	}
	if err := store.DeletePlatform(ctx, github.ID); !errors.Is(err, platform.ErrInUse) {
		t.Fatalf("DeletePlatform(in use) error = %v", err)
	}

	custom, err := store.CreatePlatform(ctx, platform.Platform{Name: "unused", CreatedAt: now, UpdatedAt: now})
	if err != nil {
		t.Fatalf("CreatePlatform() error = %v", err)
	}
	if _, err := store.CreatePlatform(ctx, platform.Platform{Name: "UNUSED", CreatedAt: now, UpdatedAt: now}); !errors.Is(err, platform.ErrAlreadyExists) {
		t.Fatalf("CreatePlatform(duplicate) error = %v", err)
	}
	if err := store.DeletePlatform(ctx, custom.ID); err != nil {
		t.Fatalf("DeletePlatform() error = %v", err)
	}
}

func TestPlatformMigrationSeedsDefaultsAndSetsVersion(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "omnicred.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	var version int
	if err := db.QueryRowContext(ctx, "PRAGMA user_version").Scan(&version); err != nil {
		t.Fatalf("read version: %v", err)
	}
	items, err := NewStore(db).ListPlatforms(ctx)
	if err != nil || version != 5 || findPlatform(items, "github") == nil || findPlatform(items, "google") == nil {
		t.Fatalf("migration version = %d, platforms = %#v, error = %v", version, items, err)
	}
}

func TestPlatformMigrationImportsExistingCredentialProviders(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "legacy.db")
	legacy, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open legacy database: %v", err)
	}
	script, err := migrationFiles.ReadFile("migrations/001_create_credentials.sql")
	if err != nil {
		t.Fatalf("read legacy migration: %v", err)
	}
	if _, err := legacy.ExecContext(ctx, string(script)); err != nil {
		t.Fatalf("create legacy schema: %v", err)
	}
	now := time.Date(2026, 8, 21, 2, 0, 0, 0, time.UTC)
	if _, err := legacy.ExecContext(ctx, `
		INSERT INTO credentials (provider, account, username, password, created_at, updated_at)
		VALUES ('custom service', 'user@example.com', '', 'secret', ?, ?)`, formatTime(now), formatTime(now)); err != nil {
		t.Fatalf("seed legacy credential: %v", err)
	}
	if err := legacy.Close(); err != nil {
		t.Fatalf("close legacy database: %v", err)
	}

	db, err := Open(ctx, path)
	if err != nil {
		t.Fatalf("upgrade legacy database: %v", err)
	}
	defer db.Close()
	items, err := NewStore(db).ListPlatforms(ctx)
	custom := findPlatform(items, "custom service")
	if err != nil || custom == nil || custom.CredentialCount != 1 {
		t.Fatalf("imported platform = %#v, error = %v", custom, err)
	}
	credentialItem, err := NewStore(db).Get(ctx, 1)
	if err != nil || credentialItem.TOTPSecret != "" || len(credentialItem.RecoveryCodes) != 0 {
		t.Fatalf("legacy credential security fields = %q / %#v, error = %v", credentialItem.TOTPSecret, credentialItem.RecoveryCodes, err)
	}
}

func findPlatform(items []platform.Platform, name string) *platform.Platform {
	for index := range items {
		if items[index].Name == name {
			return &items[index]
		}
	}
	return nil
}
