package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"omnicred/internal/credential"
	"omnicred/internal/platform"
)

type testStore struct {
	items          map[int64]credential.Credential
	platformItems  map[int64]platform.Platform
	nextID         int64
	nextPlatformID int64
}

func newTestHandler(logOutput io.Writer) *Handler {
	store := &testStore{
		items: make(map[int64]credential.Credential), platformItems: make(map[int64]platform.Platform),
		nextID: 1, nextPlatformID: 1,
	}
	credentialService := credential.NewService(store)
	platformService := platform.NewService(store)
	logger := slog.New(slog.NewTextHandler(logOutput, nil))
	return New(credentialService, platformService, &testSettingsService{}, logger)
}

func (store *testStore) Create(_ context.Context, item credential.Credential) (credential.Credential, error) {
	item.ID = store.nextID
	store.nextID++
	store.items[item.ID] = item
	return item, nil
}

func (store *testStore) Get(_ context.Context, id int64) (credential.Credential, error) {
	item, ok := store.items[id]
	if !ok {
		return credential.Credential{}, credential.ErrNotFound
	}
	return item, nil
}

func (store *testStore) List(_ context.Context, filter credential.Filter) ([]credential.Credential, error) {
	items := make([]credential.Credential, 0)
	for id := int64(1); id < store.nextID; id++ {
		item, ok := store.items[id]
		if !ok || filter.Provider != "" && item.Provider != filter.Provider {
			continue
		}
		query := strings.ToLower(filter.Query)
		if query != "" && !strings.Contains(strings.ToLower(item.Account), query) && !strings.Contains(strings.ToLower(item.Username), query) {
			continue
		}
		items = append(items, item)
	}
	return items, nil
}

func (store *testStore) Update(_ context.Context, item credential.Credential) (credential.Credential, error) {
	if _, ok := store.items[item.ID]; !ok {
		return credential.Credential{}, credential.ErrNotFound
	}
	store.items[item.ID] = item
	return item, nil
}

func (store *testStore) Delete(_ context.Context, id int64) error {
	if _, ok := store.items[id]; !ok {
		return credential.ErrNotFound
	}
	delete(store.items, id)
	return nil
}

func TestCredentialAPIFlow(t *testing.T) {
	handler := newTestHandler(io.Discard)

	created := performJSON(t, handler, http.MethodPost, "/api/v1/credentials", `{
		"provider":"GitHub","account":"user@example.com","username":"octocat","password":"secret",
		"totp_secret":"gezd gnbv gy3t qojq gezd gnbv gy3t qojq",
		"recovery_codes":[" alpha-bravo ","charlie-delta"]
	}`)
	if created.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", created.Code, created.Body.String())
	}
	var item credential.Credential
	decodeResponse(t, created, &item)
	if item.ID != 1 || item.Provider != "github" || item.Password != "secret" || item.TOTPSecret != "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ" || len(item.RecoveryCodes) != 2 {
		t.Fatalf("created item = %#v", item)
	}
	codes := performJSON(t, handler, http.MethodGet, "/api/v1/totp", "")
	if codes.Code != http.StatusOK || !strings.Contains(codes.Body.String(), `"credential_id":1`) {
		t.Fatalf("TOTP response = %d, %s", codes.Code, codes.Body.String())
	}

	listed := performJSON(t, handler, http.MethodGet, "/api/v1/credentials?provider=github&q=octo", "")
	if listed.Code != http.StatusOK || !strings.Contains(listed.Body.String(), `"account":"user@example.com"`) {
		t.Fatalf("list response = %d, %s", listed.Code, listed.Body.String())
	}

	updated := performJSON(t, handler, http.MethodPut, "/api/v1/credentials/1", `{
		"provider":"google","account":"new@example.com","username":"person","password":"changed"
	}`)
	if updated.Code != http.StatusOK || !strings.Contains(updated.Body.String(), `"provider":"google"`) {
		t.Fatalf("update response = %d, %s", updated.Code, updated.Body.String())
	}

	deleted := performJSON(t, handler, http.MethodDelete, "/api/v1/credentials/1", "")
	if deleted.Code != http.StatusNoContent || deleted.Body.Len() != 0 {
		t.Fatalf("delete response = %d, %q", deleted.Code, deleted.Body.String())
	}
	missing := performJSON(t, handler, http.MethodGet, "/api/v1/credentials/1", "")
	if missing.Code != http.StatusNotFound || !strings.Contains(missing.Body.String(), `"code":"not_found"`) {
		t.Fatalf("missing response = %d, %s", missing.Code, missing.Body.String())
	}
}

