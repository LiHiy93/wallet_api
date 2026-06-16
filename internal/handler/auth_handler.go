package handler

import (
	"net/http"

	"wallet-api/internal/service"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	result, err := h.auth.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	result, err := h.auth.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}
