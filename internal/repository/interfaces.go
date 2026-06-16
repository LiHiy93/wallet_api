package repository

import (
	"context"
	"time"

	"wallet-api/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shopspring/decimal"
)

type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type TxManager interface {
	WithinTx(ctx context.Context, fn func(q DBTX) error) error
}

type UserRepository interface {
	Create(ctx context.Context, email, passwordHash string) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.User, error)
}

type WalletRepository interface {
	Create(ctx context.Context, userID uuid.UUID, name, currency string) (*model.Wallet, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]model.Wallet, error)
	FindByID(ctx context.Context, userID, walletID uuid.UUID) (*model.Wallet, error)
	FindByIDForUpdate(ctx context.Context, q DBTX, userID, walletID uuid.UUID) (*model.Wallet, error)
	UpdateBalance(ctx context.Context, q DBTX, userID, walletID uuid.UUID, balance decimal.Decimal) error
	Delete(ctx context.Context, userID, walletID uuid.UUID) error
}

type TransactionRepository interface {
	Create(ctx context.Context, q DBTX, tx *model.Transaction) (*model.Transaction, error)
	ListByWallet(ctx context.Context, userID, walletID uuid.UUID, limit, offset int) ([]model.Transaction, error)
	ListByUser(ctx context.Context, userID uuid.UUID, from, to *time.Time) ([]model.Transaction, error)
}
