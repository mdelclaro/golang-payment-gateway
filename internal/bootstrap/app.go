package bootstrap

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	applicationpayment "example.com/payment-gateway/internal/application/payment"
	"example.com/payment-gateway/internal/bank"
	"example.com/payment-gateway/internal/config"
	"example.com/payment-gateway/internal/httpapi"
	"example.com/payment-gateway/internal/repository/postgres"
)

func Run() error {
	cfg := config.Load()
	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return err
	}
	if err := postgres.Migrate(ctx, db); err != nil {
		return err
	}

	bankAuthorizer, err := bank.NewAuthorizer(bank.AuthorizerParams{BaseURL: cfg.BankURL})
	if err != nil {
		return err
	}

	paymentService := applicationpayment.NewService(applicationpayment.ServiceParams{
		Repository: postgres.NewPayments(db),
		Authorizer: bankAuthorizer,
	})

	server := &http.Server{
		Addr:              cfg.HTTPAddress,
		Handler:           httpapi.NewRouter(paymentService),
		ReadHeaderTimeout: 5 * time.Second,
	}

	err = server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return err
}
