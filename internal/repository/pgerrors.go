package repository

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

func isMissingRelation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "42P01"
	}
	return false
}
