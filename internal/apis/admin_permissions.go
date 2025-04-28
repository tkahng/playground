package apis

import (
	"context"
	"net/http"

	"github.com/aarondl/opt/omit"
	"github.com/aarondl/opt/omitnull"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
)

func (api *Api) AdminUserPermissionsDeleteOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "admin-user-permissions-delete",
		Method:      http.MethodDelete,
		Path:        path,
		Summary:     "Delete user permission",
		Description: "Delete user permission",
		Tags:        []string{"Admin", "Permissions", "User"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (api *Api) AdminUserPermissionsDelete(ctx context.Context, input *struct {
	UserId       string `path:"user-id" format:"uuid" required:"true"`
	PermissionId string `path:"permission-id" format:"uuid" required:"true"`
}) (*struct{}, error) {
	db := api.app.Db()
	id, err := uuid.Parse(input.UserId)
	if err != nil {
		return nil, err
	}
	user, err := repository.FindUserById(ctx, db, id)
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
	permission, err := repository.FindPermissionById(ctx, db, permissionId)
	if err != nil {
		return nil, err
	}
	if permission == nil {
		return nil, huma.Error404NotFound("Permission not found")
	}
	_, err = models.UserPermissions.Delete(
		models.DeleteWhere.UserPermissions.UserID.EQ(user.ID),
		models.DeleteWhere.UserPermissions.PermissionID.EQ(permission.ID),
	).Exec(ctx, db)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (api *Api) AdminUserPermissionsCreateOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "admin-user-permissions-create",
		Method:      http.MethodPost,
		Path:        path,
		Summary:     "Create user permission",
		Description: "Create user permission",
		Tags:        []string{"Admin", "Permissions", "User"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
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
	user, err := repository.FindUserById(ctx, db, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, huma.Error404NotFound("User not found")
	}

	permissionIds := repository.ParseUUIDs(input.Body.PermissionIds)

	permissions, err := repository.FindPermissionsByIds(ctx, db, permissionIds)
	if err != nil {
		return nil, err
	}
	if len(permissions) != len(permissionIds) {
		return nil, huma.Error404NotFound("Permission not found")
	}

	err = user.AttachPermissions(ctx, db, permissions...)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (api *Api) AdminUserPermissionSourceListOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "admin-user-permission-sources",
		Method:      http.MethodGet,
		Path:        path,
		Summary:     "Admin user permission sources",
		Description: "List of permission sources",
		Tags:        []string{"Admin", "Permissions", "User"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (api *Api) AdminUserPermissionSourceList(ctx context.Context, input *struct {
	shared.UserPermissionsListParams
}) (*shared.PaginatedOutput[repository.PermissionSource], error) {
	db := api.app.Db()
	id, err := uuid.Parse(input.UserId)
	if err != nil {
		return nil, err
	}
	limit := input.PerPage
	offset := input.Page * input.PerPage
	var userPermissionSources []repository.PermissionSource
	var count int64
	if input.Reverse {
		userPermissionSources, err = repository.ListUserNotPermissionsSource(ctx, db, id, limit, offset)
		if err != nil {
			return nil, err
		}
		count, err = repository.CountNotUserPermissionSource(ctx, db, id)
		if err != nil {
			return nil, err
		}
	} else {
		userPermissionSources, err = repository.ListUserPermissionsSource(ctx, db, id, limit, offset)
		if err != nil {
			return nil, err
		}
		count, err = repository.CountUserPermissionSource(ctx, db, id)
		if err != nil {
			return nil, err
		}
	}
	return &shared.PaginatedOutput[repository.PermissionSource]{
		Body: shared.PaginatedResponse[repository.PermissionSource]{

			Data: userPermissionSources,
			Meta: shared.GenerateMeta(input.PaginatedInput, count),
		},
	}, nil
}

func (api *Api) AdminPermissionsListOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "admin-permissions",
		Method:      http.MethodGet,
		Path:        path,
		Summary:     "Admin permissions",
		Description: "List of permissions",
		Tags:        []string{"Admin", "Permissions"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (api *Api) AdminPermissionsList(ctx context.Context, input *struct {
	shared.PermissionsListParams
}) (*shared.PaginatedOutput[*shared.Permission], error) {
	db := api.app.Db()
	permissions, err := repository.ListPermissions(ctx, db, &input.PermissionsListParams)
	if err != nil {
		return nil, err
	}
	count, err := repository.CountPermissions(ctx, db, &input.PermissionsListFilter)
	if err != nil {
		return nil, err
	}

	return &shared.PaginatedOutput[*shared.Permission]{
		Body: shared.PaginatedResponse[*shared.Permission]{

			Data: mapper.Map(permissions, shared.ToPermission),
			Meta: shared.GenerateMeta(input.PaginatedInput, count),
		},
	}, nil

}

func (api *Api) AdminPermissionsCreateOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "admin-permissions-create",
		Method:      http.MethodPost,
		Path:        path,
		Summary:     "Create permission",
		Description: "Create permission",
		Tags:        []string{"Admin", "Permissions"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

type PermissionCreateInput struct {
	Name        string  `json:"name" required:"true"`
	Description *string `json:"description,omitempty"`
}

func (api *Api) AdminPermissionsCreate(ctx context.Context, input *struct {
	Body PermissionCreateInput
}) (*struct{ Body shared.Permission }, error) {
	db := api.app.Db()
	perm, err := repository.FindPermissionByName(ctx, db, input.Body.Name)
	if err != nil {
		return nil, err

	}
	if perm != nil {
		return nil, huma.Error409Conflict("Permission already exists")
	}
	data, err := repository.CreatePermission(ctx, db, &repository.CreatePermissionDto{
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
		Body: *shared.ToPermission(data),
	}, nil
}

func (api *Api) AdminPermissionsDeleteOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "admin-permissions-delete",
		Method:      http.MethodDelete,
		Path:        path,
		Summary:     "Delete permission",
		Description: "Delete permission",
		Tags:        []string{"Admin", "Permissions"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
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
	permission, err := repository.FindPermissionById(ctx, db, id)
	if err != nil {
		return nil, err
	}
	if permission == nil {
		return nil, huma.Error404NotFound("Permission not found")
	}
	if permission.Name == shared.PermissionNameAdmin {
		return nil, huma.Error400BadRequest("Cannot delete admin permission")
	}
	err = repository.DeletePermission(ctx, db, permission.ID)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (api *Api) AdminPermissionsUpdateOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "admin-permissions-update",
		Method:      http.MethodPut,
		Path:        path,
		Summary:     "Update permission",
		Description: "Update permission",
		Tags:        []string{"Admin", "Permissions"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
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
	permission, err := repository.FindPermissionById(ctx, db, id)
	if err != nil {
		return nil, err
	}
	if permission == nil {
		return nil, huma.Error404NotFound("Permission not found")
	}
	err = permission.Update(
		ctx,
		db,
		&models.PermissionSetter{
			Name:        omit.From(input.Body.Name),
			Description: omitnull.FromPtr(input.Description),
		},
	)
	if err != nil {
		return nil, err
	}
	return &struct{ Body shared.Permission }{
		Body: *shared.ToPermission(permission),
	}, nil
}

func (api *Api) AdminPermissionsGetOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "admin-permissions-get",
		Method:      http.MethodGet,
		Path:        path,
		Summary:     "Get permission",
		Description: "Get permission",
		Tags:        []string{"Admin", "Permissions"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
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
	permission, err := repository.FindPermissionById(ctx, db, id)
	if err != nil {
		return nil, err
	}
	if permission == nil {
		return nil, huma.Error404NotFound("Permission not found")
	}
	return &struct{ Body *shared.Permission }{
		Body: shared.ToPermission(permission),
	}, nil
}
