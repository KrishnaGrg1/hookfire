package helper

import (
	"encoding/json"
	"net/http"
)

type APIError struct {
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

type Response struct {
	Success bool      `json:"success"`
	Message string    `json:"message"`
	Data    any       `json:"data,omitempty"`
	Error   *APIError `json:"error,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, resp Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func Success(w http.ResponseWriter, status int, message string, data any) {
	writeJSON(w, status, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Error(w http.ResponseWriter, status int, message, code, details string) {
	writeJSON(w, status, Response{
		Success: false,
		Message: message,
		Error: &APIError{
			Code:    code,
			Details: details,
		},
	})
}
