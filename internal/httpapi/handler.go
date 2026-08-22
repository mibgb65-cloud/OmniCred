package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"omnicred/internal/appsettings"
	"omnicred/internal/credential"
	"omnicred/internal/platform"
)

type Handler struct {
	credentials *credential.Service
	platforms   *platform.Service
	settings    settingsService
	logger      *slog.Logger
	handler     http.Handler
}

type settingsService interface {
	Status(context.Context) (appsettings.RuntimeStatus, error)
	MigrateStorage(context.Context, appsettings.StorageInput) (appsettings.StorageResult, error)
	CheckUpdate(context.Context) (appsettings.UpdateInfo, error)
}

func New(credentials *credential.Service, platforms *platform.Service, settings settingsService, logger *slog.Logger) *Handler {
	api := &Handler{credentials: credentials, platforms: platforms, settings: settings, logger: logger}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", api.health)
	mux.HandleFunc("POST /api/v1/credentials", api.createCredential)
	mux.HandleFunc("GET /api/v1/credentials", api.listCredentials)
	mux.HandleFunc("/api/v1/credentials", api.methodNotAllowed)
	mux.HandleFunc("GET /api/v1/credentials/{id}", api.getCredential)
	mux.HandleFunc("PUT /api/v1/credentials/{id}", api.updateCredential)
	mux.HandleFunc("DELETE /api/v1/credentials/{id}", api.deleteCredential)
	mux.HandleFunc("/api/v1/credentials/{id}", api.methodNotAllowed)
	mux.HandleFunc("POST /api/v1/platforms", api.createPlatform)
	mux.HandleFunc("GET /api/v1/platforms", api.listPlatforms)
	mux.HandleFunc("/api/v1/platforms", api.methodNotAllowed)
	mux.HandleFunc("PUT /api/v1/platforms/{id}", api.updatePlatform)
	mux.HandleFunc("DELETE /api/v1/platforms/{id}", api.deletePlatform)
	mux.HandleFunc("/api/v1/platforms/{id}", api.methodNotAllowed)
	mux.HandleFunc("GET /api/v1/settings/status", api.settingsStatus)
	mux.HandleFunc("PUT /api/v1/settings/storage", api.migrateStorage)
	mux.HandleFunc("GET /api/v1/settings/updates", api.checkUpdate)
	mux.HandleFunc("/api/v1/settings/", api.methodNotAllowed)
	api.handler = withDesktopCORS(withSecurityHeaders(withRecovery(logger, withRequestLogging(logger, mux))))
	return api
}

func (api *Handler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	api.handler.ServeHTTP(response, request)
}

func (api *Handler) health(response http.ResponseWriter, _ *http.Request) {
	writeJSON(response, http.StatusOK, map[string]string{"status": "ok"})
}

func (api *Handler) methodNotAllowed(response http.ResponseWriter, _ *http.Request) {
	writeError(response, http.StatusMethodNotAllowed, "method_not_allowed", "HTTP method is not supported")
}

func (api *Handler) handleError(response http.ResponseWriter, request *http.Request, err error) {
	var requestErr *requestError
	if errors.As(err, &requestErr) {
		writeError(response, requestErr.status, requestErr.code, requestErr.message)
		return
	}
	var validationErr *credential.ValidationError
	if errors.As(err, &validationErr) {
		writeError(response, http.StatusBadRequest, "invalid_request", validationMessage(validationErr))
		return
	}
	if errors.Is(err, credential.ErrNotFound) {
		writeError(response, http.StatusNotFound, "not_found", "credential was not found")
		return
	}
	var platformValidationErr *platform.ValidationError
	if errors.As(err, &platformValidationErr) {
		writeError(response, http.StatusBadRequest, "invalid_request", platformValidationErr.Field+" "+platformValidationErr.Message)
		return
	}
	if errors.Is(err, platform.ErrNotFound) {
		writeError(response, http.StatusNotFound, "platform_not_found", "platform was not found")
		return
	}
	if errors.Is(err, platform.ErrAlreadyExists) {
		writeError(response, http.StatusConflict, "platform_exists", "platform name already exists")
		return
	}
	if errors.Is(err, platform.ErrInUse) {
		writeError(response, http.StatusConflict, "platform_in_use", "platform still has credentials")
		return
	}
	if errors.Is(err, appsettings.ErrTargetExists) {
		writeError(response, http.StatusConflict, "storage_target_exists", "target database already exists")
		return
	}
	var settingsValidationErr *appsettings.ValidationError
	if errors.As(err, &settingsValidationErr) {
		writeError(response, http.StatusBadRequest, "invalid_request", settingsValidationErr.Field+" "+settingsValidationErr.Message)
		return
	}
	api.logger.Error("request failed", "method", request.Method, "path", request.URL.Path, "error", err)
	writeError(response, http.StatusInternalServerError, "internal_error", "an internal error occurred")
}
