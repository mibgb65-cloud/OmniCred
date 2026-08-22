package credential

import (
	"context"
	"errors"
	"testing"
	"time"
)

type memoryStore struct {
	items       map[int64]Credential
	nextID      int64
	createCalls int
}

func newMemoryStore() *memoryStore {
	return &memoryStore{items: make(map[int64]Credential), nextID: 1}
}

func (store *memoryStore) Create(_ context.Context, item Credential) (Credential, error) {
	store.createCalls++
	item.ID = store.nextID
	store.nextID++
	store.items[item.ID] = item
	return item, nil
}

func (store *memoryStore) Get(_ context.Context, id int64) (Credential, error) {
	item, ok := store.items[id]
	if !ok {
		return Credential{}, ErrNotFound
	}
	return item, nil
}

func (store *memoryStore) List(_ context.Context, filter Filter) ([]Credential, error) {
	items := make([]Credential, 0, len(store.items))
	for _, item := range store.items {
		if filter.Provider == "" || item.Provider == filter.Provider {
			items = append(items, item)
		}
	}
	return items, nil
}

func (store *memoryStore) Update(_ context.Context, item Credential) (Credential, error) {
	if _, ok := store.items[item.ID]; !ok {
		return Credential{}, ErrNotFound
	}
	store.items[item.ID] = item
	return item, nil
}

func (store *memoryStore) Delete(_ context.Context, id int64) error {
	if _, ok := store.items[id]; !ok {
		return ErrNotFound
	}
	delete(store.items, id)
	return nil
}

func TestServiceCreateNormalizesInputAndPreservesPassword(t *testing.T) {
	store := newMemoryStore()
	service := NewService(store)
	fixed := time.Date(2026, 8, 21, 12, 0, 0, 0, time.FixedZone("SGT", 8*60*60))
	service.now = func() time.Time { return fixed }

	item, err := service.Create(context.Background(), CreateInput{
		Provider: " GitHub ", Account: " user@example.com ", Username: " octocat ", Password: " secret ",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if item.Provider != "github" || item.Account != "user@example.com" || item.Username != "octocat" {
		t.Fatalf("Create() did not normalize fields: %#v", item)
	}
	if item.Password != " secret " {
		t.Fatalf("Create() password = %q, want spaces preserved", item.Password)
	}
	if !item.CreatedAt.Equal(fixed.UTC()) || !item.UpdatedAt.Equal(fixed.UTC()) {
		t.Fatalf("Create() timestamps = %v / %v", item.CreatedAt, item.UpdatedAt)
	}
}

func TestServiceCreateRejectsInvalidInputWithoutWriting(t *testing.T) {
	tests := []struct {
		name  string
		input CreateInput
		field string
	}{
		{"provider", CreateInput{Account: "a", Password: "p"}, "provider"},
		{"account", CreateInput{Provider: "github", Password: "p"}, "account"},
		{"password", CreateInput{Provider: "github", Account: "a"}, "password"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := newMemoryStore()
			service := NewService(store)
			_, err := service.Create(context.Background(), test.input)
			var validationErr *ValidationError
			if !errors.As(err, &validationErr) || validationErr.Field != test.field {
				t.Fatalf("Create() error = %v, want validation error for %s", err, test.field)
			}
			if store.createCalls != 0 {
				t.Fatalf("Create() wrote invalid input")
			}
		})
	}
}

func TestServiceUpdatePreservesCreatedAt(t *testing.T) {
	store := newMemoryStore()
	service := NewService(store)
	created := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	updated := created.Add(time.Hour)
	store.items[1] = Credential{ID: 1, Provider: "github", Account: "old", Password: "old", CreatedAt: created, UpdatedAt: created}
	service.now = func() time.Time { return updated }

	item, err := service.Update(context.Background(), 1, UpdateInput{Provider: "Google", Account: "new", Password: "new"})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if !item.CreatedAt.Equal(created) || !item.UpdatedAt.Equal(updated) {
		t.Fatalf("Update() timestamps = %v / %v", item.CreatedAt, item.UpdatedAt)
	}
}
