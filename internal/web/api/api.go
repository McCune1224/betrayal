// Package api provides the versioned JSON API shell.
package api

import (
	"encoding/json"
	"net/http"
)

// ErrorBody is the canonical shape for API errors.
type ErrorBody struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Fields  map[string]any `json:"fields"`
}

// WriteJSON writes a JSON response with the supplied status code.
func WriteJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// WriteError writes a canonical API error response.
func WriteError(w http.ResponseWriter, status int, code, message string, fields map[string]any) {
	WriteJSON(w, status, struct {
		Error ErrorBody `json:"error"`
	}{Error: ErrorBody{Code: code, Message: message, Fields: fields}})
}

// Health reports that the API shell is available.
func Health(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// NotFound reports an API route that has not been registered.
func NotFound(w http.ResponseWriter, r *http.Request) {
	WriteError(w, http.StatusNotFound, "not_found", "API route not found", nil)
}
