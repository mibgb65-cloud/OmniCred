package httpapi

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"omnicred/internal/credential"
	"omnicred/internal/platform"
)

func (store *testStore) CreatePlatform(_ context.Context, item platform.Platform) (platform.Platform, error) {
	for _, existing := range store.platformItems {
		if strings.EqualFold(existing.Name, item.Name) {
			return platform.Platform{}, platform.ErrAlreadyExists
		}
	}
	item.ID = store.nextPlatformID
	store.nextPlatformID++
	store.platformItems[item.ID] = item
	return item, nil
}

func (store *testStore) ListPlatforms(_ context.Context) ([]platform.Platform, error) {
	items := make([]platform.Platform, 0, len(store.platformItems))
	for id := int64(1); id < store.nextPlatformID; id++ {
		item, ok := store.platformItems[id]
		if !ok {
			continue
		}
		for _, credential := range store.items {
			if strings.EqualFold(credential.Provider, item.Name) {
				item.CredentialCount++
			}
		}
		items = append(items, item)
	}
	return items, nil
}

func (store *testStore) UpdatePlatform(_ context.Context, id int64, name string, updatedAt time.Time) (platform.Platform, error) {
	item, ok := store.platformItems[id]
	if !ok {
		return platform.Platform{}, platform.ErrNotFound
	}
	for otherID, existing := range store.platformItems {
		if otherID != id && strings.EqualFold(existing.Name, name) {
			return platform.Platform{}, platform.ErrAlreadyExists
		}
	}
	oldName := item.Name
	item.Name = name
	item.UpdatedAt = updatedAt
	for credentialID, credential := range store.items {
		if strings.EqualFold(credential.Provider, oldName) {
			credential.Provider = name
			credential.UpdatedAt = updatedAt
			store.items[credentialID] = credential
			item.CredentialCount++
		}
	}
	store.platformItems[id] = item
	return item, nil
}

func (store *testStore) DeletePlatform(_ context.Context, id int64) error {
	item, ok := store.platformItems[id]
	if !ok {
		return platform.ErrNotFound
	}
	for _, credential := range store.items {
		if strings.EqualFold(credential.Provider, item.Name) {
			return platform.ErrInUse
		}
	}
	delete(store.platformItems, id)
	return nil
}

func TestPlatformAPIFlowAndDeleteProtection(t *testing.T) {
	handler := newTestHandler(io.Discard)

	created := performJSON(t, handler, http.MethodPost, "/api/v1/platforms", `{"name":"Work"}`)
	if created.Code != http.StatusCreated || !strings.Contains(created.Body.String(), `"name":"work"`) {
		t.Fatalf("create platform = %d, %s", created.Code, created.Body.String())
	}
	duplicate := performJSON(t, handler, http.MethodPost, "/api/v1/platforms", `{"name":"WORK"}`)
	if duplicate.Code != http.StatusConflict || !strings.Contains(duplicate.Body.String(), `"code":"platform_exists"`) {
		t.Fatalf("duplicate platform = %d, %s", duplicate.Code, duplicate.Body.String())
	}

	credential := performJSON(t, handler, http.MethodPost, "/api/v1/credentials", `{
		"provider":"work","account":"user@example.com","password":"secret"
	}`)
	if credential.Code != http.StatusCreated {
		t.Fatalf("create credential = %d, %s", credential.Code, credential.Body.String())
	}
	renamed := performJSON(t, handler, http.MethodPut, "/api/v1/platforms/1", `{"name":"Company"}`)
	if renamed.Code != http.StatusOK || !strings.Contains(renamed.Body.String(), `"credential_count":1`) {
		t.Fatalf("rename platform = %d, %s", renamed.Code, renamed.Body.String())
	}
	listedCredentials := performJSON(t, handler, http.MethodGet, "/api/v1/credentials?provider=company", "")
	if !strings.Contains(listedCredentials.Body.String(), `"provider":"company"`) {
		t.Fatalf("renamed credentials = %s", listedCredentials.Body.String())
	}

	inUse := performJSON(t, handler, http.MethodDelete, "/api/v1/platforms/1", "")
	if inUse.Code != http.StatusConflict || !strings.Contains(inUse.Body.String(), `"code":"platform_in_use"`) {
		t.Fatalf("delete in-use platform = %d, %s", inUse.Code, inUse.Body.String())
	}
}

func TestPlatformStoreErrorsAreDistinct(t *testing.T) {
	store := &testStore{platformItems: make(map[int64]platform.Platform), items: make(map[int64]credential.Credential)}
	if err := store.DeletePlatform(context.Background(), 99); !errors.Is(err, platform.ErrNotFound) {
		t.Fatalf("DeletePlatform() error = %v", err)
	}
}
