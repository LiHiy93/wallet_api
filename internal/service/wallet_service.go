package service

import (
	"context"
	"errors"
	"strings"

	"wallet-api/internal/model"
	"wallet-api/internal/repository"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type WalletService struct {
	wallets repository.WalletRepository
}

func NewWalletService(wallets repository.WalletRepository) *WalletService {
	return &WalletService{wallets: wallets}
}

func (s *WalletService) Create(ctx context.Context, userID uuid.UUID, name, currency string) (*model.Wallet, error) {
	name = strings.TrimSpace(name)
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == "" {
		currency = "RUB"
	}
	if name == "" || len(currency) != 3 {
		return nil, ErrInvalidInput
	}
	return s.wallets.Create(ctx, userID, name, currency)
}

func (s *WalletService) List(ctx context.Context, userID uuid.UUID) ([]model.Wallet, error) {
	return s.wallets.ListByUser(ctx, userID)
}

func (s *WalletService) Get(ctx context.Context, userID, walletID uuid.UUID) (*model.Wallet, error) {
	wallet, err := s.wallets.FindByID(ctx, userID, walletID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrWalletNotFound
		}
		return nil, err
	}
	return wallet, nil
}

func (s *WalletService) Delete(ctx context.Context, userID, walletID uuid.UUID) error {
	wallet, err := s.Get(ctx, userID, walletID)
	if err != nil {
		return err
	}
	if !wallet.Balance.Equal(decimal.Zero) {
		return ErrWalletBalanceNotZero
	}
	if err := s.wallets.Delete(ctx, userID, walletID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrWalletNotFound
		}
		return err
	}
	return nil
}
