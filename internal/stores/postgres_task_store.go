package stores

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/crudrepo"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/services"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
	"github.com/tkahng/authgo/internal/tools/types"
)

type taskStore struct {
	db database.Dbx
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
	tasks, err := crudrepo.Task.Get(
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
	task, err := crudrepo.Task.GetOne(
		ctx,
		s.db,
		&map[string]any{
			"id": map[string]any{
				"_eq": id.String(),
			},
		},
	)
	return database.OptionalRow(task, err)
}

func (s *taskStore) FindLastTaskOrder(ctx context.Context, taskProjectID uuid.UUID) (float64, error) {
	tasks, err := crudrepo.Task.Get(
		ctx,
		s.db,
		&map[string]any{
			"project_id": map[string]any{
				"_eq": taskProjectID.String(),
			},
		},
		&map[string]string{
			"order": "DESC",
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
	return task.Order + 1000, nil
}

func (s *taskStore) DeleteTask(ctx context.Context, taskID uuid.UUID) error {

	_, err := crudrepo.Task.Delete(
		ctx,
		s.db,
		&map[string]any{
			"id": map[string]any{
				"_eq": taskID.String(),
			},
		},
	)
	return err
}

func (s *taskStore) FindTaskProjectByID(ctx context.Context, id uuid.UUID) (*models.TaskProject, error) {

	task, err := crudrepo.TaskProject.GetOne(
		ctx,
		s.db,
		&map[string]any{
			"id": map[string]any{
				"_eq": id.String(),
			},
		},
	)
	return database.OptionalRow(task, err)
}
func (s *taskStore) DeleteTaskProject(ctx context.Context, taskProjectID uuid.UUID) error {
	_, err := crudrepo.TaskProject.Delete(
		ctx,
		s.db,
		&map[string]any{
			"id": map[string]any{
				"_eq": taskProjectID.String(),
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
	if len(filter.UserID) > 0 {
		where["user_id"] = map[string]any{
			"_eq": filter.UserID,
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
	data, err := crudrepo.Task.Get(
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
	return crudrepo.Task.Count(ctx, s.db, where)
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

	if filter.UserID != "" {
		where["user_id"] = map[string]any{
			"_eq": filter.UserID,
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
	data, err := crudrepo.TaskProject.Get(
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
	return crudrepo.TaskProject.Count(ctx, s.db, where)
}

func (s *taskStore) CreateTaskProject(ctx context.Context, input *shared.CreateTaskProjectDTO) (*models.TaskProject, error) {
	taskProject := models.TaskProject{
		// UserID:      input.UserID,
		TeamID:      input.TeamID,
		CreatedBy:   input.MemberID,
		Name:        input.Name,
		Description: input.Description,
		Status:      models.TaskProjectStatus(input.Status),
		Order:       input.Order,
	}
	projects, err := crudrepo.TaskProject.PostOne(ctx, s.db, &taskProject)
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
	input.CreateTaskProjectDTO.Order = float64(count * 1000)
	taskProject, err := s.CreateTaskProject(ctx, &input.CreateTaskProjectDTO)
	if err != nil {
		return nil, err
	}
	var tasks []*models.Task
	for i, task := range input.Tasks {
		task.CreatedBy = input.CreateTaskProjectDTO.MemberID
		task.TeamID = input.CreateTaskProjectDTO.TeamID
		task.Order = float64(i * 1000)
		newTask, err := s.CreateTask(ctx, taskProject.ID, &task)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, newTask)
	}
	return taskProject, nil
}

func (s *taskStore) CreateTaskWithChildren(ctx context.Context, projectID uuid.UUID, input *shared.CreateTaskWithChildrenDTO) (*models.Task, error) {
	task, err := s.CreateTask(ctx, projectID, &input.CreateTaskBaseDTO)
	if err != nil {
		return nil, err
	}
	// for _, child := range input.Children {
	// 	childTask, err := CreateTask(ctx, userID, projectID, &child)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }
	return task, nil
}

func (s *taskStore) CreateTask(ctx context.Context, projectID uuid.UUID, input *shared.CreateTaskBaseDTO) (*models.Task, error) {
	setter := models.Task{
		ProjectID: projectID,
		// UserID:      userID,
		CreatedBy:   input.CreatedBy,
		TeamID:      input.TeamID,
		Name:        input.Name,
		Description: input.Description,
		Status:      models.TaskStatus(input.Status),
		Order:       input.Order,
	}
	task, err := crudrepo.Task.PostOne(ctx, s.db, &setter)
	if err != nil {
		return nil, err
	}
	err = s.UpdateTaskProjectUpdateDate(ctx, task.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to update task project update date: %w", err)
	}
	return task, nil
}

// func DefineTaskOrderNumberByStatus(ctx context.Context, db db.s.db, taskId uuid.UUID, taskProjectId uuid.UUID, status models.TaskStatus, currentOrder float64, position int64) (float64, error) {
// 	if position == 0 {
// 		response, err := models.Tasks.Query(
// 			sm.Where(models.TaskColumns.ProjectID.EQ(psql.Arg(taskProjectId))),
// 			sm.Where(models.TaskColumns.Status.EQ(psql.Arg(status))),
// 			sm.OrderBy(models.TaskColumns.Order).Asc(),
// 			sm.Limit(1),
// 		).One(ctx, db)
// 		response, err = database.OptionalRow(response, err)
// 		if err != nil {
// 			return 0, err
// 		}
// 		if response == nil {
// 			return 0, nil
// 		}
// 		if response.ID == taskId {
// 			return response.Order, nil
// 		}
// 		return response.Order - 1000, nil
// 	}
// 	element, err := models.Tasks.Query(
// 		sm.Where(models.TaskColumns.ProjectID.EQ(psql.Arg(taskProjectId))),
// 		sm.OrderBy(models.TaskColumns.Order).Asc(),
// 		sm.Limit(1),
// 		sm.Offset(position),
// 	).One(ctx, db)
// 	element, err = database.OptionalRow(element, err)
// 	if err != nil {
// 		return 0, err
// 	}
// 	if element == nil {
// 		return 0, nil
// 	}
// 	if element.ID == taskId {
// 		return element.Order, nil
// 	}
// 	if currentOrder > element.Order {
// 		sideElements, err := models.Tasks.Query(
// 			sm.Where(models.TaskColumns.ProjectID.EQ(psql.Arg(taskProjectId))),
// 			sm.Where(models.TaskColumns.Status.EQ(psql.Arg(status))),
// 			sm.OrderBy(models.TaskColumns.Order).Asc(),
// 			sm.Limit(1),
// 			sm.Offset(position-1),
// 		).One(ctx, db)
// 		sideElements, err = database.OptionalRow(sideElements, err)
// 		if err != nil {
// 			return 0, err
// 		}
// 		if sideElements == nil {
// 			return element.Order - 1000, nil
// 		}
// 		return (element.Order + sideElements.Order) / 2, nil
// 	}
// 	sideElements, err := models.Tasks.Query(
// 		sm.Where(models.TaskColumns.ProjectID.EQ(psql.Arg(taskProjectId))),
// 		sm.Where(models.TaskColumns.Status.EQ(psql.Arg(status))),
// 		sm.OrderBy(models.TaskColumns.Order).Asc(),
// 		sm.Limit(1),
// 		sm.Offset(position+1),
// 	).One(ctx, db)
// 	sideElements, err = database.OptionalRow(sideElements, err)
// 	if err != nil {
// 		return 0, err
// 	}
// 	if sideElements == nil {
// 		return element.Order + 1000, nil
// 	}
// 	return (element.Order + sideElements.Order) / 2, nil

// }
func (s *taskStore) DefineTaskOrderNumberByStatus(ctx context.Context, taskId uuid.UUID, taskProjectId uuid.UUID, status models.TaskStatus, currentOrder float64, position int64) (float64, error) {
	if position == 0 {
		res, err := crudrepo.Task.Get(
			ctx,
			s.db,
			&map[string]any{
				"project_id": map[string]any{
					"_eq": taskProjectId.String(),
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
			return response.Order, nil
		}
		return response.Order - 1000, nil
	}
	ele, err := crudrepo.Task.Get(
		ctx,
		s.db,
		&map[string]any{
			"project_id": map[string]any{
				"_eq": taskProjectId.String(),
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
		return element.Order, nil
	}
	if currentOrder > element.Order {
		sideELe, err := crudrepo.Task.Get(
			ctx,
			s.db,
			&map[string]any{
				"project_id": map[string]any{
					"_eq": taskProjectId.String(),
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
			return element.Order - 1000, nil
		}
		sideElements := sideELe[0]
		return (element.Order + sideElements.Order) / 2, nil
	}
	sideele, err := crudrepo.Task.Get(
		ctx,
		s.db,
		&map[string]any{
			"project_id": map[string]any{
				"_eq": taskProjectId.String(),
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
		return element.Order + 1000, nil
	}
	sideElements := sideele[0]
	return (element.Order + sideElements.Order) / 2, nil
	// sideElements, err := models.Tasks.Query(
	// 	sm.Where(models.TaskColumns.ProjectID.EQ(psql.Arg(taskProjectId))),
	// 	sm.Where(models.TaskColumns.Status.EQ(psql.Arg(status))),
	// 	sm.OrderBy(models.TaskColumns.Order).Asc(),
	// 	sm.Limit(1),
	// 	sm.Offset(position+1),
	// ).One(ctx, db)
	// sideElements, err = database.OptionalRow(sideElements, err)
	// if err != nil {
	// 	return 0, err
	// }
	// if sideElements == nil {
	// 	return element.Order + 1000, nil
	// }
	// return (element.Order + sideElements.Order) / 2, nil

}

// func DefineTaskOrderNumber(ctx context.Context, db db.s.db, taskId uuid.UUID, taskProjectId uuid.UUID, currentOrder float64, position int64) (float64, error) {
// 	if position == 0 {
// 		response, err := models.Tasks.Query(
// 			sm.Where(models.TaskColumns.ProjectID.EQ(psql.Arg(taskProjectId))),
// 			sm.OrderBy(models.TaskColumns.Order).Asc(),
// 			sm.Limit(1),
// 		).One(ctx, db)
// 		response, err = database.OptionalRow(response, err)
// 		if err != nil {
// 			return 0, err
// 		}
// 		if response == nil {
// 			return 0, nil
// 		}
// 		if response.ID == taskId {
// 			return response.Order, nil
// 		}
// 		return response.Order - 1000, nil
// 	}
// 	element, err := models.Tasks.Query(
// 		sm.Where(models.TaskColumns.ProjectID.EQ(psql.Arg(taskProjectId))),
// 		sm.OrderBy(models.TaskColumns.Order).Asc(),
// 		sm.Limit(1),
// 		sm.Offset(position),
// 	).One(ctx, db)
// 	element, err = database.OptionalRow(element, err)
// 	if err != nil {
// 		return 0, err
// 	}
// 	if element == nil {
// 		return 0, nil
// 	}
// 	if element.ID == taskId {
// 		return element.Order, nil
// 	}
// 	if currentOrder > element.Order {
// 		sideElements, err := models.Tasks.Query(
// 			sm.Where(models.TaskColumns.ProjectID.EQ(psql.Arg(taskProjectId))),
// 			sm.OrderBy(models.TaskColumns.Order).Asc(),
// 			sm.Limit(1),
// 			sm.Offset(position-1),
// 		).One(ctx, db)
// 		sideElements, err = database.OptionalRow(sideElements, err)
// 		if err != nil {
// 			return 0, err
// 		}
// 		if sideElements == nil {
// 			return element.Order - 1000, nil
// 		}
// 		return (element.Order + sideElements.Order) / 2, nil
// 	}
// 	sideElements, err := models.Tasks.Query(
// 		sm.Where(models.TaskColumns.ProjectID.EQ(psql.Arg(taskProjectId))),
// 		sm.OrderBy(models.TaskColumns.Order).Asc(),
// 		sm.Limit(1),
// 		sm.Offset(position+1),
// 	).One(ctx, db)
// 	sideElements, err = database.OptionalRow(sideElements, err)
// 	if err != nil {
// 		return 0, err
// 	}
// 	if sideElements == nil {
// 		return element.Order + 1000, nil
// 	}
// 	return (element.Order + sideElements.Order) / 2, nil

// }

func (s *taskStore) UpdateTask(ctx context.Context, taskID uuid.UUID, input *shared.UpdateTaskBaseDTO) error {
	task, err := s.FindTaskByID(ctx, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("task not found")
	}
	// taskSetter := &crudModels.Task{
	// 	Name:        input.Name,
	// 	Description: input.Description,
	// 	Status:      input.Status,
	// 	Order:       input.Order,
	// 	ParentID:    input.ParentID,
	// }
	task.Name = input.Name
	task.Description = input.Description
	task.Status = models.TaskStatus(input.Status)
	task.Order = input.Order
	task.ParentID = input.ParentID
	_, err = crudrepo.Task.PutOne(ctx, s.db, task)
	if err != nil {
		return err
	}
	err = s.UpdateTaskProjectUpdateDate(ctx, task.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to update task project update date: %w", err)
	}
	return nil
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
	taskProject.Order = input.Order
	_, err = crudrepo.TaskProject.PutOne(ctx, s.db, taskProject)
	if err != nil {
		return err
	}
	return nil
}

// func UpdateTaskPosition(ctx context.Context, db db.s.db, taskID uuid.UUID, position int64) error {
// 	task, err := FindTaskByID(ctx, taskID)
// 	if err != nil {
// 		return err
// 	}
// 	if task == nil {
// 		return errors.New("task not found")
// 	}

// 	order, err := DefineTaskOrderNumber(ctx, task.ID, task.ProjectID, task.Order, position)
// 	if err != nil {
// 		return err
// 	}
// 	task.Order = order
// 	_, err = repository.Task.PutOne(ctx, task)
// 	if err != nil {
// 		return err
// 	}
// 	err = UpdateTaskProjectUpdateDate(ctx, task.ProjectID)
// 	if err != nil {
// 		return fmt.Errorf("failed to update task project update date: %w", err)
// 	}
// 	return nil
// }

func (s *taskStore) UpdateTaskPositionStatus(ctx context.Context, taskID uuid.UUID, position int64, status models.TaskStatus) error {
	task, err := s.FindTaskByID(ctx, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("task not found")
	}
	order, err := s.DefineTaskOrderNumberByStatus(ctx, task.ID, task.ProjectID, status, task.Order, position)
	if err != nil {
		return err
	}
	task.Order = order
	_, err = crudrepo.Task.PutOne(ctx, s.db, task)
	if err != nil {
		return err
	}
	err = s.UpdateTaskProjectUpdateDate(ctx, task.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to update task project update date: %w", err)
	}
	return nil
}
