//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

type paymentResponse struct {
	ID                int64  `json:"id"`
	Status            string `json:"status"`
	CardLastFour      string `json:"cardLastFour"`
	ExpiryMonth       int    `json:"expiryMonth"`
	ExpiryYear        int    `json:"expiryYear"`
	Currency          string `json:"currency"`
	Amount            int64  `json:"amount"`
	AuthorizationCode string `json:"authorizationCode"`
}

type paymentRequest struct {
	CardNumber  string `json:"cardNumber"`
	ExpiryMonth int    `json:"expiryMonth"`
	ExpiryYear  int    `json:"expiryYear"`
	Currency    string `json:"currency"`
	Amount      int64  `json:"amount"`
	CVV         string `json:"cvv"`
}

func TestPaymentGatewayEndToEnd(t *testing.T) {
	baseURL := strings.TrimRight(os.Getenv("PAYMENT_GATEWAY_URL"), "/")
	if baseURL == "" {
		baseURL = "http://localhost:8090"
	}
	client := &http.Client{Timeout: 15 * time.Second}

	t.Run("health check", func(t *testing.T) {
		response := doRequest(t, client, http.MethodGet, baseURL+"/health", nil)
		defer response.Body.Close()
		assertStatus(t, response, http.StatusOK)

		var body struct {
			Status string `json:"status"`
		}
		decodeJSON(t, response, &body)
		if body.Status != "ok" {
			t.Fatalf("health status = %q, want %q", body.Status, "ok")
		}
	})

	t.Run("authorized payment is persisted and retrievable", func(t *testing.T) {
		created := createPayment(t, client, baseURL, paymentRequestForCard("4111111111111111"), http.StatusCreated)
		if created.payment.Status != "Authorized" {
			t.Fatalf("payment status = %q, want Authorized", created.payment.Status)
		}
		if created.payment.CardLastFour != "1111" {
			t.Errorf("cardLastFour = %q, want 1111", created.payment.CardLastFour)
		}
		assertNoSensitivePaymentFields(t, created.raw)

		response := doRequest(t, client, http.MethodGet, fmt.Sprintf("%s/v1/payments/%d", baseURL, created.payment.ID), nil)
		defer response.Body.Close()
		assertStatus(t, response, http.StatusOK)

		body, err := io.ReadAll(response.Body)
		if err != nil {
			t.Fatalf("read retrieved payment: %v", err)
		}
		var fetched paymentResponse
		if err := json.Unmarshal(body, &fetched); err != nil {
			t.Fatalf("decode retrieved payment: %v", err)
		}
		if fetched != created.payment {
			t.Errorf("retrieved payment = %+v, want %+v", fetched, created.payment)
		}
		var fetchedRaw map[string]json.RawMessage
		if err := json.Unmarshal(body, &fetchedRaw); err != nil {
			t.Fatalf("decode retrieved payment fields: %v", err)
		}
		assertNoSensitivePaymentFields(t, fetchedRaw)
	})

	t.Run("declined payment is persisted", func(t *testing.T) {
		created := createPayment(t, client, baseURL, paymentRequestForCard("4111111111111112"), http.StatusCreated)
		if created.payment.Status != "Declined" {
			t.Fatalf("payment status = %q, want Declined", created.payment.Status)
		}

		response := doRequest(t, client, http.MethodGet, fmt.Sprintf("%s/v1/payments/%d", baseURL, created.payment.ID), nil)
		defer response.Body.Close()
		assertStatus(t, response, http.StatusOK)

		var fetched paymentResponse
		decodeJSON(t, response, &fetched)
		if fetched.Status != "Declined" || fetched.ID != created.payment.ID {
			t.Errorf("retrieved payment = %+v, want declined payment ID %d", fetched, created.payment.ID)
		}
	})

	t.Run("invalid request is rejected", func(t *testing.T) {
		request := paymentRequestForCard("123")
		response := doRequest(t, client, http.MethodPost, baseURL+"/v1/payments", request)
		defer response.Body.Close()
		assertStatus(t, response, http.StatusBadRequest)

		var body struct {
			Error string `json:"error"`
		}
		decodeJSON(t, response, &body)
		if body.Error != "invalid payment request" {
			t.Errorf("error = %q, want invalid payment request", body.Error)
		}
	})

	t.Run("bank unavailability is returned to the caller", func(t *testing.T) {
		request := paymentRequestForCard("4111111111111110")
		response := doRequest(t, client, http.MethodPost, baseURL+"/v1/payments", request)
		defer response.Body.Close()
		assertStatus(t, response, http.StatusServiceUnavailable)

		var body struct {
			Error string `json:"error"`
		}
		decodeJSON(t, response, &body)
		if body.Error != "payment authorization is unavailable" {
			t.Errorf("error = %q, want payment authorization is unavailable", body.Error)
		}
	})

	t.Run("missing payment returns 404", func(t *testing.T) {
		response := doRequest(t, client, http.MethodGet, baseURL+"/v1/payments/9223372036854775807", nil)
		defer response.Body.Close()
		assertStatus(t, response, http.StatusNotFound)
	})

	t.Run("invalid payment ID returns 400", func(t *testing.T) {
		response := doRequest(t, client, http.MethodGet, baseURL+"/v1/payments/0", nil)
		defer response.Body.Close()
		assertStatus(t, response, http.StatusBadRequest)
	})
}

type createdPayment struct {
	payment paymentResponse
	raw     map[string]json.RawMessage
}

func createPayment(t *testing.T, client *http.Client, baseURL string, request paymentRequest, wantStatus int) createdPayment {
	t.Helper()
	response := doRequest(t, client, http.MethodPost, baseURL+"/v1/payments", request)
	defer response.Body.Close()
	assertStatus(t, response, wantStatus)

	var result createdPayment
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read payment response: %v", err)
	}
	if err := json.Unmarshal(body, &result.payment); err != nil {
		t.Fatalf("decode payment response: %v", err)
	}
	if err := json.Unmarshal(body, &result.raw); err != nil {
		t.Fatalf("decode payment fields: %v", err)
	}
	return result
}

func paymentRequestForCard(cardNumber string) paymentRequest {
	now := time.Now()
	return paymentRequest{
		CardNumber:  cardNumber,
		ExpiryMonth: int(now.Month()),
		ExpiryYear:  now.Year() + 2,
		Currency:    "USD",
		Amount:      2500,
		CVV:         "123",
	}
}

func doRequest(t *testing.T, client *http.Client, method, url string, body any) *http.Response {
	t.Helper()

	var requestBody io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("encode request: %v", err)
		}
		requestBody = bytes.NewReader(encoded)
	}

	request, err := http.NewRequest(method, url, requestBody)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	request.Header.Set("Accept", "application/json")
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	return response
}

func assertStatus(t *testing.T, response *http.Response, want int) {
	t.Helper()
	if response.StatusCode == want {
		return
	}
	body, _ := io.ReadAll(response.Body)
	t.Fatalf("status = %d, want %d; response body: %s", response.StatusCode, want, strings.TrimSpace(string(body)))
}

func decodeJSON(t *testing.T, response *http.Response, destination any) {
	t.Helper()
	if err := json.NewDecoder(response.Body).Decode(destination); err != nil {
		t.Fatalf("decode JSON response: %v", err)
	}
}

func assertNoSensitivePaymentFields(t *testing.T, payment map[string]json.RawMessage) {
	t.Helper()
	for _, field := range []string{"cardNumber", "cvv"} {
		if _, exists := payment[field]; exists {
			t.Errorf("payment response contains sensitive field %q", field)
		}
	}
}
