package payment

import "errors"

var (
	ErrInvalidPayment  = errors.New("invalid payment")
	ErrBankUnavailable = errors.New("bank authorization is unavailable")
	ErrPaymentNotFound = errors.New("payment not found")
)
