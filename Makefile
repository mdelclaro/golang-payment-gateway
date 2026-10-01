.DEFAULT_GOAL := help

.PHONY: help fmt test test-coverage build run compose-up compose-down compose-logs

help:
	@echo "Available targets:"
	@echo "  fmt             Format Go files with gofmt"
	@echo "  test            Run all Go tests"
	@echo "  test-coverage   Run tests and show per-function and total coverage"
	@echo "  build           Compile all Go packages"
	@echo "  run             Run the payment gateway"
	@echo "  compose-up      Build and start the gateway and PostgreSQL"
	@echo "  compose-down    Stop the Docker Compose services"
	@echo "  compose-logs    Follow Docker Compose logs"

fmt:
	gofmt -w .

test:
	go test ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

build:
	go build ./...

run:
	go run ./cmd/payment-gateway

compose-up:
	docker compose up --build

compose-down:
	docker compose down

compose-logs:
	docker compose logs -f
