package resource

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/repository"
)

type PermissionsListFilter struct {
	PaginatedInput
	SortParams
	Q           string      `query:"q,omitempty" required:"false"`
	Ids         []uuid.UUID `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	Names       []string    `query:"names,omitempty" required:"false" minimum:"1" maximum:"100"`
	RoleId      uuid.UUID   `query:"role_id,omitempty" required:"false" format:"uuid"`
	RoleReverse bool        `query:"role_reverse,omitempty" required:"false" doc:"When role_id is provided, if this is true, it will return the permissions that the role does not have"`
}

func NewPermissionQueryResource(
	db database.Dbx,
) *QueryResource[models.Permission, uuid.UUID, PermissionsListFilter] {
	return NewQueryResource[models.Permission, uuid.UUID](
		db,
		repository.PermissionBuilder,
		func(qs sq.SelectBuilder, filter *PermissionsListFilter) sq.SelectBuilder {
			if filter == nil {
				return qs
			}
			if filter.Q != "" {
				qs = qs.Where(
					sq.Or{
						sq.ILike{"name": "%" + filter.Q + "%"},
						sq.ILike{"description": "%" + filter.Q + "%"},
					},
				)

			}
			if len(filter.Names) > 0 {
				qs = qs.Where(sq.Eq{"name": filter.Names})
			}
			if len(filter.Ids) > 0 {
				qs = qs.Where(sq.Eq{"id": filter.Ids})
			}

			if filter.RoleId != uuid.Nil {
				if filter.RoleReverse {
					qs = qs.LeftJoin(
						"role_permissions"+" on "+"permissions.id"+" = "+"role_permissions"+"."+"permission_id"+" and "+"role_permissions"+"."+"role_id"+" = ?",
						filter.RoleId,
					)
					qs = qs.Where("role_permissions.permission_id is null")

				} else {
					qs = qs.Join("role_permissions on permissions.id = role_permissions.permission_id and role_permissions.role_id = ?", filter.RoleId).
						Where(sq.Eq{"role_permissions.role_id": filter.RoleId})

				}
			}
			return qs
		},
		nil,
		nil,
	)
}
