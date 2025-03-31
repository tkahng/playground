package apis

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
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
