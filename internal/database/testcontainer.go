package database

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var tcOnce sync.Once

// initTestDB is a thin wrapper called from individual test helpers.
func initTestDB(_ testing.TB) {
	tcOnce.Do(startTestContainer)
}

// MustInitTestDB starts the testcontainer without requiring a testing.TB.
// Call this from TestMain so the container is ready before any tests run.
func MustInitTestDB() {
	tcOnce.Do(startTestContainer)
}

// startTestContainer starts tkahng/postgres:18 once per test binary, runs all
// embedded migrations, and overrides DATABASE_* env vars so conf.ZeroEnvConfig()
// returns the container's coordinates. Ryuk handles cleanup on process exit.
//
// tkahng/postgres:18 includes PostgreSQL 18 (uuidv7() built-in) + PostGIS.
func startTestContainer() {
	ctx := context.Background()

	ctr, err := tcpostgres.Run(ctx,
		"tkahng/postgres:18",
		tcpostgres.WithDatabase("playground_test"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
		),
	)
	if err != nil {
		panic(fmt.Sprintf("start postgres container: %v", err))
	}

	host, err := ctr.Host(ctx)
	if err != nil {
		panic(fmt.Sprintf("get postgres container host: %v", err))
	}
	port, err := ctr.MappedPort(ctx, "5432")
	if err != nil {
		panic(fmt.Sprintf("get postgres container port: %v", err))
	}

	os.Setenv("DATABASE_HOST", host)
	os.Setenv("DATABASE_PORT", port.Port())
	os.Setenv("DATABASE_USER", "postgres")
	os.Setenv("DATABASE_PASSWORD", "postgres")
	os.Setenv("DATABASE_DB", "playground_test")
	os.Setenv("DATABASE_SSL", "disable")

	connURL := fmt.Sprintf(
		"postgres://postgres:postgres@%s:%s/playground_test?sslmode=disable",
		host, port.Port(),
	)

	migrator := NewMigrator(&MigratorConfig{
		DatabaseURL:    connURL,
		AutoDumpSchema: false,
	})
	if err := migrator.CreateAndMigrate(); err != nil {
		panic(fmt.Sprintf("migrate test postgres: %v", err))
	}
}
