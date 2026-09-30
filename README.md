# Payment Gateway

Go implementation of the Checkout.com payment gateway assessment.

## Run with Docker Compose

Start the gateway and PostgreSQL:

```powershell
docker compose up --build
```

The gateway listens on `http://localhost:8090`, and PostgreSQL data is kept in the `payment_data` volume. The bank API must be running on the host at port `8080`; Compose routes the gateway to it through `host.docker.internal`. Set `BANK_API_URL` in `docker-compose.yml` if it uses another address.

## Run locally

```powershell
go run ./cmd/payment-gateway
```

The API listens on `:8090` by default. Set `HTTP_ADDR` to change the address. It expects the bank API at `http://localhost:8080`; set `BANK_API_URL` to use another address. PostgreSQL is expected at `postgres://payment:payment@localhost:5432/payment?sslmode=disable`; set `DATABASE_URL` to use another connection string. The application creates the payments table on startup.

## Payments API

`POST /payments` creates and authorizes a payment. The request uses camelCase JSON fields:

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

`amount` is a positive integer in the currency's minor units (for example, USD 10.50 is `1050`). Supported currencies are GBP, EUR, and USD. Card numbers must contain 14–19 digits, expiry must be in the future, and CVV must contain 3–4 digits. A successful request returns `201 Created` with the generated payment ID and authorization result. The response includes `cardLastFour`; it never includes the full card number or CVV. Invalid requests return `400 Bad Request`; processing failures return `500 Internal Server Error`.

`GET /payments/{id}` retrieves a previously created payment. It returns `200 OK` with the same payment details, `400 Bad Request` for an invalid ID, or `404 Not Found` when no payment has that ID.

## Structure

- `cmd/payment-gateway` is the process entry point.
- `internal/bootstrap` wires application components and starts the server.
- `internal/config` loads runtime configuration, including the bank API URL.
- `internal/bank` sends authorization requests to the bank and maps its responses.
- `internal/domain` contains payment concepts and business rules.
- `internal/application` contains use cases and the interfaces they need.
- `internal/httpapi` translates between HTTP and application behavior.
- `internal/repository/postgres` stores payment records in PostgreSQL and uses database-generated IDs.

Card number and CVV are authorization inputs only; payment records contain the last four digits, never the full card number or CVV.
