package stores

import (
	"testing"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
)

func Test_taskStore_TaskWhere(t *testing.T) {
	var id1 = uuid.New()
	var id2 = uuid.New()
	var id3 = uuid.New()
	type args struct {
		Task *TaskFilter
	}
	tests := []struct {
		name string
		args args
		want *map[string]any
	}{
		{
			name: "none",
			args: args{
				Task: nil,
			},
			want: nil,
		},
		{
			name: "name",
			args: args{
				Task: &TaskFilter{
					Names: []string{"test"},
				},
			},
			want: &map[string]any{
				"name": map[string]any{
					"_in": []string{"test"},
				},
			},
		}, {
			name: "project_id",
			args: args{
				Task: &TaskFilter{
					ProjectIds: []uuid.UUID{id1},
				},
			},
			want: &map[string]any{
				"project_id": map[string]any{
					"_in": []uuid.UUID{id1},
				},
			},
		},
		{
			name: "team_id",
			args: args{
				Task: &TaskFilter{
					TeamIds: []uuid.UUID{id2},
				},
			},
			want: &map[string]any{
				"team_id": map[string]any{
					"_in": []uuid.UUID{id2},
				},
			},
		},
		{
			name: "created_by_member_id",
			args: args{
				Task: &TaskFilter{
					CreatedByMemberIds: []uuid.UUID{id3},
				},
			},
			want: &map[string]any{
				"created_by_member_id": map[string]any{
					"_in": []uuid.UUID{id3},
				},
			},
		},
		{
			name: "name and team_id",
			args: args{
				Task: &TaskFilter{
					Names:   []string{"test"},
					TeamIds: []uuid.UUID{id2},
				},
			},
			want: &map[string]any{
				"name": map[string]any{
					"_in": []string{"test"},
				},
				"team_id": map[string]any{
					"_in": []uuid.UUID{id2},
				},
			},
		},
		{
			name: "name and project_id",
			args: args{
				Task: &TaskFilter{
					Names:      []string{"test"},
					ProjectIds: []uuid.UUID{id1},
				},
			},
			want: &map[string]any{
				"name": map[string]any{
					"_in": []string{"test"},
				},
				"project_id": map[string]any{
					"_in": []uuid.UUID{id1},
				},
			},
		},
		{
			name: "id and status",
			args: args{
				Task: &TaskFilter{
					Ids:      []uuid.UUID{id2},
					Statuses: []models.TaskStatus{models.TaskStatusDone},
				},
			},
			want: &map[string]any{
				"id": map[string]any{
					"_in": []uuid.UUID{id2},
				},
				"status": map[string]any{
					"_in": []models.TaskStatus{models.TaskStatusDone},
				},
			},
		},
		{
			name: "project_id and status",
			args: args{
				Task: &TaskFilter{
					ProjectIds: []uuid.UUID{id3},
					Statuses:   []models.TaskStatus{models.TaskStatusDone},
				},
			},
			want: &map[string]any{
				"id": map[string]any{
					"_in": []uuid.UUID{id2},
				},
				"status": map[string]any{
					"_in": []models.TaskStatus{models.TaskStatusDone},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &DbTaskStore{}
			got := tr.taskWhere(tt.args.Task)
			if got == nil {
				if tt.want != nil {
					t.Errorf("TaskWhere() = %v, want %v", got, tt.want)
				}
			}
			if tt.want == nil {
				if got != nil {
					t.Errorf("TaskWhere() = %v, want %v", got, tt.want)
				}
			}
			if got != nil && tt.want != nil && tt.args.Task != nil {
				for k := range *got {
					if k == "name" {
						if len(tt.args.Task.Names) == 0 {
							t.Errorf("have key %v, but have %v", k, tt.args.Task.Names)
						}
					}
					if k == "id" {
						if len(tt.args.Task.Ids) == 0 {
							t.Errorf("have key %v, but have %v", k, tt.args.Task.Ids)
						}
					}
					if k == "created_by_member_id" {
						if len(tt.args.Task.CreatedByMemberIds) == 0 {
							t.Errorf("have key %v, but have %v", k, tt.args.Task.CreatedByMemberIds)
						}
					}
					if k == "team_id" {
						if len(tt.args.Task.TeamIds) == 0 {
							t.Errorf("have key %v, but have %v", k, tt.args.Task.TeamIds)
						}
					}
					if k == "status" {
						if len(tt.args.Task.Statuses) == 0 {
							t.Errorf("have key %v, but have %v", k, tt.args.Task.Statuses)
						}
					}
					if k == "project_id" {
						if len(tt.args.Task.ProjectIds) == 0 {
							t.Errorf("have key %v, but have %v", k, tt.args.Task.ProjectIds)
						}
					}
				}
			}
		})
	}
}
