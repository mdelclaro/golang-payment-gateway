package bootstrap

import (
	"errors"
	"net/http"
	"time"

	"example.com/payment-gateway/internal/config"
	"example.com/payment-gateway/internal/httpapi"
)

func Run() error {
	cfg := config.Load()
	server := &http.Server{
		Addr:              cfg.HTTPAddress,
		Handler:           httpapi.NewRouter(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	err := server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return err
}
