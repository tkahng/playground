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

func (api *Api) AdminRolesOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "admin-roles",
		Method:      http.MethodGet,
		Path:        path,
		Summary:     "Admin roles",
		Description: "List of roles",
		Tags:        []string{"Auth", "Admin"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (api *Api) AdminRoles(ctx context.Context, input *struct {
	shared.RolesListParams
}) (*PaginatedOutput[*models.Role], error) {
	db := api.app.Db()
	roles, err := repository.ListRoles(ctx, db, &input.RolesListParams)
	if err != nil {
		return nil, err
	}
	count, err := repository.CountRoles(ctx, db, &input.RoleListFilter)
	if err != nil {
		return nil, err
	}

	return &PaginatedOutput[*models.Role]{
		Body: shared.PaginatedResponse[*models.Role]{

			Data: roles,
			Meta: shared.Meta{
				Page:    input.PaginatedInput.Page,
				PerPage: input.PaginatedInput.PerPage,
				Total:   int(count),
			},
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
		Tags:        []string{"Auth", "Admin"},
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
	Body models.Role
}, error) {
	db := api.app.Db()
	data, err := repository.FindRoleByName(ctx, db, input.Body.Name)
	if err != nil {
		return nil, err
	}
	if data != nil {
		return nil, huma.Error409Conflict("Role already exists")
	}
	role, err := repository.CreateRole(ctx, db, &repository.CreateRoleDto{
		Name:        input.Body.Name,
		Description: input.Body.Description,
	})
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, huma.Error500InternalServerError("Failed to create role")
	}
	return &struct{ Body models.Role }{
		Body: *role,
	}, nil
}

func (api *Api) AdminRolesDeleteOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "admin-roles-delete",
		Method:      http.MethodDelete,
		Path:        path,
		Summary:     "Delete role",
		Description: "Delete role",
		Tags:        []string{"Auth", "Admin"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (api *Api) AdminRolesDelete(ctx context.Context, input *struct {
	RoleID string `path:"id"`
}) (*struct{}, error) {
	db := api.app.Db()
	id, err := uuid.Parse(input.RoleID)
	if err != nil {
		return nil, err
	}
	role, err := repository.FindRoleById(ctx, db, id)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, huma.Error404NotFound("Role not found")
	}
	err = repository.DeleteRole(ctx, db, role.ID)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
