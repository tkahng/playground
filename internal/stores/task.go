package stores

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/tkahng/playground/internal/apierrors"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/database/repository"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/tools/mapper"
	"github.com/tkahng/playground/internal/tools/slug"
	"github.com/tkahng/playground/internal/tools/types"
)

type DbTaskStoreInterface interface { // size=16 (0x10)
	CalculateTaskRankStatus(ctx context.Context, taskId uuid.UUID, taskProjectId uuid.UUID, status models.TaskStatus, currentRank float64, position int64) (float64, error)
	CountItems(ctx context.Context, projectID uuid.UUID, status models.TaskStatus, excludeID uuid.UUID) (int64, error)
	CountTaskProjects(ctx context.Context, filter *TaskProjectsFilter) (int64, error)
	CountTasks(ctx context.Context, filter *TaskFilter) (int64, error)
	CreateTask(ctx context.Context, task *models.Task) (*models.Task, error)
	CreateTaskFromInput(ctx context.Context, teamID uuid.UUID, projectID uuid.UUID, memberID uuid.UUID, input *CreateTaskProjectTaskDTO) (*models.Task, error)
	CreateTaskProject(ctx context.Context, input *CreateTaskProjectDTO) (*models.TaskProject, error)
	CreateTaskProjectWithTasks(ctx context.Context, input *CreateTaskProjectWithTasksDTO) (*models.TaskProject, error)
	DeleteTask(ctx context.Context, taskID uuid.UUID) error
	DeleteTaskProject(ctx context.Context, taskProjectID uuid.UUID) error
	FindLastTaskRank(ctx context.Context, taskProjectID uuid.UUID) (float64, error)
	FindTask(ctx context.Context, task *TaskFilter) (*models.Task, error)
	FindTaskByID(ctx context.Context, id uuid.UUID) (*models.Task, error)
	FindTaskByIDForUpdate(ctx context.Context, id uuid.UUID) (*models.Task, error)
	FindTaskProjectByID(ctx context.Context, id uuid.UUID) (*models.TaskProject, error)
	FindWorkflowByID(ctx context.Context, id uuid.UUID) (*models.Workflow, error)
	FindWorkflowStatusByID(ctx context.Context, id uuid.UUID) (*models.WorkflowStatus, error)
	GetTaskFirstPosition(ctx context.Context, projectID uuid.UUID, status models.TaskStatus, excludeID uuid.UUID) (float64, error)
	GetTaskLastPosition(ctx context.Context, projectID uuid.UUID, status models.TaskStatus, excludeID uuid.UUID) (float64, error)
	GetTaskPositions(ctx context.Context, projectID uuid.UUID, status models.TaskStatus, excludeID uuid.UUID, offset int64) ([]float64, error)
	ListWorkflows(ctx context.Context, filter *WorkflowFilter) ([]*models.Workflow, error)
	ListTaskProjects(ctx context.Context, input *TaskProjectsFilter) ([]*models.TaskProject, error)
	ListTasks(ctx context.Context, input *TaskFilter) ([]*models.Task, error)
	LoadWorkflowStatuses(ctx context.Context, workflowIds ...uuid.UUID) ([][]*models.WorkflowStatus, error)
	CreateWorkflowStatus(ctx context.Context, workflowID uuid.UUID, input *CreateWorkflowStatusDTO) (*models.WorkflowStatus, error)
	UpdateWorkflowStatus(ctx context.Context, workflowStatusID uuid.UUID, input *UpdateWorkflowStatusDTO) (*models.WorkflowStatus, error)
	DeleteWorkflowStatus(ctx context.Context, workflowStatusID uuid.UUID) error
	LoadTaskProjectsTasks(ctx context.Context, projectIds ...uuid.UUID) ([][]*models.Task, error)
	taskWhere(task *TaskFilter) *map[string]any
	UpdateTask(ctx context.Context, task *models.Task) error
	UpdateTaskProject(ctx context.Context, taskProjectID uuid.UUID, input *UpdateTaskProjectBaseDTO) error
	UpdateTaskProjectUpdateDate(ctx context.Context, taskProjectID uuid.UUID) error
	UpdateTaskRankStatus(ctx context.Context, taskID uuid.UUID, position int64, status models.TaskStatus) error
	WithTx(dbx database.Dbx) *DbTaskStore
	GetTeamTaskStats(ctx context.Context, teamId uuid.UUID) (*models.TaskStats, error)
	FindAndUpdateTask(ctx context.Context, taskID uuid.UUID, input *UpdateTaskDto) error
	FindTasksDueToday(ctx context.Context) ([]*models.Task, error)
	FindTasksOverdue(ctx context.Context) ([]*models.Task, error)
	CreateWorkflow(ctx context.Context, input *CreateWorkflowDTO) (*models.Workflow, error)
	UpdateWorkflow(ctx context.Context, workflowID uuid.UUID, input *UpdateWorkflowDTO) (*models.Workflow, error)
}

type DbTaskStore struct {
	db database.Dbx
}

var _ DbTaskStoreInterface = (*DbTaskStore)(nil)

func (s *DbTaskStore) WithTx(dbx database.Dbx) *DbTaskStore {
	return &DbTaskStore{
		db: dbx,
	}
}

type WorkflowFilter struct {
	Ids       []uuid.UUID `query:"ids,omitempty" json:"ids,omitempty" format:"uuid" required:"false"`
	TeamIds   []uuid.UUID `query:"team_ids,omitempty" json:"team_ids,omitempty" format:"uuid" required:"false"`
	AppliesTo []string    `query:"applies_to,omitempty" json:"applies_to,omitempty" required:"false"`
	IsDefault *bool       `query:"is_default,omitempty" json:"is_default,omitempty" required:"false"`
}

type CreateWorkflowDTO struct {
	TeamID            uuid.UUID  `json:"team_id" required:"true" format:"uuid"`
	CreatedByMemberID *uuid.UUID `json:"created_by_member_id,omitempty" required:"false" format:"uuid"`
	AppliesTo         string     `json:"applies_to" required:"true" enum:"project,task"`
	Name              string     `json:"name" required:"true" minLength:"1"`
	Description       *string    `json:"description,omitempty" required:"false"`
}

type UpdateWorkflowDTO struct {
	Name        *string `json:"name,omitempty" required:"false" minLength:"1"`
	Description *string `json:"description,omitempty" required:"false"`
}

func (s *DbTaskStore) ListWorkflows(ctx context.Context, filter *WorkflowFilter) ([]*models.Workflow, error) {
	where := map[string]any{}
	if filter != nil {
		if len(filter.Ids) > 0 {
			where["id"] = map[string]any{"_in": filter.Ids}
		}
		if len(filter.TeamIds) > 0 {
			where["team_id"] = map[string]any{"_in": filter.TeamIds}
		}
		if len(filter.AppliesTo) > 0 {
			where["applies_to"] = map[string]any{"_in": filter.AppliesTo}
		}
		if filter.IsDefault != nil {
			where["is_default"] = map[string]any{"_eq": *filter.IsDefault}
		}
	}
	var wherePtr *map[string]any
	if len(where) > 0 {
		wherePtr = &where
	}
	return repository.Workflow.Get(
		ctx,
		s.db,
		wherePtr,
		&map[string]string{"applies_to": "ASC", "created_at": "ASC"},
		nil,
		nil,
	)
}

