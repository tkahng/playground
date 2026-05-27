package apis

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/playground/internal/middleware"
	"github.com/tkahng/playground/internal/middleware/humamiddleware"
	"github.com/tkahng/playground/internal/shared"
)

func bindTaskCommentApi(appApi *Api) {
	api := appApi.Api()
	app := appApi.App()

	huma.Register(
		api,
		huma.Operation{
			OperationID: "task-comment-list",
			Method:      http.MethodGet,
			Path:        "/tasks/{task-id}/comments",
			Summary:     "List task comments",
			Tags:        []string{"Task"},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.RequireTeamInfo(),
			),
		},
		appApi.TaskCommentList,
	)

	huma.Register(
		api,
		huma.Operation{
			OperationID: "task-comment-create",
			Method:      http.MethodPost,
			Path:        "/tasks/{task-id}/comments",
			Summary:     "Create task comment",
			Tags:        []string{"Task"},
			Errors:      []int{http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.RequireTeamInfo(),
				middleware.RequireTeamPermission(app, shared.TeamPermissionTasksEdit),
			),
		},
		appApi.TaskCommentCreate,
	)

	huma.Register(
		api,
		huma.Operation{
			OperationID: "task-comment-update",
			Method:      http.MethodPut,
			Path:        "/tasks/{task-id}/comments/{comment-id}",
			Summary:     "Update task comment",
			Tags:        []string{"Task"},
			Errors:      []int{http.StatusNotFound, http.StatusForbidden},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.RequireTeamInfo(),
				middleware.RequireTeamPermission(app, shared.TeamPermissionTasksEdit),
			),
		},
		appApi.TaskCommentUpdate,
	)

	huma.Register(
		api,
		huma.Operation{
			OperationID: "task-comment-delete",
			Method:      http.MethodDelete,
			Path:        "/tasks/{task-id}/comments/{comment-id}",
			Summary:     "Delete task comment",
			Tags:        []string{"Task"},
			Errors:      []int{http.StatusNotFound, http.StatusForbidden},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.RequireTeamInfo(),
				middleware.RequireTeamPermission(app, shared.TeamPermissionTasksEdit),
			),
		},
		appApi.TaskCommentDelete,
	)
}
