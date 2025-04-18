package repository

import (
	"context"
	"errors"
	"slices"

	"github.com/aarondl/opt/omit"
	"github.com/aarondl/opt/omitnull"
	"github.com/google/uuid"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/im"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/shared"
)

func FindTaskByID(ctx context.Context, db bob.Executor, id uuid.UUID) (*models.Task, error) {
	task, err := models.FindTask(ctx, db, id)
	return OptionalRow(task, err)
}

func DeleteTask(ctx context.Context, db bob.Executor, taskID uuid.UUID) error {
	task, err := models.FindTask(ctx, db, taskID)
	if err != nil {
		return err
	}
	return task.Delete(ctx, db)
}

func FindTaskProjectByID(ctx context.Context, db bob.Executor, id uuid.UUID) (*models.TaskProject, error) {
	task, err := models.FindTaskProject(ctx, db, id)
	return OptionalRow(task, err)
}
func DeleteTaskProject(ctx context.Context, db bob.Executor, taskProjectID uuid.UUID) error {
	taskProject, err := models.FindTaskProject(ctx, db, taskProjectID)
	if err != nil {
		return err
	}
	return taskProject.Delete(ctx, db)
}
func ListTasksOrderByFunc(ctx context.Context, q *psql.ViewQuery[*models.Task, models.TaskSlice], input *shared.TaskListParams) {
	if q == nil {
		return
	}
	if input == nil || input.SortBy == "" {
		q.Apply(
			sm.OrderBy(models.TaskColumns.CreatedAt).Desc(),
			sm.OrderBy(models.TaskColumns.ID).Desc(),
		)
		return
	}
	if slices.Contains(models.Tasks.Columns().Names(), input.SortBy) {
		if input.SortParams.SortOrder == "desc" {
			q.Apply(
				sm.OrderBy(psql.Quote(input.SortBy)).Desc(),
				sm.OrderBy(models.TaskColumns.ID).Desc(),
			)
		} else if input.SortParams.SortOrder == "asc" || input.SortParams.SortOrder == "" {
			q.Apply(
				sm.OrderBy(psql.Quote(input.SortBy)).Asc(),
				sm.OrderBy(models.TaskColumns.ID).Asc(),
			)
		}
		return
	}

}

func ListTasksFilterFunc(ctx context.Context, q *psql.ViewQuery[*models.Task, models.TaskSlice], filter *shared.TaskListFilter) {
	if filter == nil {
		return
	}
	if filter.Q != "" {
		q.Apply(
			psql.WhereOr(models.SelectWhere.Tasks.Name.ILike("%"+filter.Q+"%"),
				models.SelectWhere.Tasks.Description.ILike("%"+filter.Q+"%")),
		)
	}
	if len(filter.Status) > 0 {
		q.Apply(
			models.SelectWhere.Tasks.Status.In(filter.Status...),
		)
	}
	if len(filter.UserID) > 0 {
		id, err := uuid.Parse(filter.UserID)
		if err != nil {
			return
		}
		q.Apply(
			models.SelectWhere.Tasks.UserID.EQ(id),
		)
	}

	if filter.ProjectID != "" {
		id, err := uuid.Parse(filter.ProjectID)
		if err != nil {
			return
		}
		q.Apply(
			models.SelectWhere.Tasks.ProjectID.EQ(id),
		)
	}

	if len(filter.Ids) > 0 {
		var ids []uuid.UUID = ParseUUIDs(filter.Ids)
		q.Apply(
			models.SelectWhere.Tasks.ID.In(ids...),
		)
	}
	if filter.ParentID != "" {
		id, err := uuid.Parse(filter.ParentID)
		if err != nil {
			return
		}
		q.Apply(models.SelectWhere.Tasks.ParentID.EQ(id))
	}
	if filter.ParentStatus != "" {
		if filter.ParentStatus == "parent" {
			q.Apply(models.SelectWhere.Tasks.ParentID.IsNull())
		} else if filter.ParentStatus == "child" {
			q.Apply(models.SelectWhere.Tasks.ParentID.IsNotNull())
		}
	}
}

