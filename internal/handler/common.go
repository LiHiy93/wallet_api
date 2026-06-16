package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"wallet-api/internal/middleware"
	"wallet-api/internal/service"

	"github.com/google/uuid"
)

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func decodeJSON(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}

func currentUserID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return uuid.Nil, false
	}
	return userID, true
}

func mapServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "invalid input")
	case errors.Is(err, service.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, "invalid credentials")
	case errors.Is(err, service.ErrEmailAlreadyExists):
		writeError(w, http.StatusConflict, "email already exists")
	case errors.Is(err, service.ErrWalletNotFound):
		writeError(w, http.StatusNotFound, "wallet not found")
	case errors.Is(err, service.ErrInsufficientFunds):
		writeError(w, http.StatusBadRequest, "insufficient funds")
	case errors.Is(err, service.ErrWalletBalanceNotZero):
		writeError(w, http.StatusBadRequest, "wallet balance must be zero")
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
