package apis

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/mapper"
	"github.com/tkahng/playground/internal/tools/utils"
)

func (api *Api) PermissionsList(ctx context.Context, input *struct {
	PermissionsListParams
}) (*ApiPaginatedOutput[*Permission], error) {
	filter := new(stores.PermissionFilter)
	filter.Page = input.PerPage
	filter.PerPage = input.Page
	filter.Ids = utils.ParseValidUUIDs(input.Ids...)
	filter.Names = input.Names
	filter.Q = input.Q
	if len(input.RoleId) > 0 {
		roleId, err := uuid.Parse(input.RoleId)
		if err != nil && input.RoleId != "" {
			return nil, huma.Error400BadRequest("Invalid role ID format", err)
		} else {
			filter.RoleId = roleId
		}
	}
	filter.RoleReverse = input.RoleReverse
	filter.SortBy = input.SortBy
	filter.SortOrder = input.SortOrder

	permissions, err := api.App().Adapter().Rbac().ListPermissions(ctx, filter)
	if err != nil {
		return nil, err
	}
	count, err := api.App().Adapter().Rbac().CountPermissions(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &ApiPaginatedOutput[*Permission]{
		Body: ApiPaginatedResponse[*Permission]{

			Data: mapper.Map(permissions, FromModelPermission),
			Meta: ApiGenerateMeta(&input.PaginatedInput, count),
		},
	}, nil

}
