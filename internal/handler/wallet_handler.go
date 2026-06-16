package handler

import (
	"net/http"

	"wallet-api/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type WalletHandler struct {
	wallets *service.WalletService
}

func NewWalletHandler(wallets *service.WalletService) *WalletHandler {
	return &WalletHandler{wallets: wallets}
}

type createWalletRequest struct {
	Name     string `json:"name"`
	Currency string `json:"currency"`
}

func (h *WalletHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	wallets, err := h.wallets.List(r.Context(), userID)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, wallets)
}

func (h *WalletHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	var req createWalletRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	wallet, err := h.wallets.Create(r.Context(), userID, req.Name, req.Currency)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, wallet)
}

func (h *WalletHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	walletID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}
	wallet, err := h.wallets.Get(r.Context(), userID, walletID)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, wallet)
}

func (h *WalletHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	walletID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}
	if err := h.wallets.Delete(r.Context(), userID, walletID); err != nil {
		mapServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func parseUUIDParam(w http.ResponseWriter, r *http.Request, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, name))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return uuid.Nil, false
	}
	return id, true
}
