package payment

import (
	"context"

	"example.com/payment-gateway/internal/domain"
)

type Repository interface {
	Create(ctx context.Context, payment domain.Payment) (domain.Payment, error)
	FindByID(ctx context.Context, id int64) (domain.Payment, error)
}

type BankAuthorizer interface {
	Authorize(ctx context.Context, request AuthorizationRequest) (AuthorizationResult, error)
}

type Service interface {
	CreatePayment(ctx context.Context, input CreatePaymentInput) (domain.Payment, error)
	GetPayment(ctx context.Context, id int64) (domain.Payment, error)
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
