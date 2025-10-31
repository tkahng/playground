package apis

import (
	"context"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/internal/tools/types"
)

func bindApis(api huma.API, appApi *Api) {
	// Misc routes ------------------------------------
	bindMiscApi(api, appApi)
	// signup -------------------------------------------------------------
	bindAuthApi(api, appApi)
	// ---- Upload File
	bindMediaApi(api, appApi)
	// ---- Teams
	bindTeamsApi(api, appApi)
	// stats routes -------------------------------------------------------------------------------------------------
	bindStatsApi(api, appApi)
	// ---- task routes -------------------------------------------------------------------------------------------------
	bindTaskApi(api, appApi)
	// stripe routes -------------------------------------------------------------------------------------------------
	bindStripeApi(api, appApi)
	//  admin routes ----------------------------------------------------------------------------
	bindAdminApi(api, appApi)
	// admin stripe products with prices
	bindUserReactionApi(api, appApi)
}
func bindMiddlewares(api API) {
	api.Api().UseMiddleware(api.Middlewares().Recoverer)
	api.Api().UseMiddleware(api.Middlewares().Auth)
	api.Api().UseMiddleware(api.Middlewares().RequireAuth)
}

type IndexOutputBody struct {
	Access string `json:"access"`
}

type IndexOutput struct {
	Body IndexOutputBody `json:"body"`
}

func bindMiscApi(api huma.API, appApi *Api) {
	huma.Get(api, "/", func(ctx context.Context, input *struct {
		Page types.OmittableNullable[string] `query:"page" required:"false"`
	}) (*IndexOutput, error) {
		fmt.Println("input", input)
		return &IndexOutput{
			Body: IndexOutputBody{
				Access: "public",
			},
		}, nil
	})

	//  public list of permissions -----------------------------------------------------------
	huma.Register(
		api,
		huma.Operation{
			OperationID: "permissions-list",
			Method:      http.MethodGet,
			Path:        "/permissions",
			Summary:     "permissions list",
			Description: "List of permissions",
			Tags:        []string{"Permissions"},
			Errors:      []int{http.StatusNotFound},
		},
		appApi.PermissionsList,
	)
	// protected test routes -----------------------------------------------------------
	huma.Register(
		api,
		huma.Operation{
			OperationID: "api-protected",
			Method:      http.MethodGet,
			Path:        "/protected/{permission-name}",
			Summary:     "Api protected",
			Description: "Api protected",
			Tags:        []string{"Protected"},
			Errors:      []int{http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
		},
		appApi.ApiProtected,
	)
}
