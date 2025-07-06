package apis

import (
	"context"
	"slices"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/stores"
	"github.com/tkahng/authgo/internal/tools/mapper"
	"github.com/tkahng/authgo/internal/tools/utils"
)

type Role struct {
	_           struct{}      `db:"roles" json:"-"`
	ID          uuid.UUID     `db:"id" json:"id"`
	Name        string        `db:"name" json:"name"`
	Description *string       `db:"description" json:"description,omitempty"`
	CreatedAt   time.Time     `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time     `db:"updated_at" json:"updated_at"`
	Permissions []*Permission `db:"permissions" src:"id" dest:"role_id" table:"permissions" through:"role_permissions,permission_id,id" json:"permissions,omitempty"`
	Users       []*ApiUser    `db:"users" src:"id" dest:"role_id" table:"users" through:"user_roles,user_id,id" json:"users,omitempty"`
}

func FromModelRole(role *models.Role) *Role {
	if role == nil {
		return nil
	}
	return &Role{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
		Permissions: mapper.Map(role.Permissions, FromModelPermission),
		Users:       mapper.Map(role.Users, FromUserModel),
	}
}

type Permission struct {
	ID          uuid.UUID `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Description *string   `db:"description" json:"description,omitempty"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}
type PermissionSource struct {
	ID          uuid.UUID   `db:"id,pk" json:"id"`
	Name        string      `db:"name" json:"name"`
	Description *string     `db:"description" json:"description"`
	CreatedAt   time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time   `db:"updated_at" json:"updated_at"`
	RoleIDs     []uuid.UUID `db:"role_ids" json:"role_ids"`
	ProductIDs  []string    `db:"product_ids" json:"product_ids"`
	IsDirectly  bool        `db:"is_directly_assigned" json:"is_directly_assigned"`
}

func FromModelPermissionSource(permission *models.PermissionSource) *PermissionSource {
	if permission == nil {
		return nil
	}
	return &PermissionSource{
		ID:          permission.ID,
		Name:        permission.Name,
		Description: permission.Description,
		CreatedAt:   permission.CreatedAt,
		UpdatedAt:   permission.UpdatedAt,
		RoleIDs:     permission.RoleIDs,
		ProductIDs:  permission.ProductIDs,
		IsDirectly:  permission.IsDirectly,
	}
}

func FromModelPermission(permission *models.Permission) *Permission {
	return &Permission{
		ID:          permission.ID,
		Name:        permission.Name,
		Description: permission.Description,
		CreatedAt:   permission.CreatedAt,
		UpdatedAt:   permission.UpdatedAt,
	}
}

type RoleListFilter struct {
	Q         string   `query:"q,omitempty" required:"false"`
	Ids       []string `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	Names     []string `query:"names,omitempty" required:"false" minimum:"1" maximum:"100"`
	UserId    string   `query:"user_id,omitempty" required:"false" format:"uuid"`
	Reverse   string   `query:"reverse,omitempty" required:"false" doc:"When true, it will return the roles that do not match the filter criteria" enum:"user,product"`
	ProductId string   `query:"product_id,omitempty" required:"false"`
}
type RolesListParams struct {
	PaginatedInput
	RoleListFilter
	SortParams
	Expand []string `query:"expand,omitempty" required:"false" minimum:"1" maximum:"100" enum:"users,permissions"`
}

func ToRoleListFilter(input *RolesListParams) (*stores.RoleListFilter, error) {
	filter := &stores.RoleListFilter{}
	filter.Page = input.PerPage
	filter.PerPage = input.Page
	filter.Q = input.Q
	filter.Ids = utils.ParseValidUUIDs(input.Ids...)
	filter.Names = input.Names
	if input.UserId != "" {
		id, err := uuid.Parse(input.UserId)
		if err != nil {
			return nil, err
		}
		filter.UserId = id
	}
	if input.ProductId != "" {
		filter.ProductId = input.ProductId
	}
	filter.Reverse = input.Reverse
	filter.SortBy = input.SortBy
	filter.SortOrder = input.SortOrder
	filter.Expand = input.Expand
	return filter, nil
}

func (api *Api) AdminRolesList(ctx context.Context, input *struct {
	RolesListParams
}) (*ApiPaginatedOutput[*Role], error) {
	store := api.app.Adapter().Rbac()
	filter, err := ToRoleListFilter(&input.RolesListParams)
	if err != nil {
		return nil, err
	}
	roles, err := store.ListRoles(ctx, filter)
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
	count, err := store.CountRoles(ctx, filter)
	if err != nil {
		return nil, err
	}
	return &ApiPaginatedOutput[*Role]{
		Body: ApiPaginatedResponse[*Role]{
			Data: mapper.Map(roles, FromModelRole),
			Meta: ApiGenerateMeta(&input.PaginatedInput, count),
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
	Body Role
}, error) {
	data, err := api.app.Adapter().Rbac().FindRoleByName(ctx, input.Body.Name)
	if err != nil {
		return nil, err
	}
	if data != nil {
		return nil, huma.Error409Conflict("Role already exists")
	}
	role, err := api.app.Adapter().Rbac().CreateRole(ctx, &stores.CreateRoleDto{
		Name:        input.Body.Name,
		Description: input.Body.Description,
	})
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, huma.Error500InternalServerError("Failed to create role")
	}
	return &struct{ Body Role }{
		Body: *FromModelRole(role),
	}, nil
}

