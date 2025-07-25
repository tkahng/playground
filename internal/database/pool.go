package database

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	pgInstance *Queries
	pgOnce     sync.Once
)

func newQueries(ctx context.Context, connString string) (*Queries, error) {
	pgOnce.Do(func() {
		pool, err := getDbPool(ctx, connString)
		if err != nil {
			slog.Error("error creating pool.", "error", err)
		}
		pgInstance = &Queries{
			db: pool,
		}
	})
	if pgInstance == nil {
		return nil, fmt.Errorf("error at error: %w", nil)
	}
	return pgInstance, nil

}
func CreateQueriesContext(ctx context.Context, connString string) *Queries {
	pool, err := newQueries(ctx, connString)
	if err != nil {
		slog.Error("error creating pool.", "error", err)
		panic(fmt.Errorf("error creating pool: %w", err))
	}
	return pool
}

func CreateQueries(connString string) *Queries {
	return CreateQueriesContext(context.Background(), connString)
}

func getDbPool(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	// Set up a new pool with the custom types and the config.
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}
	dbpool, err := pgxpool.NewWithConfig(ctx, config)

	if err != nil {
		return nil, err
	}

	// Collect the custom data types once, store them in memory, and register them for every future connection.
	customTypes, err := getCustomDataTypes(ctx, dbpool)
	if err != nil {
		return nil, err
	}
	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		// var err error
		for _, t := range customTypes {
			conn.TypeMap().RegisterType(t)
		}
		// err = pgxvector.RegisterTypes(ctx, conn)
		// if err != nil {
		// 	fmt.Println(err)
		// 	return err
		// }
		// if err = pgxgeom.Register(ctx, conn); err != nil {
		// 	fmt.Println(err)
		// 	return err
		// }
		return nil

	}
	// Immediately close the old pool and open a new one with the new config.
	dbpool.Close()
	dbpool, err = pgxpool.NewWithConfig(ctx, config)
	return dbpool, err
}

// Any custom DB types made with CREATE TYPE need to be registered with pgx.
// https://github.com/kyleconroy/sqlc/issues/2116
// https://stackoverflow.com/questions/75658429/need-to-update-psql-row-of-a-composite-type-in-golang-with-jack-pgx
// https://pkg.go.dev/github.com/jackc/pgx/v5/pgtype
func getCustomDataTypes(ctx context.Context, pool *pgxpool.Pool) ([]*pgtype.Type, error) {
	// Get a single connection just to load type information.
	conn, err := pool.Acquire(ctx)
	defer conn.Release()
	if err != nil {
		return nil, err
	}

	dataTypeNames := []string{
		"providers",
		// An underscore prefix is an array type in pgtypes.
		"_providers",
	}

	var typesToRegister []*pgtype.Type
	for _, typeName := range dataTypeNames {
		dataType, err := conn.Conn().LoadType(ctx, typeName)
		if err != nil {
			return nil, fmt.Errorf("failed to load type %s: %v", typeName, err)
		}
		// You need to register only for this connection too, otherwise the array type will look for the register element type.
		conn.Conn().TypeMap().RegisterType(dataType)
		typesToRegister = append(typesToRegister, dataType)
	}
	return typesToRegister, nil
}
