package apis

import (
	"context"

	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
)

func (api *Api) AdminUserAccounts(ctx context.Context, input *shared.UserAccountListParams) (*shared.PaginatedOutput[*shared.UserAccountOutput], error) {
	if input == nil {
		input = &shared.UserAccountListParams{}
	}
	db := api.app.Db()

	data, err := queries.ListUserAccounts(ctx, db, input)
	if err != nil {
		return nil, err
	}
	count, err := queries.CountUserAccounts(ctx, db, &input.UserAccountListFilter)
	if err != nil {
		return nil, err
	}
	return &shared.PaginatedOutput[*shared.UserAccountOutput]{
		Body: shared.PaginatedResponse[*shared.UserAccountOutput]{
			Data: mapper.Map(data, shared.FromCrudUserAccountOutput),
			Meta: shared.GenerateMeta(input.PaginatedInput, count),
		},
	}, nil
}
