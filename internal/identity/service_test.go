package identity

import (
	"context"
	"errors"
	"testing"
	"time"
)

type memoryStore struct {
	items  map[int64]Profile
	nextID int64
}

func newMemoryStore() *memoryStore {
	return &memoryStore{items: make(map[int64]Profile), nextID: 1}
}

func (store *memoryStore) CreateIdentity(_ context.Context, item Profile) (Profile, error) {
	item.ID = store.nextID
	store.nextID++
	store.items[item.ID] = item
	return item, nil
}

func (store *memoryStore) GetIdentity(_ context.Context, id int64) (Profile, error) {
	item, ok := store.items[id]
	if !ok {
		return Profile{}, ErrNotFound
	}
	return item, nil
}

func (store *memoryStore) ListIdentities(_ context.Context, _ Filter) ([]Profile, error) {
	items := make([]Profile, 0, len(store.items))
	for _, item := range store.items {
		items = append(items, item)
	}
	return items, nil
}

func (store *memoryStore) UpdateIdentity(_ context.Context, item Profile) (Profile, error) {
	if _, ok := store.items[item.ID]; !ok {
		return Profile{}, ErrNotFound
	}
	store.items[item.ID] = item
	return item, nil
}

func (store *memoryStore) DeleteIdentity(_ context.Context, id int64) error {
	if _, ok := store.items[id]; !ok {
		return ErrNotFound
	}
	delete(store.items, id)
	return nil
}

func TestServiceCreateNormalizesInternationalProfile(t *testing.T) {
	store := newMemoryStore()
	service := NewService(store)
	fixed := time.Date(2026, 9, 3, 8, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return fixed }

	item, err := service.Create(context.Background(), CreateInput{
		Country: " Philippines ", FullName: " Angelo Santos ", LocalizedName: " 安杰洛·桑托斯 ",
		FirstName: " Angelo ", MiddleName: " Reyes ", LastName: " Santos ", Gender: " Male ",
		BirthDate: "1998-08-16", StreetAddress: " Unit 402, Sunshine Court, Aurora Blvd, Cubao ",
		City: " Quezon City ", Region: " Metro Manila ", PostalCode: " 1109 ",
		Phone: " +63 (917) 482-9301 ", Email: "angelo@example.com", Password: " secret ",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if item.Country != "Philippines" || item.FullName != "Angelo Santos" || item.Gender != "male" || item.MiddleName != "Reyes" {
		t.Fatalf("Create() did not normalize fields: %#v", item)
	}
	if item.Password != " secret " {
		t.Fatalf("Create() password = %q, want spaces preserved", item.Password)
	}
	if !item.CreatedAt.Equal(fixed) || !item.UpdatedAt.Equal(fixed) {
		t.Fatalf("Create() timestamps = %v / %v", item.CreatedAt, item.UpdatedAt)
	}
}

func TestServiceRejectsInvalidProfile(t *testing.T) {
	tests := []struct {
		name  string
		input CreateInput
		field string
	}{
		{"country", CreateInput{FullName: "Angelo"}, "country"},
		{"full name", CreateInput{Country: "Philippines"}, "full_name"},
		{"gender", CreateInput{Country: "Philippines", FullName: "Angelo", Gender: "unknown"}, "gender"},
		{"birth date", CreateInput{Country: "Philippines", FullName: "Angelo", BirthDate: "1998-13-40"}, "birth_date"},
		{"email", CreateInput{Country: "Philippines", FullName: "Angelo", Email: "invalid"}, "email"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewService(newMemoryStore()).Create(context.Background(), test.input)
			var validationErr *ValidationError
			if !errors.As(err, &validationErr) || validationErr.Field != test.field {
				t.Fatalf("Create() error = %v, want validation error for %s", err, test.field)
			}
		})
	}
}

func TestServiceUpdatePreservesCreatedAt(t *testing.T) {
	store := newMemoryStore()
	createdAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	store.items[1] = Profile{ID: 1, Country: "Philippines", FullName: "Old Name", CreatedAt: createdAt, UpdatedAt: createdAt}
	service := NewService(store)
	service.now = func() time.Time { return createdAt.Add(time.Hour) }

	item, err := service.Update(context.Background(), 1, UpdateInput{Country: "Singapore", FullName: "New Name"})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if !item.CreatedAt.Equal(createdAt) || !item.UpdatedAt.Equal(createdAt.Add(time.Hour)) {
		t.Fatalf("Update() timestamps = %v / %v", item.CreatedAt, item.UpdatedAt)
	}
}
