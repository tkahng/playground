package apis

import (
	"context"
	"fmt"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
	"github.com/tkahng/authgo/internal/tools/utils"
)

func (api *Api) AdminUserPermissionsDelete(ctx context.Context, input *struct {
	UserId       string `path:"user-id" format:"uuid" required:"true"`
	PermissionId string `path:"permission-id" format:"uuid" required:"true"`
}) (*struct{}, error) {

	id, err := uuid.Parse(input.UserId)
	if err != nil {
		return nil, err
	}
	user, err := api.app.User().Store().FindUserById(ctx, id)
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
	permission, err := api.app.Rbac().Store().FindPermissionById(ctx, permissionId)
	if err != nil {
		return nil, err
	}
	if permission == nil {
		return nil, huma.Error404NotFound("Permission not found")
	}
	err = api.app.Rbac().Store().DeletePermission(ctx, permission.ID)
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
	id, err := uuid.Parse(input.UserId)
	if err != nil {
		return nil, err
	}
	user, err := api.app.User().Store().FindUserById(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, huma.Error404NotFound("User not found")
	}

	permissionIds := utils.ParseValidUUIDs(input.Body.PermissionIds)

	permissions, err := api.app.Rbac().Store().FindPermissionsByIds(ctx, permissionIds)
	if err != nil {
		return nil, err
	}
	if len(permissions) != len(permissionIds) {
		return nil, huma.Error404NotFound("Permission not found")
	}
	err = api.app.Rbac().Store().CreateUserPermissions(ctx, user.ID, permissionIds...)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (api *Api) AdminUserPermissionSourceList(ctx context.Context, input *struct {
	shared.UserPermissionsListParams
}) (*shared.PaginatedOutput[shared.PermissionSource], error) {
	id, err := uuid.Parse(input.UserId)
	if err != nil {
		return nil, err
	}
	limit := input.PerPage
	offset := input.Page * input.PerPage
	var userPermissionSources []shared.PermissionSource
	var count int64
	if input.Reverse {
		userPermissionSources, err = api.app.Rbac().Store().ListUserNotPermissionsSource(ctx, id, limit, offset)
		if err != nil {
			return nil, err
		}
		count, err = api.app.Rbac().Store().CountNotUserPermissionSource(ctx, id)
		if err != nil {
			return nil, err
		}
	} else {
		userPermissionSources, err = api.app.Rbac().Store().ListUserPermissionsSource(ctx, id, limit, offset)
		if err != nil {
			return nil, err
		}
		count, err = api.app.Rbac().Store().CountUserPermissionSource(ctx, id)
		if err != nil {
			return nil, err
		}
	}
	return &shared.PaginatedOutput[shared.PermissionSource]{
		Body: shared.PaginatedResponse[shared.PermissionSource]{

			Data: userPermissionSources,
			Meta: shared.GenerateMeta(&input.PaginatedInput, count),
		},
	}, nil
}

func (api *Api) AdminPermissionsList(ctx context.Context, input *struct {
	shared.PermissionsListParams
}) (*shared.PaginatedOutput[*shared.Permission], error) {
	store := api.app.Rbac().Store()
	fmt.Println(input)
	permissions, err := store.ListPermissions(ctx, &input.PermissionsListParams)
	if err != nil {
		return nil, err
	}
	count, err := store.CountPermissions(ctx, &input.PermissionsListFilter)
	if err != nil {
		return nil, err
	}

	return &shared.PaginatedOutput[*shared.Permission]{
		Body: shared.PaginatedResponse[*shared.Permission]{

			Data: mapper.Map(permissions, shared.FromModelPermission),
			Meta: shared.GenerateMeta(&input.PaginatedInput, count),
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
	store := api.app.Rbac().Store()
	permission, err := store.FindPermissionByName(ctx, input.Body.Name)
	if err != nil {
		return nil, err

	}
	if permission != nil {
		return nil, huma.Error409Conflict("Permission already exists")
	}
	data, err := store.CreatePermission(ctx, input.Body.Name, input.Body.Description)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, huma.Error500InternalServerError("Failed to create permission")
	}
	return &struct{ Body shared.Permission }{
		Body: *shared.FromModelPermission(data),
	}, nil
}

func (api *Api) AdminPermissionsDelete(ctx context.Context, input *struct {
	ID string `path:"id" format:"uuid" required:"true"`
}) (*struct {
}, error) {
	store := api.app.Rbac().Store()
	id, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, err
	}
	permission, err := store.FindPermissionById(ctx, id)
	if err != nil {
		return nil, err
	}
	if permission == nil {
		return nil, huma.Error404NotFound("Permission not found")
	}
	// Check if the permission is not admin or basic
	checker := api.app.Checker()
	ok, err := checker.CannotBeAdminOrBasicName(ctx, permission.Name)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, huma.Error400BadRequest("Cannot delete the admin or basic permission")
	}
	err = store.DeletePermission(ctx, permission.ID)
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
	store := api.app.Rbac().Store()
	id, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, err
	}
	permission, err := store.FindPermissionById(ctx, id)
	if err != nil {
		return nil, err
	}
	if permission == nil {
		return nil, huma.Error404NotFound("Permission not found")
	}
	checker := api.app.Checker()
	ok, err := checker.CannotBeAdminOrBasicName(ctx, permission.Name)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, huma.Error400BadRequest("Cannot update the admin or basic permission")
	}
	err = store.UpdatePermission(ctx, permission.ID, &shared.UpdatePermissionDto{
		Name:        input.Body.Name,
		Description: input.Description,
	})

	if err != nil {
		return nil, err
	}
	return &struct{ Body shared.Permission }{
		Body: *shared.FromModelPermission(permission),
	}, nil
}

func (api *Api) AdminPermissionsGet(ctx context.Context, input *struct {
	ID string `path:"id" format:"uuid" required:"true"`
}) (*struct {
	Body *shared.Permission
}, error) {
	id, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, err
	}
	permission, err := api.app.Rbac().Store().FindPermissionById(ctx, id)
	if err != nil {
		return nil, err
	}
	if permission == nil {
		return nil, huma.Error404NotFound("Permission not found")
	}
	return &struct{ Body *shared.Permission }{
		Body: shared.FromModelPermission(permission),
	}, nil
}
