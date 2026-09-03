package httpapi

import (
	"net/http"

	"omnicred/internal/identity"
)

func (api *Handler) createIdentity(response http.ResponseWriter, request *http.Request) {
	var input identity.CreateInput
	if err := decodeJSON(response, request, &input); err != nil {
		api.handleError(response, request, err)
		return
	}
	item, err := api.identities.Create(request.Context(), input)
	if err != nil {
		api.handleError(response, request, err)
		return
	}
	writeJSON(response, http.StatusCreated, item)
}

func (api *Handler) listIdentities(response http.ResponseWriter, request *http.Request) {
	items, err := api.identities.List(request.Context(), identity.Filter{
		Country: request.URL.Query().Get("country"),
		Query:   request.URL.Query().Get("q"),
	})
	if err != nil {
		api.handleError(response, request, err)
		return
	}
	writeJSON(response, http.StatusOK, struct {
		Items []identity.Profile `json:"items"`
	}{Items: items})
}

func (api *Handler) getIdentity(response http.ResponseWriter, request *http.Request) {
	id, err := pathID(request)
	if err != nil {
		api.handleError(response, request, err)
		return
	}
	item, err := api.identities.Get(request.Context(), id)
	if err != nil {
		api.handleError(response, request, err)
		return
	}
	writeJSON(response, http.StatusOK, item)
}

func (api *Handler) updateIdentity(response http.ResponseWriter, request *http.Request) {
	id, err := pathID(request)
	if err != nil {
		api.handleError(response, request, err)
		return
	}
	var input identity.UpdateInput
	if err := decodeJSON(response, request, &input); err != nil {
		api.handleError(response, request, err)
		return
	}
	item, err := api.identities.Update(request.Context(), id, input)
	if err != nil {
		api.handleError(response, request, err)
		return
	}
	writeJSON(response, http.StatusOK, item)
}

func (api *Handler) deleteIdentity(response http.ResponseWriter, request *http.Request) {
	id, err := pathID(request)
	if err != nil {
		api.handleError(response, request, err)
		return
	}
	if err := api.identities.Delete(request.Context(), id); err != nil {
		api.handleError(response, request, err)
		return
	}
	response.WriteHeader(http.StatusNoContent)
}
