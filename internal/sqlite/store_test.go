package sqlite

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"omnicred/internal/credential"
)

func TestStoreCRUDAndPersistence(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "omnicred.db")
	db, err := Open(ctx, path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	store := NewStore(db)
	createdAt := time.Date(2026, 8, 21, 2, 0, 0, 0, time.UTC)

	created, err := store.Create(ctx, credential.Credential{
		Provider: "github", Account: "user@example.com", Username: "octocat", Password: "secret",
		CreatedAt: createdAt, UpdatedAt: createdAt,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.ID <= 0 {
		t.Fatalf("Create() ID = %d", created.ID)
	}

	got, err := store.Get(ctx, created.ID)
	if err != nil || got.Account != created.Account || !got.CreatedAt.Equal(createdAt) {
		t.Fatalf("Get() = %#v, %v", got, err)
	}

	got.Username = "new-name"
	got.Password = "new-secret"
	got.UpdatedAt = createdAt.Add(time.Hour)
	updated, err := store.Update(ctx, got)
	if err != nil || updated.Username != "new-name" {
		t.Fatalf("Update() = %#v, %v", updated, err)
	}

	if err := db.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	db, err = Open(ctx, path)
	if err != nil {
		t.Fatalf("reopen database: %v", err)
	}
	defer db.Close()
	store = NewStore(db)
	got, err = store.Get(ctx, created.ID)
	if err != nil || got.Password != "new-secret" {
		t.Fatalf("Get() after reopen = %#v, %v", got, err)
	}

	if err := store.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := store.Get(ctx, created.ID); !errors.Is(err, credential.ErrNotFound) {
		t.Fatalf("Get() after delete error = %v", err)
	}
}

func TestStoreListFilters(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "omnicred.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()
	store := NewStore(db)
	now := time.Now().UTC()

	items := []credential.Credential{
		{Provider: "github", Account: "one@example.com", Username: "octocat", Password: "a", CreatedAt: now, UpdatedAt: now},
		{Provider: "google", Account: "two@example.com", Username: "person", Password: "b", CreatedAt: now, UpdatedAt: now},
	}
	for _, item := range items {
		if _, err := store.Create(ctx, item); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	github, err := store.List(ctx, credential.Filter{Provider: "github"})
	if err != nil || len(github) != 1 || github[0].Provider != "github" {
		t.Fatalf("List(provider) = %#v, %v", github, err)
	}
	matched, err := store.List(ctx, credential.Filter{Query: "PERSON"})
	if err != nil || len(matched) != 1 || matched[0].Provider != "google" {
		t.Fatalf("List(query) = %#v, %v", matched, err)
	}
	empty, err := store.List(ctx, credential.Filter{Query: "missing"})
	if err != nil || empty == nil || len(empty) != 0 {
		t.Fatalf("List(empty) = %#v, %v", empty, err)
	}
}

func TestStoreMissingUpdateAndDelete(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "omnicred.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()
	store := NewStore(db)

	_, err = store.Update(ctx, credential.Credential{ID: 999, UpdatedAt: time.Now()})
	if !errors.Is(err, credential.ErrNotFound) {
		t.Fatalf("Update() error = %v", err)
	}
	if err := store.Delete(ctx, 999); !errors.Is(err, credential.ErrNotFound) {
		t.Fatalf("Delete() error = %v", err)
	}
}
