package jobs

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Enqueuer interface {
	Enqueue(ctx context.Context, args JobArgs, uniqueKey *string, runAfter time.Time, maxAttempts int) error
}

// DBEnqueuer implements Enqueuer using a database connection
type DBEnqueuer struct {
	DB *pgxpool.Pool
}

func NewDBEnqueuer(db *pgxpool.Pool) *DBEnqueuer {
	return &DBEnqueuer{DB: db}
}
