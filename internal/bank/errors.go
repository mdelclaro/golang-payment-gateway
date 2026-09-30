package bank

import "errors"

var (
	ErrBankRequest                = errors.New("bank request failed")
	ErrInvalidBankURL             = errors.New("bank URL must be a valid HTTP or HTTPS URL")
	ErrMissingAuthorizationStatus = errors.New("bank response is missing authorization status")
	ErrMissingAuthorizationCode   = errors.New("authorized bank response is missing authorization code")
)
