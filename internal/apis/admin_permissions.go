package apis

import (
	"context"
	"fmt"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/crudrepo"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
	"github.com/tkahng/authgo/internal/tools/utils"
)

func (api *Api) AdminUserPermissionsDelete(ctx context.Context, input *struct {
	UserId       string `path:"user-id" format:"uuid" required:"true"`
	PermissionId string `path:"permission-id" format:"uuid" required:"true"`
}) (*struct{}, error) {
	db := api.app.Db()
	id, err := uuid.Parse(input.UserId)
	if err != nil {
		return nil, err
	}
	user, err := queries.FindUserById(ctx, db, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, huma.Error404NotFound("User not found")
	}
	permissionId, err := uuid.Parse(input.PermissionId)
	if err != nil {
		return nil, err
	}
	permission, err := queries.FindPermissionById(ctx, db, permissionId)
	if err != nil {
		return nil, err
	}
	if permission == nil {
		return nil, huma.Error404NotFound("Permission not found")
	}
	_, err = crudrepo.UserPermission.DeleteReturn(
		ctx,
		db,
		&map[string]any{
			"user_id": map[string]any{
				"_eq": id.String(),
			},
			"permission_id": map[string]any{
				"_eq": permissionId.String(),
			},
		},
	)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (api *Api) AdminUserPermissionsCreate(ctx context.Context, input *struct {
	UserId string `path:"user-id" format:"uuid" required:"true"`
	Body   struct {
		PermissionIds []string `json:"permission_ids" minimum:"0" maximum:"50" format:"uuid" required:"true"`
	} `json:"body" required:"true"`
}) (*struct{}, error) {
	db := api.app.Db()
	id, err := uuid.Parse(input.UserId)
	if err != nil {
		return nil, err
	}
	user, err := queries.FindUserById(ctx, db, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, huma.Error404NotFound("User not found")
	}

	permissionIds := utils.ParseValidUUIDs(input.Body.PermissionIds)

	permissions, err := queries.FindPermissionsByIds(ctx, db, permissionIds)
	if err != nil {
		return nil, err
	}
	if len(permissions) != len(permissionIds) {
		return nil, huma.Error404NotFound("Permission not found")
	}
	err = queries.CreateUserPermissions(ctx, db, user.ID, permissionIds...)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (api *Api) AdminUserPermissionSourceList(ctx context.Context, input *struct {
	shared.UserPermissionsListParams
}) (*shared.PaginatedOutput[queries.PermissionSource], error) {
	db := api.app.Db()
	id, err := uuid.Parse(input.UserId)
	if err != nil {
		return nil, err
	}
	limit := input.PerPage
	offset := input.Page * input.PerPage
	var userPermissionSources []queries.PermissionSource
	var count int64
	if input.Reverse {
		userPermissionSources, err = queries.ListUserNotPermissionsSource(ctx, db, id, limit, offset)
		if err != nil {
			return nil, err
		}
		count, err = queries.CountNotUserPermissionSource(ctx, db, id)
		if err != nil {
			return nil, err
		}
	} else {
		userPermissionSources, err = queries.ListUserPermissionsSource(ctx, db, id, limit, offset)
		if err != nil {
			return nil, err
		}
		count, err = queries.CountUserPermissionSource(ctx, db, id)
		if err != nil {
			return nil, err
		}
	}
	return &shared.PaginatedOutput[queries.PermissionSource]{
		Body: shared.PaginatedResponse[queries.PermissionSource]{

			Data: userPermissionSources,
			Meta: shared.GenerateMeta(input.PaginatedInput, count),
		},
	}, nil
}

func (api *Api) AdminPermissionsList(ctx context.Context, input *struct {
	shared.PermissionsListParams
}) (*shared.PaginatedOutput[*shared.Permission], error) {
	db := api.app.Db()
	fmt.Println(input)
	permissions, err := queries.ListPermissions(ctx, db, &input.PermissionsListParams)
	if err != nil {
		return nil, err
	}
	count, err := queries.CountPermissions(ctx, db, &input.PermissionsListFilter)
	if err != nil {
		return nil, err
	}

	return &shared.PaginatedOutput[*shared.Permission]{
		Body: shared.PaginatedResponse[*shared.Permission]{

			Data: mapper.Map(permissions, shared.FromCrudPermission),
			Meta: shared.GenerateMeta(input.PaginatedInput, count),
		},
	}, nil

}

type PermissionCreateInput struct {
	Name        string  `json:"name" required:"true"`
	Description *string `json:"description,omitempty"`
}

func (api *Api) AdminPermissionsCreate(ctx context.Context, input *struct {
	Body PermissionCreateInput
}) (*struct{ Body shared.Permission }, error) {
	dbx := api.app.Db()
	permission, err := crudrepo.Permission.GetOne(
		ctx,
		dbx,
		&map[string]any{
			"name": map[string]any{
				"_eq": input.Body.Name,
			},
		},
	)
	if err != nil {
		return nil, err

	}
	if permission != nil {
		return nil, huma.Error409Conflict("Permission already exists")
	}
	data, err := queries.CreatePermission(ctx, dbx, &queries.CreatePermissionDto{
		Name:        input.Body.Name,
		Description: input.Body.Description,
	})
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, huma.Error500InternalServerError("Failed to create permission")
	}
	return &struct{ Body shared.Permission }{
		Body: *shared.FromCrudPermission(data),
	}, nil
}

func (api *Api) AdminPermissionsDelete(ctx context.Context, input *struct {
	ID string `path:"id" format:"uuid" required:"true"`
}) (*struct {
}, error) {
	db := api.app.Db()
	id, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, err
	}
	permission, err := queries.FindPermissionById(ctx, db, id)
	if err != nil {
		return nil, err
	}
	if permission == nil {
		return nil, huma.Error404NotFound("Permission not found")
	}
	// Check if the permission is not admin or basic
	checker := api.app.Checker()
	err = checker.CannotBeAdminOrBasicName(ctx, permission.Name)
	if err != nil {
		return nil, err
	}
	err = queries.DeletePermission(ctx, db, permission.ID)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (api *Api) AdminPermissionsUpdate(ctx context.Context, input *struct {
	ID          string `path:"id" format:"uuid" required:"true"`
	Body        PermissionCreateInput
	Description *string `json:"description,omitempty"`
}) (*struct {
	Body shared.Permission
}, error) {
	db := api.app.Db()
	id, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, err
	}
	permission, err := queries.FindPermissionById(ctx, db, id)
	if err != nil {
		return nil, err
	}
	if permission == nil {
		return nil, huma.Error404NotFound("Permission not found")
	}
	checker := api.app.Checker()
	err = checker.CannotBeAdminOrBasicName(ctx, permission.Name)
	if err != nil {
		return nil, err
	}
	err = queries.UpdatePermission(ctx, db, permission.ID, &queries.UpdatePermissionDto{
		Name:        input.Body.Name,
		Description: input.Description,
	})

	if err != nil {
		return nil, err
	}
	return &struct{ Body shared.Permission }{
		Body: *shared.FromCrudPermission(permission),
	}, nil
}

func (api *Api) AdminPermissionsGet(ctx context.Context, input *struct {
	ID string `path:"id" format:"uuid" required:"true"`
}) (*struct {
	Body *shared.Permission
}, error) {
	db := api.app.Db()
	id, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, err
	}
	permission, err := queries.FindPermissionById(ctx, db, id)
	if err != nil {
		return nil, err
	}
	if permission == nil {
		return nil, huma.Error404NotFound("Permission not found")
	}
	return &struct{ Body *shared.Permission }{
		Body: shared.FromCrudPermission(permission),
	}, nil
}
