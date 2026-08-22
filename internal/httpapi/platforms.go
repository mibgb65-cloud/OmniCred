package httpapi

import (
	"net/http"

	"omnicred/internal/platform"
)

func (api *Handler) createPlatform(response http.ResponseWriter, request *http.Request) {
	var input platform.CreateInput
	if err := decodeJSON(response, request, &input); err != nil {
		api.handleError(response, request, err)
		return
	}
	item, err := api.platforms.Create(request.Context(), input)
	if err != nil {
		api.handleError(response, request, err)
		return
	}
	writeJSON(response, http.StatusCreated, item)
}

func (api *Handler) listPlatforms(response http.ResponseWriter, request *http.Request) {
	items, err := api.platforms.List(request.Context())
	if err != nil {
		api.handleError(response, request, err)
		return
	}
	writeJSON(response, http.StatusOK, struct {
		Items []platform.Platform `json:"items"`
	}{Items: items})
}

func (api *Handler) updatePlatform(response http.ResponseWriter, request *http.Request) {
	id, err := pathID(request)
	if err != nil {
		api.handleError(response, request, err)
		return
	}
	var input platform.UpdateInput
	if err := decodeJSON(response, request, &input); err != nil {
		api.handleError(response, request, err)
		return
	}
	item, err := api.platforms.Update(request.Context(), id, input)
	if err != nil {
		api.handleError(response, request, err)
		return
	}
	writeJSON(response, http.StatusOK, item)
}

func (api *Handler) deletePlatform(response http.ResponseWriter, request *http.Request) {
	id, err := pathID(request)
	if err != nil {
		api.handleError(response, request, err)
		return
	}
	if err := api.platforms.Delete(request.Context(), id); err != nil {
		api.handleError(response, request, err)
		return
	}
	response.WriteHeader(http.StatusNoContent)
}
