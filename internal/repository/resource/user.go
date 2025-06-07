package resource

import (
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	repo "github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/tools/types"
)

type UserListFilter struct {
	PaginatedInput
	SortParams
	Providers     []models.Providers        `query:"providers,omitempty" required:"false" uniqueItems:"true" minimum:"1" maximum:"100" enum:"google,apple,facebook,github,credentials"`
	Q             string                    `query:"q,omitempty" required:"false"`
	Ids           []uuid.UUIDs              `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	Emails        []string                  `query:"emails,omitempty" required:"false" minimum:"1" maximum:"100" format:"email"`
	RoleIds       []uuid.UUID               `query:"role_ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	EmailVerified types.OptionalParam[bool] `query:"email_verified,omitempty" required:"false"`
}

var _ Resource[models.User, uuid.UUID, UserListFilter] = (*RepositoryResource[models.User, uuid.UUID, UserListFilter])(nil)

func NewUserRepositoryResource(
	db database.Dbx,
) *RepositoryResource[models.User, uuid.UUID, UserListFilter] {
	return NewRepositoryResource[models.User, uuid.UUID](
		db,
		repo.User,
		func(filter *UserListFilter) *map[string]any {
			where := make(map[string]any)
			if filter == nil {
				return &where // return empty map if no filter is provided
			}

			if filter.EmailVerified.IsSet {
				emailverified := filter.EmailVerified.Value
				if emailverified {
					where["email_verified_at"] = map[string]any{
						"_neq": nil,
					}
				} else {
					where["email_verified_at"] = map[string]any{
						"_eq": nil,
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
		func(filter *UserListFilter) *map[string]string {
			if filter == nil || filter.SortBy == "" || filter.SortOrder == "" {
				return nil
			}
			order := make(map[string]string)

			if slices.Contains(models.UserTable.Columns, filter.SortBy) {
				order[filter.SortBy] = strings.ToUpper(filter.SortOrder)
			}

			return &order
		},
		func(input *UserListFilter) (limit int, offset int) {
			var page, perPage int
			if input == nil {
				page, perPage = 0, 10 // default values
			} else {
				page, perPage = int(input.Page), int(input.PerPage)
			}
			if perPage < 1 {
				perPage = 10
			}
			if page < 0 {
				page = 0
			}
			return perPage, (page) * perPage
		},
	)

}
