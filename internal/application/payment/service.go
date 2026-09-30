package payment

import (
	"context"

	"example.com/payment-gateway/internal/domain"
	"github.com/go-playground/validator/v10"
)

var inputValidator = validator.New()

type ServiceParams struct {
	Repository Repository
	Authorizer BankAuthorizer
}

type service struct {
	repository Repository
	authorizer BankAuthorizer
}

func NewService(params ServiceParams) Service {
	return &service{
		repository: params.Repository,
		authorizer: params.Authorizer,
	}
}

func (s *service) CreatePayment(ctx context.Context, input CreatePaymentInput) (domain.Payment, error) {
	if err := inputValidator.Struct(input); err != nil {
		return domain.Payment{}, ErrInvalidPayment
	}

	result, err := s.authorizer.Authorize(ctx, AuthorizationRequest{
		CardNumber:  input.CardNumber,
		ExpiryMonth: input.ExpiryMonth,
		ExpiryYear:  input.ExpiryYear,
		Currency:    input.Currency,
		Amount:      input.Amount,
		CVV:         input.CVV,
	})
	if err != nil {
		return domain.Payment{}, err
	}

	cardLastFour := input.CardNumber[len(input.CardNumber)-4:]

	status := domain.PaymentDeclined
	if result.Authorized {
		status = domain.PaymentAuthorized
	}

	payment := domain.Payment{
		Status:            status,
		CardLastFour:      cardLastFour,
		ExpiryMonth:       input.ExpiryMonth,
		ExpiryYear:        input.ExpiryYear,
		Currency:          input.Currency,
		Amount:            input.Amount,
		AuthorizationCode: result.AuthorizationCode,
	}

	return s.repository.Create(ctx, payment)
}

func (s *service) GetPayment(ctx context.Context, id int64) (domain.Payment, error) {
	return s.repository.FindByID(ctx, id)
}
