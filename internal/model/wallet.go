package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Wallet struct {
	ID        uuid.UUID       `json:"id"`
	UserID    uuid.UUID       `json:"user_id"`
	Name      string          `json:"name"`
	Balance   decimal.Decimal `json:"balance"`
	Currency  string          `json:"currency"`
	CreatedAt time.Time       `json:"created_at"`
}
