package stores

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/services"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
	"github.com/tkahng/authgo/internal/tools/types"
)

type taskStore struct {
	db database.Dbx
}

// CreateTask implements services.TaskStore.
func (s *taskStore) CreateTask(ctx context.Context, task *models.Task) (*models.Task, error) {
	return repository.Task.PostOne(ctx, s.db, task)
}

// FindTask implements services.TaskStore.
func (s *taskStore) FindTask(ctx context.Context, task *models.Task) (*models.Task, error) {

	where := s.TaskWhere(task)

	return repository.Task.GetOne(ctx, s.db, where)
}

func (*taskStore) TaskWhere(task *models.Task) *map[string]any {
	if task == nil {
		return nil
	}
	where := map[string]any{}
	if task.ID != uuid.Nil {
		where["id"] = map[string]any{
			"_eq": task.ID,
		}
	}
	if task.ProjectID != uuid.Nil {
		where["project_id"] = map[string]any{
			"_eq": task.ProjectID,
		}
	}
	if task.TeamID != uuid.Nil {
		where["team_id"] = map[string]any{
			"_eq": task.TeamID,
		}
	}
	if task.CreatedByMemberID != nil {
		where["created_by_member_id"] = map[string]any{
			"_eq": task.CreatedByMemberID,
		}
	}
	if task.Status != "" {
		where["status"] = map[string]any{
			"_eq": task.Status,
		}
	}
	if task.Name != "" {
		where["name"] = map[string]any{
			"_like": fmt.Sprintf("%%%s%%", task.Name),
		}
	}
	if task.Description != nil {
		where["description"] = map[string]any{
			"_like": fmt.Sprintf("%%%s%%", *task.Description),
		}
	}
	if task.ParentID != nil {
		where["parent_id"] = map[string]any{
			"_eq": task.ParentID,
		}
	}
	return &where
}

// UpdateTask implements services.TaskStore.
func (s *taskStore) UpdateTask(ctx context.Context, task *models.Task) error {
	_, err := repository.Task.PutOne(ctx, s.db, task)
	return err
}

func (s *taskStore) CountItems(ctx context.Context, projectID uuid.UUID, status models.TaskStatus, excludeID uuid.UUID) (int64, error) {
	var count int64
	query := `
		SELECT COUNT(*) 
		FROM tasks 
		WHERE project_id = $1 AND status = $2 AND id != $3
	`
	err := s.db.QueryRow(ctx, query, projectID, status, excludeID).Scan(&count)
	return count, err
}

func (s *taskStore) GetTaskFirstPosition(ctx context.Context, projectID uuid.UUID, status models.TaskStatus, excludeID uuid.UUID) (float64, error) {
	var rank float64
	query := `
		SELECT rank 
		FROM tasks 
		WHERE project_id = $1 AND status = $2 AND id != $3
		ORDER BY rank ASC 
		LIMIT 1
	`
	err := s.db.QueryRow(ctx, query, projectID, status, excludeID).Scan(&rank)
	return rank, err
}

func (s *taskStore) GetTaskLastPosition(ctx context.Context, projectID uuid.UUID, status models.TaskStatus, excludeID uuid.UUID) (float64, error) {
	var rank float64
	query := `
		SELECT rank 
		FROM tasks 
		WHERE project_id = $1 AND status = $2 AND id != $3
		ORDER BY rank DESC 
		LIMIT 1
	`
	err := s.db.QueryRow(ctx, query, projectID, status, excludeID).Scan(&rank)
	return rank, err
}

