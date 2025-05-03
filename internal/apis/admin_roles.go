package apis

import (
	"context"
	"net/http"
	"slices"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
)

func (api *Api) AdminRolesOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "admin-roles",
		Method:      http.MethodGet,
		Path:        path,
		Summary:     "Admin roles",
		Description: "List of roles",
		Tags:        []string{"Admin", "Roles"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (api *Api) AdminRolesList(ctx context.Context, input *struct {
	shared.RolesListParams
}) (*shared.PaginatedOutput[*shared.RoleWithPermissions], error) {
	db := api.app.Db()
	roles, err := queries.ListRoles(ctx, db, &input.RolesListParams)
	if err != nil {
		return nil, err
	}
	if slices.Contains(input.Expand, "permissions") {
		roleIds := mapper.Map(roles, func(r *models.Role) uuid.UUID {
			return r.ID
		})
		data, err := queries.LoadRolePermissions(ctx, db, roleIds...)
		if err != nil {
			return nil, err
		}
		for idx, role := range roles {
			role.Permissions = data[idx]
		}

	}
	count, err := queries.CountRoles(ctx, db, &input.RoleListFilter)
	if err != nil {
		return nil, err
	}
	return &shared.PaginatedOutput[*shared.RoleWithPermissions]{
		Body: shared.PaginatedResponse[*shared.RoleWithPermissions]{
			Data: mapper.Map(roles, func(r *models.Role) *shared.RoleWithPermissions {
				return &shared.RoleWithPermissions{
					Role:        shared.FromCrudRole(r),
					Permissions: mapper.Map(r.Permissions, shared.FromCrudPermission),
				}
			}),
			Meta: shared.GenerateMeta(input.PaginatedInput, count),
		},
	}, nil

}

func (api *Api) AdminRolesCreateOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "admin-roles-create",
		Method:      http.MethodPost,
		Path:        path,
		Summary:     "Create role",
		Description: "Create role",
		Tags:        []string{"Admin", "Roles"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
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
	dbx := api.app.Db()
	data, err := repository.Role.GetOne(
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
	if data != nil {
		return nil, huma.Error409Conflict("Role already exists")
	}
	role, err := queries.CreateRole(ctx, dbx, &queries.CreateRoleDto{
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
		Body: *shared.FromCrudRole(role),
	}, nil
}

func (api *Api) AdminRolesDeleteOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "admin-roles-delete",
		Method:      http.MethodDelete,
		Path:        path,
		Summary:     "Delete role",
		Description: "Delete role",
		Tags:        []string{"Admin", "Roles"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (api *Api) AdminRolesDelete(ctx context.Context, input *struct {
	RoleID string `path:"id" format:"uuid" required:"true"`
}) (*struct{}, error) {
	db := api.app.Db()
	id, err := uuid.Parse(input.RoleID)
	if err != nil {
		return nil, err
	}
	role, err := repository.Role.GetOne(ctx, db, &map[string]any{
		"id": map[string]any{
			"_eq": id.String(),
		},
	})
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, huma.Error404NotFound("Role not found")
	}
	// Check if the user is trying to delete the admin or basic role
	checker := api.app.NewChecker(ctx)
	err = checker.CannotBeAdminOrBasicName(role.Name)
	if err != nil {
		return nil, err
	}
	err = queries.DeleteRole(ctx, db, role.ID)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (api *Api) AdminRolesUpdateOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "admin-roles-update",
		Method:      http.MethodPut,
		Path:        path,
		Summary:     "Update role",
		Description: "Update role",
		Tags:        []string{"Admin", "Roles"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (api *Api) AdminRolesUpdate(ctx context.Context, input *struct {
	RoleID string `path:"id" format:"uuid" required:"true"`
	Body   RoleCreateInput
}) (*struct {
	Body *shared.Role
}, error) {
	db := api.app.Db()
	id, err := uuid.Parse(input.RoleID)
	if err != nil {
		return nil, err
	}
	role, err := repository.Role.GetOne(ctx, db, &map[string]any{
		"id": map[string]any{
			"_eq": id.String(),
		},
	})
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, huma.Error404NotFound("Role not found")
	}
	checker := api.app.NewChecker(ctx)
	err = checker.CannotBeAdminOrBasicName(role.Name)
	if err != nil {
		return nil, err
	}
	err = queries.UpdateRole(ctx, db, role.ID, &queries.UpdateRoleDto{
		Name:        input.Body.Name,
		Description: input.Body.Description,
	})

	if err != nil {
		return nil, err
	}
	return &struct{ Body *shared.Role }{
		Body: shared.FromCrudRole(role),
	}, nil
}
func (api *Api) AdminUserRolesDeleteOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "admin-user-roles-delete",
		Method:      http.MethodDelete,
		Path:        path,
		Summary:     "Delete user roles",
		Description: "Delete user roles",
		Tags:        []string{"Admin", "Roles", "User"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (api *Api) AdminUserRolesDelete(ctx context.Context, input *struct {
	UserID string `path:"user-id" format:"uuid" required:"true"`
	RoleID string `path:"role-id" format:"uuid" required:"true"`
}) (*struct{}, error) {
	db := api.app.Db()
	id, err := uuid.Parse(input.UserID)
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
	roleID, err := uuid.Parse(input.RoleID)
	if err != nil {
		return nil, err
	}
	role, err := repository.Role.GetOne(ctx, db, &map[string]any{
		"id": map[string]any{
			"_eq": roleID.String(),
		},
	})
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, huma.Error404NotFound("Role not found")
	}
	// Check if the user is trying to remove the super user role from their own account
	checker := api.app.NewChecker(ctx)
	err = checker.CannotBeSuperUserEmailAndRoleName(user.Email, role.Name)
	if err != nil {
		return nil, err
	}

	_, err = repository.UserRole.DeleteReturn(
		ctx,
		db,
		&map[string]any{
			"user_id": map[string]any{
				"_eq": id.String(),
			},
			"role_id": map[string]any{
				"_eq": roleID.String(),
			},
		},
	)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
func (api *Api) AdminUserRolesCreateOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "admin-create-user-roles",
		Method:      http.MethodPost,
		Path:        path,
		Summary:     "Create user roles",
		Description: "Create user roles",
		Tags:        []string{"Admin", "Roles", "User"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (api *Api) AdminUserRolesCreate(ctx context.Context, input *struct {
	UserID string `path:"user-id" format:"uuid" required:"true"`
	Body   RoleIdsInput
}) (*struct{}, error) {
	db := api.app.Db()
	id, err := uuid.Parse(input.UserID)
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

	roleIds := queries.ParseUUIDs(input.Body.RolesIds)

	roles, err := repository.Role.Get(
		ctx,
		db,
		&map[string]any{
			"id": map[string]any{
				"_in": roleIds,
			},
		},
		nil,
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}
	newRoleIds := []uuid.UUID{}
	for _, role := range roles {
		newRoleIds = append(newRoleIds, role.ID)
	}
	err = queries.CreateUserRoles(ctx, db, user.ID, newRoleIds...)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (api *Api) AdminUserRolesUpdateOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "admin-update-user-roles",
		Method:      http.MethodPut,
		Path:        path,
		Summary:     "Update user roles",
		Description: "Update user roles",
		Tags:        []string{"Admin", "Roles"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

type RoleIdsInput struct {
	RolesIds []string `json:"role_ids" minimum:"1" maximum:"100" format:"uuid" required:"true"`
}

func (api *Api) AdminUserRolesUpdate(ctx context.Context, input *struct {
	UserID string       `path:"user-id" format:"uuid" required:"true"`
	Body   RoleIdsInput `json:"body" required:"true"`
}) (*shared.PaginatedOutput[*shared.Role], error) {
	db := api.app.Db()
	id, err := uuid.Parse(input.UserID)
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
	roleIds := make([]uuid.UUID, len(input.Body.RolesIds))
	for i, id := range input.Body.RolesIds {
		id, err := uuid.Parse(id)
		if err != nil {
			return nil, err
		}
		roleIds[i] = id
	}
	roles, err := repository.Role.Get(
		ctx,
		db,
		&map[string]any{
			"id": map[string]any{
				"_in": roleIds,
			},
		},
		nil,
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}
	_, err = repository.UserRole.DeleteReturn(
		ctx,
		db,
		&map[string]any{
			"user_id": map[string]any{
				"_eq": id.String(),
			},
		},
	)
	if err != nil {
		return nil, err
	}
	newRoleIds := []uuid.UUID{}
	for _, role := range roles {
		newRoleIds = append(newRoleIds, role.ID)
	}
	err = queries.CreateUserRoles(ctx, db, user.ID, newRoleIds...)
	if err != nil {
		return nil, err
	}
	output := shared.PaginatedOutput[*shared.Role]{
		Body: shared.PaginatedResponse[*shared.Role]{
			Data: mapper.Map(roles, shared.FromCrudRole),
		},
	}
	return &output, nil
}

func (api *Api) AdminRolesUpdatePermissionsOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "admin-roles-update-permissions",
		Method:      http.MethodPut,
		Path:        path,
		Summary:     "Update role permissions",
		Description: "Update role permissions",
		Tags:        []string{"Admin", "Roles"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

type RolePermissionsUpdateInput struct {
	PermissionIDs []string `json:"permission_ids" minimum:"0" maximum:"100" format:"uuid" required:"true"`
}

func (api *Api) AdminRolesUpdatePermissions(ctx context.Context, input *struct {
	RoleID string `path:"id" format:"uuid" required:"true"`
	Body   RolePermissionsUpdateInput
}) (*struct {
	Body *shared.Role
}, error) {
	db := api.app.Db()
	id, err := uuid.Parse(input.RoleID)
	if err != nil {
		return nil, err
	}
	role, err := repository.Role.GetOne(ctx, db, &map[string]any{
		"id": map[string]any{
			"_eq": id.String(),
		},
	})
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, huma.Error404NotFound("Role not found")
	}
	permissionIds := make([]uuid.UUID, len(input.Body.PermissionIDs))
	for i, id := range input.Body.PermissionIDs {
		id, err := uuid.Parse(id)
		if err != nil {
			continue
		}
		permissionIds[i] = id
	}
	err = queries.DeleteRolePermissions(ctx, db, role.ID)
	if err != nil {
		return nil, err
	}
	err = queries.CreateRolePermissions(ctx, db, role.ID, permissionIds...)
	if err != nil {
		return nil, err
	}
	return &struct{ Body *shared.Role }{
		Body: shared.FromCrudRole(role),
	}, nil
}

func (api *Api) AdminRolesGetOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "admin-roles-get",
		Method:      http.MethodGet,
		Path:        path,
		Summary:     "Get role",
		Description: "Get role",
		Tags:        []string{"Admin", "Roles"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (api *Api) AdminRolesGet(ctx context.Context, input *struct {
	RoleID string   `path:"id" format:"uuid" required:"true"`
	Expand []string `query:"expand" required:"false" minimum:"1" maximum:"100" enum:"permissions"`
}) (*struct {
	Body shared.RoleWithPermissions
}, error) {
	db := api.app.Db()
	id, err := uuid.Parse(input.RoleID)
	if err != nil {
		return nil, err
	}
	role, err := repository.Role.GetOne(ctx, db, &map[string]any{
		"id": map[string]any{
			"_eq": id.String(),
		},
	})
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, huma.Error404NotFound("Role not found")
	}
	if len(input.Expand) > 0 {
		if slices.Contains(input.Expand, "permissions") {
			rolePermissions, err := queries.LoadRolePermissions(ctx, db, role.ID)
			if err != nil {
				return nil, err
			}
			if len(rolePermissions) > 0 {
				role.Permissions = rolePermissions[0]
			}
		}
	}
	return &struct{ Body shared.RoleWithPermissions }{
		Body: *shared.FromCrudRoleWithPermissions(role),
	}, nil
}

func (api *Api) AdminRolesCreatePermissionsOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "admin-roles-create-permissions",
		Method:      http.MethodPost,
		Path:        path,
		Summary:     "Create role permissions",
		Description: "Create role permissions",
		Tags:        []string{"Admin", "Roles", "Permissions"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (api *Api) AdminRolesCreatePermissions(ctx context.Context, input *struct {
	RoleID string `path:"id" format:"uuid" required:"true"`
	Body   RolePermissionsUpdateInput
}) (*struct {
}, error) {
	db := api.app.Db()
	id, err := uuid.Parse(input.RoleID)
	if err != nil {
		return nil, err
	}
	role, err := repository.Role.GetOne(ctx, db, &map[string]any{
		"id": map[string]any{
			"_eq": id.String(),
		},
	})
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, huma.Error404NotFound("Role not found")
	}
	permissionIds := queries.ParseUUIDs(input.Body.PermissionIDs)

	err = queries.CreateRolePermissions(ctx, db, role.ID, permissionIds...)

	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (api *Api) AdminRolesDeletePermissionsOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "admin-roles-delete-permissions",
		Method:      http.MethodDelete,
		Path:        path,
		Summary:     "Delete role permissions",
		Description: "Delete role permissions",
		Tags:        []string{"Admin", "Roles", "Permissions"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (api *Api) AdminRolesDeletePermissions(ctx context.Context, input *struct {
	RoleId       string `path:"roleId" format:"uuid" required:"true"`
	PermissionId string `path:"permissionId" format:"uuid" required:"true"`
}) (*struct {
}, error) {
	db := api.app.Db()
	id, err := uuid.Parse(input.RoleId)
	if err != nil {
		return nil, err
	}
	role, err := repository.Role.GetOne(ctx, db, &map[string]any{
		"id": map[string]any{
			"_eq": id.String(),
		},
	})
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
	permission, err := queries.FindPermissionById(ctx, db, permissionId)
	if err != nil {
		return nil, err
	}
	if permission == nil {
		return nil, huma.Error404NotFound("Permission not found")
	}
	// Check if the user is trying to remove the admin permission from the admin role
	checker := api.app.NewChecker(ctx)
	err = checker.CannotBeAdminOrBasicRoleAndPermissionName(role.Name, permission.Name)
	if err != nil {
		return nil, err
	}
	_, err = repository.RolePermission.DeleteReturn(
		ctx,
		db,
		&map[string]any{
			"role_id": map[string]any{
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
