package handler

import (
	"net/http"
	"strconv"
	"time"

	"wallet-api/internal/model"
	"wallet-api/internal/service"

	"github.com/shopspring/decimal"
)

type TransactionHandler struct {
	transactions *service.TransactionService
}

func NewTransactionHandler(transactions *service.TransactionService) *TransactionHandler {
	return &TransactionHandler{transactions: transactions}
}

type createTransactionRequest struct {
	Type        model.TransactionType `json:"type"`
	Amount      decimal.Decimal       `json:"amount"`
	Description string                `json:"description"`
}

func (h *TransactionHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	walletID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}
	var req createTransactionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	tx, err := h.transactions.Create(r.Context(), userID, walletID, req.Type, req.Amount, req.Description)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, tx)
}

func (h *TransactionHandler) ListByWallet(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	walletID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}
	page := parsePositiveInt(r.URL.Query().Get("page"), 1)
	limit := parsePositiveInt(r.URL.Query().Get("limit"), 20)
	transactions, err := h.transactions.ListByWallet(r.Context(), userID, walletID, page, limit)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, transactions)
}

func (h *TransactionHandler) ListByUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	from, ok := parseDateQuery(w, r, "from")
	if !ok {
		return
	}
	to, ok := parseDateQuery(w, r, "to")
	if !ok {
		return
	}
	if to != nil {
		nextDay := to.AddDate(0, 0, 1)
		to = &nextDay
	}
	transactions, err := h.transactions.ListByUser(r.Context(), userID, from, to)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, transactions)
}

func parsePositiveInt(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return fallback
	}
	return value
}

func parseDateQuery(w http.ResponseWriter, r *http.Request, name string) (*time.Time, bool) {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		return nil, true
	}
	parsed, err := time.Parse("2006-01-02", raw)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid "+name+" date")
		return nil, false
	}
	return &parsed, true
}
