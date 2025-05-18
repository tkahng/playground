package queries

import (
	"context"
	"fmt"

	"github.com/tkahng/authgo/internal/database"
)

type Executor[T any] func(ctx context.Context, exec database.Dbx) (T, error)

func ErrorWrapper[T any](ctx context.Context, db database.Dbx, returnFirstErr bool, exec ...Executor[T]) error {
	var e error
	for _, ex := range exec {
		_, err := ex(ctx, db)
		if err != nil {
			e = fmt.Errorf("error executing query: %w", e)
			if returnFirstErr {
				return e
			}
		}
	}
	return e
}

func DefaultCountWrapper[T any](ctx context.Context, db database.Dbx, exec Executor[T]) T {
	t, _ := exec(ctx, db)
	return t
}
