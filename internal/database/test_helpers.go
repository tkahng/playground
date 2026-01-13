package database

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
	"strings"
	"sync"
	"testing"

	_ "github.com/amacneil/dbmate/v2/pkg/driver/postgres"

	"github.com/tkahng/playground/internal/conf"
	"github.com/tkahng/playground/internal/tools/logger"
	"github.com/tkahng/playground/internal/tools/security"
)

var (
	ErrEndTest  = errors.New("end test. rollback transaction")
	ctxInstance context.Context
	ctxOnce     sync.Once
	dbx         *Queries
)

func DbSetup(cfg *conf.EnvConfig) (context.Context, *Queries) {
	ctxOnce.Do(func() {
		ctxInstance = context.Background()
		dbx = CreateNewQueriesContext(ctxInstance, cfg.Db.GetDatabaseUrl())
	})
	return ctxInstance, dbx
}

func WithSingletonTestTx(t *testing.T, fn func(ctx context.Context, db Dbx)) {
	DbSetup(conf.ZeroEnvConfig())
	ctx := ctxInstance
	tx, err := dbx.BeginTx(ctx)
	if err != nil {
		t.Fatal(err)
	}
	//nolint:errcheck
	defer tx.Rollback(ctx)
	// panic handle
	defer func() {
		if recErr := recover(); recErr != nil {
			slog.ErrorContext(ctx, "recovered from panic in transaction.", slog.Any("error", fmt.Sprint(recErr)), slog.Any("stacktrace", string(debug.Stack())))
			rollBackErr := tx.Rollback(ctx)
			if rollBackErr != nil {
				slog.ErrorContext(ctx, "recovered from panic in transaction.", slog.Any("error", fmt.Sprint(recErr)), slog.Any("stacktrace", string(debug.Stack())))
				t.Error(rollBackErr)
			}
			t.Fatal(fmt.Sprint(recErr))
		}
	}()
	fn(ctx, NewTxQueries(tx))
}

// WithNewTestTx creates a new pool connection, runs the test within that transactions, rolls back, and closes the pool.
func WithNewTestTx(t *testing.T, fn func(ctx context.Context, db Dbx)) {
	t.Helper()
	// TODO: add context timeout
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()
	// ctx := t.Context()
	_ = logger.GetDefaultLogger()
	ctx := context.Background()
	cfg := conf.ZeroEnvConfig()
	dbx := CreateNewQueriesContext(ctx, cfg.Db.GetDatabaseUrl())
	tx, beginErr := dbx.BeginTx(ctx)
	if beginErr != nil {
		t.Fatal(beginErr)
	}
	defer dbx.Close()
	//nolint:errcheck
	defer tx.Rollback(ctx)
	// panic handle
	defer func() {
		if recErr := recover(); recErr != nil {
			slog.ErrorContext(ctx, "recovered from panic in transaction.", slog.String("error", strings.ToValidUTF8(string(debug.Stack()), "")))
			rollBackErr := tx.Rollback(ctx)
			if rollBackErr != nil {
				slog.ErrorContext(ctx, "error during rollback.", slog.Any("error", rollBackErr))
				t.Error(rollBackErr)
			}
			t.Fatal(fmt.Sprint(recErr))
		}
	}()
	fn(ctx, NewTxQueries(tx))
}

// WithNewTestTx creates a new pool connection, runs the test within that transactions, rolls back, and closes the pool.
func WithNewDatabase(t *testing.T, fn func(ctx context.Context, db Dbx)) {
	t.Helper()
	_ = logger.GetDefaultLogger()
	uid := security.RandomString(16)
	ctx := context.Background()
	cfg := conf.ZeroEnvConfig()
	dbCfg := cfg.Db
	clonedDbCfg := conf.DBConfig{
		User:     dbCfg.User,
		Password: dbCfg.Password,
		Host:     dbCfg.Host,
		Port:     dbCfg.Port,
		Db:       fmt.Sprintf("%s_%s", cfg.Db.Db, uid),
		SSL:      dbCfg.SSL,
	}
	mConfig := MigratorConfig{
		DatabaseUrl: clonedDbCfg.GetDatabaseUrl(),
	}
	migrator := NewMigrator(&mConfig)
	if err := migrator.CreateAndMigrate(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := migrator.Drop(); err != nil {
			t.Fatal(err)
		}
	}()
	dbx := CreateNewQueriesContext(ctx, clonedDbCfg.GetDatabaseUrl())
	defer dbx.Close()
	// panic handle
	defer func() {
		if recErr := recover(); recErr != nil {
			slog.ErrorContext(ctx, "recovered from panic in transaction.", slog.String("error", strings.ToValidUTF8(string(debug.Stack()), "")))
			t.Fatal(fmt.Sprint(recErr))
		}
	}()
	fn(ctx, dbx)
}

func WithNewDatabase2(t *testing.T, fn func(ctx context.Context, db Dbx)) {
	t.Helper()
	_ = logger.GetDefaultLogger()
	ctx := context.Background()
	cfg := conf.ZeroEnvConfig()
	defaultDbCfg := cfg.Db
	defaultDbCfg.Db = "postgres"
	clonedDbCfg := conf.DBConfig{
		User:     cfg.Db.User,
		Password: cfg.Db.Password,
		Host:     cfg.Db.Host,
		Port:     cfg.Db.Port,
		Db:       fmt.Sprintf("%s_%s", cfg.Db.Db, security.RandomString(16)),
		SSL:      cfg.Db.SSL,
	}
	pool, err := CreatePool(ctx, defaultDbCfg.GetDatabaseUrl())
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	name := clonedDbCfg.Db
	err = CreateDatabaseWithTemplate(ctx, pool, name, "playground_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := DeleteDatabase(ctx, pool, name)
		if err != nil {
			t.Fatal(err)
		}
	}()
	dbx := CreateNewQueriesContext(ctx, clonedDbCfg.GetDatabaseUrl())
	defer dbx.Close()
	// panic handle
	defer func() {
		if recErr := recover(); recErr != nil {
			slog.ErrorContext(ctx, "recovered from panic in transaction.", slog.String("error", strings.ToValidUTF8(string(debug.Stack()), "")))
			t.Fatal(fmt.Sprint(recErr))
		}
	}()
	fn(ctx, dbx)
}
