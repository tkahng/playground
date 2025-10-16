package resource

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/database/repository"
	"github.com/tkahng/playground/internal/models"
)

type PermissionsFilter struct {
	repository.PaginatedInput
	repository.SortParams
	Q           string      `query:"q,omitempty" required:"false"`
	Ids         []uuid.UUID `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	Names       []string    `query:"names,omitempty" required:"false" minimum:"1" maximum:"100"`
	RoleId      uuid.UUID   `query:"role_id,omitempty" required:"false" format:"uuid"`
	RoleReverse bool        `query:"role_reverse,omitempty" required:"false" doc:"When role_id is provided, if this is true, it will return the permissions that the role does not have"`
}

func NewPermissionQueryResource(
	db database.Dbx,
) *QueryResource[models.Permission, uuid.UUID, PermissionsFilter] {
	return NewQueryResource[models.Permission, uuid.UUID](
		db,
		repository.PermissionBuilder,
		func(qs sq.SelectBuilder, filter *PermissionsFilter) sq.SelectBuilder {
			if filter == nil {
				return qs
			}
			if filter.Q != "" {
				qs = qs.Where(
					sq.Or{
						sq.ILike{"auth.permissions.name": "%" + filter.Q + "%"},
						sq.ILike{"auth.permissions.description": "%" + filter.Q + "%"},
					},
				)

			}
			if len(filter.Names) > 0 {
				qs = qs.Where(sq.Eq{"auth.permissions.name": filter.Names})
			}
			if len(filter.Ids) > 0 {
				qs = qs.Where(sq.Eq{"auth.permissions.id": filter.Ids})
			}

			if filter.RoleId != uuid.Nil {
				if filter.RoleReverse {
					qs = qs.LeftJoin(
						"auth.role_permissions"+" on "+"auth.permissions.id"+" = "+"auth.role_permissions"+"."+"permission_id"+" and "+"auth.role_permissions"+"."+"role_id"+" = ?",
						filter.RoleId,
					)
					qs = qs.Where("auth.role_permissions.permission_id is null")

				} else {
					qs = qs.Join("auth.role_permissions on auth.permissions.id = auth.role_permissions.permission_id and auth.role_permissions.role_id = ?", filter.RoleId).
						Where(sq.Eq{"auth.role_permissions.role_id": filter.RoleId})

				}
			}
			return qs
		},
		nil,
		nil,
	)
}
