package apis

import (
	"context"
	"math"

	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/mapper"
	"github.com/tkahng/playground/internal/tools/utils"
)

type UserAccountFilter struct {
	PaginatedInput
	SortParams
	Providers     []models.Providers     `query:"providers,omitempty" required:"false" uniqueItems:"true" minimum:"1" maximum:"100" enum:"google,apple,facebook,github,credentials"`
	ProviderTypes []models.ProviderTypes `query:"provider_types,omitempty" required:"false" uniqueItems:"true" minimum:"1" maximum:"100" enum:"oauth,credentials"`
	Q             string                 `query:"q,omitempty" required:"false"`
	Ids           []string               `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	UserIds       []string               `query:"user_ids,omitempty" minimum:"1" maximum:"100" required:"false" format:"uuid"`
}

func GenerateMeta(input *shared.PaginatedInput, total int64) Meta {
	var meta = Meta{
		Page:    input.Page,
		PerPage: input.PerPage,
		Total:   total,
	}
	nextPage, prevPage := input.Page+1, input.Page-1

	perPage := input.PerPage
	if perPage == 0 {
		perPage = 10
	}
	pageCount := int64(math.Ceil(float64(total) / float64(perPage)))

	if prevPage >= 0 {
		meta.PrevPage = &prevPage
	} else {
		meta.PrevPage = nil
	}
	if nextPage < pageCount-1 {
		meta.HasMore = true
		meta.NextPage = &nextPage
	} else {
		meta.NextPage = nil
	}
	return meta
}

func (api *Api) AdminUserAccounts(ctx context.Context, input *UserAccountFilter) (*ApiPaginatedOutput[*UserAccountOutput], error) {
	filter := &stores.UserAccountFilter{}
	filter.Page = input.Page
	filter.PerPage = input.PerPage
	filter.SortBy, filter.SortOrder = input.Sort()
	filter.Q = input.Q
	filter.Ids = utils.ParseValidUUIDs(input.Ids...)
	filter.UserIds = utils.ParseValidUUIDs(input.UserIds...)
	filter.Providers = input.Providers
	filter.ProviderTypes = input.ProviderTypes

	data, err := api.App().Adapter().UserAccount().ListUserAccounts(ctx, filter)
	if err != nil {
		return nil, err
	}
	count, err := api.App().Adapter().UserAccount().CountUserAccounts(ctx, filter)
	if err != nil {
		return nil, err
	}
	return &ApiPaginatedOutput[*UserAccountOutput]{
		Body: ApiPaginatedResponse[*UserAccountOutput]{
			Data: mapper.Map(data, FromModelUserAccountOutput),
			Meta: ApiGenerateMeta(&input.PaginatedInput, count),
		},
	}, nil
}
