package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"wallet-api/internal/model"
	"wallet-api/internal/repository"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransactionService struct {
	txManager    repository.TxManager
	wallets      repository.WalletRepository
	transactions repository.TransactionRepository
}

func NewTransactionService(txManager repository.TxManager, wallets repository.WalletRepository, transactions repository.TransactionRepository) *TransactionService {
	return &TransactionService{txManager: txManager, wallets: wallets, transactions: transactions}
}

func (s *TransactionService) Create(ctx context.Context, userID, walletID uuid.UUID, txType model.TransactionType, amount decimal.Decimal, description string) (*model.Transaction, error) {
	description = strings.TrimSpace(description)
	if err := validateTransaction(txType, amount); err != nil {
		return nil, err
	}

	var created *model.Transaction
	err := s.txManager.WithinTx(ctx, func(q repository.DBTX) error {
		wallet, err := s.wallets.FindByIDForUpdate(ctx, q, userID, walletID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return ErrWalletNotFound
			}
			return err
		}

		newBalance := wallet.Balance
		switch txType {
		case model.TransactionTypeIncome:
			newBalance = wallet.Balance.Add(amount)
		case model.TransactionTypeExpense:
			if wallet.Balance.LessThan(amount) {
				return ErrInsufficientFunds
			}
			newBalance = wallet.Balance.Sub(amount)
		case model.TransactionTypeTransfer:
			return ErrInvalidInput
		default:
			return ErrInvalidInput
		}

		if err := s.wallets.UpdateBalance(ctx, q, userID, walletID, newBalance); err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return ErrWalletNotFound
			}
			return err
		}

		created, err = s.transactions.Create(ctx, q, &model.Transaction{
			WalletID:    walletID,
			UserID:      userID,
			Type:        txType,
			Amount:      amount,
			Description: description,
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s *TransactionService) ListByWallet(ctx context.Context, userID, walletID uuid.UUID, page, limit int) ([]model.Transaction, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	if _, err := s.wallets.FindByID(ctx, userID, walletID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrWalletNotFound
		}
		return nil, err
	}
	offset := (page - 1) * limit
	return s.transactions.ListByWallet(ctx, userID, walletID, limit, offset)
}

func (s *TransactionService) ListByUser(ctx context.Context, userID uuid.UUID, from, to *time.Time) ([]model.Transaction, error) {
	return s.transactions.ListByUser(ctx, userID, from, to)
}

func validateTransaction(txType model.TransactionType, amount decimal.Decimal) error {
	if !amount.GreaterThan(decimal.Zero) {
		return ErrInvalidInput
	}
	if !amount.Equal(amount.Round(2)) {
		return ErrInvalidInput
	}
	if txType != model.TransactionTypeIncome && txType != model.TransactionTypeExpense && txType != model.TransactionTypeTransfer {
		return ErrInvalidInput
	}
	return nil
}
