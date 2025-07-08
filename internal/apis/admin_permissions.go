package apis

import (
	"context"
	"fmt"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/stores"
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
	user, err := api.App().Adapter().User().FindUserByID(ctx, id)
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
	permission, err := api.App().Adapter().Rbac().FindPermissionById(ctx, permissionId)
	if err != nil {
		return nil, err
	}
	if permission == nil {
		return nil, huma.Error404NotFound("Permission not found")
	}
	err = api.App().Adapter().Rbac().DeletePermission(ctx, permission.ID)
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
	user, err := api.App().Adapter().User().FindUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, huma.Error404NotFound("User not found")
	}

	permissionIds := utils.ParseValidUUIDs(input.Body.PermissionIds...)

	permissions, err := api.App().Adapter().Rbac().FindPermissionsByIds(ctx, permissionIds)
	if err != nil {
		return nil, err
	}
	if len(permissions) != len(permissionIds) {
		return nil, huma.Error404NotFound("Permission not found")
	}
	err = api.App().Adapter().Rbac().CreateUserPermissions(ctx, user.ID, permissionIds...)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

type UserPermissionsListFilter struct {
	UserId  string `path:"user-id" format:"uuid"`
	Reverse bool   `query:"reverse,omitempty"`
}
type UserPermissionsListParams struct {
	PaginatedInput
	UserPermissionsListFilter
	SortParams
}

func (api *Api) AdminUserPermissionSourceList(ctx context.Context, input *struct {
	UserPermissionsListParams
}) (*ApiPaginatedOutput[*PermissionSource], error) {
	id, err := uuid.Parse(input.UserId)
	if err != nil {
		return nil, err
	}
	limit := input.PerPage
	offset := input.Page * input.PerPage
	var userPermissionSources []*models.PermissionSource
	var count int64
	if input.Reverse {
		userPermissionSources, err = api.App().Adapter().Rbac().ListUserNotPermissionsSource(ctx, id, limit, offset)
		if err != nil {
			return nil, err
		}
		count, err = api.App().Adapter().Rbac().CountNotUserPermissionSource(ctx, id)
		if err != nil {
			return nil, err
		}
	} else {
		userPermissionSources, err = api.App().Adapter().Rbac().ListUserPermissionsSource(ctx, id, limit, offset)
		if err != nil {
			return nil, err
		}
		count, err = api.App().Adapter().Rbac().CountUserPermissionSource(ctx, id)
		if err != nil {
			return nil, err
		}
	}
	return &ApiPaginatedOutput[*PermissionSource]{
		Body: ApiPaginatedResponse[*PermissionSource]{

			Data: mapper.Map(userPermissionSources, FromModelPermissionSource),
			Meta: ApiGenerateMeta(&input.PaginatedInput, count),
		},
	}, nil
}

type PermissionsListFilter struct {
	Q              string   `query:"q,omitempty" required:"false"`
	Ids            []string `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	Names          []string `query:"names,omitempty" required:"false" minimum:"1" maximum:"100"`
	RoleId         string   `query:"role_id,omitempty" required:"false" format:"uuid"`
	RoleReverse    bool     `query:"role_reverse,omitempty" required:"false" doc:"When role_id is provided, if this is true, it will return the permissions that the role does not have"`
	ProductID      string   `query:"product_id,omitempty" required:"false"`
	ProductReverse bool     `query:"product_reverse,omitempty" required:"false" doc:"When product_id is provided, if this is true, it will return the permissions that the product does not have"`
}
type PermissionsListParams struct {
	PaginatedInput
	PermissionsListFilter
	SortParams
}

func (api *Api) AdminPermissionsList(ctx context.Context, input *struct {
	PermissionsListParams
}) (*ApiPaginatedOutput[*Permission], error) {
	store := api.App().Adapter().Rbac()
	fmt.Println(input)
	filter := new(stores.PermissionFilter)
	filter.Page = input.PerPage
	filter.PerPage = input.Page
	filter.Ids = utils.ParseValidUUIDs(input.Ids...)
	filter.Names = input.Names
	filter.Q = input.Q
	if len(input.RoleId) > 0 {
		roleId, err := uuid.Parse(input.RoleId)
		if err != nil && input.RoleId != "" {
			return nil, huma.Error400BadRequest("Invalid role ID format", err)
		} else {
			filter.RoleId = roleId
		}
	}
	filter.RoleReverse = input.RoleReverse
	filter.SortBy = input.SortBy
	filter.SortOrder = input.SortOrder
	filter.ProductID = input.ProductID
	filter.ProductReverse = input.ProductReverse
	permissions, err := store.ListPermissions(ctx, filter)
	if err != nil {
		return nil, err
	}
	count, err := store.CountPermissions(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &ApiPaginatedOutput[*Permission]{
		Body: ApiPaginatedResponse[*Permission]{

			Data: mapper.Map(permissions, FromModelPermission),
			Meta: ApiGenerateMeta(&input.PaginatedInput, count),
		},
	}, nil

}

type PermissionCreateInput struct {
	Name        string  `json:"name" required:"true"`
	Description *string `json:"description,omitempty"`
}

func (api *Api) AdminPermissionsCreate(ctx context.Context, input *struct {
	Body PermissionCreateInput
}) (*struct{ Body Permission }, error) {
	store := api.App().Adapter().Rbac()
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
	return &struct{ Body Permission }{
		Body: *FromModelPermission(data),
	}, nil
}

func (api *Api) AdminPermissionsDelete(ctx context.Context, input *struct {
	ID string `path:"id" format:"uuid" required:"true"`
}) (*struct {
}, error) {
	store := api.App().Adapter().Rbac()
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
	checker := api.App().Checker()
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
	Body Permission
}, error) {
	store := api.App().Adapter().Rbac()
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
	checker := api.App().Checker()
	ok, err := checker.CannotBeAdminOrBasicName(ctx, permission.Name)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, huma.Error400BadRequest("Cannot update the admin or basic permission")
	}
	err = store.UpdatePermission(ctx, permission.ID, &stores.UpdatePermissionDto{
		Name:        input.Body.Name,
		Description: input.Description,
	})

	if err != nil {
		return nil, err
	}
	return &struct{ Body Permission }{
		Body: *FromModelPermission(permission),
	}, nil
}

func (api *Api) AdminPermissionsGet(ctx context.Context, input *struct {
	ID string `path:"id" format:"uuid" required:"true"`
}) (*struct {
	Body *Permission
}, error) {
	id, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, err
	}
	permission, err := api.App().Adapter().Rbac().FindPermissionById(ctx, id)
	if err != nil {
		return nil, err
	}
	if permission == nil {
		return nil, huma.Error404NotFound("Permission not found")
	}
	return &struct{ Body *Permission }{
		Body: FromModelPermission(permission),
	}, nil
}
