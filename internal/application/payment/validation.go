package payment

import "time"

func validateExpiryAndGetLastFour(cardNumber string, expiryMonth, expiryYear int) (string, error) {
	if len(cardNumber) < 4 || expiryMonth < 1 || expiryMonth > 12 || expiryYear < 1 {
		return "", ErrInvalidPayment
	}

	now := time.Now()
	if expiryYear < now.Year() || (expiryYear == now.Year() && expiryMonth < int(now.Month())) {
		return "", ErrInvalidPayment
	}

	return cardNumber[len(cardNumber)-4:], nil
}
