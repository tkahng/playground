package resource

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/repository"
	repo "github.com/tkahng/playground/internal/repository"
	"github.com/tkahng/playground/internal/tools/types"
)

type UserFilter struct {
	repository.PaginatedInput
	repository.SortParams
	Providers     []models.Providers        `query:"providers,omitempty" required:"false" uniqueItems:"true" minimum:"1" maximum:"100" enum:"google,apple,facebook,github,credentials"`
	Q             string                    `query:"q,omitempty" required:"false"`
	Ids           []uuid.UUID               `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	Emails        []string                  `query:"emails,omitempty" required:"false" minimum:"1" maximum:"100" format:"email"`
	RoleIds       []uuid.UUID               `query:"role_ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	EmailVerified types.OptionalParam[bool] `query:"email_verified,omitempty" required:"false"`
}
type dsad interface {
	Find(ctx context.Context, filter *UserFilter) ([]*models.User, error)
}

var _ dsad = (*RepositoryResource[models.User, uuid.UUID, UserFilter])(nil)
var _ Resource[models.User, uuid.UUID, UserFilter] = (*RepositoryResource[models.User, uuid.UUID, UserFilter])(nil)

func NewUserRepositoryResource(
	db database.Dbx,
) *RepositoryResource[models.User, uuid.UUID, UserFilter] {
	return NewRepositoryResource[models.User, uuid.UUID](
		db,
		repo.User,
		func(filter *UserFilter) *map[string]any {
			where := make(map[string]any)
			if filter == nil {
				return &where // return empty map if no filter is provided
			}

			if filter.EmailVerified.IsSet {
				emailverified := filter.EmailVerified.Value
				if emailverified {
					where[models.UserTable.EmailVerifiedAt] = map[string]any{
						repo.IsNotNull: nil,
					}
				} else {
					where[models.UserTable.EmailVerifiedAt] = map[string]any{
						repo.IsNull: nil,
					}
				}
			}
			if len(filter.Emails) > 0 {
				where["email"] = map[string]any{
					"_in": filter.Emails,
				}
			}
			if len(filter.Ids) > 0 {
				where["id"] = map[string]any{
					"_in": filter.Ids,
				}
			}
			if len(filter.Providers) > 0 {
				where["accounts"] = map[string]any{
					"provider": map[string]any{
						"_in": filter.Providers,
					},
				}
			}
			if len(filter.RoleIds) > 0 {
				where["roles"] = map[string]any{
					"id": map[string]any{
						"_in": filter.RoleIds,
					},
				}
			}
			if filter.Q != "" {
				where["_or"] = []map[string]any{
					{
						"email": map[string]any{
							"_ilike": "%" + filter.Q + "%",
						},
					},
					{
						"name": map[string]any{
							"_ilike": "%" + filter.Q + "%",
						},
					},
				}
			}
			if len(where) == 0 {
				return nil
			}
			return &where
		},
		nil,
		nil,
	)

}
