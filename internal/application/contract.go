package application

import (
	"context"

	"example.com/payment-gateway/internal/domain"
)

type PaymentRepository interface {
	Save(ctx context.Context, payment domain.Payment) error
	FindByID(ctx context.Context, id string) (domain.Payment, error)
}

type BankAuthorizer interface {
	Authorize(ctx context.Context, request AuthorizationRequest) (AuthorizationResult, error)
}

type AuthorizationRequest struct {
	CardNumber  string
	ExpiryMonth int
	ExpiryYear  int
	Currency    string
	Amount      int64
	CVV         string
}

type AuthorizationResult struct {
	Authorized        bool
	AuthorizationCode string
}
