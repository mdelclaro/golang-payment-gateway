package payment

type CreatePaymentInput struct {
	CardNumber  string `validate:"required,numeric,min=12,max=19"`
	ExpiryMonth int    `validate:"gte=1,lte=12"`
	ExpiryYear  int    `validate:"gte=1"`
	Currency    string `validate:"required"`
	Amount      int64  `validate:"gt=0"`
	CVV         string `validate:"required"`
}
