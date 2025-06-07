package resource

import (
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
	repo "github.com/tkahng/authgo/internal/repository"
)

type UserListFilter struct {
	repo.PaginatedInput
	repo.SortParams
}

var _ repo.Resource[models.User, uuid.UUID, UserListFilter] = (*repo.RepositoryResource[models.User, uuid.UUID, UserListFilter])(nil)

// func NewUserRepositoryResource(
// 	db database.Dbx,
// ) *repo.RepositoryResource[models.User, uuid.UUID, UserListFilter] {
// 	return repo.NewRepositoryResource[models.User, uuid.UUID](
// 		db,
// 	)
// }
