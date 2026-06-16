package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransactionType string

const (
	TransactionTypeIncome   TransactionType = "income"
	TransactionTypeExpense  TransactionType = "expense"
	TransactionTypeTransfer TransactionType = "transfer"
)

type Transaction struct {
	ID          uuid.UUID       `json:"id"`
	WalletID    uuid.UUID       `json:"wallet_id"`
	UserID      uuid.UUID       `json:"user_id"`
	Type        TransactionType `json:"type"`
	Amount      decimal.Decimal `json:"amount"`
	Description string          `json:"description"`
	CreatedAt   time.Time       `json:"created_at"`
}
