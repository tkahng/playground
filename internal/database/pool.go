package database

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	pgxgeom "github.com/twpayne/pgx-geom"
)

var (
	cachedCustomTypes     []*pgtype.Type
	cachedCustomTypesOnce sync.Once
	cachedCustomTypesErr  error
)

func CreatePool(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}
	dbpool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	return dbpool, nil
}

// CreateNewQueriesContext creates a new pool.
func CreateNewQueriesContext(ctx context.Context, connString string) *Queries {
	pool, err := CreatePoolWithCustomDataTypes(ctx, connString)
	if err != nil {
		slog.Error("CreateNewQueriesContext: error creating pool.", "error", err)
		panic(err.Error())
	}
	return &Queries{
		db: pool,
	}
}

func CreatePoolWithCustomDataTypes(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}
	// Load custom types once per process; subsequent pool creations reuse the cache.
	cachedCustomTypesOnce.Do(func() {
		conn, err := pgx.Connect(ctx, connString)
		if err != nil {
			cachedCustomTypesErr = err
			return
		}
		cachedCustomTypes, cachedCustomTypesErr = getCustomDataTypes(ctx, conn)
		conn.Close(ctx)
	})
	if cachedCustomTypesErr != nil {
		return nil, cachedCustomTypesErr
	}
	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		_, err := conn.Exec(ctx, `SET search_path TO gis, public, postgis, "$user";`)
		if err != nil {
			return err
		}
		for _, t := range cachedCustomTypes {
			conn.TypeMap().RegisterType(t)
		}
		if err := pgxgeom.Register(ctx, conn); err != nil {
			return err
		}
		return nil
	}
	return pgxpool.NewWithConfig(ctx, config)
}

// Any custom DB types made with CREATE TYPE need to be registered with pgx.
// https://github.com/kyleconroy/sqlc/issues/2116
// https://stackoverflow.com/questions/75658429/need-to-update-psql-row-of-a-composite-type-in-golang-with-jack-pgx
// https://pkg.go.dev/github.com/jackc/pgx/v5/pgtype
func getCustomDataTypes(ctx context.Context, conn *pgx.Conn) ([]*pgtype.Type, error) {
	dataTypeNames := []string{
		"auth.providers",
		// An underscore prefix is an array type in pgtypes.
		"auth._providers",
	}

	typesToRegister := []*pgtype.Type{}
	for _, typeName := range dataTypeNames {
		dataType, err := conn.LoadType(ctx, typeName)
		if err != nil {
			return nil, fmt.Errorf("failed to load type %s: %v", typeName, err)
		}
		// Register on this connection too so array type can find element type.
		conn.TypeMap().RegisterType(dataType)
		typesToRegister = append(typesToRegister, dataType)
	}
	return typesToRegister, nil
}
