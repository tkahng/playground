package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/tkahng/authgo/internal/crud/crudModels"
	"github.com/tkahng/authgo/internal/crud/repository"
)

type TaskService struct {
	task        *repository.PostgresRepository[crudModels.Task]
	taskProject *repository.PostgresRepository[crudModels.TaskProject]
}

func (service *TaskService) FindTaskByID(ctx context.Context, db repository.DBTX, id uuid.UUID) (*crudModels.Task, error) {
	task, err := service.task.GetOne(ctx, db, &map[string]any{
		"id": map[string]any{
			"_eq": id.String(),
		},
	})
	return repository.OptionalRow(task, err)
}

// func (service *TaskService) FindLastTaskOrder(ctx context.Context, db repository.DBTX, taskProjectID uuid.UUID) (float64, error) {
// 	task, err := models.Tasks.Query(
// 		sm.Where(models.TaskColumns.ProjectID.EQ(psql.Arg(taskProjectID))),
// 		sm.OrderBy(models.TaskColumns.Order).Desc(),
// 		sm.Limit(1),
// 	).One(ctx, db)
// 	task, err = repository.OptionalRow(task, err)
// 	if err != nil {
// 		return 0, err
// 	}
// 	if task == nil {
// 		return 0, nil
// 	}
// 	return task.Order + 1000, nil
// }

