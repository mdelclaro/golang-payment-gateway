package bank

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"example.com/payment-gateway/internal/application/payment"
)

type AuthorizerParams struct {
	BaseURL string
	Client  *http.Client
}

type Authorizer struct {
	endpoint string
	client   *http.Client
}

func NewAuthorizer(params AuthorizerParams) (*Authorizer, error) {
	baseURL := strings.TrimSpace(params.BaseURL)
	parsedURL, err := url.Parse(baseURL)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") || parsedURL.Host == "" {
		return nil, ErrInvalidBankURL
	}

	client := params.Client
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}

	return &Authorizer{
		endpoint: strings.TrimRight(baseURL, "/") + "/payments",
		client:   client,
	}, nil
}

func (a *Authorizer) Authorize(ctx context.Context, input payment.AuthorizationRequest) (payment.AuthorizationResult, error) {
	requestBody := bankRequest{
		CardNumber: input.CardNumber,
		ExpiryDate: fmt.Sprintf("%02d/%04d", input.ExpiryMonth, input.ExpiryYear),
		Currency:   input.Currency,
		Amount:     input.Amount,
		CVV:        input.CVV,
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return payment.AuthorizationResult{}, fmt.Errorf("encode bank authorization request: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, a.endpoint, bytes.NewReader(body))
	if err != nil {
		return payment.AuthorizationResult{}, fmt.Errorf("create bank authorization request: %w", err)
	}

	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")

	response, err := a.client.Do(request)
	if err != nil {
		return payment.AuthorizationResult{}, fmt.Errorf("send bank authorization request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusServiceUnavailable {
		return payment.AuthorizationResult{}, payment.ErrBankUnavailable
	}

	if response.StatusCode != http.StatusOK {
		return payment.AuthorizationResult{}, fmt.Errorf("%w: HTTP %d", ErrBankRequest, response.StatusCode)
	}

	var bankResponse bankResponse
	if err := json.NewDecoder(response.Body).Decode(&bankResponse); err != nil {
		return payment.AuthorizationResult{}, fmt.Errorf("decode bank authorization response: %w", err)
	}

	if bankResponse.Authorized == nil {
		return payment.AuthorizationResult{}, ErrMissingAuthorizationStatus
	}

	if *bankResponse.Authorized && bankResponse.AuthorizationCode == "" {
		return payment.AuthorizationResult{}, ErrMissingAuthorizationCode
	}

	return payment.AuthorizationResult{
		Authorized:        *bankResponse.Authorized,
		AuthorizationCode: bankResponse.AuthorizationCode,
	}, nil
}