func TestAPIRejectsInvalidRequests(t *testing.T) {
	handler := newTestHandler(io.Discard)
	tests := []struct {
		name        string
		method      string
		path        string
		contentType string
		body        string
		status      int
		code        string
	}{
		{"content type", http.MethodPost, "/api/v1/credentials", "text/plain", `{}`, 415, "unsupported_media_type"},
		{"unknown field", http.MethodPost, "/api/v1/credentials", "application/json", `{"provider":"x","account":"a","password":"p","extra":true}`, 400, "invalid_request"},
		{"missing password", http.MethodPost, "/api/v1/credentials", "application/json", `{"provider":"x","account":"a"}`, 400, "invalid_request"},
		{"invalid TOTP secret", http.MethodPost, "/api/v1/credentials", "application/json", `{"provider":"x","account":"a","password":"p","totp_secret":"invalid!"}`, 400, "invalid_request"},
		{"bad id", http.MethodGet, "/api/v1/credentials/nope", "", "", 400, "invalid_request"},
		{"method", http.MethodPatch, "/api/v1/credentials/1", "application/json", `{}`, 405, "method_not_allowed"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.path, strings.NewReader(test.body))
			if test.contentType != "" {
				request.Header.Set("Content-Type", test.contentType)
			}
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != test.status || !strings.Contains(response.Body.String(), `"code":"`+test.code+`"`) {
				t.Fatalf("response = %d, %s", response.Code, response.Body.String())
			}
		})
	}
}

func TestAPISecurityHeadersAndLogsOmitPassword(t *testing.T) {
	var logs bytes.Buffer
	handler := newTestHandler(&logs)
	response := performJSON(t, handler, http.MethodPost, "/api/v1/credentials", `{
		"provider":"github","account":"user@example.com","password":"never-log-this",
		"totp_secret":"GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ",
		"recovery_codes":["never-log-recovery"]
	}`)
	if response.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("CORS must not be enabled")
	}
	if response.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("Cache-Control = %q", response.Header().Get("Cache-Control"))
	}
	if strings.Contains(logs.String(), "never-log-this") || strings.Contains(logs.String(), "user@example.com") || strings.Contains(logs.String(), "GEZDGNBV") || strings.Contains(logs.String(), "never-log-recovery") {
		t.Fatalf("logs contain credential data: %s", logs.String())
	}
}

func TestAPIDesktopCORSIsNarrowlyScoped(t *testing.T) {
	handler := newTestHandler(io.Discard)

	allowedRequest := httptest.NewRequest(http.MethodOptions, "/api/v1/credentials", nil)
	allowedRequest.Header.Set("Origin", "http://wails.localhost")
	allowedRequest.Header.Set("Access-Control-Request-Method", http.MethodPost)
	allowed := httptest.NewRecorder()
	handler.ServeHTTP(allowed, allowedRequest)
	if allowed.Code != http.StatusNoContent || allowed.Header().Get("Access-Control-Allow-Origin") != "http://wails.localhost" {
		t.Fatalf("allowed preflight = %d, origin %q", allowed.Code, allowed.Header().Get("Access-Control-Allow-Origin"))
	}

	blockedRequest := httptest.NewRequest(http.MethodGet, "/api/v1/credentials", nil)
	blockedRequest.Header.Set("Origin", "https://example.com")
	blocked := httptest.NewRecorder()
	handler.ServeHTTP(blocked, blockedRequest)
	if blocked.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatalf("unexpected CORS origin %q", blocked.Header().Get("Access-Control-Allow-Origin"))
	}
}

func performJSON(t *testing.T, handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func decodeResponse(t *testing.T, response *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.Unmarshal(response.Body.Bytes(), target); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}
