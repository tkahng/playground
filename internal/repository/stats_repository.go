package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/scan"
	"github.com/tkahng/authgo/internal/shared"
)

const TaskStatsQuery = `
WITH project_stats AS (
    SELECT COUNT(*) as total_projects,
        COUNT(*) FILTER (
            WHERE tp.status = 'done'
        ) as completed_projects
    FROM task_projects tp
    WHERE tp.user_id = ?
),
task_stats AS (
    SELECT COUNT(*) as total_tasks,
        COUNT(*) FILTER (
            WHERE t.status = 'done'
        ) as completed_tasks
    FROM tasks t
    WHERE t.user_id = ?
)
SELECT ps.total_projects,
    ps.completed_projects,
    ts.total_tasks,
    ts.completed_tasks
FROM project_stats ps
    CROSS JOIN task_stats ts;
	`

func GetUserTaskStats(ctx context.Context, db Queryer, userID uuid.UUID) (*shared.TaskStats, error) {
	query := psql.RawQuery(TaskStatsQuery, userID, userID)
	res, err := bob.All(ctx, db, query, scan.StructMapper[shared.TaskStats]())
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, nil
	}
	return &res[0], nil
}
