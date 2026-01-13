package apis

import (
	"context"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/playground/internal/middleware"
	"github.com/tkahng/playground/internal/middleware/humamiddleware"
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/internal/tools/types"
)

func bindApis(appApi *Api) {
	// Misc routes ------------------------------------
	bindMiscApi(appApi)
	// signup -------------------------------------------------------------
	bindAuthApi(appApi)
	// ---- Upload File
	bindMediaApi(appApi)
	// ---- Teams
	bindTeamsApi(appApi)
	// stats routes -------------------------------------------------------------------------------------------------
	bindStatsApi(appApi)
	// ---- task routes -------------------------------------------------------------------------------------------------
	bindTaskApi(appApi)
	// stripe routes -------------------------------------------------------------------------------------------------
	bindStripeApi(appApi)
	//  admin routes ----------------------------------------------------------------------------
	bindAdminApi(appApi)
	// admin stripe products with prices
	bindUserReactionApi(appApi)
	// bind game api
	bindGameApi(appApi)
}
func bindMiddlewares(api API) {
	api.Api().UseMiddleware(humamiddleware.HumaOperationSecurityMiddleware())
	api.Api().UseMiddleware(
		humamiddleware.HumaChiMiddlewares(
			middleware.RecovererMiddleware(),
			middleware.AuthMiddleware(api.App()),
			middleware.RequireAuthMiddleware(),
		)...,
	)
}

type IndexOutputBody struct {
	Access string `json:"access"`
}

type IndexOutput struct {
	Body IndexOutputBody `json:"body"`
}

func bindMiscApi(appApi *Api) {
	api := appApi.Api()
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
