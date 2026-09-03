package sqlite

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"omnicred/internal/identity"
)

func TestIdentityStoreCRUDAndSearch(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "omnicred.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()
	store := NewStore(db)
	now := time.Date(2026, 9, 3, 8, 0, 0, 0, time.UTC)

	created, err := store.CreateIdentity(ctx, identity.Profile{
		Country: "Philippines", FullName: "Angelo Santos", LocalizedName: "安杰洛·桑托斯",
		FirstName: "Angelo", MiddleName: "Reyes", LastName: "Santos", Gender: "male",
		BirthDate: "1998-08-16", StreetAddress: "Unit 402, Sunshine Court, Aurora Blvd, Cubao",
		City: "Quezon City", Region: "Metro Manila", PostalCode: "1109", Phone: "+63 (917) 482-9301",
		Email: "angelo@example.com", Password: "secret", CreatedAt: now, UpdatedAt: now,
	})
	if err != nil || created.ID <= 0 {
		t.Fatalf("CreateIdentity() = %#v, %v", created, err)
	}

	got, err := store.GetIdentity(ctx, created.ID)
	if err != nil || got.MiddleName != "Reyes" || got.LocalizedName != "安杰洛·桑托斯" || !got.CreatedAt.Equal(now) {
		t.Fatalf("GetIdentity() = %#v, %v", got, err)
	}
	matched, err := store.ListIdentities(ctx, identity.Filter{Query: "QUEZON"})
	if err != nil || len(matched) != 1 || matched[0].ID != created.ID {
		t.Fatalf("ListIdentities(query) = %#v, %v", matched, err)
	}
	matched, err = store.ListIdentities(ctx, identity.Filter{Country: "philippines"})
	if err != nil || len(matched) != 1 {
		t.Fatalf("ListIdentities(country) = %#v, %v", matched, err)
	}

	got.City = "Manila"
	got.UpdatedAt = now.Add(time.Hour)
	if _, err := store.UpdateIdentity(ctx, got); err != nil {
		t.Fatalf("UpdateIdentity() error = %v", err)
	}
	if err := store.DeleteIdentity(ctx, created.ID); err != nil {
		t.Fatalf("DeleteIdentity() error = %v", err)
	}
	if _, err := store.GetIdentity(ctx, created.ID); !errors.Is(err, identity.ErrNotFound) {
		t.Fatalf("GetIdentity() after delete error = %v", err)
	}
}
