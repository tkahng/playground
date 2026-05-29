package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateDatabaseWithTemplate(ctx context.Context, dbx *pgxpool.Pool, targetDb, sourceDb string) error {
	_, err := dbx.Exec(ctx, fmt.Sprintf(
		`CREATE DATABASE %s WITH TEMPLATE %s`,
		pgx.Identifier{targetDb}.Sanitize(),
		pgx.Identifier{sourceDb}.Sanitize(),
	))
	return err
}

func DeleteDatabase(ctx context.Context, dbx *pgxpool.Pool, targetDb string) error {
	_, err := dbx.Exec(ctx, fmt.Sprintf(
		`DROP DATABASE %s WITH (FORCE)`,
		pgx.Identifier{targetDb}.Sanitize(),
	))
	return err
}