func (s *DbTaskStore) LoadWorkflowStatuses(ctx context.Context, workflowIds ...uuid.UUID) ([][]*models.WorkflowStatus, error) {
	if len(workflowIds) == 0 {
		return [][]*models.WorkflowStatus{}, nil
	}
	statuses, err := repository.WorkflowStatus.Get(
		ctx,
		s.db,
		&map[string]any{
			"workflow_id": map[string]any{
				"_in": workflowIds,
			},
		},
		&map[string]string{
			"rank": "ASC",
		},
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return mapper.MapToManyPointer(statuses, workflowIds, func(status *models.WorkflowStatus) uuid.UUID {
		return status.WorkflowID
	}), nil
}

func (s *DbTaskStore) CreateWorkflow(ctx context.Context, input *CreateWorkflowDTO) (*models.Workflow, error) {
	if input == nil {
		return nil, apierrors.BadRequest("workflow input is required")
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, apierrors.BadRequest("workflow name is required")
	}
	appliesTo, err := normalizeWorkflowAppliesTo(input.AppliesTo)
	if err != nil {
		return nil, err
	}
	workflow, err := repository.Workflow.PostOne(ctx, s.db, &models.Workflow{
		TeamID:            input.TeamID,
		CreatedByMemberID: input.CreatedByMemberID,
		AppliesTo:         appliesTo,
		Name:              name,
		Description:       input.Description,
		IsDefault:         false,
	})
	if database.IsUniqConstraintErr(err) {
		return nil, apierrors.Conflict("workflow already exists")
	}
	return workflow, err
}

func (s *DbTaskStore) UpdateWorkflow(ctx context.Context, workflowID uuid.UUID, input *UpdateWorkflowDTO) (*models.Workflow, error) {
	if input == nil {
		return nil, apierrors.BadRequest("workflow input is required")
	}
	workflow, err := s.FindWorkflowByID(ctx, workflowID)
	if err != nil {
		return nil, err
	}
	if workflow == nil {
		return nil, apierrors.NotFound("workflow not found")
	}
	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return nil, apierrors.BadRequest("workflow name is required")
		}
		workflow.Name = name
	}
	if input.Description != nil {
		workflow.Description = input.Description
	}
	updated, err := repository.Workflow.PutOne(ctx, s.db, workflow)
	if database.IsUniqConstraintErr(err) {
		return nil, apierrors.Conflict("workflow already exists")
	}
	return updated, err
}

type CreateWorkflowStatusDTO struct {
	Name        string   `json:"name" required:"true" minLength:"1"`
	Slug        *string  `json:"slug,omitempty" required:"false"`
	Description *string  `json:"description,omitempty" required:"false"`
	Category    string   `json:"category" required:"true" enum:"todo,in_progress,done"`
	Color       *string  `json:"color,omitempty" required:"false"`
	Rank        *float64 `json:"rank,omitempty" required:"false"`
	IsCompleted *bool    `json:"is_completed,omitempty" required:"false"`
}

type UpdateWorkflowStatusDTO struct {
	Name        *string  `json:"name,omitempty" required:"false" minLength:"1"`
	Slug        *string  `json:"slug,omitempty" required:"false"`
	Description *string  `json:"description,omitempty" required:"false"`
	Category    *string  `json:"category,omitempty" required:"false" enum:"todo,in_progress,done"`
	Color       *string  `json:"color,omitempty" required:"false"`
	Rank        *float64 `json:"rank,omitempty" required:"false"`
	IsCompleted *bool    `json:"is_completed,omitempty" required:"false"`
}

func (s *DbTaskStore) FindWorkflowByID(ctx context.Context, id uuid.UUID) (*models.Workflow, error) {
	return repository.Workflow.GetOne(
		ctx,
		s.db,
		&map[string]any{
			"id": map[string]any{
				"_eq": id,
			},
		},
	)
}

func (s *DbTaskStore) FindWorkflowStatusByID(ctx context.Context, id uuid.UUID) (*models.WorkflowStatus, error) {
	return s.findWorkflowStatusByID(ctx, id)
}

func (s *DbTaskStore) CreateWorkflowStatus(ctx context.Context, workflowID uuid.UUID, input *CreateWorkflowStatusDTO) (*models.WorkflowStatus, error) {
	if input == nil {
		return nil, apierrors.BadRequest("workflow status input is required")
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, apierrors.BadRequest("workflow status name is required")
	}
	category, err := normalizeWorkflowStatusCategory(input.Category)
	if err != nil {
		return nil, err
	}
	statusSlug := slug.NewSlug(name)
	if input.Slug != nil {
		statusSlug = slug.NewSlug(*input.Slug)
	}
	if statusSlug == "" {
		return nil, apierrors.BadRequest("workflow status slug is required")
	}
	rank := input.Rank
	if rank == nil {
		nextRank, err := s.nextWorkflowStatusRank(ctx, workflowID)
		if err != nil {
			return nil, err
		}
		rank = &nextRank
	}
	isCompleted := category == string(models.TaskStatusDone)
	if input.IsCompleted != nil {
		isCompleted = *input.IsCompleted
	}
	status, err := repository.WorkflowStatus.PostOne(ctx, s.db, &models.WorkflowStatus{
		WorkflowID:  workflowID,
		Name:        name,
		Slug:        statusSlug,
		Description: input.Description,
		Category:    category,
		Color:       input.Color,
		Rank:        *rank,
		IsCompleted: isCompleted,
	})
	if database.IsUniqConstraintErr(err) {
		return nil, apierrors.Conflict("workflow status already exists")
	}
	return status, err
}

func (s *DbTaskStore) UpdateWorkflowStatus(ctx context.Context, workflowStatusID uuid.UUID, input *UpdateWorkflowStatusDTO) (*models.WorkflowStatus, error) {
	if input == nil {
		return nil, apierrors.BadRequest("workflow status input is required")
	}
	status, err := s.findWorkflowStatusByID(ctx, workflowStatusID)
	if err != nil {
		return nil, err
	}
	if status == nil {
		return nil, apierrors.NotFound("workflow status not found")
	}
	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return nil, apierrors.BadRequest("workflow status name is required")
		}
		status.Name = name
	}
	if input.Slug != nil {
		statusSlug := slug.NewSlug(*input.Slug)
		if statusSlug == "" {
			return nil, apierrors.BadRequest("workflow status slug is required")
		}
		status.Slug = statusSlug
	}
	if input.Description != nil {
		status.Description = input.Description
	}
	if input.Category != nil {
		category, err := normalizeWorkflowStatusCategory(*input.Category)
		if err != nil {
			return nil, err
		}
		status.Category = category
		if input.IsCompleted == nil {
			status.IsCompleted = category == string(models.TaskStatusDone)
		}
	}
	if input.Color != nil {
		status.Color = input.Color
	}
	if input.Rank != nil {
		status.Rank = *input.Rank
	}
	if input.IsCompleted != nil {
		status.IsCompleted = *input.IsCompleted
	}
	updated, err := repository.WorkflowStatus.PutOne(ctx, s.db, status)
	if database.IsUniqConstraintErr(err) {
		return nil, apierrors.Conflict("workflow status already exists")
	}
	return updated, err
}

