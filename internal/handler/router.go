package handler

import (
	"log/slog"
	"net/http"

	"wallet-api/internal/middleware"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func NewRouter(auth *AuthHandler, wallets *WalletHandler, transactions *TransactionHandler, jwtSecret string, logger *slog.Logger) http.Handler {
	r := chi.NewRouter()
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.RequestLogger(logger))

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", auth.Register)
			r.Post("/login", auth.Login)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.JWTAuth(jwtSecret))
			r.Get("/wallets", wallets.List)
			r.Post("/wallets", wallets.Create)
			r.Get("/wallets/{id}", wallets.Get)
			r.Delete("/wallets/{id}", wallets.Delete)
			r.Get("/wallets/{id}/transactions", transactions.ListByWallet)
			r.Post("/wallets/{id}/transactions", transactions.Create)
			r.Get("/transactions", transactions.ListByUser)
		})
	})

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	return r
}