func (service *TaskService) DeleteTask(ctx context.Context, db repository.DBTX, taskID uuid.UUID) error {
	_, err := service.task.Delete(ctx, db, &map[string]any{
		"id": map[string]any{
			"_eq": taskID.String(),
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (service *TaskService) FindTaskProjectByID(ctx context.Context, db repository.DBTX, id uuid.UUID) (*crudModels.TaskProject, error) {
	task, err := service.taskProject.GetOne(ctx, db, &map[string]any{
		"id": map[string]any{
			"_eq": id.String(),
		},
	})
	return repository.OptionalRow(task, err)
}
func (service *TaskService) DeleteTaskProject(ctx context.Context, db repository.DBTX, taskProjectID uuid.UUID) error {
	_, err := service.taskProject.Delete(ctx, db, &map[string]any{
		"id": map[string]any{
			"_eq": taskProjectID.String(),
		},
	})
	if err != nil {
		return err
	}
	return nil
}

// func (service *TaskService) ListTasksOrderByFunc(ctx context.Context, q *psql.ViewQuery[*models.Task, models.TaskSlice], input *shared.TaskListParams) {
// 	if q == nil {
// 		return
// 	}
// 	if input == nil || input.SortBy == "" {
// 		q.Apply(
// 			sm.OrderBy(models.TaskColumns.CreatedAt).Desc(),
// 			sm.OrderBy(models.TaskColumns.ID).Desc(),
// 		)
// 		return
// 	}
// 	if slices.Contains(models.Tasks.Columns().Names(), input.SortBy) {
// 		if input.SortParams.SortOrder == "desc" {
// 			q.Apply(
// 				sm.OrderBy(psql.Quote(input.SortBy)).Desc(),
// 				sm.OrderBy(models.TaskColumns.ID).Desc(),
// 			)
// 		} else if input.SortParams.SortOrder == "asc" || input.SortParams.SortOrder == "" {
// 			q.Apply(
// 				sm.OrderBy(psql.Quote(input.SortBy)).Asc(),
// 				sm.OrderBy(models.TaskColumns.ID).Asc(),
// 			)
// 		}
// 		return
// 	}

// }

// func (service *TaskService) ListTasksFilterFunc(ctx context.Context, q *psql.ViewQuery[*models.Task, models.TaskSlice], filter *shared.TaskListFilter) {
// 	if filter == nil {
// 		return
// 	}
// 	if filter.Q != "" {
// 		q.Apply(
// 			psql.WhereOr(models.SelectWhere.Tasks.Name.ILike("%"+filter.Q+"%"),
// 				models.SelectWhere.Tasks.Description.ILike("%"+filter.Q+"%")),
// 		)
// 	}
// 	if len(filter.Status) > 0 {
// 		q.Apply(
// 			models.SelectWhere.Tasks.Status.In(filter.Status...),
// 		)
// 	}
// 	if len(filter.UserID) > 0 {
// 		id, err := uuid.Parse(filter.UserID)
// 		if err != nil {
// 			return
// 		}
// 		q.Apply(
// 			models.SelectWhere.Tasks.UserID.EQ(id),
// 		)
// 	}

// 	if filter.ProjectID != "" {
// 		id, err := uuid.Parse(filter.ProjectID)
// 		if err != nil {
// 			return
// 		}
// 		q.Apply(
// 			models.SelectWhere.Tasks.ProjectID.EQ(id),
// 		)
// 	}

// 	if len(filter.Ids) > 0 {
// 		var ids []uuid.UUID = ParseUUIDs(filter.Ids)
// 		q.Apply(
// 			models.SelectWhere.Tasks.ID.In(ids...),
// 		)
// 	}
// 	if filter.ParentID != "" {
// 		id, err := uuid.Parse(filter.ParentID)
// 		if err != nil {
// 			return
// 		}
// 		q.Apply(models.SelectWhere.Tasks.ParentID.EQ(id))
// 	}
// 	if filter.ParentStatus != "" {
// 		if filter.ParentStatus == "parent" {
// 			q.Apply(models.SelectWhere.Tasks.ParentID.IsNull())
// 		} else if filter.ParentStatus == "child" {
// 			q.Apply(models.SelectWhere.Tasks.ParentID.IsNotNull())
// 		}
// 	}
// }

// ListTasks implements AdminCrudActions.
// func (service *TaskService) ListTasks(ctx context.Context, db repository.DBTX, input *shared.TaskListParams) ([]*models.Task, error) {
// 	// q := models.Tasks.Query()
// 	// filter := input.TaskListFilter
// 	// pageInput := &input.PaginatedInput

// 	// ViewApplyPagination(q, pageInput)
// 	// ListTasksOrderByFunc(ctx, q, input)
// 	// ListTasksFilterFunc(ctx, q, &filter)
// 	data, err := q.All(ctx, db)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return data, nil
// }

// CountTasks implements AdminCrudActions.
//
//	func (service *TaskService) CountTasks(ctx context.Context, db repository.DBTX, filter *shared.TaskListFilter) (int64, error) {
//		q := models.Tasks.Query()
//		ListTasksFilterFunc(ctx, q, filter)
//		return CountExec(ctx, db, q)
//	}
// func (service *TaskService) ListTaskProjectsFilterFunc(ctx context.Context, q *psql.ViewQuery[*models.TaskProject, models.TaskProjectSlice], filter *shared.TaskProjectsListFilter) {
// 	if filter == nil {
// 		return
// 	}
// 	if filter.Q != "" {
// 		q.Apply(
// 			psql.WhereOr(models.SelectWhere.TaskProjects.Name.ILike("%"+filter.Q+"%"),
// 				models.SelectWhere.TaskProjects.Description.ILike("%"+filter.Q+"%")),
// 		)
// 	}

// 	if len(filter.Ids) > 0 {
// 		var ids []uuid.UUID = ParseUUIDs(filter.Ids)
// 		q.Apply(
// 			models.SelectWhere.TaskProjects.ID.In(ids...),
// 		)
// 	}

// 	if filter.UserID != "" {
// 		id, err := uuid.Parse(filter.UserID)
// 		if err != nil {
// 			return
// 		}
// 		q.Apply(
// 			models.SelectJoins.TaskProjects.InnerJoin.User(ctx),
// 			models.SelectWhere.Users.ID.EQ(id),
// 		)
// 	}
// }

// func (service *TaskService) ListTaskProjectsOrderByFunc(ctx context.Context, q *psql.ViewQuery[*models.TaskProject, models.TaskProjectSlice], input *shared.TaskProjectsListParams) {
// 	if q == nil {
// 		return
// 	}
// 	if input == nil || input.SortBy == "" {
// 		q.Apply(
// 			sm.OrderBy(models.TaskProjectColumns.CreatedAt).Desc(),
// 			sm.OrderBy(models.TaskProjectColumns.ID).Desc(),
// 		)
// 		return
// 	}
// 	if slices.Contains(models.TaskProjects.Columns().Names(), input.SortBy) {
// 		if input.SortParams.SortOrder == "desc" {
// 			q.Apply(
// 				sm.OrderBy(psql.Quote(input.SortBy)).Desc(),
// 				sm.OrderBy(models.TaskProjectColumns.ID).Desc(),
// 			)
// 		} else if input.SortParams.SortOrder == "asc" || input.SortParams.SortOrder == "" {
// 			q.Apply(
// 				sm.OrderBy(psql.Quote(input.SortBy)).Asc(),
// 				sm.OrderBy(models.TaskProjectColumns.ID).Asc(),
// 			)
// 		}
// 		return
// 	}
// }

// ListTaskProjects implements AdminCrudActions.
// func (service *TaskService) ListTaskProjects(ctx context.Context, db repository.DBTX, input *shared.TaskProjectsListParams) (models.TaskProjectSlice, error) {
// 	q := models.TaskProjects.Query()
// 	filter := input.TaskProjectsListFilter
// 	pageInput := &input.PaginatedInput

// 	ViewApplyPagination(q, pageInput)
// 	ListTaskProjectsOrderByFunc(ctx, q, input)
// 	ListTaskProjectsFilterFunc(ctx, q, &filter)
// 	data, err := q.All(ctx, db)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return data, nil
// }

// CountTaskProjects implements AdminCrudActions.
// func (service *TaskService) CountTaskProjects(ctx context.Context, db repository.DBTX, filter *shared.TaskProjectsListFilter) (int64, error) {
// 	q := models.TaskProjects.Query()
// 	// ListTaskProjectsFilterFunc(ctx, q, filter)
// 	data, err := q.Count(ctx, db)
// 	if err != nil {
// 		return 0, err
// 	}
// 	return data, nil
// }

// func (service *TaskService) CreateTaskProject(ctx context.Context, db repository.DBTX, userID uuid.UUID, input *shared.CreateTaskProjectDTO) (*models.TaskProject, error) {
// 	taskProject, err := service.taskProject.PostOne(ctx, &models.TaskProject{
// 		UserID:      userID,
// 		Name:        input.Name,
// 		Status:      models.TaskProjectStatus(input.Status),
// 		Order:       input.Order,
// 		Description: input.Description,
// 	})
// 	if err != nil {
// 		return nil, err
// 	}
// 	return taskProject, nil
// }

// func (service *TaskService) CreateTaskProjectWithTasks(ctx context.Context, db repository.DBTX, userID uuid.UUID, input *shared.CreateTaskProjectWithTasksDTO) (*models.TaskProject, error) {
// 	count, err := service.task.Count(ctx, nil)
// 	if err != nil {
// 		return nil, err
// 	}
// 	input.CreateTaskProjectDTO.Order = float64(count * 1000)
// 	taskProject, err := service.taskProject.PostOne(ctx, &models.TaskProject{
// 		UserID:      userID,
// 		Name:        input.CreateTaskProjectDTO.Name,
// 		Status:      models.TaskProjectStatus(input.CreateTaskProjectDTO.Status),
// 		Order:       input.CreateTaskProjectDTO.Order,
// 		Description: input.CreateTaskProjectDTO.Description,
// 	})
// 	if err != nil {
// 		return nil, err
// 	}
// 	var tasks []*models.Task
// 	for i, task := range input.Tasks {
// 		task.Order = float64(i * 1000)
// 		newTask := &models.Task{
// 			UserID:      userID,
// 			ProjectID:   taskProject.ID,
// 			Name:        task.Name,
// 			Description: task.Description,
// 			Status:      models.TaskStatus(task.Status),
// 			Order:       task.Order,
// 		}
// 		tasks = append(tasks, newTask)
// 	}
// 	_, err = service.task.Post(ctx, tasks)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return taskProject, nil
// }

// func (service *TaskService) CreateTaskWithChildren(ctx context.Context, db repository.DBTX, userID uuid.UUID, projectID uuid.UUID, input *shared.CreateTaskWithChildrenDTO) (*models.Task, error) {
// 	task, err := CreateTask(ctx, db, userID, projectID, &input.CreateTaskBaseDTO)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// for _, child := range input.Children {
// 	// 	childTask, err := CreateTask(ctx, db, userID, projectID, &child)
// 	// 	if err != nil {
// 	// 		return nil, err
// 	// 	}
// 	// }
// 	return task, nil
// }

// func (service *TaskService) CreateTask(ctx context.Context, db repository.DBTX, userID uuid.UUID, projectID uuid.UUID, input *shared.CreateTaskBaseDTO) (*models.Task, error) {
// 	setter := models.TaskSetter{
// 		ProjectID:   omit.From(projectID),
// 		UserID:      omit.From(userID),
// 		Name:        omit.From(input.Name),
// 		Description: omitnull.FromPtr(input.Description),
// 		Status:      omit.From(input.Status),
// 		Order:       omit.From(input.Order),
// 	}
// 	task, err := models.Tasks.Insert(&setter, im.Returning("*")).One(ctx, db)
// 	if err != nil {
// 		return nil, err
// 	}
// 	err = UpdateTaskProjectUpdateDate(ctx, db, task.ProjectID)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to update task project update date: %w", err)
// 	}
// 	return task, nil
// }

// func (service *TaskService) DefineTaskOrderNumberByStatus(ctx context.Context, db repository.DBTX, taskId uuid.UUID, taskProjectId uuid.UUID, status models.TaskStatus, currentOrder float64, position int64) (float64, error) {
// 	if position == 0 {
// 		response, err := models.Tasks.Query(
// 			sm.Where(models.TaskColumns.ProjectID.EQ(psql.Arg(taskProjectId))),
// 			sm.Where(models.TaskColumns.Status.EQ(psql.Arg(status))),
// 			sm.OrderBy(models.TaskColumns.Order).Asc(),
// 			sm.Limit(1),
// 		).One(ctx, db)
// 		response, err = repository.OptionalRow(response, err)
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
// 	element, err = repository.OptionalRow(element, err)
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
// 		sideElements, err = repository.OptionalRow(sideElements, err)
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
// 	sideElements, err = repository.OptionalRow(sideElements, err)
// 	if err != nil {
// 		return 0, err
// 	}
// 	if sideElements == nil {
// 		return element.Order + 1000, nil
// 	}
// 	return (element.Order + sideElements.Order) / 2, nil

// }
// func (service *TaskService) DefineTaskOrderNumberByStatusCrud(ctx context.Context, repo crud.Repository[crudModels.Task], taskId uuid.UUID, taskProjectId uuid.UUID, status models.TaskStatus, currentOrder float64, position int64) (float64, error) {
// 	if position == 0 {
// 		res, err := repo.Get(
// 			ctx,
// 			&map[string]any{
// 				"project_id": taskProjectId,
// 				"status":     status,
// 			},
// 			&map[string]any{
// 				"order": "ASC",
// 			},
// 			types.Pointer(1),
// 			nil,
// 		)
// 		if err != nil {
// 			return 0, err
// 		}
// 		if len(res) == 0 {
// 			return 0, nil
// 		}
// 		response := res[0]

// 		if response.ID == taskId {
// 			return response.Order, nil
// 		}
// 		return response.Order - 1000, nil
// 	}
// 	ele, err := repo.Get(
// 		ctx,
// 		&map[string]any{
// 			"project_id": taskProjectId,
// 		},
// 		&map[string]any{
// 			"order": "ASC",
// 		},
// 		types.Pointer(1),
// 		types.Pointer(int(position)),
// 	)
// 	if err != nil {
// 		return 0, err
// 	}
// 	if len(ele) == 0 {
// 		return 0, nil
// 	}
// 	element := ele[0]

// 	if element.ID == taskId {
// 		return element.Order, nil
// 	}
// 	if currentOrder > element.Order {
// 		sideELe, err := repo.Get(
// 			ctx,
// 			&map[string]any{
// 				"project_id": taskProjectId,
// 				"status":     status,
// 			},
// 			&map[string]any{
// 				"order": "ASC",
// 			},
// 			types.Pointer(1),
// 			types.Pointer(int(position-1)),
// 		)
// 		if err != nil {
// 			return 0, err
// 		}
// 		if len(sideELe) == 0 {
// 			return element.Order - 1000, nil
// 		}
// 		sideElements := sideELe[0]
// 		return (element.Order + sideElements.Order) / 2, nil
// 	}
// 	sideele, err := repo.Get(
// 		ctx,
// 		&map[string]any{
// 			"project_id": taskProjectId,
// 			"status":     status,
// 		},
// 		&map[string]any{
// 			"order": "ASC",
// 		},
// 		types.Pointer(1),
// 		types.Pointer(int(position+1)),
// 	)
// 	if err != nil {
// 		return 0, err
// 	}
// 	if len(sideele) == 0 {
// 		return element.Order + 1000, nil
// 	}
// 	sideElements := sideele[0]
// 	return (element.Order + sideElements.Order) / 2, nil
// 	// sideElements, err := models.Tasks.Query(
// 	// 	sm.Where(models.TaskColumns.ProjectID.EQ(psql.Arg(taskProjectId))),
// 	// 	sm.Where(models.TaskColumns.Status.EQ(psql.Arg(status))),
// 	// 	sm.OrderBy(models.TaskColumns.Order).Asc(),
// 	// 	sm.Limit(1),
// 	// 	sm.Offset(position+1),
// 	// ).One(ctx, db)
// 	// sideElements, err = repository.OptionalRow(sideElements, err)
// 	// if err != nil {
// 	// 	return 0, err
// 	// }
// 	// if sideElements == nil {
// 	// 	return element.Order + 1000, nil
// 	// }
// 	// return (element.Order + sideElements.Order) / 2, nil

// }

// func (service *TaskService) DefineTaskOrderNumber(ctx context.Context, db repository.DBTX, taskId uuid.UUID, taskProjectId uuid.UUID, currentOrder float64, position int64) (float64, error) {
// 	if position == 0 {
// 		response, err := models.Tasks.Query(
// 			sm.Where(models.TaskColumns.ProjectID.EQ(psql.Arg(taskProjectId))),
// 			sm.OrderBy(models.TaskColumns.Order).Asc(),
// 			sm.Limit(1),
// 		).One(ctx, db)
// 		response, err = repository.OptionalRow(response, err)
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
// 	element, err = repository.OptionalRow(element, err)
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
// 		sideElements, err = repository.OptionalRow(sideElements, err)
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
// 	sideElements, err = repository.OptionalRow(sideElements, err)
// 	if err != nil {
// 		return 0, err
// 	}
// 	if sideElements == nil {
// 		return element.Order + 1000, nil
// 	}
// 	return (element.Order + sideElements.Order) / 2, nil

// }

// func (service *TaskService) UpdateTask(ctx context.Context, db repository.DBTX, taskID uuid.UUID, input *shared.UpdateTaskBaseDTO) error {
// 	task, err := FindTaskByID(ctx, db, taskID)
// 	if err != nil {
// 		return err
// 	}
// 	if task == nil {
// 		return errors.New("task not found")
// 	}
// 	taskSetter := &models.TaskSetter{
// 		Name:        omit.From(input.Name),
// 		Description: omitnull.FromPtr(input.Description),
// 		Status:      omit.From(input.Status),
// 		Order:       omit.From(input.Order),
// 		ParentID:    omitnull.FromPtr(input.ParentID),
// 	}
// 	err = task.Update(ctx, db, taskSetter)
// 	if err != nil {
// 		return err
// 	}
// 	err = UpdateTaskProjectUpdateDate(ctx, db, task.ProjectID)
// 	if err != nil {
// 		return fmt.Errorf("failed to update task project update date: %w", err)
// 	}
// 	return nil
// }

// func (service *TaskService) UpdateTaskProjectUpdateDate(ctx context.Context, db repository.DBTX, taskProjectID uuid.UUID) error {
// 	q := models.TaskProjects.Update(
// 		models.UpdateWhere.TaskProjects.ID.EQ(taskProjectID),
// 		models.TaskProjectSetter{
// 			UpdatedAt: omit.From(time.Now()),
// 		}.UpdateMod(),
// 	)
// 	_, err := q.Exec(ctx, db)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (service *TaskService) UpdateTaskProject(ctx context.Context, db repository.DBTX, taskProjectID uuid.UUID, input *shared.UpdateTaskProjectBaseDTO) error {
// 	taskProject, err := FindTaskProjectByID(ctx, db, taskProjectID)
// 	if err != nil {
// 		return err
// 	}
// 	taskProjectSetter := &models.TaskProjectSetter{
// 		Name:        omit.From(input.Name),
// 		Description: omitnull.FromPtr(input.Description),
// 		Status:      omit.From(input.Status),
// 		Order:       omit.From(input.Order),
// 	}
// 	err = taskProject.Update(ctx, db, taskProjectSetter)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (service *TaskService) UpdateTaskPosition(ctx context.Context, db repository.DBTX, taskID uuid.UUID, position int64) error {
// 	task, err := FindTaskByID(ctx, db, taskID)
// 	if err != nil {
// 		return err
// 	}
// 	if task == nil {
// 		return errors.New("task not found")
// 	}
// 	taskSetter := &models.TaskSetter{}

// 	order, err := DefineTaskOrderNumber(ctx, db, task.ID, task.ProjectID, task.Order, position)
// 	if err != nil {
// 		return err
// 	}
// 	taskSetter.Order = omit.From(order)
// 	err = task.Update(ctx, db, taskSetter)
// 	if err != nil {
// 		return err
// 	}
// 	err = UpdateTaskProjectUpdateDate(ctx, db, task.ProjectID)
// 	if err != nil {
// 		return fmt.Errorf("failed to update task project update date: %w", err)
// 	}
// 	return nil
// }

// func (service *TaskService) UpdateTaskPositionStatus(ctx context.Context, db repository.DBTX, taskID uuid.UUID, position int64, status models.TaskStatus) error {
// 	task, err := FindTaskByID(ctx, db, taskID)
// 	if err != nil {
// 		return err
// 	}
// 	if task == nil {
// 		return errors.New("task not found")
// 	}
// 	taskSetter := &models.TaskSetter{
// 		Status: omit.From(status),
// 	}
// 	order, err := DefineTaskOrderNumberByStatusCrud(ctx, db, task.ID, task.ProjectID, status, task.Order, position)
// 	if err != nil {
// 		return err
// 	}
// 	taskSetter.Order = omit.From(order)
// 	err = task.Update(ctx, db, taskSetter)
// 	if err != nil {
// 		return err
// 	}
// 	err = UpdateTaskProjectUpdateDate(ctx, db, task.ProjectID)
// 	if err != nil {
// 		return fmt.Errorf("failed to update task project update date: %w", err)
// 	}
// 	return nil
// }
