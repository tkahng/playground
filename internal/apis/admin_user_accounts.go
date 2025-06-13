package apis

import (
	"context"
	"math"

	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/stores"
	"github.com/tkahng/authgo/internal/tools/mapper"
	"github.com/tkahng/authgo/internal/tools/utils"
)

type ApiProviderTypes string

const (
	ProviderTypeOAuth       ApiProviderTypes = "oauth"
	ProviderTypeCredentials ApiProviderTypes = "credentials"
)

func (p ApiProviderTypes) String() string {
	return string(p)
}

type ApiProviders string

const (
	ProvidersGoogle      ApiProviders = "google"
	ProvidersApple       ApiProviders = "apple"
	ProvidersFacebook    ApiProviders = "facebook"
	ProvidersGithub      ApiProviders = "github"
	ProvidersCredentials ApiProviders = "credentials"
)

func (p ApiProviders) String() string {
	return string(p)
}

type UserAccountFilter struct {
	PaginatedInput
	SortParams
	Providers     []ApiProviders     `query:"providers,omitempty" required:"false" uniqueItems:"true" minimum:"1" maximum:"100" enum:"google,apple,facebook,github,credentials"`
	ProviderTypes []ApiProviderTypes `query:"provider_types,omitempty" required:"false" uniqueItems:"true" minimum:"1" maximum:"100" enum:"oauth,credentials"`
	Q             string             `query:"q,omitempty" required:"false"`
	Ids           []string           `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
	UserIds       []string           `query:"user_ids,omitempty" minimum:"1" maximum:"100" required:"false" format:"uuid"`
}

func GenerateMeta(input *PaginatedInput, total int64) Meta {
	var meta Meta = Meta{
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

func (api *Api) AdminUserAccounts(ctx context.Context, input *UserAccountFilter) (*ApiPaginatedOutput[*shared.UserAccountOutput], error) {
	filter := &stores.UserAccountFilter{}
	filter.Page = input.Page
	filter.PerPage = input.PerPage
	filter.SortBy, filter.SortOrder = input.Sort()
	filter.Q = input.Q
	filter.Ids = utils.ParseValidUUIDs(input.Ids...)
	filter.UserIds = utils.ParseValidUUIDs(input.UserIds...)

	data, err := api.app.Adapter().UserAccount().ListUserAccounts(ctx, filter)
	if err != nil {
		return nil, err
	}
	count, err := api.app.Adapter().UserAccount().CountUserAccounts(ctx, filter)
	if err != nil {
		return nil, err
	}
	return &ApiPaginatedOutput[*shared.UserAccountOutput]{
		Body: ApiPaginatedResponse[*shared.UserAccountOutput]{
			Data: mapper.Map(data, shared.FromModelUserAccountOutput),
			Meta: ApiGenerateMeta(&input.PaginatedInput, count),
		},
	}, nil
}
