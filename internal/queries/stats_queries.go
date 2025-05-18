package queries

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/shared"
)

const TaskStatsQuery = `
WITH project_stats AS (
    SELECT COUNT(*) as total_projects,
        COUNT(*) FILTER (
            WHERE tp.status = 'done'
        ) as completed_projects
    FROM task_projects tp
    WHERE tp.user_id = $1
),
task_stats AS (
    SELECT COUNT(*) as total_tasks,
        COUNT(*) FILTER (
            WHERE t.status = 'done'
        ) as completed_tasks
    FROM tasks t
    WHERE t.user_id = $1
)
SELECT ps.total_projects,
    ps.completed_projects,
    ts.total_tasks,
    ts.completed_tasks
FROM project_stats ps
    CROSS JOIN task_stats ts;
	`

func GetUserTaskStats(ctx context.Context, db database.Dbx, userID uuid.UUID) (*shared.TaskStats, error) {
	res, err := QueryAll[shared.TaskStats](ctx, db, TaskStatsQuery, userID)
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, nil
	}
	return &res[0], nil
}
