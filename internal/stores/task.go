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
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
	"github.com/tkahng/authgo/internal/tools/types"
	"github.com/tkahng/authgo/internal/tools/utils"
)

type DbTaskStoreInterface interface { // size=16 (0x10)
	CalculateTaskRankStatus(ctx context.Context, taskId uuid.UUID, taskProjectId uuid.UUID, status models.TaskStatus, currentRank float64, position int64) (float64, error)
	CountItems(ctx context.Context, projectID uuid.UUID, status models.TaskStatus, excludeID uuid.UUID) (int64, error)
	CountTaskProjects(ctx context.Context, filter *TaskProjectsFilter) (int64, error)
	CountTasks(ctx context.Context, filter *TaskFilter) (int64, error)
	CreateTask(ctx context.Context, task *models.Task) (*models.Task, error)
	CreateTaskFromInput(ctx context.Context, teamID uuid.UUID, projectID uuid.UUID, memberID uuid.UUID, input *shared.CreateTaskProjectTaskDTO) (*models.Task, error)
	CreateTaskProject(ctx context.Context, input *shared.CreateTaskProjectDTO) (*models.TaskProject, error)
	CreateTaskProjectWithTasks(ctx context.Context, input *shared.CreateTaskProjectWithTasksDTO) (*models.TaskProject, error)
	DeleteTask(ctx context.Context, taskID uuid.UUID) error
	DeleteTaskProject(ctx context.Context, taskProjectID uuid.UUID) error
	FindLastTaskRank(ctx context.Context, taskProjectID uuid.UUID) (float64, error)
	FindTask(ctx context.Context, task *TaskFilter) (*models.Task, error)
	FindTaskByID(ctx context.Context, id uuid.UUID) (*models.Task, error)
	FindTaskProjectByID(ctx context.Context, id uuid.UUID) (*models.TaskProject, error)
	GetTaskFirstPosition(ctx context.Context, projectID uuid.UUID, status models.TaskStatus, excludeID uuid.UUID) (float64, error)
	GetTaskLastPosition(ctx context.Context, projectID uuid.UUID, status models.TaskStatus, excludeID uuid.UUID) (float64, error)
	GetTaskPositions(ctx context.Context, projectID uuid.UUID, status models.TaskStatus, excludeID uuid.UUID, offset int64) ([]float64, error)
	ListTaskProjects(ctx context.Context, input *TaskProjectsFilter) ([]*models.TaskProject, error)
	ListTasks(ctx context.Context, input *TaskFilter) ([]*models.Task, error)
	LoadTaskProjectsTasks(ctx context.Context, projectIds ...uuid.UUID) ([][]*models.Task, error)
	taskWhere(task *TaskFilter) *map[string]any
	UpdateTask(ctx context.Context, task *models.Task) error
	UpdateTaskProject(ctx context.Context, taskProjectID uuid.UUID, input *shared.UpdateTaskProjectBaseDTO) error
	UpdateTaskProjectUpdateDate(ctx context.Context, taskProjectID uuid.UUID) error
	UpdateTaskRankStatus(ctx context.Context, taskID uuid.UUID, position int64, status models.TaskStatus) error
	WithTx(dbx database.Dbx) *DbTaskStore
}

type DbTaskStore struct {
	db database.Dbx
}

func (s *DbTaskStore) WithTx(dbx database.Dbx) *DbTaskStore {
	return &DbTaskStore{
		db: dbx,
	}
}

func (s *DbTaskStore) CreateTask(ctx context.Context, task *models.Task) (*models.Task, error) {
	return repository.Task.PostOne(ctx, s.db, task)
}

func (s *DbTaskStore) FindTask(ctx context.Context, task *TaskFilter) (*models.Task, error) {

	where := s.taskWhere(task)

	return repository.Task.GetOne(ctx, s.db, where)
}

