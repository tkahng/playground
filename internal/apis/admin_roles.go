package apis

import (
	"context"
	"slices"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
	"github.com/tkahng/authgo/internal/tools/utils"
)

func (api *Api) AdminRolesList(ctx context.Context, input *struct {
	shared.RolesListParams
}) (*shared.PaginatedOutput[*shared.RoleWithPermissions], error) {
	store := api.app.Rbac().Store()
	roles, err := store.ListRoles(ctx, &input.RolesListParams)
	if err != nil {
		return nil, err
	}
	if slices.Contains(input.Expand, "permissions") {
		roleIds := mapper.Map(roles, func(r *models.Role) uuid.UUID {
			return r.ID
		})

		data, err := store.LoadRolePermissions(ctx, roleIds...)
		if err != nil {
			return nil, err
		}
		for idx, role := range roles {
			role.Permissions = data[idx]
		}

	}
	count, err := store.CountRoles(ctx, &input.RoleListFilter)
	if err != nil {
		return nil, err
	}
	return &shared.PaginatedOutput[*shared.RoleWithPermissions]{
		Body: shared.PaginatedResponse[*shared.RoleWithPermissions]{
			Data: mapper.Map(roles, func(r *models.Role) *shared.RoleWithPermissions {
				return &shared.RoleWithPermissions{
					Role:        shared.FromModelRole(r),
					Permissions: mapper.Map(r.Permissions, shared.FromModelPermission),
				}
			}),
			Meta: shared.GenerateMeta(&input.PaginatedInput, count),
		},
	}, nil

}

type RoleCreateInput struct {
	Name        string  `json:"name" required:"true"`
	Description *string `json:"description,omitempty"`
}

func (api *Api) AdminRolesCreate(ctx context.Context, input *struct {
	Body RoleCreateInput
}) (*struct {
	Body shared.Role
}, error) {
	data, err := api.app.Rbac().Store().FindRoleByName(ctx, input.Body.Name)
	if err != nil {
		return nil, err
	}
	if data != nil {
		return nil, huma.Error409Conflict("Role already exists")
	}
	role, err := api.app.Rbac().Store().CreateRole(ctx, &shared.CreateRoleDto{
		Name:        input.Body.Name,
		Description: input.Body.Description,
	})
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, huma.Error500InternalServerError("Failed to create role")
	}
	return &struct{ Body shared.Role }{
		Body: *shared.FromModelRole(role),
	}, nil
}

func (api *Api) AdminRolesDelete(ctx context.Context, input *struct {
	RoleID string `path:"id" format:"uuid" required:"true"`
}) (*struct{}, error) {
	id, err := uuid.Parse(input.RoleID)
	if err != nil {
		return nil, err
	}
	role, err := api.app.Rbac().Store().FindRoleById(ctx, id)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, huma.Error404NotFound("Role not found")
	}
	// Check if the user is trying to delete the admin or basic role
	checker := api.app.Checker()
	err = checker.CannotBeAdminOrBasicName(ctx, role.Name)
	if err != nil {
		return nil, err
	}
	err = api.app.Rbac().Store().DeleteRole(ctx, role.ID)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (api *Api) AdminRolesUpdate(ctx context.Context, input *struct {
	RoleID string `path:"id" format:"uuid" required:"true"`
	Body   RoleCreateInput
}) (*struct {
	Body *shared.Role
}, error) {
	id, err := uuid.Parse(input.RoleID)
	if err != nil {
		return nil, err
	}
	role, err := api.app.Rbac().Store().FindRoleById(ctx, id)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, huma.Error404NotFound("Role not found")
	}
	checker := api.app.Checker()
	err = checker.CannotBeAdminOrBasicName(ctx, role.Name)
	if err != nil {
		return nil, err
	}
	err = api.app.Rbac().Store().UpdateRole(ctx, role.ID, &shared.UpdateRoleDto{
		Name:        input.Body.Name,
		Description: input.Body.Description,
	})

	if err != nil {
		return nil, err
	}
	return &struct{ Body *shared.Role }{
		Body: shared.FromModelRole(role),
	}, nil
}
func (api *Api) AdminUserRolesDelete(ctx context.Context, input *struct {
	UserID string `path:"user-id" format:"uuid" required:"true"`
	RoleID string `path:"role-id" format:"uuid" required:"true"`
}) (*struct{}, error) {
	id, err := uuid.Parse(input.UserID)
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
	roleID, err := uuid.Parse(input.RoleID)
	if err != nil {
		return nil, err
	}
	role, err := api.app.Rbac().Store().FindRoleById(ctx, roleID)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, huma.Error404NotFound("Role not found")
	}
	// Check if the user is trying to remove the super user role from their own account
	checker := api.app.Checker()
	err = checker.CannotBeSuperUserEmailAndRoleName(ctx, user.Email, role.Name)
	if err != nil {
		return nil, err
	}

	err = api.app.Rbac().Store().DeleteUserRole(ctx, user.ID, role.ID)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
