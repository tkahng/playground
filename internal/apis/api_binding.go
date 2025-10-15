package apis

import (
	"context"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/middleware"
	"github.com/tkahng/playground/internal/middleware/humamiddleware"
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/internal/tools/types"
)

func (appApi *Api) RegisterRoutes() {
	bindMiddlewares(appApi.Api(), appApi.App())
	bindApis(appApi.Api(), appApi)
}

func bindMiddlewares(api huma.API, app core.App) {
	api.UseMiddleware(humamiddleware.HumaChiMiddleware(middleware.RecovererMiddleware(app)))
	api.UseMiddleware(humamiddleware.HumaAuthMiddleware(api, app))
	api.UseMiddleware(humamiddleware.HumaRequireAuthMiddleware(api, app))
}

type IndexOutputBody struct {
	Access string `json:"access"`
}

type IndexOutput struct {
	Body IndexOutputBody `json:"body"`
}

func bindApis(api huma.API, appApi *Api) {
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

	// signup -------------------------------------------------------------
	BindAuthApi(api, appApi)

	// ---- Upload File
	BindMediaApi(api, appApi)

	// ---- Teams
	BindTeamsApi(api, appApi)

	// stats routes -------------------------------------------------------------------------------------------------
	BindStatsApi(api, appApi)

	// ---- task routes -------------------------------------------------------------------------------------------------
	BindTaskApi(api, appApi)

	// stripe routes -------------------------------------------------------------------------------------------------

	BindStripeApi(api, appApi)

	//  admin routes ----------------------------------------------------------------------------
	BindAdminApi(api, appApi)
	// admin stripe products with prices
	BindUserReactionApi(api, appApi)
}
