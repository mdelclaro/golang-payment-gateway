package bootstrap

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	applicationpayment "example.com/payment-gateway/internal/application/payment"
	"example.com/payment-gateway/internal/bank"
	"example.com/payment-gateway/internal/config"
	"example.com/payment-gateway/internal/httpapi"
	"example.com/payment-gateway/internal/migrations"
	"example.com/payment-gateway/internal/repository/postgres"
)

func Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()
	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	startupCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(startupCtx); err != nil {
		return err
	}
	if err := migrations.Migrate(startupCtx, db); err != nil {
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

	return serve(ctx, server)
}

func serve(ctx context.Context, server *http.Server) error {
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		_ = server.Close()
		return fmt.Errorf("shut down HTTP server: %w", err)
	}

	if err := <-serverErr; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("HTTP server stopped: %w", err)
	}

	return nil
}
