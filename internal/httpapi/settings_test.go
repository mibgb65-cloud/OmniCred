package httpapi

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"omnicred/internal/appsettings"
)

type testSettingsService struct{}

func (*testSettingsService) Status(_ context.Context) (appsettings.RuntimeStatus, error) {
	return appsettings.RuntimeStatus{
		Version: "0.1.0", DatabasePath: `C:\data\omnicred.db`, APIAddress: "127.0.0.1:8787",
		RepositoryURL: "https://github.com/mibgb65-cloud/OmniCred", DatabaseOK: true, UninstallAvailable: true,
	}, nil
}

func (*testSettingsService) MigrateStorage(_ context.Context, input appsettings.StorageInput) (appsettings.StorageResult, error) {
	return appsettings.StorageResult{DatabasePath: input.DatabasePath, RestartRequired: true}, nil
}

func (*testSettingsService) CheckUpdate(_ context.Context) (appsettings.UpdateInfo, error) {
	return appsettings.UpdateInfo{
		CurrentVersion: "0.1.0", LatestVersion: "v0.2.0", UpdateAvailable: true,
		ReleaseURL: "https://github.com/mibgb65-cloud/OmniCred/releases/tag/v0.2.0", CheckedAt: time.Now(), Status: "ok",
	}, nil
}

func TestSettingsAPI(t *testing.T) {
	handler := newTestHandler(io.Discard)
	status := performJSON(t, handler, http.MethodGet, "/api/v1/settings/status", "")
	if status.Code != http.StatusOK || !strings.Contains(status.Body.String(), `"version":"0.1.0"`) {
		t.Fatalf("status response = %d, %s", status.Code, status.Body.String())
	}
	storage := performJSON(t, handler, http.MethodPut, "/api/v1/settings/storage", `{"database_path":"D:\\vault.db"}`)
	if storage.Code != http.StatusOK || !strings.Contains(storage.Body.String(), `"restart_required":true`) {
		t.Fatalf("storage response = %d, %s", storage.Code, storage.Body.String())
	}
	update := performJSON(t, handler, http.MethodGet, "/api/v1/settings/updates", "")
	if update.Code != http.StatusOK || !strings.Contains(update.Body.String(), `"update_available":true`) {
		t.Fatalf("update response = %d, %s", update.Code, update.Body.String())
	}
}