func (api *Api) AdminRolesDelete(ctx context.Context, input *struct {
	RoleID string `path:"id" format:"uuid" required:"true"`
}) (*struct{}, error) {
	id, err := uuid.Parse(input.RoleID)
	if err != nil {
		return nil, err
	}
	role, err := api.app.Adapter().Rbac().FindRoleById(ctx, id)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, huma.Error404NotFound("Role not found")
	}
	// Check if the user is trying to delete the admin or basic role
	checker := api.app.Checker()
	ok, err := checker.CannotBeAdminOrBasicName(ctx, role.Name)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, huma.Error400BadRequest("Cannot delete the admin or basic role")
	}
	err = api.app.Adapter().Rbac().DeleteRole(ctx, role.ID)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (api *Api) AdminRolesUpdate(ctx context.Context, input *struct {
	RoleID string `path:"id" format:"uuid" required:"true"`
	Body   RoleCreateInput
}) (*struct {
	Body *Role
}, error) {
	id, err := uuid.Parse(input.RoleID)
	if err != nil {
		return nil, err
	}
	role, err := api.app.Adapter().Rbac().FindRoleById(ctx, id)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, huma.Error404NotFound("Role not found")
	}
	checker := api.app.Checker()
	ok, err := checker.CannotBeAdminOrBasicName(ctx, role.Name)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, huma.Error400BadRequest("Cannot update the admin or basic role")
	}
	err = api.app.Adapter().Rbac().UpdateRole(ctx, role.ID, &stores.UpdateRoleDto{
		Name:        input.Body.Name,
		Description: input.Body.Description,
	})

	if err != nil {
		return nil, err
	}
	return &struct{ Body *Role }{
		Body: FromModelRole(role),
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
	user, err := api.app.Adapter().User().FindUserByID(ctx, id)
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
	role, err := api.app.Adapter().Rbac().FindRoleById(ctx, roleID)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, huma.Error404NotFound("Role not found")
	}
	// Check if the user is trying to remove the super user role from their own account
	checker := api.app.Checker()
	ok, err := checker.CannotBeSuperUserEmailAndRoleName(ctx, user.Email, role.Name)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, huma.Error400BadRequest("Cannot remove the super user role from your own account")
	}

	err = api.app.Adapter().Rbac().DeleteUserRole(ctx, user.ID, role.ID)
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
	user, err := api.app.Adapter().User().FindUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, huma.Error404NotFound("User not found")
	}

	roleIds := utils.ParseValidUUIDs(input.Body.RolesIds...)

	roles, err := api.app.Adapter().Rbac().FindRolesByIds(ctx, roleIds)
	if err != nil {
		return nil, err
	}
	newRoleIds := []uuid.UUID{}
	for _, role := range roles {
		newRoleIds = append(newRoleIds, role.ID)
	}
	err = api.app.Adapter().Rbac().CreateUserRoles(ctx, user.ID, newRoleIds...)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

type RoleIdsInput struct {
	RolesIds []string `json:"role_ids" minimum:"1" maximum:"100" format:"uuid" required:"true"`
}

type PermissionIdsInput struct {
	PermissionIDs []string `json:"permission_ids" minimum:"0" maximum:"100" format:"uuid" required:"true"`
}

func (api *Api) AdminRolesGet(ctx context.Context, input *struct {
	RoleID string   `path:"id" format:"uuid" required:"true"`
	Expand []string `query:"expand" required:"false" minimum:"1" maximum:"100" enum:"permissions"`
}) (*struct {
	Body Role
}, error) {
	id, err := uuid.Parse(input.RoleID)
	if err != nil {
		return nil, err
	}
	role, err := api.app.Adapter().Rbac().FindRoleById(ctx, id)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, huma.Error404NotFound("Role not found")
	}
	if len(input.Expand) > 0 {
		if slices.Contains(input.Expand, "permissions") {
			rolePermissions, err := api.app.Adapter().Rbac().LoadRolePermissions(ctx, role.ID)
			if err != nil {
				return nil, err
			}
			if len(rolePermissions) > 0 {
				role.Permissions = rolePermissions[0]
			}
		}
	}
	return &struct{ Body Role }{
		Body: *FromModelRole(role),
	}, nil
}

func (api *Api) AdminRolesCreatePermissions(ctx context.Context, input *struct {
	RoleID string `path:"id" format:"uuid" required:"true"`
	Body   PermissionIdsInput
}) (*struct {
}, error) {
	id, err := uuid.Parse(input.RoleID)
	if err != nil {
		return nil, err
	}
	role, err := api.app.Adapter().Rbac().FindRoleById(ctx, id)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, huma.Error404NotFound("Role not found")
	}
	permissionIds := utils.ParseValidUUIDs(input.Body.PermissionIDs...)

	err = api.app.Adapter().Rbac().CreateRolePermissions(ctx, role.ID, permissionIds...)

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
	role, err := api.app.Adapter().Rbac().FindRoleById(ctx, id)
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
	permission, err := api.app.Adapter().Rbac().FindPermissionById(ctx, permissionId)
	if err != nil {
		return nil, err
	}
	if permission == nil {
		return nil, huma.Error404NotFound("Permission not found")
	}
	// Check if the user is trying to remove the admin permission from the admin role
	checker := api.app.Checker()
	ok, err := checker.CannotBeAdminOrBasicRoleAndPermissionName(ctx, role.Name, permission.Name)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, huma.Error400BadRequest("Cannot remove the admin permission from the admin role")
	}
	err = api.app.Adapter().Rbac().DeleteRolePermissions(ctx, role.ID, permission.ID)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
