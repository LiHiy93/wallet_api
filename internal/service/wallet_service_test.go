package service_test

import (
	"context"
	"testing"

	"wallet-api/internal/model"
	"wallet-api/internal/repository"
	"wallet-api/internal/service"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type walletRepoMock struct{ mock.Mock }

func (m *walletRepoMock) Create(ctx context.Context, userID uuid.UUID, name, currency string) (*model.Wallet, error) {
	args := m.Called(ctx, userID, name, currency)
	wallet, _ := args.Get(0).(*model.Wallet)
	return wallet, args.Error(1)
}

func (m *walletRepoMock) ListByUser(ctx context.Context, userID uuid.UUID) ([]model.Wallet, error) {
	args := m.Called(ctx, userID)
	wallets, _ := args.Get(0).([]model.Wallet)
	return wallets, args.Error(1)
}

func (m *walletRepoMock) FindByID(ctx context.Context, userID, walletID uuid.UUID) (*model.Wallet, error) {
	args := m.Called(ctx, userID, walletID)
	wallet, _ := args.Get(0).(*model.Wallet)
	return wallet, args.Error(1)
}

func (m *walletRepoMock) FindByIDForUpdate(ctx context.Context, q repository.DBTX, userID, walletID uuid.UUID) (*model.Wallet, error) {
	args := m.Called(ctx, q, userID, walletID)
	wallet, _ := args.Get(0).(*model.Wallet)
	return wallet, args.Error(1)
}

func (m *walletRepoMock) UpdateBalance(ctx context.Context, q repository.DBTX, userID, walletID uuid.UUID, balance decimal.Decimal) error {
	args := m.Called(ctx, q, userID, walletID, balance)
	return args.Error(0)
}

func (m *walletRepoMock) Delete(ctx context.Context, userID, walletID uuid.UUID) error {
	args := m.Called(ctx, userID, walletID)
	return args.Error(0)
}

func TestWalletService_Create_DefaultCurrency(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	walletID := uuid.New()
	repo := new(walletRepoMock)
	expected := &model.Wallet{ID: walletID, UserID: userID, Name: "Main", Currency: "RUB", Balance: decimal.Zero}
	repo.On("Create", ctx, userID, "Main", "RUB").Return(expected, nil).Once()

	svc := service.NewWalletService(repo)
	got, err := svc.Create(ctx, userID, " Main ", "")

	require.NoError(t, err)
	require.Equal(t, expected, got)
	repo.AssertExpectations(t)
}

func TestWalletService_Delete_RejectsNonZeroBalance(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	walletID := uuid.New()
	repo := new(walletRepoMock)
	repo.On("FindByID", ctx, userID, walletID).Return(&model.Wallet{ID: walletID, UserID: userID, Balance: decimal.NewFromInt(100)}, nil).Once()

	svc := service.NewWalletService(repo)
	err := svc.Delete(ctx, userID, walletID)

	require.ErrorIs(t, err, service.ErrWalletBalanceNotZero)
	repo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything, mock.Anything)
	repo.AssertExpectations(t)
}

func TestWalletService_Delete_ZeroBalance(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	walletID := uuid.New()
	repo := new(walletRepoMock)
	repo.On("FindByID", ctx, userID, walletID).Return(&model.Wallet{ID: walletID, UserID: userID, Balance: decimal.Zero}, nil).Once()
	repo.On("Delete", ctx, userID, walletID).Return(nil).Once()

	svc := service.NewWalletService(repo)
	err := svc.Delete(ctx, userID, walletID)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}