func (s *DbTaskStore) DeleteWorkflowStatus(ctx context.Context, workflowStatusID uuid.UUID) error {
	status, err := s.findWorkflowStatusByID(ctx, workflowStatusID)
	if err != nil {
		return err
	}
	if status == nil {
		return apierrors.NotFound("workflow status not found")
	}
	taskCount, err := repository.Task.Count(ctx, s.db, &map[string]any{
		"workflow_status_id": map[string]any{
			"_eq": workflowStatusID,
		},
	})
	if err != nil {
		return err
	}
	projectCount, err := repository.TaskProject.Count(ctx, s.db, &map[string]any{
		"workflow_status_id": map[string]any{
			"_eq": workflowStatusID,
		},
	})
	if err != nil {
		return err
	}
	if taskCount > 0 || projectCount > 0 {
		return apierrors.Conflict("workflow status is in use")
	}
	_, err = repository.WorkflowStatus.Delete(ctx, s.db, &map[string]any{
		"id": map[string]any{
			"_eq": workflowStatusID,
		},
	})
	return err
}

func (s *DbTaskStore) CreateTask(ctx context.Context, task *models.Task) (*models.Task, error) {
	status, workflowStatusID, err := s.resolveTaskWorkflowStatus(ctx, task.ProjectID, task.WorkflowStatusID, task.Status)
	if err != nil {
		return nil, err
	}
	task.Status = status
	task.WorkflowStatusID = workflowStatusID
	return repository.Task.PostOne(ctx, s.db, task)
}

func (s *DbTaskStore) FindTask(ctx context.Context, task *TaskFilter) (*models.Task, error) {
	where := s.taskWhere(task)

	return repository.Task.GetOne(ctx, s.db, where)
}

type TaskFilter struct {
	PaginatedInput
	SortParams
	Q                  string              `query:"q,omitempty" json:"q,omitempty" required:"false"`
	Ids                []uuid.UUID         `query:"ids,omitempty" json:"ids,omitempty" format:"uuid" required:"false"`
	ProjectIds         []uuid.UUID         `query:"project_ids,omitempty" json:"project_ids,omitempty" format:"uuid" required:"false"`
	Names              []string            `query:"names,omitempty" json:"names,omitempty" required:"false"`
	Statuses           []models.TaskStatus `query:"statuses,omitempty" json:"statuses,omitempty" required:"false"`
	TeamIds            []uuid.UUID         `query:"team_ids,omitempty" json:"team_ids,omitempty" format:"uuid" required:"false"`
	CreatedByMemberIds []uuid.UUID         `query:"created_by_member_ids,omitempty" json:"created_by_member_ids,omitempty" format:"uuid" required:"false"`
	ParentIds          []uuid.UUID         `query:"parent_ids,omitempty" json:"parent_ids,omitempty" format:"uuid" required:"false"`
}

func (*DbTaskStore) taskWhere(task *TaskFilter) *map[string]any {
	if task == nil {
		return nil
	}
	where := map[string]any{}
	if task.Q != "" {
		where["_or"] = []map[string]any{
			{
				"_and": []map[string]any{
					{
						"name": map[string]any{
							"_ilike": "%" + task.Q + "%",
						},
					},
				},
			},
			{
				"_and": []map[string]any{
					{
						"description": map[string]any{
							"_ilike": "%" + task.Q + "%",
						},
					},
				},
			},
		}
	}
	if len(task.Ids) > 0 {
		where["id"] = map[string]any{
			"_in": task.Ids,
		}
	}
	if len(task.Names) > 0 {
		where["name"] = map[string]any{
			"_in": task.Names,
		}
	}
	if len(task.ProjectIds) > 0 {
		where["project_id"] = map[string]any{
			"_in": task.ProjectIds,
		}
	}
	if len(task.TeamIds) > 0 {
		where["team_id"] = map[string]any{
			"_in": task.TeamIds,
		}
	}
	if len(task.CreatedByMemberIds) > 0 {
		where["created_by_member_id"] = map[string]any{
			"_in": task.CreatedByMemberIds,
		}
	}
	if len(task.Statuses) > 0 {
		where["status"] = map[string]any{
			"_in": task.Statuses,
		}
	}

	if len(task.ParentIds) > 0 {
		where["parent_id"] = map[string]any{
			"_in": task.ParentIds,
		}
	}
	return &where
}

type UpdateTaskDto struct {
	Name             string            `db:"name" json:"name"`
	Description      *string           `db:"description" json:"description"`
	Status           models.TaskStatus `db:"status" json:"status" enum:"todo,in_progress,done"`
	WorkflowStatusID *uuid.UUID        `db:"workflow_status_id" json:"workflow_status_id" nullable:"true"`
	StartAt          *time.Time        `db:"start_at" json:"start_at" nullable:"true"`
	EndAt            *time.Time        `db:"end_at" json:"end_at" nullable:"true"`
	AssigneeID       *uuid.UUID        `db:"assignee_id" json:"assignee_id" nullable:"true"`
	ReporterID       *uuid.UUID        `db:"reporter_id" json:"reporter_id" nullable:"true"`
	ParentID         *uuid.UUID        `db:"parent_id" json:"parent_id" nullable:"true"`
}

func (s *DbTaskStore) FindAndUpdateTask(ctx context.Context, taskID uuid.UUID, input *UpdateTaskDto) error {
	task, err := s.FindTaskByID(ctx, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return apierrors.NotFound("task not found")
	}

	task.Name = input.Name
	task.Description = input.Description
	task.Status = models.TaskStatus(input.Status)
	task.WorkflowStatusID = input.WorkflowStatusID
	task.StartAt = input.StartAt
	task.EndAt = input.EndAt
	task.AssigneeID = input.AssigneeID
	task.ReporterID = input.ReporterID
	task.ParentID = input.ParentID
	err = s.UpdateTask(ctx, task)
	if err != nil {
		return err
	}
	return nil
}

func (s *DbTaskStore) UpdateTask(ctx context.Context, task *models.Task) error {
	status, workflowStatusID, err := s.resolveTaskWorkflowStatus(ctx, task.ProjectID, task.WorkflowStatusID, task.Status)
	if err != nil {
		return err
	}
	task.Status = status
	task.WorkflowStatusID = workflowStatusID
	_, err = repository.Task.PutOne(ctx, s.db, task)
	return err
}

