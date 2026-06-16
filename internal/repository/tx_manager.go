package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresTxManager struct {
	pool *pgxpool.Pool
}

func NewPostgresTxManager(pool *pgxpool.Pool) *PostgresTxManager {
	return &PostgresTxManager{pool: pool}
}

func (m *PostgresTxManager) WithinTx(ctx context.Context, fn func(q DBTX) error) error {
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := fn(tx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
