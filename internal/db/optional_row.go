package db

import (
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
)

func OptionalRow[T any](record *T, err error) (*T, error) {
	if err == nil {
		return record, nil
	} else if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	} else if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else {
		return nil, err
	}
}
