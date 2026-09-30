package config

import "os"

type Config struct {
	HTTPAddress string
}

func Load() Config {
	address := os.Getenv("HTTP_ADDR")
	if address == "" {
		address = ":8090"
	}

	return Config{HTTPAddress: address}
}
