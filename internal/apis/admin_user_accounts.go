package apis

import (
	"context"

	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
)

func (api *Api) AdminUserAccounts(ctx context.Context, input *shared.UserAccountListParams) (*shared.PaginatedOutput[*shared.UserAccountOutput], error) {
	if input == nil {
		input = &shared.UserAccountListParams{}
	}

	data, err := api.app.UserAccount().Store().ListUserAccounts(ctx, input)
	if err != nil {
		return nil, err
	}
	count, err := api.app.UserAccount().Store().CountUserAccounts(ctx, &input.UserAccountListFilter)
	if err != nil {
		return nil, err
	}
	return &shared.PaginatedOutput[*shared.UserAccountOutput]{
		Body: shared.PaginatedResponse[*shared.UserAccountOutput]{
			Data: mapper.Map(data, shared.FromModelUserAccountOutput),
			Meta: shared.GenerateMeta(&input.PaginatedInput, count),
		},
	}, nil
}