// ListTasks implements AdminCrudActions.
func ListTasks(ctx context.Context, db bob.Executor, input *shared.TaskListParams) ([]*models.Task, error) {
	q := models.Tasks.Query()
	filter := input.TaskListFilter
	pageInput := &input.PaginatedInput

	ViewApplyPagination(q, pageInput)
	ListTasksOrderByFunc(ctx, q, input)
	ListTasksFilterFunc(ctx, q, &filter)
	data, err := q.All(ctx, db)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// CountTasks implements AdminCrudActions.
func CountTasks(ctx context.Context, db bob.Executor, filter *shared.TaskListFilter) (int64, error) {
	q := models.Tasks.Query()
	ListTasksFilterFunc(ctx, q, filter)
	return CountExec(ctx, db, q)
}
func ListTaskProjectsFilterFunc(ctx context.Context, q *psql.ViewQuery[*models.TaskProject, models.TaskProjectSlice], filter *shared.TaskProjectsListFilter) {
	if filter == nil {
		return
	}
	if filter.Q != "" {
		q.Apply(
			psql.WhereOr(models.SelectWhere.TaskProjects.Name.ILike("%"+filter.Q+"%"),
				models.SelectWhere.TaskProjects.Description.ILike("%"+filter.Q+"%")),
		)
	}

	if len(filter.Ids) > 0 {
		var ids []uuid.UUID = ParseUUIDs(filter.Ids)
		q.Apply(
			models.SelectWhere.TaskProjects.ID.In(ids...),
		)
	}

	if filter.UserID != "" {
		id, err := uuid.Parse(filter.UserID)
		if err != nil {
			return
		}
		q.Apply(
			models.SelectJoins.TaskProjects.InnerJoin.User(ctx),
			models.SelectWhere.Users.ID.EQ(id),
		)
	}
}

func ListTaskProjectsOrderByFunc(ctx context.Context, q *psql.ViewQuery[*models.TaskProject, models.TaskProjectSlice], input *shared.TaskProjectsListParams) {
	if q == nil {
		return
	}
	if input == nil || input.SortBy == "" {
		q.Apply(
			sm.OrderBy(models.TaskProjectColumns.CreatedAt).Desc(),
			sm.OrderBy(models.TaskProjectColumns.ID).Desc(),
		)
		return
	}
	if slices.Contains(models.TaskProjects.Columns().Names(), input.SortBy) {
		if input.SortParams.SortOrder == "desc" {
			q.Apply(
				sm.OrderBy(psql.Quote(input.SortBy)).Desc(),
				sm.OrderBy(models.TaskProjectColumns.ID).Desc(),
			)
		} else if input.SortParams.SortOrder == "asc" || input.SortParams.SortOrder == "" {
			q.Apply(
				sm.OrderBy(psql.Quote(input.SortBy)).Asc(),
				sm.OrderBy(models.TaskProjectColumns.ID).Asc(),
			)
		}
		return
	}
}

// ListTaskProjects implements AdminCrudActions.
func ListTaskProjects(ctx context.Context, db bob.Executor, input *shared.TaskProjectsListParams) (models.TaskProjectSlice, error) {
	q := models.TaskProjects.Query()
	filter := input.TaskProjectsListFilter
	pageInput := &input.PaginatedInput

	ViewApplyPagination(q, pageInput)
	ListTaskProjectsOrderByFunc(ctx, q, input)
	ListTaskProjectsFilterFunc(ctx, q, &filter)
	data, err := q.All(ctx, db)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// CountTaskProjects implements AdminCrudActions.
func CountTaskProjects(ctx context.Context, db bob.Executor, filter *shared.TaskProjectsListFilter) (int64, error) {
	q := models.TaskProjects.Query()
	ListTaskProjectsFilterFunc(ctx, q, filter)
	data, err := q.Count(ctx, db)
	if err != nil {
		return 0, err
	}
	return data, nil
}

func CreateTaskProject(ctx context.Context, db bob.Executor, userID uuid.UUID, input *shared.CreateTaskProjectDTO) (*models.TaskProject, error) {
	taskProject := models.TaskProjectSetter{
		UserID:      omit.From(userID),
		Name:        omit.From(input.Name),
		Description: omitnull.FromPtr(input.Description),
		Status:      omit.From(input.Status),
		Order:       omit.From(input.Order),
	}
	projects, err := models.TaskProjects.Insert(&taskProject, im.Returning("*")).One(ctx, db)
	if err != nil {
		return nil, err
	}
	return projects, nil
}

func CreateTaskProjectWithTasks(ctx context.Context, db bob.Executor, userID uuid.UUID, input *shared.CreateTaskProjectWithTasksDTO) (*models.TaskProject, error) {
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
	err = taskProject.AttachProjectTasks(ctx, db, tasks...)
	if err != nil {
		return nil, err
	}
	return taskProject, nil
}

func CreateTaskWithChildren(ctx context.Context, db bob.Executor, userID uuid.UUID, projectID uuid.UUID, input *shared.CreateTaskWithChildrenDTO) (*models.Task, error) {
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

func CreateTask(ctx context.Context, db bob.Executor, userID uuid.UUID, projectID uuid.UUID, input *shared.CreateTaskBaseDTO) (*models.Task, error) {
	setter := models.TaskSetter{
		ProjectID:   omit.From(projectID),
		UserID:      omit.From(userID),
		Name:        omit.From(input.Name),
		Description: omitnull.FromPtr(input.Description),
		Status:      omit.From(input.Status),
		Order:       omit.From(input.Order),
	}
	task, err := models.Tasks.Insert(&setter, im.Returning("*")).One(ctx, db)
	if err != nil {
		return nil, err
	}
	return task, nil
}

func DefineTaskOrderNumber(ctx context.Context, db bob.Executor, taskId uuid.UUID, taskProjectId uuid.UUID, currentOrder float64, position int64) (float64, error) {
	if position == 0 {
		response, err := models.Tasks.Query(
			sm.Where(models.TaskColumns.ProjectID.EQ(psql.Arg(taskProjectId))),
			sm.OrderBy(models.TaskColumns.Order).Asc(),
			sm.Limit(1),
		).One(ctx, db)
		if err != nil {
			return 0, err
		}
		if response.ID == taskId {
			return response.Order, nil
		}
		return response.Order - 1000, nil
	}
	element, err := models.Tasks.Query(
		sm.Where(models.TaskColumns.ProjectID.EQ(psql.Arg(taskProjectId))),
		sm.OrderBy(models.TaskColumns.Order).Asc(),
		sm.Limit(1),
		sm.Offset(position),
	).One(ctx, db)
	if err != nil {
		return 0, err
	}
	if element.ID == taskId {
		return element.Order, nil
	}
	if currentOrder > element.Order {
		sideElements, err := models.Tasks.Query(
			sm.Where(models.TaskColumns.ID.EQ(psql.Arg(taskProjectId))),
			sm.OrderBy(models.TaskColumns.Order).Asc(),
			sm.Limit(1),
			sm.Offset(position-1),
		).One(ctx, db)
		if err != nil {
			return 0, err
		}
		return (element.Order + sideElements.Order) / 2, nil
	}
	sideElements, err := models.Tasks.Query(
		sm.Where(models.TaskColumns.ID.EQ(psql.Arg(taskProjectId))),
		sm.OrderBy(models.TaskColumns.Order).Asc(),
		sm.Limit(1),
		sm.Offset(position+1),
	).One(ctx, db)
	if err != nil {
		return 0, err
	}
	if sideElements == nil {
		return element.Order + 1000, nil
	}
	return (element.Order + sideElements.Order) / 2, nil

}

func UpdateTask(ctx context.Context, db bob.Executor, taskID uuid.UUID, input *shared.UpdateTaskBaseDTO) error {
	task, err := FindTaskByID(ctx, db, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("task not found")
	}
	taskSetter := &models.TaskSetter{
		Name:        omit.From(input.Name),
		Description: omitnull.FromPtr(input.Description),
		Status:      omit.From(input.Status),
		Order:       omit.From(input.Order),
		ParentID:    omitnull.FromPtr(input.ParentID),
	}
	if input.Position != nil {
		position := *input.Position
		order, err := DefineTaskOrderNumber(ctx, db, task.ID, task.ProjectID, task.Order, position)
		if err != nil {
			return err
		}
		taskSetter.Order = omit.From(order)
	}
	err = task.Update(ctx, db, taskSetter)
	if err != nil {
		return err
	}
	return nil
}

// private defineOrderNumber = async (
//
//		  id: number,
//		  todoListId: number,
//		  currentOrder: number,
//		  position: number,
//		) => {
//	    if (position === 0) {
//	      const response = await this.findByTodoListId(
//	        todoListId,
//	        { take: 1, skip: 0 },
//	        { property: 'order', direction: 'ASC' },
//	      );
//	      if (response.elements[0].id === id) return response.elements[0].order;
//	      return response.elements[0].order - 1000;
//	    }
//	    const elements = await this.todoGroupRepository.find(
//	      { take: 1, skip: position },
//	      {
//	        filter: { todoListId },
//	        sorting: { property: 'order', direction: 'ASC' },
//	      },
//	    );
//	    if (elements[0].id === id) return elements[0].order;
//	    if (currentOrder > elements[0].order) {
//	      const sideElements = await this.todoGroupRepository.find(
//	        { take: 1, skip: position - 1 },
//	        {
//	          filter: { todoListId },
//	          sorting: { property: 'order', direction: 'ASC' },
//	        },
//	      );
//	      return (elements[0].order + sideElements[0].order) / 2;
//	    }
//	    const sideElements = await this.todoGroupRepository.find(
//	      { take: 1, skip: position + 1 },
//	      {
//	        filter: { todoListId },
//	        sorting: { property: 'order', direction: 'ASC' },
//	      },
//	    );
//	    if (!sideElements.length) return elements[0].order + 1000;
//	    return (elements[0].order + sideElements[0].order) / 2;
//	  };
