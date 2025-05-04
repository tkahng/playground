package queries

import (
	"context"

	"github.com/tkahng/authgo/internal/db"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/tools/mapper"
)

type DelFunc func(ctx context.Context, dbx repository.DBTX, where *map[string]any) (int64, error)

func TruncateModels(ctx context.Context, db db.Dbx) error {
	return ErrorWrapper(ctx, db, false,
		Convert(
			repository.User.Delete,
			repository.Role.Delete,
			repository.Permission.Delete,
			repository.UserPermission.Delete,
			repository.UserRole.Delete,
		)...,
	)
}
func Convert(dels ...DelFunc) []Executor[int64] {
	return mapper.Map(dels, func(del DelFunc) Executor[int64] {
		return func(ctx context.Context, dbx db.Dbx) (int64, error) {
			return del(ctx, dbx, nil)
		}
	})

}
