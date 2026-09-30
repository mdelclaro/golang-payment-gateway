package postgres

import (
	"context"
	"database/sql"
	"errors"

	applicationpayment "example.com/payment-gateway/internal/application/payment"
	"example.com/payment-gateway/internal/domain"
)

type Payments struct {
	db *sql.DB
}

func NewPayments(db *sql.DB) *Payments {
	return &Payments{db: db}
}

func (r *Payments) Create(ctx context.Context, payment domain.Payment) (domain.Payment, error) {
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO payments (
			status,
			card_last_four,
			expiry_month,
			expiry_year,
			currency,
			amount,
			authorization_code
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`,
		payment.Status,
		payment.CardLastFour,
		payment.ExpiryMonth,
		payment.ExpiryYear,
		payment.Currency,
		payment.Amount,
		payment.AuthorizationCode,
	).Scan(&payment.ID)
	if err != nil {
		return domain.Payment{}, err
	}

	return payment, nil
}

func (r *Payments) FindByID(ctx context.Context, id int64) (domain.Payment, error) {
	var payment domain.Payment
	err := r.db.QueryRowContext(ctx, `
		SELECT id, status, card_last_four, expiry_month, expiry_year, currency, amount, authorization_code
		FROM payments
		WHERE id = $1`, id).Scan(
		&payment.ID,
		&payment.Status,
		&payment.CardLastFour,
		&payment.ExpiryMonth,
		&payment.ExpiryYear,
		&payment.Currency,
		&payment.Amount,
		&payment.AuthorizationCode,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Payment{}, applicationpayment.ErrPaymentNotFound
	}
	if err != nil {
		return domain.Payment{}, err
	}

	return payment, nil
}
