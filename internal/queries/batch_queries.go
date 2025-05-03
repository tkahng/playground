package queries

import (
	"context"

	"github.com/tkahng/authgo/internal/crud/crudrepo"
	"github.com/tkahng/authgo/internal/crud/repository"
	"github.com/tkahng/authgo/internal/tools/mapper"
)

type DelFunc func(ctx context.Context, dbx repository.DBTX, where *map[string]any) (int64, error)

func TruncateModels(ctx context.Context, db Queryer) error {
	return ErrorWrapper(ctx, db, false,
		Convert(
			crudrepo.User.Delete,
			crudrepo.Role.Delete,
			crudrepo.Permission.Delete,
			crudrepo.UserPermission.Delete,
			crudrepo.UserRole.Delete,
		)...,
	)
}
func Convert(dels ...DelFunc) []Executor[int64] {
	return mapper.Map(dels, func(del DelFunc) Executor[int64] {
		return func(ctx context.Context, dbx Queryer) (int64, error) {
			return del(ctx, dbx, nil)
		}
	})

}
