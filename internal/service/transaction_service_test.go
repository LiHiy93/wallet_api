package service_test

import (
	"context"
	"time"

	"testing"

	"wallet-api/internal/model"
	"wallet-api/internal/repository"
	"wallet-api/internal/service"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type txManagerMock struct{ mock.Mock }

func (m *txManagerMock) WithinTx(ctx context.Context, fn func(q repository.DBTX) error) error {
	args := m.Called(ctx, fn)
	if err := args.Error(0); err != nil {
		return err
	}
	return fn(nil)
}

type transactionRepoMock struct{ mock.Mock }

func (m *transactionRepoMock) Create(ctx context.Context, q repository.DBTX, tx *model.Transaction) (*model.Transaction, error) {
	args := m.Called(ctx, q, tx)
	created, _ := args.Get(0).(*model.Transaction)
	return created, args.Error(1)
}

func (m *transactionRepoMock) ListByWallet(ctx context.Context, userID, walletID uuid.UUID, limit, offset int) ([]model.Transaction, error) {
	args := m.Called(ctx, userID, walletID, limit, offset)
	transactions, _ := args.Get(0).([]model.Transaction)
	return transactions, args.Error(1)
}

func (m *transactionRepoMock) ListByUser(ctx context.Context, userID uuid.UUID, from, to *time.Time) ([]model.Transaction, error) {
	args := m.Called(ctx, userID, from, to)
	transactions, _ := args.Get(0).([]model.Transaction)
	return transactions, args.Error(1)
}

func TestTransactionService_CreateIncome_UpdatesBalanceAndCreatesTransaction(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	walletID := uuid.New()
	wallets := new(walletRepoMock)
	transactions := new(transactionRepoMock)
	txManager := new(txManagerMock)
	txManager.On("WithinTx", ctx, mock.Anything).Return(nil).Once()

	wallets.On("FindByIDForUpdate", ctx, mock.Anything, userID, walletID).
		Return(&model.Wallet{ID: walletID, UserID: userID, Balance: decimal.NewFromInt(100)}, nil).Once()
	wallets.On("UpdateBalance", ctx, mock.Anything, userID, walletID, decimal.NewFromInt(150)).Return(nil).Once()

	created := &model.Transaction{ID: uuid.New(), WalletID: walletID, UserID: userID, Type: model.TransactionTypeIncome, Amount: decimal.NewFromInt(50), Description: "salary"}
	transactions.On("Create", ctx, mock.Anything, mock.MatchedBy(func(tx *model.Transaction) bool {
		return tx.WalletID == walletID && tx.UserID == userID && tx.Type == model.TransactionTypeIncome && tx.Amount.Equal(decimal.NewFromInt(50)) && tx.Description == "salary"
	})).Return(created, nil).Once()

	svc := service.NewTransactionService(txManager, wallets, transactions)
	got, err := svc.Create(ctx, userID, walletID, model.TransactionTypeIncome, decimal.NewFromInt(50), " salary ")

	require.NoError(t, err)
	require.Equal(t, created, got)
	wallets.AssertExpectations(t)
	transactions.AssertExpectations(t)
	txManager.AssertExpectations(t)
}

func TestTransactionService_CreateExpense_RejectsInsufficientFunds(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	walletID := uuid.New()
	wallets := new(walletRepoMock)
	transactions := new(transactionRepoMock)
	txManager := new(txManagerMock)
	txManager.On("WithinTx", ctx, mock.Anything).Return(nil).Once()
	wallets.On("FindByIDForUpdate", ctx, mock.Anything, userID, walletID).
		Return(&model.Wallet{ID: walletID, UserID: userID, Balance: decimal.NewFromInt(10)}, nil).Once()

	svc := service.NewTransactionService(txManager, wallets, transactions)
	got, err := svc.Create(ctx, userID, walletID, model.TransactionTypeExpense, decimal.NewFromInt(20), "food")

	require.Nil(t, got)
	require.ErrorIs(t, err, service.ErrInsufficientFunds)
	wallets.AssertNotCalled(t, "UpdateBalance", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	transactions.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
	wallets.AssertExpectations(t)
	txManager.AssertExpectations(t)
}

func TestTransactionService_ListByWallet_UsesPagination(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	walletID := uuid.New()
	wallets := new(walletRepoMock)
	transactions := new(transactionRepoMock)
	txManager := new(txManagerMock)
	wallets.On("FindByID", ctx, userID, walletID).Return(&model.Wallet{ID: walletID, UserID: userID}, nil).Once()
	expected := []model.Transaction{{ID: uuid.New(), WalletID: walletID, UserID: userID}}
	transactions.On("ListByWallet", ctx, userID, walletID, 20, 20).Return(expected, nil).Once()

	svc := service.NewTransactionService(txManager, wallets, transactions)
	got, err := svc.ListByWallet(ctx, userID, walletID, 2, 20)

	require.NoError(t, err)
	require.Equal(t, expected, got)
	wallets.AssertExpectations(t)
	transactions.AssertExpectations(t)
}
