package payment

import (
	"testing"
	"time"
)

func TestValidateExpiryAndGetLastFourReturnsCardSuffix(t *testing.T) {
	got, err := validateExpiryAndGetLastFour("4111111111111111", 12, 2030)
	if err != nil {
		t.Fatalf("validateExpiryAndGetLastFour() error = %v", err)
	}
	if got != "1111" {
		t.Errorf("last four = %q, want 1111", got)
	}
}

func TestValidateExpiryAndGetLastFourRejectsExpiredCard(t *testing.T) {
	previousMonth := time.Now().AddDate(0, -1, 0)
	got, err := validateExpiryAndGetLastFour("4111111111111111", int(previousMonth.Month()), previousMonth.Year())
	if err != ErrInvalidPayment {
		t.Errorf("validateExpiryAndGetLastFour() error = %v, want %v", err, ErrInvalidPayment)
	}
	if got != "" {
		t.Errorf("last four = %q, want empty value for expired card", got)
	}
}