func (api *Api) AdminUserRolesCreate(ctx context.Context, input *struct {
	UserID string `path:"user-id" format:"uuid" required:"true"`
	Body   RoleIdsInput
}) (*struct{}, error) {
	id, err := uuid.Parse(input.UserID)
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

	roleIds := utils.ParseValidUUIDs(input.Body.RolesIds)

	roles, err := api.app.Rbac().Store().FindRolesByIds(ctx, roleIds)
	if err != nil {
		return nil, err
	}
	newRoleIds := []uuid.UUID{}
	for _, role := range roles {
		newRoleIds = append(newRoleIds, role.ID)
	}
	err = api.app.Rbac().Store().CreateUserRoles(ctx, user.ID, newRoleIds...)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

type RoleIdsInput struct {
	RolesIds []string `json:"role_ids" minimum:"1" maximum:"100" format:"uuid" required:"true"`
}

// func (api *Api) AdminUserRolesUpdate(ctx context.Context, input *struct {
// 	UserID string       `path:"user-id" format:"uuid" required:"true"`
// 	Body   RoleIdsInput `json:"body" required:"true"`
// }) (*shared.PaginatedOutput[*shared.Role], error) {
// 	db := api.app.Db()
// 	id, err := uuid.Parse(input.UserID)
// 	if err != nil {
// 		return nil, err
// 	}
// 	user, err := api.app.User().Store().FindUserById(ctx, id)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if user == nil {
// 		return nil, huma.Error404NotFound("User not found")
// 	}
// 	roleIds := make([]uuid.UUID, len(input.Body.RolesIds))
// 	for i, id := range input.Body.RolesIds {
// 		id, err := uuid.Parse(id)
// 		if err != nil {
// 			return nil, err
// 		}
// 		roleIds[i] = id
// 	}
// 	roles, err := api.app.Rbac().Store().FindRolesByIds(ctx, roleIds)
// 	if err != nil {
// 		return nil, err
// 	}
// 	_, err = crudrepo.UserRole.DeleteReturn(
// 		ctx,
// 		db,
// 		&map[string]any{
// 			"user_id": map[string]any{
// 				"_eq": id.String(),
// 			},
// 		},
// 	)
// 	if err != nil {
// 		return nil, err
// 	}
// 	newRoleIds := []uuid.UUID{}
// 	for _, role := range roles {
// 		newRoleIds = append(newRoleIds, role.ID)
// 	}
// 	err = api.app.Rbac().Store().CreateUserRoles(ctx, user.ID, newRoleIds...)
// 	if err != nil {
// 		return nil, err
// 	}
// 	output := shared.PaginatedOutput[*shared.Role]{
// 		Body: shared.PaginatedResponse[*shared.Role]{
// 			Data: mapper.Map(roles, shared.FromCrudRole),
// 		},
// 	}
// 	return &output, nil
// }

type RolePermissionsUpdateInput struct {
	PermissionIDs []string `json:"permission_ids" minimum:"0" maximum:"100" format:"uuid" required:"true"`
}

func (api *Api) AdminRolesGet(ctx context.Context, input *struct {
	RoleID string   `path:"id" format:"uuid" required:"true"`
	Expand []string `query:"expand" required:"false" minimum:"1" maximum:"100" enum:"permissions"`
}) (*struct {
	Body shared.RoleWithPermissions
}, error) {
	id, err := uuid.Parse(input.RoleID)
	if err != nil {
		return nil, err
	}
	role, err := api.app.Rbac().Store().FindRoleById(ctx, id)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, huma.Error404NotFound("Role not found")
	}
	if len(input.Expand) > 0 {
		if slices.Contains(input.Expand, "permissions") {
			rolePermissions, err := api.app.Rbac().Store().LoadRolePermissions(ctx, role.ID)
			if err != nil {
				return nil, err
			}
			if len(rolePermissions) > 0 {
				role.Permissions = rolePermissions[0]
			}
		}
	}
	return &struct{ Body shared.RoleWithPermissions }{
		Body: *shared.FromModelRoleWithPermissions(role),
	}, nil
}

func (api *Api) AdminRolesCreatePermissions(ctx context.Context, input *struct {
	RoleID string `path:"id" format:"uuid" required:"true"`
	Body   RolePermissionsUpdateInput
}) (*struct {
}, error) {
	id, err := uuid.Parse(input.RoleID)
	if err != nil {
		return nil, err
	}
	role, err := api.app.Rbac().Store().FindRoleById(ctx, id)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, huma.Error404NotFound("Role not found")
	}
	permissionIds := utils.ParseValidUUIDs(input.Body.PermissionIDs)

	err = api.app.Rbac().Store().CreateRolePermissions(ctx, role.ID, permissionIds...)

	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (api *Api) AdminRolesDeletePermissions(ctx context.Context, input *struct {
	RoleId       string `path:"roleId" format:"uuid" required:"true"`
	PermissionId string `path:"permissionId" format:"uuid" required:"true"`
}) (*struct {
}, error) {
	id, err := uuid.Parse(input.RoleId)
	if err != nil {
		return nil, err
	}
	role, err := api.app.Rbac().Store().FindRoleById(ctx, id)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, huma.Error404NotFound("Role not found")
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
	// Check if the user is trying to remove the admin permission from the admin role
	checker := api.app.Checker()
	err = checker.CannotBeAdminOrBasicRoleAndPermissionName(ctx, role.Name, permission.Name)
	if err != nil {
		return nil, err
	}
	err = api.app.Rbac().Store().DeleteRolePermissions(ctx, role.ID, permission.ID)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
