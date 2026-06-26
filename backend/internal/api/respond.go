package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

type errorEnvelope struct {
	Error apiError `json:"error"`
}

type successEnvelope struct {
	Data any `json:"data"`
}

type apiError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

func WriteJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(successEnvelope{Data: value})
}

func WriteEmpty(w http.ResponseWriter, status int) {
	WriteJSON(w, status, nil)
}

func WriteError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	ctx := context.Background()
	if r != nil {
		ctx = r.Context()
	}
	WriteRawJSON(w, status, errorEnvelope{
		Error: apiError{
			Code:      code,
			Message:   message,
			RequestID: middleware.GetReqID(ctx),
		},
	})
}

func WriteRawJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	WriteJSON(w, status, value)
}

func writeError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	WriteError(w, r, status, code, message)
}
