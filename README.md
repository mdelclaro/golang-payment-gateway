# Payment Gateway

Go implementation of the Checkout.com payment gateway assessment.

## Run with Docker Compose

Start the gateway and PostgreSQL:

```powershell
make compose-up
```

The gateway listens on `http://localhost:8090`, and PostgreSQL data is kept in the `payment_data` volume. The assessment's bank API is a separate Compose project; from its repository root, run `docker compose up -d bank_simulator` so it listens on the host at port `8080`. The gateway Compose setup routes to it through `host.docker.internal`. To use a different bank API address, set `BANK_API_URL` in your shell before starting Compose.

Swagger UI is available at <http://localhost:8090/swagger/index.html>. The OpenAPI document is served at <http://localhost:8090/openapi.yaml>.

## Run locally

Start PostgreSQL in the background:

```powershell
docker compose up -d postgres
```

Make sure the assessment's bank API is running at `http://localhost:8080`, then start the gateway:

```powershell
make run
```

The API listens on `:8090` by default. Start PostgreSQL and the bank API before starting the gateway. The application creates the payments table on startup.

| Environment variable | Default | Purpose |
| --- | --- | --- |
| `HTTP_ADDR` | `:8090` | HTTP listener address |
| `BANK_API_URL` | `http://localhost:8080` | Bank API base URL |
| `DATABASE_URL` | `postgres://payment:payment@localhost:5432/payment?sslmode=disable` | PostgreSQL connection string |

Swagger UI is available at <http://localhost:8090/swagger/index.html>. The OpenAPI document is served at <http://localhost:8090/openapi.yaml>.

## Development commands

The Makefile provides shortcuts for common tasks. Run `make` or `make help` to list them. GNU Make, Go, and Docker Compose are required for the corresponding targets.

Windows does not include GNU Make by default. If `make` is unavailable, run the underlying commands directly, such as `go test ./...`, `go build ./...`, `go run ./cmd/payment-gateway`, and `docker compose up --build`.

| Command | Action |
| --- | --- |
| `make fmt` | Format Go files with `gofmt -w .` |
| `make test` | Run all Go tests with `go test ./...` |
| `make test-coverage` | Run all Go tests with coverage summaries |
| `make build` | Compile all Go packages |
| `make run` | Start the gateway locally |
| `make compose-up` | Build and start the gateway and PostgreSQL |
| `make compose-down` | Stop the Compose services |
| `make compose-logs` | Follow the Compose logs |

## End-to-end tests

The E2E suite sends HTTP requests to a running gateway and uses the real PostgreSQL and assessment bank simulator services. Start the bank simulator from its repository in one terminal:

```powershell
docker compose up -d bank_simulator
```

Then start the gateway stack in another terminal:

```powershell
make compose-up
```

In a second terminal, run:

```powershell
make test-e2e
```

The suite defaults to `http://localhost:8090`; set `PAYMENT_GATEWAY_URL` to target another running gateway. It creates test payment rows in the configured database. The suite covers approval and retrieval, decline persistence, validation errors, bank unavailability, missing IDs, and invalid IDs. These tests are opt-in and are not included in `go test ./...`.

## Payments API

The [OpenAPI (Swagger) specification](internal/httpapi/docs/openapi.yaml) documents the available endpoints, examples, and request and response schemas. Payment operations are versioned under `/v1`; future breaking API changes should use a new major prefix such as `/v2`. The operational `/health` endpoint remains unversioned.

`POST /v1/payments` creates and authorizes a payment. The request uses camelCase JSON fields:

```json
{
  "cardNumber": "4111111111111111",
  "expiryMonth": 12,
  "expiryYear": 2030,
  "currency": "USD",
  "amount": 2500,
  "cvv": "123"
}
```

`amount` is a positive integer in the currency's minor units (for example, USD 10.50 is `1050`). Supported currencies are GBP, EUR, and USD. Card numbers must contain 14–19 digits, expiry must be in the future, and CVV must contain 3–4 digits. A valid request returns `201 Created` with the generated payment ID and authorization result, including declined payments. The response includes `cardLastFour`; it never includes the full card number or CVV. Invalid requests return `400 Bad Request`, unavailable authorization returns `503 Service Unavailable`, and other processing failures return `500 Internal Server Error`.

`GET /v1/payments/{id}` retrieves a previously created payment. It returns `200 OK` with the same payment details, `400 Bad Request` for an invalid ID, or `404 Not Found` when no payment has that ID.

## Design decisions and assumptions

- The API accepts GBP, EUR, and USD as its three supported currency codes, within the assessment's limit of no more than three ISO currency codes.
- Payment input is validated before contacting the bank. Invalid input is rejected with `400 Bad Request` and is not sent for authorization.
- A declined authorization still creates a payment record and returns `201 Created`; its payment status is `Declined`.
- The assessment permits an in-memory test double for storage. This implementation uses PostgreSQL so payment records survive restarts and IDs are generated by the database.
- Full card numbers and CVVs are used only for the authorization request. Only the last four card digits and non-sensitive payment details are stored and returned.
- Authentication and idempotency are outside the initial assessment scope.

## Structure

- `cmd/payment-gateway` is the process entry point.
- `internal/bootstrap` wires application components and starts the server.
- `internal/config` loads runtime configuration, including the bank API URL.
- `internal/bank` sends authorization requests to the bank and maps its responses.
- `internal/domain` contains payment concepts and business rules.
- `internal/application` contains use cases and the interfaces they need.
- `internal/httpapi` translates between HTTP and application behavior.
- `internal/migrations` embeds and applies database schema migrations at startup.
- `internal/repository/postgres` stores payment records in PostgreSQL and uses database-generated IDs.

Card number and CVV are authorization inputs only; payment records contain the last four digits, never the full card number or CVV.
