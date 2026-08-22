package httpapi

import (
	"net/http"

	"omnicred/internal/appsettings"
)

func (api *Handler) settingsStatus(response http.ResponseWriter, request *http.Request) {
	status, err := api.settings.Status(request.Context())
	if err != nil {
		api.handleError(response, request, err)
		return
	}
	writeJSON(response, http.StatusOK, status)
}

func (api *Handler) migrateStorage(response http.ResponseWriter, request *http.Request) {
	var input appsettings.StorageInput
	if err := decodeJSON(response, request, &input); err != nil {
		api.handleError(response, request, err)
		return
	}
	result, err := api.settings.MigrateStorage(request.Context(), input)
	if err != nil {
		api.handleError(response, request, err)
		return
	}
	writeJSON(response, http.StatusOK, result)
}

func (api *Handler) checkUpdate(response http.ResponseWriter, request *http.Request) {
	result, err := api.settings.CheckUpdate(request.Context())
	if err != nil {
		api.handleError(response, request, err)
		return
	}
	writeJSON(response, http.StatusOK, result)
}
