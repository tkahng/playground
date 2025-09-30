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
	Migrations embed.FS

	//go:embed migrations
	migrations embed.FS
)

type Migrator interface {
	Migrate() error
	Reset() error
}
type MigratorConfig struct {
	DatabaseUrl string
	AutoDumpSchema bool
}

func NewMigrator(config *MigratorConfig) Migrator {
	u, err := url.Parse(config.DatabaseUrl)
	if err != nil {
		slog.Error("error parsing database url", "error", err)
		panic(fmt.Errorf("error parsing database url for migrator: %w", err))
	}
	dm := dbmate.New(u)
	dm.FS = migrations
	dm.AutoDumpSchema = config.AutoDumpSchema
	dm.MigrationsDir = []string{"./migrations"}
	dm.SchemaFile = "./internal/database/schema.sql"
	return &migrator{dm: dm}
}

type migrator struct {
	dm *dbmate.DB
}

// Migrate implements Migrator.
func (m *migrator) Migrate() error {
	return m.dm.CreateAndMigrate()
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


