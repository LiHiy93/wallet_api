package service

import "errors"

var (
	ErrInvalidInput         = errors.New("invalid input")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrEmailAlreadyExists   = errors.New("email already exists")
	ErrWalletNotFound       = errors.New("wallet not found")
	ErrTransactionNotFound  = errors.New("transaction not found")
	ErrInsufficientFunds    = errors.New("insufficient funds")
	ErrWalletBalanceNotZero = errors.New("wallet balance must be zero")
)
