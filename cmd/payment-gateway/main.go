package main

import (
	"log/slog"
	"os"

	"example.com/payment-gateway/internal/bootstrap"
)

func main() {
	if err := bootstrap.Run(); err != nil {
		slog.Error("application stopped", "error", err)
		os.Exit(1)
	}
}
