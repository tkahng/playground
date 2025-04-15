package repository

import (
	"context"
	"slices"

	"github.com/google/uuid"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/shared"
)

func GetTaskByID(ctx context.Context, db bob.Executor, id uuid.UUID) (*models.Task, error) {
	task, err := models.FindTask(ctx, db, id)
	return OptionalRow(task, err)
}

func GetTaskProjectByID(ctx context.Context, db bob.Executor, id uuid.UUID) (*models.TaskProject, error) {
	task, err := models.FindTaskProject(ctx, db, id)
	return OptionalRow(task, err)
}

func ListTasksOrderByFunc(ctx context.Context, q *psql.ViewQuery[*models.Task, models.TaskSlice], input *shared.TaskListParams) {
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
				sm.OrderBy(input.SortBy).Desc(),
				sm.OrderBy(models.TaskProjectColumns.ID).Desc(),
			)
		} else if input.SortParams.SortOrder == "asc" || input.SortParams.SortOrder == "" {
			q.Apply(
				sm.OrderBy(input.SortBy).Asc(),
				sm.OrderBy(models.TaskProjectColumns.ID).Asc(),
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
		id, err := uuid.Parse(filter.ProjectID)
		if err != nil {
			return
		}
		q.Apply(
			models.SelectWhere.Tasks.ID.EQ(id),
		)
	}

	if filter.ProjectID != "" {
		id, err := uuid.Parse(filter.ProjectID)
		if err != nil {
			return
		}
		q.Apply(
			models.SelectJoins.Tasks.InnerJoin.ProjectTaskProject(ctx),
			models.SelectWhere.TaskProjects.ID.EQ(id),
		)
	}

	if len(filter.Ids) > 0 {
		var ids []uuid.UUID = ParseUUIDs(filter.Ids)
		q.Apply(
			models.SelectWhere.Tasks.ID.In(ids...),
		)
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
				sm.OrderBy(input.SortBy).Desc(),
				sm.OrderBy(models.TaskProjectColumns.ID).Desc(),
			)
		} else if input.SortParams.SortOrder == "asc" || input.SortParams.SortOrder == "" {
			q.Apply(
				sm.OrderBy(input.SortBy).Asc(),
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
