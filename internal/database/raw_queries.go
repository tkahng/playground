package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateDatabaseWithTemplate(ctx context.Context, dbx *pgxpool.Pool, targetDb, sourceDb string) error {
	_, err := dbx.Exec(ctx, `CREATE DATABASE "`+targetDb+`" WITH TEMPLATE "`+sourceDb+`"`)
	return err
}

func DeleteDatabase(ctx context.Context, dbx *pgxpool.Pool, targetDb string) error {
	_, err := dbx.Exec(ctx, `DROP DATABASE "`+targetDb+`" WITH (FORCE);`)
	return err
}
