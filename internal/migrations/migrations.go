package migrations

import (
	"context"
	"database/sql"
	_ "embed"
)

//go:embed sql/000001_create_payments.sql
var createPaymentsTable string

func Migrate(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, createPaymentsTable)
	return err
}
