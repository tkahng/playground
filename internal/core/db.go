package core

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/stephenafamo/bob"
	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/pool"
)

type DBX interface {
	Pool() *pgxpool.Pool
	Db() bob.DB
}

type Dbx struct {
	pool *pgxpool.Pool
	db   bob.DB
}

// Db implements DBX.
func (d *Dbx) Db() bob.DB {
	// if d.pool == nil {
	// 	panic("pgx pool is nil")
	// }
	// return bob.NewDB(stdlib.OpenDBFromPool(d.pool))
	return d.db
}

// Pool implements DBX.
func (d *Dbx) Pool() *pgxpool.Pool {
	return d.pool
}

var _ DBX = (*Dbx)(nil)

func NewDBX(pool *pgxpool.Pool) DBX {
	return &Dbx{
		pool: pool,
		db:   NewBobFromPool(pool),
	}
}

func NewPoolFromConf(ctx context.Context, conf conf.DBConfig) *pgxpool.Pool {
	return pool.CreatePool(ctx, conf.DatabaseUrl)
}

func NewBobFromPool(pool *pgxpool.Pool) bob.DB {
	return bob.NewDB(stdlib.OpenDBFromPool(pool))
}
