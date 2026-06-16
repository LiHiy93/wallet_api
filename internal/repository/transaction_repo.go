package repository

import (
	"context"
	"fmt"
	"time"

	"wallet-api/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionRepo struct {
	pool *pgxpool.Pool
}

func NewTransactionRepo(pool *pgxpool.Pool) *TransactionRepo { return &TransactionRepo{pool: pool} }

func (r *TransactionRepo) Create(ctx context.Context, q DBTX, tx *model.Transaction) (*model.Transaction, error) {
	const query = `INSERT INTO transactions (wallet_id, user_id, type, amount, description) VALUES ($1, $2, $3, $4, $5) RETURNING id, wallet_id, user_id, type, amount, description, created_at`
	var created model.Transaction
	if err := q.QueryRow(ctx, query, tx.WalletID, tx.UserID, tx.Type, tx.Amount, tx.Description).Scan(&created.ID, &created.WalletID, &created.UserID, &created.Type, &created.Amount, &created.Description, &created.CreatedAt); err != nil {
		return nil, fmt.Errorf("create transaction: %w", err)
	}
	return &created, nil
}

func (r *TransactionRepo) ListByWallet(ctx context.Context, userID, walletID uuid.UUID, limit, offset int) ([]model.Transaction, error) {
	const query = `SELECT id, wallet_id, user_id, type, amount, description, created_at FROM transactions WHERE user_id = $1 AND wallet_id = $2 ORDER BY created_at DESC LIMIT $3 OFFSET $4`
	rows, err := r.pool.Query(ctx, query, userID, walletID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list wallet transactions: %w", err)
	}
	defer rows.Close()
	return scanTransactions(rows)
}

func (r *TransactionRepo) ListByUser(ctx context.Context, userID uuid.UUID, from, to *time.Time) ([]model.Transaction, error) {
	query := `SELECT id, wallet_id, user_id, type, amount, description, created_at FROM transactions WHERE user_id = $1`
	args := []any{userID}
	if from != nil {
		args = append(args, *from)
		query += fmt.Sprintf(" AND created_at >= $%d", len(args))
	}
	if to != nil {
		args = append(args, *to)
		query += fmt.Sprintf(" AND created_at < $%d", len(args))
	}
	query += ` ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list user transactions: %w", err)
	}
	defer rows.Close()
	return scanTransactions(rows)
}

func scanTransactions(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}) ([]model.Transaction, error) {
	transactions := make([]model.Transaction, 0)
	for rows.Next() {
		var tx model.Transaction
		if err := rows.Scan(&tx.ID, &tx.WalletID, &tx.UserID, &tx.Type, &tx.Amount, &tx.Description, &tx.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan transaction: %w", err)
		}
		transactions = append(transactions, tx)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate transactions: %w", err)
	}
	return transactions, nil
}
