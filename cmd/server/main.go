package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"wallet-api/internal/config"
	"wallet-api/internal/db"
	"wallet-api/internal/handler"
	"wallet-api/internal/repository"
	"wallet-api/internal/service"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfgPath := ".env"
	if _, err := os.Stat(cfgPath); errors.Is(err, os.ErrNotExist) {
		cfgPath = ""
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		logger.Error("load config", "error", err)
		os.Exit(1)
	}

	if err := runMigrations(cfg.Migrations, cfg.DatabaseURL); err != nil {
		logger.Error("run migrations", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := db.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("connect database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	userRepo := repository.NewUserRepo(pool)
	walletRepo := repository.NewWalletRepo(pool)
	transactionRepo := repository.NewTransactionRepo(pool)
	txManager := repository.NewPostgresTxManager(pool)

	authService := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTTTL)
	walletService := service.NewWalletService(walletRepo)
	transactionService := service.NewTransactionService(txManager, walletRepo, transactionRepo)

	authHandler := handler.NewAuthHandler(authService)
	walletHandler := handler.NewWalletHandler(walletService)
	transactionHandler := handler.NewTransactionHandler(transactionService)

	router := handler.NewRouter(authHandler, walletHandler, transactionHandler, cfg.JWTSecret, logger)
	server := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("server started", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown", "error", err)
	}
}

func runMigrations(sourceURL, databaseURL string) error {
	m, err := migrate.New(sourceURL, databaseURL)
	if err != nil {
		return err
	}
	defer func() {
		_, _ = m.Close()
	}()
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}
