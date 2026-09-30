package bank

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/payment-gateway/internal/application/payment"
)

func TestAuthorizeSendsExpectedRequestAndMapsApproval(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/payments" {
			t.Errorf("request = %s %s, want POST /payments", r.Method, r.URL.Path)
		}

		var body bankRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode bank request: %v", err)
		}

		if body.CardNumber != "4111111111111111" || body.ExpiryDate != "04/2030" || body.Currency != "GBP" || body.Amount != 1050 || body.CVV != "123" {
			t.Errorf("bank request = %+v, want mapped payment fields", body)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"authorized":true,"authorization_code":"auth-123"}`))
	}))
	defer server.Close()

	authorizer, err := NewAuthorizer(AuthorizerParams{BaseURL: server.URL, Client: server.Client()})
	if err != nil {
		t.Fatalf("NewAuthorizer() error = %v", err)
	}

	got, err := authorizer.Authorize(context.Background(), validAuthorizationRequest())
	if err != nil {
		t.Fatalf("Authorize() error = %v", err)
	}
	if !got.Authorized || got.AuthorizationCode != "auth-123" {
		t.Errorf("Authorize() = %+v, want approved auth-123", got)
	}
}

func TestAuthorizeMapsDecline(t *testing.T) {
	authorizer := newAuthorizerForResponse(t, http.StatusOK, `{"authorized":false}`)
	got, err := authorizer.Authorize(context.Background(), validAuthorizationRequest())
	if err != nil {
		t.Fatalf("Authorize() error = %v", err)
	}
	if got.Authorized || got.AuthorizationCode != "" {
		t.Errorf("Authorize() = %+v, want decline without authorization code", got)
	}
}

func TestAuthorizeMapsUnavailable(t *testing.T) {
	authorizer := newAuthorizerForResponse(t, http.StatusServiceUnavailable, `{"error":"unavailable"}`)
	_, err := authorizer.Authorize(context.Background(), validAuthorizationRequest())
	if !errors.Is(err, payment.ErrBankUnavailable) {
		t.Errorf("Authorize() error = %v, want %v", err, payment.ErrBankUnavailable)
	}
}

func TestAuthorizeMapsUnexpectedStatus(t *testing.T) {
	authorizer := newAuthorizerForResponse(t, http.StatusBadRequest, `{"error":"missing field"}`)
	_, err := authorizer.Authorize(context.Background(), validAuthorizationRequest())
	if !errors.Is(err, ErrBankRequest) {
		t.Errorf("Authorize() error = %v, want %v", err, ErrBankRequest)
	}
}

func TestNewAuthorizerValidatesURL(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
	}{
		{name: "empty URL", baseURL: ""},
		{name: "unsupported URL scheme", baseURL: "ftp://localhost:8080"},
		{name: "missing host", baseURL: "http:///payments"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := NewAuthorizer(AuthorizerParams{BaseURL: tt.baseURL}); err == nil {
				t.Error("NewAuthorizer() succeeded with invalid URL")
			}
		})
	}
}

func newAuthorizerForResponse(t *testing.T, status int, body string) *Authorizer {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	authorizer, err := NewAuthorizer(AuthorizerParams{BaseURL: server.URL, Client: server.Client()})
	if err != nil {
		t.Fatalf("NewAuthorizer() error = %v", err)
	}
	return authorizer
}

func validAuthorizationRequest() payment.AuthorizationRequest {
	return payment.AuthorizationRequest{
		CardNumber:  "4111111111111111",
		ExpiryMonth: 4,
		ExpiryYear:  2030,
		Currency:    "GBP",
		Amount:      1050,
		CVV:         "123",
	}
}
