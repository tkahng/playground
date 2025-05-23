package apis

import (
	"context"

	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
)

func (api *Api) PermissionsList(ctx context.Context, input *struct {
	shared.PermissionsListParams
}) (*shared.PaginatedOutput[*shared.Permission], error) {
	permissions, err := api.app.Rbac().Store().ListPermissions(ctx, &input.PermissionsListParams)
	if err != nil {
		return nil, err
	}
	count, err := api.app.Rbac().Store().CountPermissions(ctx, &input.PermissionsListFilter)
	if err != nil {
		return nil, err
	}

	return &shared.PaginatedOutput[*shared.Permission]{
		Body: shared.PaginatedResponse[*shared.Permission]{

			Data: mapper.Map(permissions, shared.FromModelPermission),
			Meta: shared.GenerateMeta(&input.PaginatedInput, count),
		},
	}, nil

}
