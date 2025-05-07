package apis

import (
	"context"

	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
)

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

			Data: mapper.Map(permissions, shared.FromCrudPermission),
			Meta: shared.GenerateMeta(input.PaginatedInput, count),
		},
	}, nil

}
