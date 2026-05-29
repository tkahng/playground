package database

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	pgxgeom "github.com/twpayne/pgx-geom"
)

// customTypeCache holds the result of a successful custom-type load.
// It is replaced atomically on success and is never set to a non-nil value
// until the types have been loaded without error, so callers never see a
// partial or errored cache.
type customTypeCache struct {
	types []*pgtype.Type
}

var customTypeCachePtr atomic.Pointer[customTypeCache]
var customTypeMu sync.Mutex

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
func CreateNewQueriesContext(ctx context.Context, connString string) (*Queries, error) {
	pool, err := CreatePoolWithCustomDataTypes(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("creating database pool: %w", err)
	}
	return &Queries{
		db: pool,
	}, nil
}

// loadCustomTypesOnce loads custom Postgres types, using a cached result when
// available. Unlike sync.Once, a failed load is retried on the next call so
// that a transient DB-unavailable error at startup does not permanently poison
// all subsequent pool creations.
func loadCustomTypesOnce(ctx context.Context, connString string) ([]*pgtype.Type, error) {
	if cached := customTypeCachePtr.Load(); cached != nil {
		return cached.types, nil
	}

	customTypeMu.Lock()
	defer customTypeMu.Unlock()

	// Double-check under lock.
	if cached := customTypeCachePtr.Load(); cached != nil {
		return cached.types, nil
	}

	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		return nil, err
	}
	types, err := getCustomDataTypes(ctx, conn)
	conn.Close(ctx)
	if err != nil {
		return nil, err
	}

	// Store only on success so errors are always retried.
	customTypeCachePtr.Store(&customTypeCache{types: types})
	return types, nil
}

func CreatePoolWithCustomDataTypes(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}

	customTypes, err := loadCustomTypesOnce(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("loading custom db types: %w", err)
	}

	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		_, err := conn.Exec(ctx, `SET search_path TO gis, public, postgis, "$user";`)
		if err != nil {
			return err
		}
		for _, t := range customTypes {
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
