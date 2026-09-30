package httpapi

type createPaymentRequest struct {
	CardNumber  string `json:"cardNumber" binding:"required,numeric,min=12,max=19"`
	ExpiryMonth int    `json:"expiryMonth" binding:"gte=1,lte=12"`
	ExpiryYear  int    `json:"expiryYear" binding:"gte=1"`
	Currency    string `json:"currency" binding:"required"`
	Amount      int64  `json:"amount" binding:"gt=0"`
	CVV         string `json:"cvv" binding:"required,numeric,min=3,max=4"`
}

type paymentResponse struct {
	ID                int64  `json:"id"`
	Status            string `json:"status"`
	CardLastFour      string `json:"cardLastFour"`
	ExpiryMonth       int    `json:"expiryMonth"`
	ExpiryYear        int    `json:"expiryYear"`
	Currency          string `json:"currency"`
	Amount            int64  `json:"amount"`
	AuthorizationCode string `json:"authorizationCode,omitempty"`
}

type errorResponse struct {
	Error string `json:"error"`
}
