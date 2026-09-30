package bank

type bankRequest struct {
	CardNumber string `json:"card_number"`
	ExpiryDate string `json:"expiry_date"`
	Currency   string `json:"currency"`
	Amount     int64  `json:"amount"`
	CVV        string `json:"cvv"`
}

type bankResponse struct {
	Authorized        *bool  `json:"authorized"`
	AuthorizationCode string `json:"authorization_code"`
}
