package apis

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/playground/internal/middleware"
	"github.com/tkahng/playground/internal/middleware/humamiddleware"
	"github.com/tkahng/playground/internal/shared"
)

func bindTaskApi(appApi *Api) {
	api := appApi.Api()
	app := appApi.App()
	// checkTaskOwnerMiddleware := middleware.CheckTaskOwnerMiddleware(api, appApi.App())

	taskGroup := huma.NewGroup(api)
	taskGroup.UseMiddleware(
		humamiddleware.HumaChiMiddlewares(
			middleware.TeamFromParam(app),
			middleware.TaskFromParam(app),
			middleware.TaskProjectFromParam(app),
			middleware.TeamInfoFromContext(app),
		)...,
	)
	// taskGroup.UseMiddleware(checkTaskOwnerMiddleware)
	// task list
	appApi.TeamTaskListBind(taskGroup)
	// task create
	// task update
	appApi.TeamTaskUpdateBind(taskGroup)
	// task position
	// task position status
	huma.Register(
		taskGroup,
		huma.Operation{
			OperationID: "update-task-position-status",
			Method:      http.MethodPut,
			Path:        "/tasks/{task-id}/position-status",
			Summary:     "Update task position and status",
			Description: "Update task position and status",
			Tags:        []string{"Task"},
			Errors:      []int{http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.RequireTeamInfo(),
			),
		},
		appApi.UpdateTaskPositionStatus,
	)
	// // task delete
	huma.Register(
		taskGroup,
		huma.Operation{
			OperationID: "task-delete",
			Method:      http.MethodDelete,
			Path:        "/tasks/{task-id}",
			Summary:     "Task delete",
			Description: "Delete a task",
			Tags:        []string{"Task"},
			Errors:      []int{http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.RequireTeamInfo(),
			),
		},
		appApi.TaskDelete,
	)
	// // task get
	huma.Register(
		taskGroup,
		huma.Operation{
			OperationID: "task-get",
			Method:      http.MethodGet,
			Path:        "/tasks/{task-id}",
			Summary:     "Task get",
			Description: "Get a task",
			Tags:        []string{"Task"},
			Errors:      []int{http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.RequireTeamInfo(),
			),
		},
		appApi.TaskGet,
	)

	// task project routes -------------------------------------------------------------------------------------------------
	taskProjectGroup := huma.NewGroup(api)
	taskProjectGroup.UseMiddleware(
		humamiddleware.HumaChiMiddlewares(
			middleware.TeamFromParam(app),
			middleware.TaskProjectFromParam(app),
			middleware.TeamInfoFromContext(app),
		)...,
	)
	// task project list
	appApi.TeamTaskProjectListBind(taskProjectGroup)
	// task project create
	appApi.TeamTaskProjectCreateBind(taskProjectGroup)
	// task project create with ai
	appApi.TeamTaskProjectCreateWithAiBind(taskProjectGroup)
	// task project update
	huma.Register(
		taskProjectGroup,
		huma.Operation{
			OperationID: "task-project-update",
			Method:      http.MethodPut,
			Path:        "/task-projects/{task-project-id}",
			Summary:     "Task project update",
			Description: "Update a task project",
			Tags:        []string{"Task"},
			Errors:      []int{http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.RequireTeamInfo(),
			),
		},
		appApi.TeamTaskProjectUpdate,
	)
	// // task project delete
	huma.Register(
		taskProjectGroup,
		huma.Operation{
			OperationID: "task-project-delete",
			Method:      http.MethodDelete,
			Path:        "/task-projects/{task-project-id}",
			Summary:     "Task project delete",
			Description: "Delete a task project",
			Tags:        []string{"Task"},
			Errors:      []int{http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.RequireTeamInfo(),
			),
		},
		appApi.TeamTaskProjectDelete,
	)
	// // task project get
	huma.Register(
		taskProjectGroup,
		huma.Operation{
			OperationID: "task-project-get",
			Method:      http.MethodGet,
			Path:        "/task-projects/{task-project-id}",
			Summary:     "Task project get",
			Description: "Get a task project",
			Tags:        []string{"Task"},
			Errors:      []int{http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.RequireTeamInfo(),
			),
		},
		appApi.TeamTaskProjectGet,
	)
	// task project tasks create
	huma.Register(
		taskProjectGroup,
		huma.Operation{
			OperationID: "task-project-tasks-create",
			Method:      http.MethodPost,
			Path:        "/task-projects/{task-project-id}",
			Summary:     "Task project tasks create",
			Description: "Create a new task project task",
			Tags:        []string{"Task"},
			Errors:      []int{http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.RequireTeamInfo(),
			),
		},
		appApi.TeamTaskProjectTasksCreate,
	)
}
