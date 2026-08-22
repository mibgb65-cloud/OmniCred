package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"

	"omnicred/internal/credential"
)

const maxRequestBody = 64 << 10

type errorBody struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type requestError struct {
	status  int
	code    string
	message string
}

func (err *requestError) Error() string {
	return err.message
}

func writeJSON(response http.ResponseWriter, status int, value any) {
	response.Header().Set("Content-Type", "application/json; charset=utf-8")
	response.WriteHeader(status)
	_ = json.NewEncoder(response).Encode(value)
}

func writeError(response http.ResponseWriter, status int, code, message string) {
	writeJSON(response, status, errorBody{Error: errorDetail{Code: code, Message: message}})
}

func decodeJSON(response http.ResponseWriter, request *http.Request, target any) error {
	mediaType, _, err := mime.ParseMediaType(request.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return &requestError{http.StatusUnsupportedMediaType, "unsupported_media_type", "Content-Type must be application/json"}
	}
	request.Body = http.MaxBytesReader(response, request.Body, maxRequestBody)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return &requestError{http.StatusBadRequest, "invalid_request", "request body must be a valid JSON object"}
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return &requestError{http.StatusBadRequest, "invalid_request", "request body must contain one JSON object"}
	}
	return nil
}

func validationMessage(err *credential.ValidationError) string {
	return err.Field + " " + err.Message
}
