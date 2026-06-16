package repository

import (
	"context"
	"errors"
	"fmt"

	"wallet-api/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

var ErrNotFound = errors.New("not found")

type WalletRepo struct {
	pool *pgxpool.Pool
}

func NewWalletRepo(pool *pgxpool.Pool) *WalletRepo { return &WalletRepo{pool: pool} }

func (r *WalletRepo) Create(ctx context.Context, userID uuid.UUID, name, currency string) (*model.Wallet, error) {
	const query = `INSERT INTO wallets (user_id, name, currency) VALUES ($1, $2, $3) RETURNING id, user_id, name, balance, currency, created_at`
	var wallet model.Wallet
	if err := r.pool.QueryRow(ctx, query, userID, name, currency).Scan(&wallet.ID, &wallet.UserID, &wallet.Name, &wallet.Balance, &wallet.Currency, &wallet.CreatedAt); err != nil {
		return nil, fmt.Errorf("create wallet: %w", err)
	}
	return &wallet, nil
}

func (r *WalletRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]model.Wallet, error) {
	const query = `SELECT id, user_id, name, balance, currency, created_at FROM wallets WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list wallets: %w", err)
	}
	defer rows.Close()

	wallets := make([]model.Wallet, 0)
	for rows.Next() {
		var wallet model.Wallet
		if err := rows.Scan(&wallet.ID, &wallet.UserID, &wallet.Name, &wallet.Balance, &wallet.Currency, &wallet.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan wallet: %w", err)
		}
		wallets = append(wallets, wallet)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate wallets: %w", err)
	}
	return wallets, nil
}

func (r *WalletRepo) FindByID(ctx context.Context, userID, walletID uuid.UUID) (*model.Wallet, error) {
	return scanWallet(ctx, r.pool, `SELECT id, user_id, name, balance, currency, created_at FROM wallets WHERE id = $1 AND user_id = $2`, walletID, userID)
}

func (r *WalletRepo) FindByIDForUpdate(ctx context.Context, q DBTX, userID, walletID uuid.UUID) (*model.Wallet, error) {
	return scanWallet(ctx, q, `SELECT id, user_id, name, balance, currency, created_at FROM wallets WHERE id = $1 AND user_id = $2 FOR UPDATE`, walletID, userID)
}

func (r *WalletRepo) UpdateBalance(ctx context.Context, q DBTX, userID, walletID uuid.UUID, balance decimal.Decimal) error {
	const query = `UPDATE wallets SET balance = $1 WHERE id = $2 AND user_id = $3`
	tag, err := q.Exec(ctx, query, balance, walletID, userID)
	if err != nil {
		return fmt.Errorf("update wallet balance: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *WalletRepo) Delete(ctx context.Context, userID, walletID uuid.UUID) error {
	const query = `DELETE FROM wallets WHERE id = $1 AND user_id = $2`
	tag, err := r.pool.Exec(ctx, query, walletID, userID)
	if err != nil {
		return fmt.Errorf("delete wallet: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func scanWallet(ctx context.Context, q DBTX, query string, args ...any) (*model.Wallet, error) {
	var wallet model.Wallet
	if err := q.QueryRow(ctx, query, args...).Scan(&wallet.ID, &wallet.UserID, &wallet.Name, &wallet.Balance, &wallet.Currency, &wallet.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("scan wallet: %w", err)
	}
	return &wallet, nil
}
