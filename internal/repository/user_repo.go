package repository

import (
	"context"
	"errors"
	"fmt"

	"wallet-api/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrDuplicateEmail = errors.New("email already exists")

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo { return &UserRepo{pool: pool} }

func (r *UserRepo) Create(ctx context.Context, email, passwordHash string) (*model.User, error) {
	const query = `INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id, email, password_hash, created_at`
	var user model.User
	if err := r.pool.QueryRow(ctx, query, email, passwordHash).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrDuplicateEmail
		}
		return nil, fmt.Errorf("create user: %w", err)
	}
	return &user, nil
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	const query = `SELECT id, email, password_hash, created_at FROM users WHERE email = $1`
	var user model.User
	if err := r.pool.QueryRow(ctx, query, email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	return &user, nil
}

func (r *UserRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	const query = `SELECT id, email, password_hash, created_at FROM users WHERE id = $1`
	var user model.User
	if err := r.pool.QueryRow(ctx, query, id).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return &user, nil
}
