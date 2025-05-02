package apis

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
)

func (api *Api) PermissionsListOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "permissions-list",
		Method:      http.MethodGet,
		Path:        path,
		Summary:     "permissions list",
		Description: "List of permissions",
		Tags:        []string{"Permissions"},
		Errors:      []int{http.StatusNotFound},
	}
}

func (api *Api) PermissionsList(ctx context.Context, input *struct {
	shared.PermissionsListParams
}) (*shared.PaginatedOutput[*shared.Permission], error) {
	db := api.app.Db()
	permissions, err := queries.ListPermissions(ctx, db, &input.PermissionsListParams)
	if err != nil {
		return nil, err
	}
	count, err := queries.CountPermissions(ctx, db, &input.PermissionsListFilter)
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
