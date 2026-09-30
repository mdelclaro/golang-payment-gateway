package config

import "os"

type Config struct {
	HTTPAddress string
	BankURL     string
	DatabaseURL string
}

func Load() Config {
	address := os.Getenv("HTTP_ADDR")
	if address == "" {
		address = ":8090"
	}

	bankURL := os.Getenv("BANK_API_URL")
	if bankURL == "" {
		bankURL = "http://localhost:8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://payment:payment@localhost:5432/payment?sslmode=disable"
	}

	return Config{HTTPAddress: address, BankURL: bankURL, DatabaseURL: databaseURL}
}
