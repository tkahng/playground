package apis

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
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
