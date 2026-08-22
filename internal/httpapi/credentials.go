package httpapi

import (
	"net/http"
	"strconv"

	"omnicred/internal/credential"
)

func (api *Handler) createCredential(response http.ResponseWriter, request *http.Request) {
	var input credential.CreateInput
	if err := decodeJSON(response, request, &input); err != nil {
		api.handleError(response, request, err)
		return
	}
	item, err := api.credentials.Create(request.Context(), input)
	if err != nil {
		api.handleError(response, request, err)
		return
	}
	writeJSON(response, http.StatusCreated, item)
}

func (api *Handler) listCredentials(response http.ResponseWriter, request *http.Request) {
	items, err := api.credentials.List(request.Context(), credential.Filter{
		Provider: request.URL.Query().Get("provider"),
		Query:    request.URL.Query().Get("q"),
	})
	if err != nil {
		api.handleError(response, request, err)
		return
	}
	writeJSON(response, http.StatusOK, struct {
		Items []credential.Credential `json:"items"`
	}{Items: items})
}

func (api *Handler) listTOTPCodes(response http.ResponseWriter, request *http.Request) {
	codes, err := api.credentials.TOTPCodes(request.Context())
	if err != nil {
		api.handleError(response, request, err)
		return
	}
	writeJSON(response, http.StatusOK, codes)
}

func (api *Handler) getCredential(response http.ResponseWriter, request *http.Request) {
	id, err := pathID(request)
	if err != nil {
		api.handleError(response, request, err)
		return
	}
	item, err := api.credentials.Get(request.Context(), id)
	if err != nil {
		api.handleError(response, request, err)
		return
	}
	writeJSON(response, http.StatusOK, item)
}

func (api *Handler) updateCredential(response http.ResponseWriter, request *http.Request) {
	id, err := pathID(request)
	if err != nil {
		api.handleError(response, request, err)
		return
	}
	var input credential.UpdateInput
	if err := decodeJSON(response, request, &input); err != nil {
		api.handleError(response, request, err)
		return
	}
	item, err := api.credentials.Update(request.Context(), id, input)
	if err != nil {
		api.handleError(response, request, err)
		return
	}
	writeJSON(response, http.StatusOK, item)
}

func (api *Handler) deleteCredential(response http.ResponseWriter, request *http.Request) {
	id, err := pathID(request)
	if err != nil {
		api.handleError(response, request, err)
		return
	}
	if err := api.credentials.Delete(request.Context(), id); err != nil {
		api.handleError(response, request, err)
		return
	}
	response.WriteHeader(http.StatusNoContent)
}

func pathID(request *http.Request) (int64, error) {
	id, err := strconv.ParseInt(request.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, &requestError{http.StatusBadRequest, "invalid_request", "id must be a positive integer"}
	}
	return id, nil
}
