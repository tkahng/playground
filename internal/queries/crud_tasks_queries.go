package queries

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/repository"

	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
	"github.com/tkahng/authgo/internal/types"
)

const (
	LoadTaskProjectsTasksQuery = `
SELECT tp.id as key,
        json_agg(to_json(t.*)) AS "data"
FROM public.task_projects tp
        LEFT JOIN public.tasks t ON tp.id = t.project_id
WHERE tp.id = ANY ($1::uuid [])
GROUP BY tp.id;`
)

func LoadTaskProjectsTasks(ctx context.Context, db Queryer, projectIds ...uuid.UUID) ([][]*models.Task, error) {
	tasks, err := repository.Task.Get(
		ctx,
		db,
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

func FindTaskByID(ctx context.Context, db Queryer, id uuid.UUID) (*models.Task, error) {
	task, err := repository.Task.GetOne(
		ctx,
		db,
		&map[string]any{
			"id": map[string]any{
				"_eq": id.String(),
			},
		},
	)
	return OptionalRow(task, err)
}

func FindLastTaskOrder(ctx context.Context, db Queryer, taskProjectID uuid.UUID) (float64, error) {
	// task, err := models.Tasks.Query(
	// 	sm.Where(models.TaskColumns.ProjectID.EQ(psql.Arg(taskProjectID))),
	// 	sm.OrderBy(models.TaskColumns.Order).Desc(),
	// 	sm.Limit(1),
	// ).One(ctx, db)
	tasks, err := repository.Task.Get(
		ctx,
		db,
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

func DeleteTask(ctx context.Context, db Queryer, taskID uuid.UUID) error {
	// task, err := models.FindTask(ctx, db, taskID)
	// if err != nil {
	// 	return err
	// }
	// return task.Delete(ctx, db)
	_, err := repository.Task.DeleteReturn(
		ctx,
		db,
		&map[string]any{
			"id": map[string]any{
				"_eq": taskID.String(),
			},
		},
	)
	return err
}

func FindTaskProjectByID(ctx context.Context, db Queryer, id uuid.UUID) (*models.TaskProject, error) {
	// task, err := models.FindTaskProject(ctx, db, id)
	// return OptionalRow(task, err)
	task, err := repository.TaskProject.GetOne(
		ctx,
		db,
		&map[string]any{
			"id": map[string]any{
				"_eq": id.String(),
			},
		},
	)
	return OptionalRow(task, err)
}
func DeleteTaskProject(ctx context.Context, db Queryer, taskProjectID uuid.UUID) error {
	// taskProject, err := models.FindTaskProject(ctx, db, taskProjectID)
	// if err != nil {
	// 	return err
	// }
	// return taskProject.Delete(ctx, db)
	_, err := repository.TaskProject.DeleteReturn(
		ctx,
		db,
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
		// q.Apply(
		// 	psql.WhereOr(models.SelectWhere.Tasks.Name.ILike("%"+filter.Q+"%"),
		// 		models.SelectWhere.Tasks.Description.ILike("%"+filter.Q+"%")),
		// )
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
		// q.Apply(
		// 	models.SelectWhere.Tasks.Status.In(filter.Status...),
		// )
		where["status"] = map[string]any{
			"_in": filter.Status,
		}
	}
	if len(filter.UserID) > 0 {
		// id, err := uuid.Parse(filter.UserID)
		// if err != nil {
		// 	return
		// }
		// q.Apply(
		// 	models.SelectWhere.Tasks.UserID.EQ(id),
		// )
		where["user_id"] = map[string]any{
			"_eq": filter.UserID,
		}

	}

	if filter.ProjectID != "" {
		// id, err := uuid.Parse(filter.ProjectID)
		// if err != nil {
		// 	return
		// }
		// q.Apply(
		// 	models.SelectWhere.Tasks.ProjectID.EQ(id),
		// )
		where["project_id"] = map[string]any{
			"_eq": filter.ProjectID,
		}
	}

	if len(filter.Ids) > 0 {
		// var ids []uuid.UUID = ParseUUIDs(filter.Ids)
		// q.Apply(
		// 	models.SelectWhere.Tasks.ID.In(ids...),
		// )
		where["id"] = map[string]any{
			"_in": filter.Ids,
		}
	}
	if filter.ParentID != "" {
		// id, err := uuid.Parse(filter.ParentID)
		// if err != nil {
		// 	return
		// }
		// q.Apply(models.SelectWhere.Tasks.ParentID.EQ(id))
		where["parent_id"] = map[string]any{
			"_eq": filter.ParentID,
		}
	}
	if filter.ParentStatus != "" {
		if filter.ParentStatus == "parent" {
			// q.Apply(models.SelectWhere.Tasks.ParentID.IsNull())
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
func ListTasks(ctx context.Context, db Queryer, input *shared.TaskListParams) ([]*models.Task, error) {
	filter := input.TaskListFilter
	pageInput := &input.PaginatedInput

	iimit, offset := PaginateRepo(pageInput)
	order := ListTasksOrderByFunc(input)
	where := ListTasksFilterFunc(&filter)
	data, err := repository.Task.Get(
		ctx,
		db,
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
func CountTasks(ctx context.Context, db Queryer, filter *shared.TaskListFilter) (int64, error) {
	where := ListTasksFilterFunc(filter)
	return repository.Task.Count(ctx, db, where)
}
func ListTaskProjectsFilterFunc(filter *shared.TaskProjectsListFilter) *map[string]any {
	if filter == nil {
		return nil
	}
	where := make(map[string]any)
	if filter.Q != "" {
		// q.Apply(
		// 	psql.WhereOr(models.SelectWhere.TaskProjects.Name.ILike("%"+filter.Q+"%"),
		// 		models.SelectWhere.TaskProjects.Description.ILike("%"+filter.Q+"%")),
		// )
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
		// var ids []uuid.UUID = ParseUUIDs(filter.Ids)
		// q.Apply(
		// 	models.SelectWhere.TaskProjects.ID.In(ids...),
		// )
		where["id"] = map[string]any{
			"_in": filter.Ids,
		}
	}

	if filter.UserID != "" {
		// id, err := uuid.Parse(filter.UserID)
		// if err != nil {
		// 	return
		// }
		// q.Apply(
		// 	models.SelectJoins.TaskProjects.InnerJoin.User(ctx),
		// 	models.SelectWhere.Users.ID.EQ(id),
		// )
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
	// if slices.Contains(models.TaskProjects.Columns().Names(), input.SortBy) {
	order[input.SortBy] = strings.ToUpper(input.SortOrder)
	// }
	return &order
}

// ListTaskProjects implements AdminCrudActions.
func ListTaskProjects(ctx context.Context, db Queryer, input *shared.TaskProjectsListParams) ([]*models.TaskProject, error) {
	filter := input.TaskProjectsListFilter
	pageInput := &input.PaginatedInput

	limit, offset := PaginateRepo(pageInput)
	oredr := ListTaskProjectsOrderByFunc(input)
	where := ListTaskProjectsFilterFunc(&filter)
	data, err := repository.TaskProject.Get(
		ctx,
		db,
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
func CountTaskProjects(ctx context.Context, db Queryer, filter *shared.TaskProjectsListFilter) (int64, error) {
	where := ListTaskProjectsFilterFunc(filter)
	return repository.TaskProject.Count(ctx, db, where)
}

func CreateTaskProject(ctx context.Context, db Queryer, userID uuid.UUID, input *shared.CreateTaskProjectDTO) (*models.TaskProject, error) {
	taskProject := models.TaskProject{
		UserID:      userID,
		Name:        input.Name,
		Description: input.Description,
		Status:      models.TaskProjectStatus(input.Status),
		Order:       input.Order,
	}
	projects, err := repository.TaskProject.PostOne(ctx, db, &taskProject)
	if err != nil {
		return nil, err
	}
	return projects, nil
}

func CreateTaskProjectWithTasks(ctx context.Context, db Queryer, userID uuid.UUID, input *shared.CreateTaskProjectWithTasksDTO) (*models.TaskProject, error) {
	count, err := CountTaskProjects(ctx, db, nil)
	if err != nil {
		return nil, err
	}
	input.CreateTaskProjectDTO.Order = float64(count * 1000)
	taskProject, err := CreateTaskProject(ctx, db, userID, &input.CreateTaskProjectDTO)
	if err != nil {
		return nil, err
	}
	var tasks []*models.Task
	for i, task := range input.Tasks {
		task.Order = float64(i * 1000)
		newTask, err := CreateTask(ctx, db, userID, taskProject.ID, &task)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, newTask)
	}
	return taskProject, nil
}

func CreateTaskWithChildren(ctx context.Context, db Queryer, userID uuid.UUID, projectID uuid.UUID, input *shared.CreateTaskWithChildrenDTO) (*models.Task, error) {
	task, err := CreateTask(ctx, db, userID, projectID, &input.CreateTaskBaseDTO)
	if err != nil {
		return nil, err
	}
	// for _, child := range input.Children {
	// 	childTask, err := CreateTask(ctx, db, userID, projectID, &child)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }
	return task, nil
}

func CreateTask(ctx context.Context, db Queryer, userID uuid.UUID, projectID uuid.UUID, input *shared.CreateTaskBaseDTO) (*models.Task, error) {
	setter := models.Task{
		ProjectID:   projectID,
		UserID:      userID,
		Name:        input.Name,
		Description: input.Description,
		Status:      models.TaskStatus(input.Status),
		Order:       input.Order,
	}
	task, err := repository.Task.PostOne(ctx, db, &setter)
	if err != nil {
		return nil, err
	}
	err = UpdateTaskProjectUpdateDate(ctx, db, task.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to update task project update date: %w", err)
	}
	return task, nil
}

// func DefineTaskOrderNumberByStatus(ctx context.Context, db Queryer, taskId uuid.UUID, taskProjectId uuid.UUID, status models.TaskStatus, currentOrder float64, position int64) (float64, error) {
// 	if position == 0 {
// 		response, err := models.Tasks.Query(
// 			sm.Where(models.TaskColumns.ProjectID.EQ(psql.Arg(taskProjectId))),
// 			sm.Where(models.TaskColumns.Status.EQ(psql.Arg(status))),
// 			sm.OrderBy(models.TaskColumns.Order).Asc(),
// 			sm.Limit(1),
// 		).One(ctx, db)
// 		response, err = OptionalRow(response, err)
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
// 	element, err = OptionalRow(element, err)
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
// 		sideElements, err = OptionalRow(sideElements, err)
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
// 	sideElements, err = OptionalRow(sideElements, err)
// 	if err != nil {
// 		return 0, err
// 	}
// 	if sideElements == nil {
// 		return element.Order + 1000, nil
// 	}
// 	return (element.Order + sideElements.Order) / 2, nil

// }
func DefineTaskOrderNumberByStatus(ctx context.Context, dbx Queryer, taskId uuid.UUID, taskProjectId uuid.UUID, status models.TaskStatus, currentOrder float64, position int64) (float64, error) {
	if position == 0 {
		res, err := repository.Task.Get(
			ctx,
			nil,
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
	ele, err := repository.Task.Get(
		ctx,
		nil,
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
		sideELe, err := repository.Task.Get(
			ctx,
			nil,
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
	sideele, err := repository.Task.Get(
		ctx,
		nil,
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
	// sideElements, err = OptionalRow(sideElements, err)
	// if err != nil {
	// 	return 0, err
	// }
	// if sideElements == nil {
	// 	return element.Order + 1000, nil
	// }
	// return (element.Order + sideElements.Order) / 2, nil

}

// func DefineTaskOrderNumber(ctx context.Context, db Queryer, taskId uuid.UUID, taskProjectId uuid.UUID, currentOrder float64, position int64) (float64, error) {
// 	if position == 0 {
// 		response, err := models.Tasks.Query(
// 			sm.Where(models.TaskColumns.ProjectID.EQ(psql.Arg(taskProjectId))),
// 			sm.OrderBy(models.TaskColumns.Order).Asc(),
// 			sm.Limit(1),
// 		).One(ctx, db)
// 		response, err = OptionalRow(response, err)
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
// 	element, err = OptionalRow(element, err)
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
// 		sideElements, err = OptionalRow(sideElements, err)
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
// 	sideElements, err = OptionalRow(sideElements, err)
// 	if err != nil {
// 		return 0, err
// 	}
// 	if sideElements == nil {
// 		return element.Order + 1000, nil
// 	}
// 	return (element.Order + sideElements.Order) / 2, nil

// }

func UpdateTask(ctx context.Context, db Queryer, taskID uuid.UUID, input *shared.UpdateTaskBaseDTO) error {
	task, err := FindTaskByID(ctx, db, taskID)
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
	_, err = repository.Task.PutOne(ctx, db, task)
	if err != nil {
		return err
	}
	err = UpdateTaskProjectUpdateDate(ctx, db, task.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to update task project update date: %w", err)
	}
	return nil
}

func UpdateTaskProjectUpdateDate(ctx context.Context, db Queryer, taskProjectID uuid.UUID) error {
	q := squirrel.Update("task_projects").
		Where("id = ?", taskProjectID).
		Set("updated_at", time.Now())

	err := ExecWithBuilder(ctx, db, q.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return err
	}
	return nil
}

func UpdateTaskProject(ctx context.Context, db Queryer, taskProjectID uuid.UUID, input *shared.UpdateTaskProjectBaseDTO) error {
	taskProject, err := FindTaskProjectByID(ctx, db, taskProjectID)
	if err != nil {
		return err
	}
	taskProject.Name = input.Name
	taskProject.Description = input.Description
	taskProject.Status = models.TaskProjectStatus(input.Status)
	taskProject.Order = input.Order
	_, err = repository.TaskProject.PutOne(ctx, db, taskProject)
	if err != nil {
		return err
	}
	return nil
}

// func UpdateTaskPosition(ctx context.Context, db Queryer, taskID uuid.UUID, position int64) error {
// 	task, err := FindTaskByID(ctx, db, taskID)
// 	if err != nil {
// 		return err
// 	}
// 	if task == nil {
// 		return errors.New("task not found")
// 	}

// 	order, err := DefineTaskOrderNumber(ctx, db, task.ID, task.ProjectID, task.Order, position)
// 	if err != nil {
// 		return err
// 	}
// 	task.Order = order
// 	_, err = repository.Task.PutOne(ctx, db, task)
// 	if err != nil {
// 		return err
// 	}
// 	err = UpdateTaskProjectUpdateDate(ctx, db, task.ProjectID)
// 	if err != nil {
// 		return fmt.Errorf("failed to update task project update date: %w", err)
// 	}
// 	return nil
// }

func UpdateTaskPositionStatus(ctx context.Context, db Queryer, taskID uuid.UUID, position int64, status models.TaskStatus) error {
	task, err := FindTaskByID(ctx, db, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("task not found")
	}
	order, err := DefineTaskOrderNumberByStatus(ctx, db, task.ID, task.ProjectID, status, task.Order, position)
	if err != nil {
		return err
	}
	task.Order = order
	_, err = repository.Task.PutOne(ctx, db, task)
	if err != nil {
		return err
	}
	err = UpdateTaskProjectUpdateDate(ctx, db, task.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to update task project update date: %w", err)
	}
	return nil
}
