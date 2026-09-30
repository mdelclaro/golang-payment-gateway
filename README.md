# Payment Gateway

Go implementation of the Checkout.com payment gateway assessment.

## Run

```powershell
go run ./cmd/payment-gateway
```

The API listens on `:8090` by default. Set `HTTP_ADDR` to change the address.

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

`amount` is a positive integer. A successful request returns `201 Created` with the generated payment ID and authorization result. The response includes `cardLastFour`; it never includes the full card number or CVV. Invalid requests return `400 Bad Request`; processing failures return `500 Internal Server Error`.

## Structure

- `cmd/payment-gateway` is the process entry point.
- `internal/bootstrap` wires application components and starts the server.
- `internal/config` loads runtime configuration.
- `internal/domain` contains payment concepts and business rules.
- `internal/application` contains use cases and the interfaces they need.
- `internal/httpapi` translates between HTTP and application behavior.
- `internal/repository` will contain the PostgreSQL repository implementation.
- `internal/database` will manage the PostgreSQL connection pool.

Card number and CVV are authorization inputs only; payment records contain the last four digits, never the full card number or CVV.