type TaskFilter struct {
	shared.PaginatedInput
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

func (s *DbTaskStore) UpdateTask(ctx context.Context, task *models.Task) error {
	_, err := repository.Task.PutOne(ctx, s.db, task)
	return err
}

func (s *DbTaskStore) CountItems(ctx context.Context, projectID uuid.UUID, status models.TaskStatus, excludeID uuid.UUID) (int64, error) {
	var count int64
	query := `
		SELECT COUNT(*) 
		FROM tasks 
		WHERE project_id = $1 AND status = $2 AND id != $3
	`
	err := s.db.QueryRow(ctx, query, projectID, status, excludeID).Scan(&count)
	return count, err
}

func (s *DbTaskStore) GetTaskFirstPosition(ctx context.Context, projectID uuid.UUID, status models.TaskStatus, excludeID uuid.UUID) (float64, error) {
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

func (s *DbTaskStore) GetTaskLastPosition(ctx context.Context, projectID uuid.UUID, status models.TaskStatus, excludeID uuid.UUID) (float64, error) {
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

func (s *DbTaskStore) GetTaskPositions(ctx context.Context, projectID uuid.UUID, status models.TaskStatus, excludeID uuid.UUID, offset int64) ([]float64, error) {
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
func NewDbTaskStore(db database.Dbx) *DbTaskStore {
	return &DbTaskStore{
		db: db,
	}
}

const (
	LoadTaskProjectsTasksQuery = `
SELECT tp.id as key,
        json_agg(to_json(t.*)) AS "data"
FROM public.task_projects tp
        LEFT JOIN public.tasks t ON tp.id = t.project_id
WHERE tp.id = ANY ($1::uuid [])
GROUP BY tp.id;`
)

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
	shared.PaginatedInput
	SortParams
	Q        string                     `query:"q,omitempty" json:"q,omitempty" required:"false"`
	Ids      []uuid.UUID                `query:"ids,omitempty" json:"ids,omitempty" format:"uuid" required:"false"`
	TeamIds  []uuid.UUID                `query:"team_ids,omitempty" json:"team_ids,omitempty" format:"uuid" required:"false"`
	Statuses []models.TaskProjectStatus `query:"statuses,omitempty" json:"statuses,omitempty" required:"false"`
}

func (s *DbTaskStore) FindTaskProjectByID(ctx context.Context, id uuid.UUID) (*models.TaskProject, error) {

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
	if slices.Contains(repository.TaskBuilder.ColumnNames(), utils.Quote(sortBy)) {
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
	if len(task.Statuses) > 0 {
		where["status"] = map[string]any{
			"_in": task.Statuses,
		}
	}

	return &where
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

func ListTaskProjectsOrderByFunc(input *TaskProjectsFilter) *map[string]string {
	sortBy, sortOrder := input.Sort()
	if slices.Contains(repository.TaskProjectBuilder.ColumnNames(), utils.Quote(sortBy)) {
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

func (s *DbTaskStore) CreateTaskProject(ctx context.Context, input *shared.CreateTaskProjectDTO) (*models.TaskProject, error) {
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

func (s *DbTaskStore) CreateTaskProjectWithTasks(ctx context.Context, input *shared.CreateTaskProjectWithTasksDTO) (*models.TaskProject, error) {
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
	var tasks []*models.Task
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

func (s *DbTaskStore) CreateTaskFromInput(ctx context.Context, teamID uuid.UUID, projectID uuid.UUID, memberID uuid.UUID, input *shared.CreateTaskProjectTaskDTO) (*models.Task, error) {
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

func (s *DbTaskStore) UpdateTaskProjectUpdateDate(ctx context.Context, taskProjectID uuid.UUID) error {
	q := squirrel.Update("task_projects").
		Where("id = ?", taskProjectID).
		Set("updated_at", time.Now())

	_, err := database.ExecWithBuilder(ctx, s.db, q.PlaceholderFormat(squirrel.Dollar))
	return err
}

func (s *DbTaskStore) UpdateTaskProject(ctx context.Context, taskProjectID uuid.UUID, input *shared.UpdateTaskProjectBaseDTO) error {
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

func (s *DbTaskStore) UpdateTaskRankStatus(ctx context.Context, taskID uuid.UUID, position int64, status models.TaskStatus) error {
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
