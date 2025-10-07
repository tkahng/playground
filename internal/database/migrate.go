package database

import (
	"embed"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/amacneil/dbmate/v2/pkg/dbmate"
)

var (
	//go:embed migrations
	migrationsFs embed.FS
)

type Migrator interface {
	CreateAndMigrate() error
	DumpSchema() error
	Reset() error
	Rollback() error
	Drop() error
}
type MigratorConfig struct {
	DatabaseUrl    string
	AutoDumpSchema bool
}

func NewMigrator(config *MigratorConfig) Migrator {
	u, err := url.Parse(config.DatabaseUrl)
	if err != nil {
		slog.Error("error parsing database url", "error", err)
		panic(fmt.Errorf("error parsing database url for migrator: %w", err))
	}
	dm := dbmate.New(u)
	dm.FS = migrationsFs
	dm.AutoDumpSchema = config.AutoDumpSchema
	dm.MigrationsDir = []string{"./migrations"}
	dm.SchemaFile = "./internal/database/schema.sql"
	return &migrator{dm: dm}
}

type migrator struct {
	dm *dbmate.DB
}

// CreateAndMigrate implements Migrator.
func (m *migrator) CreateAndMigrate() error {
	return m.dm.CreateAndMigrate()
}
func (m *migrator) DumpSchema() error {
	return m.dm.DumpSchema()
}
func (m *migrator) Rollback() error {
	return m.dm.Rollback()
}

func (m *migrator) Drop() error {
	return m.dm.Drop()
}

// Reset implements Migrator.
func (m *migrator) Reset() error {
	err := m.dm.Drop()
	if err != nil {
		return err
	}
	err = m.dm.CreateAndMigrate()
	if err != nil {
		return err
	}
	return nil
}
