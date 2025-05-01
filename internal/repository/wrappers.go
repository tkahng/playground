package repository

import (
	"context"
	"fmt"

	"github.com/stephenafamo/bob"
)

type Exec[T any] func(ctx context.Context, exec bob.Executor) (T, error)

func ErrorWrapper[T any](ctx context.Context, db Queryer, returnFirstErr bool, exec ...Exec[T]) error {
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

func DefaultCountWrapper[T any](ctx context.Context, db Queryer, exec Exec[T]) T {
	t, _ := exec(ctx, db)
	return t
}
