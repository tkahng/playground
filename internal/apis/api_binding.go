package apis

import (
	"context"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/core"
	"github.com/tkahng/authgo/internal/middleware"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/types"
)

func BindMiddlewares(api huma.API, app core.App) {
	api.UseMiddleware(middleware.AuthMiddleware(api, app))
	api.UseMiddleware(middleware.RequireAuthMiddleware(api))
}

type IndexOutputBody struct {
	Access string `json:"access"`
}

type IndexOutput struct {
	Body IndexOutputBody `json:"body"`
}

func BindApis(api huma.API, appApi *Api) {
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

	// ---- notifications
	// sse.Register(
	// 	api,
	// 	huma.Operation{
	// 		OperationID: "notifications-sse",
	// 		Method:      http.MethodGet,
	// 		Path:        "/notifications/sse",
	// 		Summary:     "Notifications SSE",
	// 		Description: "Notifications SSE",
	// 		Tags:        []string{"Notifications"},
	// 		Errors:      []int{http.StatusNotFound},
	// 		Security: []map[string][]string{{
	// 			shared.BearerAuthSecurityKey: {},
	// 		}},
	// 	}, map[string]any{
	// 		// Mapping of event type name to Go struct for that event.
	// 		"message": models.Notification{},
	// 	},
	// 	appApi.NotificationsSsefunc)
	// stats routes -------------------------------------------------------------------------------------------------
	BindStatsApi(api, appApi)

	// ---- task routes -------------------------------------------------------------------------------------------------
	BindTaskApi(api, appApi)

	// stripe routes -------------------------------------------------------------------------------------------------

	BindStripeApi(api, appApi)

	//  admin routes ----------------------------------------------------------------------------
	BindAdminApi(api, appApi)
	// admin stripe products with prices

}

func AddRoutes(api huma.API, appApi *Api) {
	BindMiddlewares(api, appApi.App())
	BindApis(api, appApi)
}