func (s *DbTaskStore) CountItems(ctx context.Context, projectID uuid.UUID, status models.TaskStatus, excludeID uuid.UUID) (int64, error) {
	return repository.Task.Count(
		ctx,
		s.db,
		&map[string]any{
			"project_id": map[string]any{
				"_eq": projectID,
			},
			"status": map[string]any{
				"_eq": status,
			},
			"id": map[string]any{
				"_neq": excludeID,
			},
		},
	)
}

type rankRow struct {
	Rank float64 `db:"rank"`
}

func (s *DbTaskStore) GetTaskFirstPosition(ctx context.Context, projectID uuid.UUID, status models.TaskStatus, excludeID uuid.UUID) (float64, error) {
	q := squirrel.Select("rank").
		From(repository.TaskBuilder.TableName()).
		Where(squirrel.Eq{
			"project_id": projectID,
			"status":     status,
		}).
		Where(squirrel.NotEq{"id": excludeID}).
		OrderBy("rank ASC").
		Limit(1).
		PlaceholderFormat(squirrel.Dollar)

	rows, err := database.QueryWithBuilder[rankRow](ctx, s.db, q)
	if err != nil {
		return 0, err
	}
	if len(rows) == 0 {
		return 0, nil
	}
	return rows[0].Rank, nil
}

func (s *DbTaskStore) GetTaskLastPosition(ctx context.Context, projectID uuid.UUID, status models.TaskStatus, excludeID uuid.UUID) (float64, error) {
	q := squirrel.Select("rank").
		From(repository.TaskBuilder.TableName()).
		Where(squirrel.Eq{
			"project_id": projectID,
			"status":     status,
		}).
		Where(squirrel.NotEq{"id": excludeID}).
		OrderBy("rank DESC").
		Limit(1).
		PlaceholderFormat(squirrel.Dollar)

	rows, err := database.QueryWithBuilder[rankRow](ctx, s.db, q)
	if err != nil {
		return 0, err
	}
	if len(rows) == 0 {
		return 0, nil
	}
	return rows[0].Rank, nil
}

func (s *DbTaskStore) GetTaskPositions(ctx context.Context, projectID uuid.UUID, status models.TaskStatus, excludeID uuid.UUID, offset int64) ([]float64, error) {
	q := squirrel.Select("rank").
		From(repository.TaskBuilder.TableName()).
		Where(squirrel.Eq{
			"project_id": projectID,
			"status":     status,
		}).
		Where(squirrel.NotEq{"id": excludeID}).
		OrderBy("rank ASC").
		Limit(2).
		Offset(uint64(offset)).
		PlaceholderFormat(squirrel.Dollar)

	rows, err := database.QueryWithBuilder[rankRow](ctx, s.db, q)
	if err != nil {
		return nil, err
	}
	return mapper.Map(rows, func(row rankRow) float64 {
		return row.Rank
	}), nil
}
func NewDbTaskStore(db database.Dbx) *DbTaskStore {
	return &DbTaskStore{
		db: db,
	}
}

