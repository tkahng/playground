package apis

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/shared"
)

func (api *Api) AdminPermissionsOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "admin-permissions",
		Method:      http.MethodGet,
		Path:        path,
		Summary:     "Admin permissions",
		Description: "List of permissions",
		Tags:        []string{"Auth", "Admin"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (api *Api) AdminPermissions(ctx context.Context, input *struct {
	shared.PermissionsListParams
}) (*PaginatedOutput[*models.Permission], error) {
	db := api.app.Db()
	permissions, err := repository.ListPermissions(ctx, db, &input.PermissionsListParams)
	if err != nil {
		return nil, err
	}
	count, err := repository.CountPermissions(ctx, db, &input.PermissionsListFilter)
	if err != nil {
		return nil, err
	}

	return &PaginatedOutput[*models.Permission]{
		Body: shared.PaginatedResponse[*models.Permission]{

			Data: permissions,
			Meta: shared.Meta{
				Page:    input.PaginatedInput.Page,
				PerPage: input.PaginatedInput.PerPage,
				Total:   int(count),
			},
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
		Tags:        []string{"Auth", "Admin"},
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
}) (*struct{ Body models.Permission }, error) {
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
	return &struct{ Body models.Permission }{
		Body: *data,
	}, nil
}

func (api *Api) AdminPermissionsDeleteOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "admin-permissions-delete",
		Method:      http.MethodDelete,
		Path:        path,
		Summary:     "Delete permission",
		Description: "Delete permission",
		Tags:        []string{"Auth", "Admin"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (api *Api) AdminPermissionsDelete(ctx context.Context, input *struct {
	ID string `path:"id"`
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
	err = repository.DeletePermission(ctx, db, permission.ID)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
