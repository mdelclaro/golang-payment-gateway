package payment

import (
	"context"
	"errors"
	"testing"

	"example.com/payment-gateway/internal/domain"
)

type repositoryStub struct {
	created     domain.Payment
	createCalls int
	createErr   error
	found       domain.Payment
	findCalls   int
	findID      int64
	findErr     error
}

func (r *repositoryStub) Create(_ context.Context, payment domain.Payment) (domain.Payment, error) {
	r.created = payment
	r.createCalls++
	if r.createErr != nil {
		return domain.Payment{}, r.createErr
	}
	payment.ID = 42
	return payment, nil
}

func (r *repositoryStub) FindByID(_ context.Context, id int64) (domain.Payment, error) {
	r.findCalls++
	r.findID = id
	return r.found, r.findErr
}

type authorizerStub struct {
	request AuthorizationRequest
	result  AuthorizationResult
	err     error
	calls   int
}

func (a *authorizerStub) Authorize(_ context.Context, request AuthorizationRequest) (AuthorizationResult, error) {
	a.calls++
	a.request = request
	return a.result, a.err
}

func TestCreatePaymentAuthorizedPersistsOnlyCardLastFour(t *testing.T) {
	repository := &repositoryStub{}
	authorizer := &authorizerStub{result: AuthorizationResult{
		Authorized:        true,
		AuthorizationCode: "auth-123",
	}}
	service := NewService(ServiceParams{Repository: repository, Authorizer: authorizer})

	got, err := service.CreatePayment(context.Background(), validInput())
	if err != nil {
		t.Fatalf("CreatePayment() error = %v", err)
	}
	if got.ID != 42 {
		t.Errorf("payment ID = %d, want 42", got.ID)
	}
	if got.Status != domain.PaymentAuthorized {
		t.Errorf("payment status = %q, want %q", got.Status, domain.PaymentAuthorized)
	}
	if got.CardLastFour != "1111" {
		t.Errorf("card last four = %q, want 1111", got.CardLastFour)
	}
	if got.AuthorizationCode != "auth-123" {
		t.Errorf("authorization code = %q, want auth-123", got.AuthorizationCode)
	}
	if repository.createCalls != 1 {
		t.Errorf("repository Create calls = %d, want 1", repository.createCalls)
	}
	if repository.created.CardLastFour != "1111" {
		t.Errorf("persisted card last four = %q, want 1111", repository.created.CardLastFour)
	}
	if authorizer.calls != 1 {
		t.Errorf("authorizer calls = %d, want 1", authorizer.calls)
	}
	if authorizer.request.CardNumber != validInput().CardNumber || authorizer.request.CVV != validInput().CVV {
		t.Error("authorizer did not receive card authorization inputs")
	}
}

func TestCreatePaymentDeclinedIsPersisted(t *testing.T) {
	repository := &repositoryStub{}
	authorizer := &authorizerStub{result: AuthorizationResult{Authorized: false}}
	service := NewService(ServiceParams{Repository: repository, Authorizer: authorizer})

	got, err := service.CreatePayment(context.Background(), validInput())
	if err != nil {
		t.Fatalf("CreatePayment() error = %v", err)
	}
	if got.Status != domain.PaymentDeclined {
		t.Errorf("payment status = %q, want %q", got.Status, domain.PaymentDeclined)
	}
	if repository.createCalls != 1 {
		t.Errorf("repository Create calls = %d, want 1", repository.createCalls)
	}
}

func TestCreatePaymentRejectsInvalidInputBeforeCallingDependencies(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*CreatePaymentInput)
	}{
		{name: "non-numeric card number", mutate: func(input *CreatePaymentInput) { input.CardNumber = "4111abcd11111111" }},
		{name: "card number too short", mutate: func(input *CreatePaymentInput) { input.CardNumber = "4111" }},
		{name: "month below range", mutate: func(input *CreatePaymentInput) { input.ExpiryMonth = 0 }},
		{name: "month above range", mutate: func(input *CreatePaymentInput) { input.ExpiryMonth = 13 }},
		{name: "year is zero", mutate: func(input *CreatePaymentInput) { input.ExpiryYear = 0 }},
		{name: "currency missing", mutate: func(input *CreatePaymentInput) { input.Currency = "" }},
		{name: "amount is zero", mutate: func(input *CreatePaymentInput) { input.Amount = 0 }},
		{name: "CVV missing", mutate: func(input *CreatePaymentInput) { input.CVV = "" }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository := &repositoryStub{}
			authorizer := &authorizerStub{}
			service := NewService(ServiceParams{Repository: repository, Authorizer: authorizer})
			input := validInput()
			tt.mutate(&input)

			_, err := service.CreatePayment(context.Background(), input)
			if !errors.Is(err, ErrInvalidPayment) {
				t.Errorf("CreatePayment() error = %v, want %v", err, ErrInvalidPayment)
			}
			if authorizer.calls != 0 {
				t.Errorf("authorizer calls = %d, want 0", authorizer.calls)
			}
			if repository.createCalls != 0 {
				t.Errorf("repository Create calls = %d, want 0", repository.createCalls)
			}
		})
	}
}

func TestCreatePaymentDoesNotPersistWhenAuthorizationFails(t *testing.T) {
	wantErr := errors.New("authorization unavailable")
	repository := &repositoryStub{}
	authorizer := &authorizerStub{err: wantErr}
	service := NewService(ServiceParams{Repository: repository, Authorizer: authorizer})

	_, err := service.CreatePayment(context.Background(), validInput())
	if !errors.Is(err, wantErr) {
		t.Errorf("CreatePayment() error = %v, want %v", err, wantErr)
	}
	if repository.createCalls != 0 {
		t.Errorf("repository Create calls = %d, want 0", repository.createCalls)
	}
}

func TestCreatePaymentReturnsRepositoryError(t *testing.T) {
	wantErr := errors.New("database unavailable")
	repository := &repositoryStub{createErr: wantErr}
	authorizer := &authorizerStub{result: AuthorizationResult{Authorized: true}}
	service := NewService(ServiceParams{Repository: repository, Authorizer: authorizer})

	_, err := service.CreatePayment(context.Background(), validInput())
	if !errors.Is(err, wantErr) {
		t.Errorf("CreatePayment() error = %v, want %v", err, wantErr)
	}
}

func validInput() CreatePaymentInput {
	return CreatePaymentInput{
		CardNumber:  "4111111111111111",
		ExpiryMonth: 12,
		ExpiryYear:  2030,
		Currency:    "USD",
		Amount:      2500,
		CVV:         "123",
	}
}
