package shared

type UserStats struct {
	Task TaskStats `json:"task_stats" db:"task_stats"`
}

type TaskStats struct {
	TotalProjects     int64 `db:"total_projects" json:"total_projects"`
	CompletedProjects int64 `db:"completed_projects" json:"completed_projects"`
	TotalTasks        int64 `db:"total_tasks" json:"total_tasks"`
	CompletedTasks    int64 `db:"completed_tasks" json:"completed_tasks"`
}