func (s *DbTaskStore) LoadTaskProjectsTasks(ctx context.Context, projectIds ...uuid.UUID) ([][]*models.Task, error) {
	tasks, err := repository.Task.Get(
		ctx,
		s.db,
		&map[string]any{
			"project_id": map[string]any{
				"_in": projectIds,
			},
		},
		nil,
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return mapper.MapToManyPointer(tasks, projectIds, func(t *models.Task) uuid.UUID {
		return t.ProjectID
	}), nil
}

func (s *DbTaskStore) FindTaskByID(ctx context.Context, id uuid.UUID) (*models.Task, error) {
	return repository.Task.GetOne(
		ctx,
		s.db,
		&map[string]any{
			"id": map[string]any{
				"_eq": id,
			},
		},
	)
}

// FindTaskByIDForUpdate fetches the task and acquires a row-level lock (SELECT … FOR UPDATE).
// Must be called inside a transaction.
func (s *DbTaskStore) FindTaskByIDForUpdate(ctx context.Context, id uuid.UUID) (*models.Task, error) {
	return repository.Task.GetOneForUpdate(
		ctx,
		s.db,
		&map[string]any{
			"id": map[string]any{
				"_eq": id,
			},
		},
	)
}

func (s *DbTaskStore) FindLastTaskRank(ctx context.Context, taskProjectID uuid.UUID) (float64, error) {
	tasks, err := repository.Task.Get(
		ctx,
		s.db,
		&map[string]any{
			"project_id": map[string]any{
				"_eq": taskProjectID,
			},
		},
		&map[string]string{
			"rank": "DESC",
		},
		types.Pointer(1),
		nil,
	)
	if err != nil {
		return 0, err
	}
	if len(tasks) == 0 {
		return 0, nil
	}
	task := tasks[0]
	return task.Rank + 1000, nil
}

func (s *DbTaskStore) DeleteTask(ctx context.Context, taskID uuid.UUID) error {
	_, err := repository.Task.Delete(
		ctx,
		s.db,
		&map[string]any{
			"id": map[string]any{
				"_eq": taskID,
			},
		},
	)
	return err
}

type TaskProjectsFilter struct {
	PaginatedInput
	SortParams
	Q       string                     `query:"q,omitempty" json:"q,omitempty" required:"false"`
	Ids     []uuid.UUID                `query:"ids,omitempty" json:"ids,omitempty" format:"uuid" required:"false"`
	TeamIds []uuid.UUID                `query:"team_ids,omitempty" json:"team_ids,omitempty" format:"uuid" required:"false"`
	Status  []models.TaskProjectStatus `query:"status,omitempty" json:"statuses,omitempty" required:"false"`
}

func (s *DbTaskStore) FindTaskProjectByID(ctx context.Context, id uuid.UUID) (*models.TaskProject, error) {
	return repository.TaskProject.GetOne(
		ctx,
		s.db,
		&map[string]any{
			"id": map[string]any{
				"_eq": id,
			},
		},
	)
}
func (s *DbTaskStore) DeleteTaskProject(ctx context.Context, taskProjectID uuid.UUID) error {
	_, err := repository.TaskProject.Delete(
		ctx,
		s.db,
		&map[string]any{
			"id": map[string]any{
				"_eq": taskProjectID,
			},
		},
	)
	return err
}
func ListTasksOrderByFunc(input *TaskFilter) *map[string]string {
	sortBy, sortOrder := input.Sort()
	if slices.Contains(repository.TaskBuilder.FieldNames(), sortBy) {
		return &map[string]string{
			sortBy: strings.ToUpper(sortOrder),
		}
	}
	return nil
}

// ListTasks implements AdminCrudActions.
func (s *DbTaskStore) ListTasks(ctx context.Context, input *TaskFilter) ([]*models.Task, error) {
	iimit, offset := pagination(input)
	order := ListTasksOrderByFunc(input)
	where := s.taskWhere(input)
	data, err := repository.Task.Get(
		ctx,
		s.db,
		where,
		order,
		&iimit,
		&offset,
	)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// CountTasks implements AdminCrudActions.
func (s *DbTaskStore) CountTasks(ctx context.Context, filter *TaskFilter) (int64, error) {
	where := s.taskWhere(filter)
	return repository.Task.Count(ctx, s.db, where)
}
func (*DbTaskStore) TaskProjectWhere(task *TaskProjectsFilter) *map[string]any {
	if task == nil {
		return nil
	}
	where := map[string]any{}
	if task.Q != "" {
		where["_or"] = []map[string]any{
			{
				"_and": []map[string]any{
					{
						"name": map[string]any{
							"_ilike": "%" + task.Q + "%",
						},
					},
				},
			},
			{
				"_and": []map[string]any{
					{
						"description": map[string]any{
							"_ilike": "%" + task.Q + "%",
						},
					},
				},
			},
		}
	}
	if len(task.Ids) > 0 {
		where["id"] = map[string]any{
			"_in": task.Ids,
		}
	}

	if len(task.TeamIds) > 0 {
		where["team_id"] = map[string]any{
			"_in": task.TeamIds,
		}
	}
	// if len(task.CreatedByMemberIds) > 0 {
	// 	where["created_by_member_id"] = map[string]any{
	// 		"_in": task.CreatedByMemberIds,
	// 	}
	// }
	if len(task.Status) > 0 {
		where["status"] = map[string]any{
			"_in": task.Status,
		}
	}

	return &where
}

func ListTaskProjectsOrderByFunc(input *TaskProjectsFilter) *map[string]string {
	sortBy, sortOrder := input.Sort()
	if slices.Contains(repository.TaskProjectBuilder.FieldNames(), sortBy) {
		return &map[string]string{
			sortBy: strings.ToUpper(sortOrder),
		}
	}
	return nil
}

// ListTaskProjects implements AdminCrudActions.
func (s *DbTaskStore) ListTaskProjects(ctx context.Context, input *TaskProjectsFilter) ([]*models.TaskProject, error) {
	limit, offset := input.LimitOffset()
	oredr := ListTaskProjectsOrderByFunc(input)
	where := s.TaskProjectWhere(input)
	data, err := repository.TaskProject.Get(
		ctx,
		s.db,
		where,
		oredr,
		&limit,
		&offset,
	)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// CountTaskProjects implements AdminCrudActions.
func (s *DbTaskStore) CountTaskProjects(ctx context.Context, filter *TaskProjectsFilter) (int64, error) {
	where := s.TaskProjectWhere(filter)
	return repository.TaskProject.Count(ctx, s.db, where)
}

const (
	workflowAppliesToProject = "project"
	workflowAppliesToTask    = "task"
)

type defaultWorkflowStatus struct {
	name        string
	slug        string
	category    string
	color       string
	rank        float64
	isCompleted bool
}

var defaultWorkflowStatuses = []defaultWorkflowStatus{
	{name: "To do", slug: "todo", category: "todo", color: "#6b7280", rank: 1000, isCompleted: false},
	{name: "In progress", slug: "in_progress", category: "in_progress", color: "#2563eb", rank: 2000, isCompleted: false},
	{name: "Done", slug: "done", category: "done", color: "#16a34a", rank: 3000, isCompleted: true},
}

func (s *DbTaskStore) ensureDefaultWorkflow(ctx context.Context, teamID uuid.UUID, memberID *uuid.UUID, appliesTo string) (*models.Workflow, error) {
	workflow, err := s.findDefaultWorkflow(ctx, teamID, appliesTo)
	if err == nil {
		return workflow, s.ensureDefaultWorkflowStatuses(ctx, workflow.ID)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	name := "Default task workflow"
	description := "Default workflow for task board status."
	if appliesTo == workflowAppliesToProject {
		name = "Default project workflow"
		description = "Default workflow for project lifecycle status."
	}

	q := squirrel.Insert(repository.WorkflowBuilder.TableName()).
		Columns("team_id", "created_by_member_id", "applies_to", "name", "description", "is_default").
		Values(teamID, memberID, appliesTo, name, description, true).
		Suffix("on conflict (team_id, applies_to, name) do update set is_default = true returning " + repository.WorkflowBuilder.ColumnNamesJoined()).
		PlaceholderFormat(squirrel.Dollar)

	created, err := database.QueryWithBuilder[models.Workflow](ctx, s.db, q)
	if err != nil {
		return nil, err
	}
	if len(created) == 0 {
		return nil, pgx.ErrNoRows
	}
	if err := s.ensureDefaultWorkflowStatuses(ctx, created[0].ID); err != nil {
		return nil, err
	}
	return &created[0], nil
}

func (s *DbTaskStore) findDefaultWorkflow(ctx context.Context, teamID uuid.UUID, appliesTo string) (*models.Workflow, error) {
	q := squirrel.Select(repository.WorkflowBuilder.ColumnNames()...).
		From(repository.WorkflowBuilder.TableName()).
		Where(squirrel.Eq{
			"team_id":    teamID,
			"applies_to": appliesTo,
			"is_default": true,
		}).
		Limit(1).
		PlaceholderFormat(squirrel.Dollar)

	workflows, err := database.QueryWithBuilder[models.Workflow](ctx, s.db, q)
	if err != nil {
		return nil, err
	}
	if len(workflows) == 0 {
		return nil, pgx.ErrNoRows
	}
	return &workflows[0], nil
}

func (s *DbTaskStore) ensureDefaultWorkflowStatuses(ctx context.Context, workflowID uuid.UUID) error {
	for _, status := range defaultWorkflowStatuses {
		q := squirrel.Insert(repository.WorkflowStatusBuilder.TableName()).
			Columns("workflow_id", "name", "slug", "category", "color", "rank", "is_completed").
			Values(workflowID, status.name, status.slug, status.category, status.color, status.rank, status.isCompleted).
			Suffix("on conflict (workflow_id, slug) do nothing").
			PlaceholderFormat(squirrel.Dollar)
		if _, err := database.ExecWithBuilder(ctx, s.db, q); err != nil {
			return err
		}
	}
	return nil
}

func (s *DbTaskStore) findWorkflowStatusIDByCategory(ctx context.Context, workflowID uuid.UUID, category string) (*uuid.UUID, error) {
	q := squirrel.Select("id").
		From(repository.WorkflowStatusBuilder.TableName()).
		Where(squirrel.Eq{
			"workflow_id": workflowID,
			"category":    category,
		}).
		OrderBy("rank ASC").
		Limit(1).
		PlaceholderFormat(squirrel.Dollar)

	id, err := database.PgxQuerySingleScalar[uuid.UUID](ctx, s.db, q)
	if err != nil {
		return nil, err
	}
	if id == uuid.Nil {
		return nil, pgx.ErrNoRows
	}
	return &id, nil
}

func normalizeWorkflowStatusCategory(category string) (string, error) {
	category = strings.TrimSpace(category)
	if slices.Contains([]string{
		string(models.TaskStatusTodo),
		string(models.TaskStatusInProgress),
		string(models.TaskStatusDone),
	}, category) {
		return category, nil
	}
	return "", apierrors.BadRequest("workflow status category must be todo, in_progress, or done")
}

func normalizeWorkflowAppliesTo(appliesTo string) (string, error) {
	appliesTo = strings.TrimSpace(appliesTo)
	if slices.Contains([]string{workflowAppliesToProject, workflowAppliesToTask}, appliesTo) {
		return appliesTo, nil
	}
	return "", apierrors.BadRequest("workflow applies_to must be project or task")
}

func (s *DbTaskStore) nextWorkflowStatusRank(ctx context.Context, workflowID uuid.UUID) (float64, error) {
	q := squirrel.Select("coalesce(max(rank), 0) + 1000").
		From(repository.WorkflowStatusBuilder.TableName()).
		Where(squirrel.Eq{
			"workflow_id": workflowID,
		}).
		PlaceholderFormat(squirrel.Dollar)

	return database.PgxQuerySingleScalar[float64](ctx, s.db, q)
}

func (s *DbTaskStore) findWorkflowStatusByID(ctx context.Context, id uuid.UUID) (*models.WorkflowStatus, error) {
	return repository.WorkflowStatus.GetOne(
		ctx,
		s.db,
		&map[string]any{
			"id": map[string]any{
				"_eq": id,
			},
		},
	)
}

func (s *DbTaskStore) validateWorkflowStatus(ctx context.Context, id uuid.UUID, workflowID uuid.UUID) (*models.WorkflowStatus, error) {
	status, err := s.findWorkflowStatusByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if status == nil {
		return nil, apierrors.BadRequest("workflow status not found")
	}
	if status.WorkflowID != workflowID {
		return nil, apierrors.BadRequest("workflow status must belong to the workflow")
	}
	return status, nil
}

func (s *DbTaskStore) validateWorkflowForTeam(ctx context.Context, workflowID uuid.UUID, teamID uuid.UUID, appliesTo string) (*models.Workflow, error) {
	workflow, err := s.FindWorkflowByID(ctx, workflowID)
	if err != nil {
		return nil, err
	}
	if workflow == nil || workflow.TeamID != teamID {
		return nil, apierrors.BadRequest("workflow not found")
	}
	if workflow.AppliesTo != appliesTo {
		return nil, apierrors.BadRequest("workflow applies_to does not match")
	}
	return workflow, nil
}

func (s *DbTaskStore) validateTaskWorkflowAssignable(ctx context.Context, workflowID uuid.UUID) error {
	statuses, err := s.LoadWorkflowStatuses(ctx, workflowID)
	if err != nil {
		return err
	}
	if len(statuses) == 0 {
		return apierrors.BadRequest("workflow must have statuses before assignment")
	}
	categories := map[string]bool{}
	for _, status := range statuses[0] {
		categories[status.Category] = true
	}
	for _, category := range []string{
		string(models.TaskStatusTodo),
		string(models.TaskStatusInProgress),
		string(models.TaskStatusDone),
	} {
		if !categories[category] {
			return apierrors.BadRequest("workflow must have statuses for todo, in_progress, and done before assignment")
		}
	}
	return nil
}

func (s *DbTaskStore) findProjectWorkflowStatusID(ctx context.Context, teamID uuid.UUID, memberID *uuid.UUID, status models.TaskProjectStatus) (*uuid.UUID, error) {
	workflow, err := s.ensureDefaultWorkflow(ctx, teamID, memberID, workflowAppliesToProject)
	if err != nil {
		return nil, err
	}
	return s.findWorkflowStatusIDByCategory(ctx, workflow.ID, string(status))
}

func (s *DbTaskStore) resolveProjectWorkflowStatus(ctx context.Context, teamID uuid.UUID, memberID *uuid.UUID, workflowStatusID *uuid.UUID, status models.TaskProjectStatus) (models.TaskProjectStatus, *uuid.UUID, error) {
	workflow, err := s.ensureDefaultWorkflow(ctx, teamID, memberID, workflowAppliesToProject)
	if err != nil {
		return status, nil, err
	}
	if workflowStatusID == nil {
		id, err := s.findWorkflowStatusIDByCategory(ctx, workflow.ID, string(status))
		return status, id, err
	}
	workflowStatus, err := s.validateWorkflowStatus(ctx, *workflowStatusID, workflow.ID)
	if err != nil {
		return status, nil, err
	}
	return models.TaskProjectStatus(workflowStatus.Category), workflowStatusID, nil
}

func (s *DbTaskStore) findTaskWorkflowStatusID(ctx context.Context, projectID uuid.UUID, status models.TaskStatus) (*uuid.UUID, error) {
	project, err := s.FindTaskProjectByID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, apierrors.NotFound("task project not found")
	}
	workflowID := project.WorkflowID
	if workflowID == nil {
		workflow, err := s.ensureDefaultWorkflow(ctx, project.TeamID, project.CreatedByMemberID, workflowAppliesToTask)
		if err != nil {
			return nil, err
		}
		workflowID = &workflow.ID
		project.WorkflowID = workflowID
		if _, err := repository.TaskProject.PutOne(ctx, s.db, project); err != nil {
			return nil, err
		}
	}
	return s.findWorkflowStatusIDByCategory(ctx, *workflowID, string(status))
}

func (s *DbTaskStore) resolveTaskWorkflowStatus(ctx context.Context, projectID uuid.UUID, workflowStatusID *uuid.UUID, status models.TaskStatus) (models.TaskStatus, *uuid.UUID, error) {
	project, err := s.FindTaskProjectByID(ctx, projectID)
	if err != nil {
		return status, nil, err
	}
	if project == nil {
		return status, nil, apierrors.NotFound("task project not found")
	}
	workflowID := project.WorkflowID
	if workflowID == nil {
		workflow, err := s.ensureDefaultWorkflow(ctx, project.TeamID, project.CreatedByMemberID, workflowAppliesToTask)
		if err != nil {
			return status, nil, err
		}
		workflowID = &workflow.ID
		project.WorkflowID = workflowID
		if _, err := repository.TaskProject.PutOne(ctx, s.db, project); err != nil {
			return status, nil, err
		}
	}
	if workflowStatusID == nil {
		id, err := s.findWorkflowStatusIDByCategory(ctx, *workflowID, string(status))
		return status, id, err
	}
	workflowStatus, err := s.validateWorkflowStatus(ctx, *workflowStatusID, *workflowID)
	if err != nil {
		return status, nil, err
	}
	return models.TaskStatus(workflowStatus.Category), workflowStatusID, nil
}

func (s *DbTaskStore) CreateTaskProject(ctx context.Context, input *CreateTaskProjectDTO) (*models.TaskProject, error) {
	var taskWorkflow *models.Workflow
	var err error
	if input.WorkflowID != nil {
		taskWorkflow, err = s.validateWorkflowForTeam(ctx, *input.WorkflowID, input.TeamID, workflowAppliesToTask)
		if err != nil {
			return nil, err
		}
		if err := s.validateTaskWorkflowAssignable(ctx, taskWorkflow.ID); err != nil {
			return nil, err
		}
	} else {
		taskWorkflow, err = s.ensureDefaultWorkflow(ctx, input.TeamID, &input.MemberID, workflowAppliesToTask)
		if err != nil {
			return nil, err
		}
	}
	projectStatus, projectWorkflowStatusID, err := s.resolveProjectWorkflowStatus(ctx, input.TeamID, &input.MemberID, input.WorkflowStatusID, input.Status)
	if err != nil {
		return nil, err
	}
	taskProject := models.TaskProject{
		TeamID:            input.TeamID,
		CreatedByMemberID: &input.MemberID,
		WorkflowID:        &taskWorkflow.ID,
		WorkflowStatusID:  projectWorkflowStatusID,
		Name:              input.Name,
		Description:       input.Description,
		Status:            projectStatus,
		Rank:              input.Rank,
	}
	projects, err := repository.TaskProject.PostOne(ctx, s.db, &taskProject)
	if err != nil {
		return nil, err
	}
	return projects, nil
}

type CreateTaskProjectDTO struct {
	TeamID           uuid.UUID                `json:"team_id" required:"true" format:"uuid"`
	MemberID         uuid.UUID                `json:"member_id" required:"true" format:"uuid"`
	Name             string                   `json:"name" required:"true"`
	Description      *string                  `json:"description,omitempty" required:"false"`
	Status           models.TaskProjectStatus `json:"status" required:"false" enum:"todo,in_progress,done" default:"todo"`
	WorkflowStatusID *uuid.UUID               `json:"workflow_status_id,omitempty" required:"false" format:"uuid"`
	WorkflowID       *uuid.UUID               `json:"workflow_id,omitempty" required:"false" format:"uuid"`
	Rank             float64                  `json:"rank,omitempty" required:"false"`
}
type CreateTaskProjectTaskDTO struct {
	Name             string            `json:"name" required:"true"`
	Description      *string           `json:"description,omitempty" required:"false"`
	Status           models.TaskStatus `json:"status" required:"false" enum:"todo,in_progress,done" default:"todo"`
	WorkflowStatusID *uuid.UUID        `json:"workflow_status_id,omitempty" required:"false" format:"uuid"`
	Rank             float64           `json:"rank,omitempty" required:"false"`
}
type CreateTaskProjectWithTasksDTO struct {
	CreateTaskProjectDTO
	Tasks []CreateTaskProjectTaskDTO `json:"tasks,omitempty" required:"false"`
}

func (s *DbTaskStore) CreateTaskProjectWithTasks(ctx context.Context, input *CreateTaskProjectWithTasksDTO) (*models.TaskProject, error) {
	count, err := s.CountTaskProjects(ctx, nil)
	if err != nil {
		return nil, err
	}
	input.Rank = float64(count * 1000)
	taskProject, err := s.CreateTaskProject(ctx, &input.CreateTaskProjectDTO)
	if err != nil {
		return nil, err
	}
	if taskProject == nil {
		return nil, errors.New("task project not created")
	}
	tasks := []*models.Task{}
	for i, task := range input.Tasks {
		task.Rank = float64(i * 1000)
		newTask, err := s.CreateTaskFromInput(ctx, taskProject.TeamID, taskProject.ID, input.MemberID, &task)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, newTask)
	}
	taskProject.Tasks = tasks
	return taskProject, nil
}

func (s *DbTaskStore) CreateTaskFromInput(ctx context.Context, teamID uuid.UUID, projectID uuid.UUID, memberID uuid.UUID, input *CreateTaskProjectTaskDTO) (*models.Task, error) {
	setter := models.Task{
		ProjectID:         projectID,
		CreatedByMemberID: &memberID,
		TeamID:            teamID,
		Name:              input.Name,
		Description:       input.Description,
		Status:            models.TaskStatus(input.Status),
		WorkflowStatusID:  input.WorkflowStatusID,
		Rank:              input.Rank,
	}
	task, err := s.CreateTask(ctx, &setter)
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (s *DbTaskStore) CalculateTaskRankStatus(ctx context.Context, taskId uuid.UUID, taskProjectId uuid.UUID, status models.TaskStatus, currentRank float64, position int64) (float64, error) {
	if position == 0 {
		res, err := repository.Task.Get(
			ctx,
			s.db,
			&map[string]any{
				"project_id": map[string]any{
					"_eq": taskProjectId,
				},
				"status": map[string]any{
					"_eq": status,
				},
			},
			&map[string]string{
				"rank": "ASC",
			},
			types.Pointer(1),
			nil,
		)
		if err != nil {
			return 0, err
		}
		if len(res) == 0 {
			return 0, nil
		}
		response := res[0]

		if response.ID == taskId {
			return response.Rank, nil
		}
		return response.Rank - 1000, nil
	}
	ele, err := repository.Task.Get(
		ctx,
		s.db,
		&map[string]any{
			"project_id": map[string]any{
				"_eq": taskProjectId,
			},
		},
		&map[string]string{
			"rank": "ASC",
		},
		types.Pointer(1),
		types.Pointer(int(position)),
	)
	if err != nil {
		return 0, err
	}
	if len(ele) == 0 {
		return 0, nil
	}
	element := ele[0]

	if element.ID == taskId {
		return element.Rank, nil
	}
	if currentRank > element.Rank {
		sideELe, err := repository.Task.Get(
			ctx,
			s.db,
			&map[string]any{
				"project_id": map[string]any{
					"_eq": taskProjectId,
				},
				"status": map[string]any{
					"_eq": status,
				},
			},
			&map[string]string{
				"rank": "ASC",
			},
			types.Pointer(1),
			types.Pointer(int(position-1)),
		)
		if err != nil {
			return 0, err
		}
		if len(sideELe) == 0 {
			return element.Rank - 1000, nil
		}
		sideElements := sideELe[0]
		return (element.Rank + sideElements.Rank) / 2, nil
	}
	sideele, err := repository.Task.Get(
		ctx,
		s.db,
		&map[string]any{
			"project_id": map[string]any{
				"_eq": taskProjectId,
			},
			"status": map[string]any{
				"_eq": status,
			},
		},
		&map[string]string{
			"rank": "ASC",
		},
		types.Pointer(1),
		types.Pointer(int(position+1)),
	)
	if err != nil {
		return 0, err
	}
	if len(sideele) == 0 {
		return element.Rank + 1000, nil
	}
	sideElements := sideele[0]
	return (element.Rank + sideElements.Rank) / 2, nil
}

func (s *DbTaskStore) UpdateTaskProjectUpdateDate(ctx context.Context, taskProjectID uuid.UUID) error {
	q := squirrel.Update("task.task_projects").
		Where("id = ?", taskProjectID).
		Set("updated_at", time.Now())

	_, err := database.ExecWithBuilder(ctx, s.db, q.PlaceholderFormat(squirrel.Dollar))
	return err
}

type UpdateTaskProjectBaseDTO struct {
	Name             string                   `json:"name" required:"true"`
	Description      *string                  `json:"description,omitempty" required:"false"`
	Status           models.TaskProjectStatus `json:"status" enum:"todo,in_progress,done"`
	WorkflowStatusID *uuid.UUID               `json:"workflow_status_id,omitempty" required:"false" format:"uuid"`
	WorkflowID       *uuid.UUID               `json:"workflow_id,omitempty" required:"false" format:"uuid"`
	Rank             float64                  `json:"rank"`
	Position         *int64                   `json:"position,omitempty" required:"false"`
}

func (s *DbTaskStore) UpdateTaskProject(ctx context.Context, taskProjectID uuid.UUID, input *UpdateTaskProjectBaseDTO) error {
	taskProject, err := s.FindTaskProjectByID(ctx, taskProjectID)
	if err != nil {
		return err
	}
	if taskProject == nil {
		return apierrors.NotFound("task project not found")
	}
	taskProject.Name = input.Name
	taskProject.Description = input.Description
	taskProject.Status = models.TaskProjectStatus(input.Status)
	workflowChanged := false
	if input.WorkflowID != nil && (taskProject.WorkflowID == nil || *taskProject.WorkflowID != *input.WorkflowID) {
		workflow, err := s.validateWorkflowForTeam(ctx, *input.WorkflowID, taskProject.TeamID, workflowAppliesToTask)
		if err != nil {
			return err
		}
		if err := s.validateTaskWorkflowAssignable(ctx, workflow.ID); err != nil {
			return err
		}
		taskProject.WorkflowID = &workflow.ID
		workflowChanged = true
	}
	status, workflowStatusID, err := s.resolveProjectWorkflowStatus(ctx, taskProject.TeamID, taskProject.CreatedByMemberID, input.WorkflowStatusID, taskProject.Status)
	if err != nil {
		return err
	}
	taskProject.Status = status
	input.Status = status
	taskProject.WorkflowStatusID = workflowStatusID
	taskProject.Rank = input.Rank
	_, err = repository.TaskProject.PutOne(ctx, s.db, taskProject)
	if err != nil {
		return err
	}
	if workflowChanged {
		if err := s.remapTaskWorkflowStatuses(ctx, taskProject.ID, *taskProject.WorkflowID); err != nil {
			return err
		}
	}
	return nil
}

func (s *DbTaskStore) remapTaskWorkflowStatuses(ctx context.Context, taskProjectID uuid.UUID, workflowID uuid.UUID) error {
	q := squirrel.Update(repository.TaskBuilder.TableName()).
		Set("workflow_status_id", squirrel.Expr(
			"(select id from task.workflow_statuses where workflow_id = ? and category = task.tasks.status::text order by rank asc limit 1)",
			workflowID,
		)).
		Where(squirrel.Eq{"project_id": taskProjectID}).
		PlaceholderFormat(squirrel.Dollar)

	_, err := database.ExecWithBuilder(ctx, s.db, q)
	return err
}

func (s *DbTaskStore) UpdateTaskRankStatus(ctx context.Context, taskID uuid.UUID, position int64, status models.TaskStatus) error {
	task, err := s.FindTaskByID(ctx, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return apierrors.NotFound("task not found")
	}
	rank, err := s.CalculateTaskRankStatus(ctx, task.ID, task.ProjectID, status, task.Rank, position)
	if err != nil {
		return err
	}
	task.Rank = rank
	task.Status = status
	workflowStatusID, err := s.findTaskWorkflowStatusID(ctx, task.ProjectID, status)
	if err != nil {
		return err
	}
	task.WorkflowStatusID = workflowStatusID
	_, err = repository.Task.PutOne(ctx, s.db, task)
	if err != nil {
		return err
	}
	err = s.UpdateTaskProjectUpdateDate(ctx, task.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to update task project update date: %w", err)
	}
	return nil
}

func (s *DbTaskStore) GetTeamTaskStats(ctx context.Context, teamId uuid.UUID) (*models.TaskStats, error) {
	projectCounts, err := database.QueryWithBuilder[database.CountOutput](
		ctx,
		s.db,
		squirrel.Select("count(*) as count").
			From(repository.TaskProjectBuilder.TableName()).
			Where(squirrel.Eq{"team_id": teamId}).
			PlaceholderFormat(squirrel.Dollar),
	)
	if err != nil {
		return nil, err
	}
	completedProjectCounts, err := database.QueryWithBuilder[database.CountOutput](
		ctx,
		s.db,
		squirrel.Select("count(*) as count").
			From(repository.TaskProjectBuilder.TableName()).
			Where(squirrel.Eq{
				"team_id": teamId,
				"status":  models.TaskProjectStatusDone,
			}).
			PlaceholderFormat(squirrel.Dollar),
	)
	if err != nil {
		return nil, err
	}
	taskCounts, err := database.QueryWithBuilder[database.CountOutput](
		ctx,
		s.db,
		squirrel.Select("count(*) as count").
			From(repository.TaskBuilder.TableName()).
			Where(squirrel.Eq{"team_id": teamId}).
			PlaceholderFormat(squirrel.Dollar),
	)
	if err != nil {
		return nil, err
	}
	completedTaskCounts, err := database.QueryWithBuilder[database.CountOutput](
		ctx,
		s.db,
		squirrel.Select("count(*) as count").
			From(repository.TaskBuilder.TableName()).
			Where(squirrel.Eq{
				"team_id": teamId,
				"status":  models.TaskStatusDone,
			}).
			PlaceholderFormat(squirrel.Dollar),
	)
	if err != nil {
		return nil, err
	}
	return &models.TaskStats{
		TotalProjects:     countOutputValue(projectCounts),
		CompletedProjects: countOutputValue(completedProjectCounts),
		TotalTasks:        countOutputValue(taskCounts),
		CompletedTasks:    countOutputValue(completedTaskCounts),
	}, nil
}

func countOutputValue(counts []database.CountOutput) int64 {
	if len(counts) == 0 {
		return 0
	}
	return counts[0].Count
}

// FindTasksDueToday returns non-done tasks whose end_at falls within today (UTC).
func (s *DbTaskStore) FindTasksDueToday(ctx context.Context) ([]*models.Task, error) {
	now := time.Now().UTC()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour)
	return repository.Task.Get(
		ctx,
		s.db,
		&map[string]any{
			"_and": []map[string]any{
				{"end_at": map[string]any{repository.Gte: startOfDay}},
				{"end_at": map[string]any{repository.Lt: endOfDay}},
				{"status": map[string]any{repository.Neq: models.TaskStatusDone}},
			},
		},
		nil,
		nil,
		nil,
	)
}

// FindTasksOverdue returns non-done tasks whose end_at is before today (UTC).
func (s *DbTaskStore) FindTasksOverdue(ctx context.Context) ([]*models.Task, error) {
	startOfDay := time.Now().UTC().Truncate(24 * time.Hour)
	return repository.Task.Get(
		ctx,
		s.db,
		&map[string]any{
			"_and": []map[string]any{
				{"end_at": map[string]any{repository.Lt: startOfDay}},
				{"status": map[string]any{repository.Neq: models.TaskStatusDone}},
			},
		},
		nil,
		nil,
		nil,
	)
}