func (s *taskStore) GetTaskPositions(ctx context.Context, projectID uuid.UUID, status models.TaskStatus, excludeID uuid.UUID, offset int64) ([]float64, error) {
	query := `
		SELECT rank 
		FROM tasks 
		WHERE project_id = $1 AND status = $2 AND id != $3
		ORDER BY rank ASC 
		LIMIT $4 OFFSET $5
	`

	rows, err := s.db.Query(ctx, query, projectID, status, excludeID, 2, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ranks []float64
	for rows.Next() {
		var pos float64
		if err := rows.Scan(&pos); err != nil {
			return nil, err
		}
		ranks = append(ranks, pos)
	}

	return ranks, rows.Err()
}
func NewTaskStore(db database.Dbx) *taskStore {
	return &taskStore{
		db: db,
	}
}

var _ services.TaskStore = &taskStore{}

const (
	LoadTaskProjectsTasksQuery = `
SELECT tp.id as key,
        json_agg(to_json(t.*)) AS "data"
FROM public.task_projects tp
        LEFT JOIN public.tasks t ON tp.id = t.project_id
WHERE tp.id = ANY ($1::uuid [])
GROUP BY tp.id;`
)

func (s *taskStore) LoadTaskProjectsTasks(ctx context.Context, projectIds ...uuid.UUID) ([][]*models.Task, error) {
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

func (s *taskStore) FindTaskByID(ctx context.Context, id uuid.UUID) (*models.Task, error) {
	task, err := repository.Task.GetOne(
		ctx,
		s.db,
		&map[string]any{
			"id": map[string]any{
				"_eq": id,
			},
		},
	)
	return database.OptionalRow(task, err)
}

func (s *taskStore) FindLastTaskRank(ctx context.Context, taskProjectID uuid.UUID) (float64, error) {
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

func (s *taskStore) DeleteTask(ctx context.Context, taskID uuid.UUID) error {

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

func (s *taskStore) FindTaskProjectByID(ctx context.Context, id uuid.UUID) (*models.TaskProject, error) {

	task, err := repository.TaskProject.GetOne(
		ctx,
		s.db,
		&map[string]any{
			"id": map[string]any{
				"_eq": id,
			},
		},
	)
	return database.OptionalRow(task, err)
}
func (s *taskStore) DeleteTaskProject(ctx context.Context, taskProjectID uuid.UUID) error {
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
func ListTasksOrderByFunc(input *shared.TaskListParams) *map[string]string {
	order := map[string]string{}

	if input == nil || input.SortBy == "" || input.SortOrder == "" {
		return nil
	}
	order[input.SortBy] = strings.ToUpper(input.SortOrder)
	return &order
}

func ListTasksFilterFunc(filter *shared.TaskListFilter) *map[string]any {
	if filter == nil {
		return nil
	}
	where := make(map[string]any)
	if filter.Q != "" {
		where["_or"] = []map[string]any{
			{
				"name": map[string]any{
					"_ilike": "%" + filter.Q + "%",
				},
			},
			{
				"description": map[string]any{
					"_ilike": "%" + filter.Q + "%",
				},
			},
		}
	}
	if len(filter.Status) > 0 {
		where["status"] = map[string]any{
			"_in": filter.Status,
		}
	}
	if len(filter.CreatedByMemberID) > 0 {
		where["created_by_member_id"] = map[string]any{
			"_eq": filter.CreatedByMemberID,
		}
	}

	if filter.ProjectID != "" {
		where["project_id"] = map[string]any{
			"_eq": filter.ProjectID,
		}
	}

	if len(filter.Ids) > 0 {
		where["id"] = map[string]any{
			"_in": filter.Ids,
		}
	}
	if filter.ParentID != "" {
		where["parent_id"] = map[string]any{
			"_eq": filter.ParentID,
		}
	}
	if filter.ParentStatus != "" {
		if filter.ParentStatus == "parent" {
			where["parent_id"] = map[string]any{
				"_eq": "NULL",
			}
		} else if filter.ParentStatus == "child" {
			where["parent_id"] = map[string]any{
				"_neq": "NULL",
			}
		}
	}

	return &where
}

// ListTasks implements AdminCrudActions.
func (s *taskStore) ListTasks(ctx context.Context, input *shared.TaskListParams) ([]*models.Task, error) {
	filter := input.TaskListFilter
	pageInput := &input.PaginatedInput

	iimit, offset := database.PaginateRepo(pageInput)
	order := ListTasksOrderByFunc(input)
	where := ListTasksFilterFunc(&filter)
	data, err := repository.Task.Get(
		ctx,
		s.db,
		where,
		order,
		iimit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// CountTasks implements AdminCrudActions.
func (s *taskStore) CountTasks(ctx context.Context, filter *shared.TaskListFilter) (int64, error) {
	where := ListTasksFilterFunc(filter)
	return repository.Task.Count(ctx, s.db, where)
}
func ListTaskProjectsFilterFunc(filter *shared.TaskProjectsListFilter) *map[string]any {
	if filter == nil {
		return nil
	}
	where := make(map[string]any)
	if filter.Q != "" {
		where["_or"] = []map[string]any{
			{
				"name": map[string]any{
					"_ilike": "%" + filter.Q + "%",
				},
			},
			{
				"description": map[string]any{
					"_ilike": "%" + filter.Q + "%",
				},
			},
		}
	}

	if len(filter.Ids) > 0 {
		where["id"] = map[string]any{
			"_in": filter.Ids,
		}
	}

	if filter.TeamID != "" {
		where["user_id"] = map[string]any{
			"_eq": filter.TeamID,
		}
	}

	return &where
}

func ListTaskProjectsOrderByFunc(input *shared.TaskProjectsListParams) *map[string]string {

	if input == nil || input.SortBy == "" || input.SortOrder == "" {
		return nil
	}
	order := make(map[string]string)
	order[input.SortBy] = strings.ToUpper(input.SortOrder)
	return &order
}

// ListTaskProjects implements AdminCrudActions.
func (s *taskStore) ListTaskProjects(ctx context.Context, input *shared.TaskProjectsListParams) ([]*models.TaskProject, error) {
	filter := input.TaskProjectsListFilter
	pageInput := &input.PaginatedInput

	limit, offset := database.PaginateRepo(pageInput)
	oredr := ListTaskProjectsOrderByFunc(input)
	where := ListTaskProjectsFilterFunc(&filter)
	data, err := repository.TaskProject.Get(
		ctx,
		s.db,
		where,
		oredr,
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// CountTaskProjects implements AdminCrudActions.
func (s *taskStore) CountTaskProjects(ctx context.Context, filter *shared.TaskProjectsListFilter) (int64, error) {
	where := ListTaskProjectsFilterFunc(filter)
	return repository.TaskProject.Count(ctx, s.db, where)
}

func (s *taskStore) CreateTaskProject(ctx context.Context, input *shared.CreateTaskProjectDTO) (*models.TaskProject, error) {
	taskProject := models.TaskProject{
		TeamID:            input.TeamID,
		CreatedByMemberID: &input.MemberID,
		Name:              input.Name,
		Description:       input.Description,
		Status:            models.TaskProjectStatus(input.Status),
		Rank:              input.Rank,
	}
	projects, err := repository.TaskProject.PostOne(ctx, s.db, &taskProject)
	if err != nil {
		return nil, err
	}
	return projects, nil
}

func (s *taskStore) CreateTaskProjectWithTasks(ctx context.Context, input *shared.CreateTaskProjectWithTasksDTO) (*models.TaskProject, error) {
	count, err := s.CountTaskProjects(ctx, nil)
	if err != nil {
		return nil, err
	}
	input.CreateTaskProjectDTO.Rank = float64(count * 1000)
	taskProject, err := s.CreateTaskProject(ctx, &input.CreateTaskProjectDTO)
	if err != nil {
		return nil, err
	}
	if taskProject == nil {
		return nil, errors.New("task project not created")
	}
	var tasks []*models.Task
	for i, task := range input.Tasks {
		task.Rank = float64(i * 1000)
		newTask, err := s.CreateTaskFromInput(ctx, taskProject.TeamID, taskProject.ID, input.CreateTaskProjectDTO.MemberID, &task)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, newTask)
	}
	return taskProject, nil
}

func (s *taskStore) CreateTaskFromInput(ctx context.Context, teamID uuid.UUID, projectID uuid.UUID, memberID uuid.UUID, input *shared.CreateTaskProjectTaskDTO) (*models.Task, error) {
	setter := models.Task{
		ProjectID:         projectID,
		CreatedByMemberID: &memberID,
		TeamID:            teamID,
		Name:              input.Name,
		Description:       input.Description,
		Status:            models.TaskStatus(input.Status),
		Rank:              input.Rank,
	}
	task, err := s.CreateTask(ctx, &setter)
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (s *taskStore) CalculateTaskRankStatus(ctx context.Context, taskId uuid.UUID, taskProjectId uuid.UUID, status models.TaskStatus, currentRank float64, position int64) (float64, error) {
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
				"order": "ASC",
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
			"order": "ASC",
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
				"order": "ASC",
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
			"order": "ASC",
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

func (s *taskStore) UpdateTaskProjectUpdateDate(ctx context.Context, taskProjectID uuid.UUID) error {
	q := squirrel.Update("task_projects").
		Where("id = ?", taskProjectID).
		Set("updated_at", time.Now())

	err := database.ExecWithBuilder(ctx, s.db, q.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return err
	}
	return nil
}

func (s *taskStore) UpdateTaskProject(ctx context.Context, taskProjectID uuid.UUID, input *shared.UpdateTaskProjectBaseDTO) error {
	taskProject, err := s.FindTaskProjectByID(ctx, taskProjectID)
	if err != nil {
		return err
	}
	if taskProject == nil {
		return errors.New("task project not found")
	}
	taskProject.Name = input.Name
	taskProject.Description = input.Description
	taskProject.Status = models.TaskProjectStatus(input.Status)
	taskProject.Rank = input.Rank
	_, err = repository.TaskProject.PutOne(ctx, s.db, taskProject)
	if err != nil {
		return err
	}
	return nil
}

func (s *taskStore) UpdateTaskRankStatus(ctx context.Context, taskID uuid.UUID, position int64, status models.TaskStatus) error {
	task, err := s.FindTaskByID(ctx, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("task not found")
	}
	rank, err := s.CalculateTaskRankStatus(ctx, task.ID, task.ProjectID, status, task.Rank, position)
	if err != nil {
		return err
	}
	task.Rank = rank
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
