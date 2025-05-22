package apis

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/shared"
)

func BindTaskApi(api huma.API, appApi *Api) {
	checkTaskOwnerMiddleware := CheckTaskOwnerMiddleware(api, appApi.app)
	taskGroup := huma.NewGroup(api)
	taskGroup.UseMiddleware(checkTaskOwnerMiddleware)
	// task list
	huma.Register(
		taskGroup,
		huma.Operation{
			OperationID: "task-list",
			Method:      http.MethodGet,
			Path:        "/tasks",
			Summary:     "Task list",
			Description: "List of tasks",
			Tags:        []string{"Task"},
			Errors:      []int{http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
		},
		appApi.TaskList,
	)
	// task create
	// huma.Register(taskGroup, appApi.TaskCreateOperation("/task"), appApi.TaskCreate)
	// task update
	huma.Register(
		taskGroup,
		huma.Operation{
			OperationID: "task-update",
			Method:      http.MethodPut,
			Path:        "/tasks/{task-id}",
			Summary:     "Task update",
			Description: "Update a task",
			Tags:        []string{"Task"},
			Errors:      []int{http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
		},
		appApi.TaskUpdate,
	)
	// task position
	// huma.Register(taskGroup, appApi.UpdateTaskPositionOperation("/tasks/{task-id}/position"), appApi.UpdateTaskPosition)
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
		},
		appApi.TaskGet,
	)

	// task project routes -------------------------------------------------------------------------------------------------
	taskProjectGroup := huma.NewGroup(api)
	// task project list
	huma.Register(
		taskProjectGroup,
		huma.Operation{
			OperationID: "task-project-list",
			Method:      http.MethodGet,
			Path:        "/task-projects",
			Summary:     "Task project list",
			Description: "List of task projects",
			Tags:        []string{"Task"},
			Errors:      []int{http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
		},
		appApi.TaskProjectList,
	)
	// task project create
	huma.Register(
		taskProjectGroup,
		huma.Operation{
			OperationID: "task-project-create",
			Method:      http.MethodPost,
			Path:        "/task-projects",
			Summary:     "Task project create",
			Description: "Create a new task project",
			Tags:        []string{"Task"},
			Errors:      []int{http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
		},
		appApi.TaskProjectCreate,
	)
	// task project create with ai
	huma.Register(
		taskProjectGroup,
		huma.Operation{
			OperationID: "task-project-create-with-ai",
			Method:      http.MethodPost,
			Path:        "/task-projects/ai",
			Summary:     "Task project create with ai",
			Description: "Create a new task project with ai",
			Tags:        []string{"Task"},
			Errors:      []int{http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
		},
		appApi.TaskProjectCreateWithAi,
	)
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
		},
		appApi.TaskProjectUpdate,
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
		},
		appApi.TaskProjectDelete,
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
		},
		appApi.TaskProjectGet,
	)
	// task project tasks create
	huma.Register(
		taskProjectGroup,
		huma.Operation{
			OperationID: "task-project-tasks-create",
			Method:      http.MethodPost,
			Path:        "/task-projects/{task-project-id}/tasks",
			Summary:     "Task project tasks create",
			Description: "Create a new task project task",
			Tags:        []string{"Task"},
			Errors:      []int{http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
		},
		appApi.TaskProjectTasksCreate,
	)
}
