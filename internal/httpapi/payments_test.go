package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	applicationpayment "example.com/payment-gateway/internal/application/payment"
	"example.com/payment-gateway/internal/domain"
)

type paymentServiceStub struct {
	input   applicationpayment.CreatePaymentInput
	created domain.Payment
	err     error
}

func (s *paymentServiceStub) CreatePayment(_ context.Context, input applicationpayment.CreatePaymentInput) (domain.Payment, error) {
	s.input = input
	return s.created, s.err
}

func (*paymentServiceStub) GetPayment(context.Context, int64) (domain.Payment, error) {
	return domain.Payment{}, nil
}

func TestCreatePaymentReturnsSanitizedPayment(t *testing.T) {
	service := &paymentServiceStub{created: domain.Payment{
		ID:           42,
		Status:       domain.PaymentAuthorized,
		CardLastFour: "1111",
		ExpiryMonth:  12,
		ExpiryYear:   2030,
		Currency:     "USD",
		Amount:       2500,
	}}
	requestBody := `{"cardNumber":"4111111111111111","expiryMonth":12,"expiryYear":2030,"currency":"USD","amount":2500,"cvv":"123"}`
	request := httptest.NewRequest(http.MethodPost, "/payments", strings.NewReader(requestBody))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	NewRouter(service).ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body: %s", response.Code, http.StatusCreated, response.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body["id"] != float64(42) || body["cardLastFour"] != "1111" {
		t.Errorf("response = %v, expected generated ID and card last four", body)
	}
	if body["status"] != string(domain.PaymentAuthorized) {
		t.Errorf("response status = %v, want %q", body["status"], domain.PaymentAuthorized)
	}
	if _, exists := body["card_number"]; exists {
		t.Error("response included card number")
	}
	if _, exists := body["cvv"]; exists {
		t.Error("response included CVV")
	}
	if service.input.CardNumber != "4111111111111111" || service.input.CVV != "123" {
		t.Error("service did not receive authorization inputs")
	}
}

func TestCreatePaymentRejectsMalformedRequest(t *testing.T) {
	service := &paymentServiceStub{}
	request := httptest.NewRequest(http.MethodPost, "/payments", strings.NewReader(`{"cardNumber":"123","cvv":"1"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	NewRouter(service).ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
	if service.input.CardNumber != "" {
		t.Error("service was called for malformed request")
	}
}

func TestCreatePaymentMapsServiceErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "invalid payment", err: applicationpayment.ErrInvalidPayment, want: http.StatusBadRequest},
		{name: "bank unavailable", err: applicationpayment.ErrBankUnavailable, want: http.StatusServiceUnavailable},
		{name: "processing failure", err: errors.New("database details"), want: http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &paymentServiceStub{err: tt.err}
			request := httptest.NewRequest(http.MethodPost, "/payments", strings.NewReader(`{"cardNumber":"4111111111111111","expiryMonth":12,"expiryYear":2030,"currency":"USD","amount":2500,"cvv":"123"}`))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()

			NewRouter(service).ServeHTTP(response, request)

			if response.Code != tt.want {
				t.Errorf("status = %d, want %d", response.Code, tt.want)
			}
			if strings.Contains(response.Body.String(), "database details") {
				t.Error("response exposed internal error details")
			}
		})
	}
}
