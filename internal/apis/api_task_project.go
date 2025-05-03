package apis

import (
	"context"
	"net/http"
	"slices"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/core"
	"github.com/tkahng/authgo/internal/db"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/ai/googleai"
	"github.com/tkahng/authgo/internal/tools/mapper"
)

func (api *Api) TaskProjectListOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "task-project-list",
		Method:      http.MethodGet,
		Path:        path,
		Summary:     "Task project list",
		Description: "List of task projects",
		Tags:        []string{"Task"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

type TaskProjectListResponse struct {
	Body *shared.PaginatedResponse[*shared.TaskProjectWithTasks]
}

func (api *Api) TaskProjectList(ctx context.Context, input *shared.TaskProjectsListParams) (*TaskProjectListResponse, error) {
	db := api.app.Db()
	userInfo := core.GetContextUserInfo(ctx)
	if userInfo == nil {
		return nil, huma.Error401Unauthorized("Unauthorized")
	}
	input.TaskProjectsListFilter.UserID = userInfo.User.ID.String()
	taskProject, err := queries.ListTaskProjects(ctx, db, input)
	if err != nil {
		return nil, err
	}
	total, err := queries.CountTaskProjects(ctx, db, &input.TaskProjectsListFilter)
	if err != nil {
		return nil, err
	}
	taskProjectIds := mapper.Map(taskProject, func(taskProject *models.TaskProject) uuid.UUID {
		return taskProject.ID
	})

	if input.Expand != nil && slices.Contains(input.Expand, "tasks") {
		tasks, err := queries.LoadTaskProjectsTasks(ctx, db, taskProjectIds...)
		if err != nil {
			return nil, err
		}
		for idx, taskProject := range taskProject {
			taskProject.Tasks = tasks[idx]
		}
	}
	return &TaskProjectListResponse{
		Body: &shared.PaginatedResponse[*shared.TaskProjectWithTasks]{
			Data: mapper.Map(taskProject, func(taskProject *models.TaskProject) *shared.TaskProjectWithTasks {
				return &shared.TaskProjectWithTasks{
					TaskProject: shared.CrudToProject(taskProject),
					Tasks: mapper.Map(taskProject.Tasks, func(task *models.Task) *shared.TaskWithSubtask {
						return &shared.TaskWithSubtask{
							Task: shared.CrudModelToTask(task),
							Children: mapper.Map(task.Children, func(child *models.Task) *shared.Task {
								return shared.CrudModelToTask(child)
							}),
						}
					}),
				}
			}),
			Meta: shared.GenerateMeta(input.PaginatedInput, total),
		},
	}, nil
}

func (api *Api) TaskProjectCreateOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "task-project-create",
		Method:      http.MethodPost,
		Path:        path,
		Summary:     "Task project create",
		Description: "Create a new task project",
		Tags:        []string{"Task"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (api *Api) TaskProjectCreate(ctx context.Context, input *struct {
	Body shared.CreateTaskProjectWithTasksDTO
}) (*struct {
	Body *shared.TaskProject
}, error) {
	userInfo := core.GetContextUserInfo(ctx)
	if userInfo == nil {
		return nil, huma.Error401Unauthorized("Unauthorized")
	}
	db := api.app.Db()
	taskProject, err := createTaskProjectWithTasks(ctx, db, userInfo.User.ID, input.Body)
	if err != nil {
		return nil, err
	}
	return &struct {
		Body *shared.TaskProject
	}{
		Body: shared.CrudToProject(taskProject),
	}, nil
}

func createTaskProjectWithTasks(ctx context.Context, db *db.Queries, userId uuid.UUID, input shared.CreateTaskProjectWithTasksDTO) (*models.TaskProject, error) {
	taskProject, err := queries.CreateTaskProjectWithTasks(ctx, db, userId, &input)
	if err != nil {
		return nil, err
	}
	return taskProject, nil
}

func (api *Api) TaskProjectCreateWithAiOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "task-project-create-with-ai",
		Method:      http.MethodPost,
		Path:        path,
		Summary:     "Task project create with ai",
		Description: "Create a new task project with ai",
		Tags:        []string{"Task"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

type TaskProjectCreateWithAiDto struct {
	Input string `json:"input"`
}
type TaskProjectCreateWithAiInput struct {
	Body TaskProjectCreateWithAiDto `json:"body"`
}

func (api *Api) TaskProjectCreateWithAi(ctx context.Context, input *TaskProjectCreateWithAiInput) (*struct {
	Body *shared.TaskProject
}, error) {
	userInfo := core.GetContextUserInfo(ctx)
	if userInfo == nil {
		return nil, huma.Error401Unauthorized("Unauthorized")
	}
	db := api.app.Db()
	aiService := googleai.NewAiService(ctx, api.app.Cfg().AiConfig)
	taskProjectPlan, err := aiService.GenerateProjectPlan(ctx, input.Body.Input)
	if err != nil {
		return nil, err
	}
	args := shared.CreateTaskProjectWithTasksDTO{
		CreateTaskProjectDTO: shared.CreateTaskProjectDTO{
			Name:        taskProjectPlan.Project.Name,
			Description: &taskProjectPlan.Project.Description,
			Status:      shared.TaskProjectStatusTodo,
		},
		Tasks: mapper.Map(taskProjectPlan.Tasks, func(task googleai.Task) shared.CreateTaskBaseDTO {
			return shared.CreateTaskBaseDTO{
				Name:        task.Name,
				Description: &task.Description,
				Status:      shared.TaskStatusTodo,
			}
		}),
	}
	taskProject, err := createTaskProjectWithTasks(ctx, db, userInfo.User.ID, args)
	if err != nil {
		return nil, err
	}
	return &struct {
		Body *shared.TaskProject
	}{
		Body: shared.CrudToProject(taskProject),
	}, nil
}

func (api *Api) TaskProjectUpdateOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "task-project-update",
		Method:      http.MethodPut,
		Path:        path,
		Summary:     "Task project update",
		Description: "Update a task project",
		Tags:        []string{"Task"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

type TaskProjectResponse struct {
	Body *shared.TaskProjectWithTasks
}

func (api *Api) TaskProjectUpdate(ctx context.Context, input *shared.UpdateTaskProjectDTO) (*struct{}, error) {
	userInfo := core.GetContextUserInfo(ctx)
	if userInfo == nil {
		return nil, huma.Error401Unauthorized("Unauthorized")
	}
	db := api.app.Db()
	id, err := uuid.Parse(input.TaskProjectID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid task project id")
	}
	payload := input.Body
	err = queries.UpdateTaskProject(ctx, db, id, &payload)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (api *Api) TaskProjectDeleteOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "task-project-delete",
		Method:      http.MethodDelete,
		Path:        path,
		Summary:     "Task project delete",
		Description: "Delete a task project",
		Tags:        []string{"Task"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (api *Api) TaskProjectDelete(ctx context.Context, input *struct {
	TaskProjectID string `path:"task-project-id"`
}) (*struct{}, error) {
	userInfo := core.GetContextUserInfo(ctx)
	if userInfo == nil {
		return nil, huma.Error401Unauthorized("Unauthorized")
	}
	db := api.app.Db()
	id, err := uuid.Parse(input.TaskProjectID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid task project id")
	}
	err = queries.DeleteTaskProject(ctx, db, id)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (api *Api) TaskProjectGetOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "task-project-get",
		Method:      http.MethodGet,
		Path:        path,
		Summary:     "Task project get",
		Description: "Get a task project",
		Tags:        []string{"Task"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (api *Api) TaskProjectGet(ctx context.Context, input *struct {
	TaskProjectID string   `path:"task-project-id" json:"task_project_id" required:"true" format:"uuid"`
	Expand        []string `query:"expand,omitempty" required:"false" minimum:"1" maximum:"100" enum:"tasks"`
}) (*TaskProjectResponse, error) {
	userInfo := core.GetContextUserInfo(ctx)
	if userInfo == nil {
		return nil, huma.Error401Unauthorized("Unauthorized")
	}
	db := api.app.Db()
	id, err := uuid.Parse(input.TaskProjectID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid task project id")
	}
	taskProject, err := queries.FindTaskProjectByID(ctx, db, id)
	if err != nil {
		return nil, err
	}
	if input.Expand != nil && slices.Contains(input.Expand, "tasks") {
		tasks, err := queries.LoadTaskProjectsTasks(ctx, db, taskProject.ID)
		if err != nil {
			return nil, err
		}
		if len(tasks) > 0 {
			taskProject.Tasks = tasks[0]
		}
	}
	return &TaskProjectResponse{
		Body: &shared.TaskProjectWithTasks{
			TaskProject: shared.CrudToProject(taskProject),
			Tasks: mapper.Map(taskProject.Tasks, func(task *models.Task) *shared.TaskWithSubtask {
				return &shared.TaskWithSubtask{
					Task: shared.CrudModelToTask(task),
					Children: mapper.Map(task.Children, func(child *models.Task) *shared.Task {
						return shared.CrudModelToTask(child)
					}),
				}
			}),
		},
	}, nil
}

func (api *Api) TaskProjectTasksCreateOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "task-project-tasks-create",
		Method:      http.MethodPost,
		Path:        path,
		Summary:     "Task project tasks create",
		Description: "Create a new task project task",
		Tags:        []string{"Task"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (api *Api) TaskProjectTasksCreate(ctx context.Context, input *shared.CreateTaskWithProjectIdInput) (*TaskResposne, error) {
	userInfo := core.GetContextUserInfo(ctx)
	if userInfo == nil {
		return nil, huma.Error401Unauthorized("Unauthorized")
	}
	db := api.app.Db()
	id, err := uuid.Parse(input.TaskProjectID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid task project id")
	}
	payload := input.Body
	order, err := queries.FindLastTaskOrder(ctx, db, id)
	if err != nil {
		return nil, err
	}
	payload.Order = order
	task, err := queries.CreateTaskWithChildren(ctx, db, userInfo.User.ID, id, &payload)
	if err != nil {
		return nil, err
	}
	return &TaskResposne{
		Body: &shared.TaskWithSubtask{
			Task: shared.CrudModelToTask(task),
			Children: mapper.Map(task.Children, func(child *models.Task) *shared.Task {
				return shared.CrudModelToTask(child)
			}),
		},
	}, nil
}
