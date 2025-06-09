package stores

import (
	"testing"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/tools/types"
)

func Test_taskStore_TaskWhere(t *testing.T) {
	type args struct {
		Task *models.Task
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
				Task: &models.Task{
					Name: "test",
				},
			},
			want: &map[string]any{
				"name": "test",
			},
		}, {
			name: "project_id",
			args: args{
				Task: &models.Task{
					ProjectID: uuid.New(),
				},
			},
			want: &map[string]any{
				"project_id": uuid.New(),
			},
		},
		{
			name: "team_id",
			args: args{
				Task: &models.Task{
					TeamID: uuid.New(),
				},
			},
			want: &map[string]any{
				"team_id": uuid.New(),
			},
		},
		{
			name: "created_by_member_id",
			args: args{
				Task: &models.Task{
					CreatedByMemberID: types.Pointer(uuid.New()),
				},
			},
			want: &map[string]any{
				"created_by_member_id": uuid.New(),
			},
		},
		{
			name: "name and team_id",
			args: args{
				Task: &models.Task{
					Name:   "test",
					TeamID: uuid.New(),
				},
			},
			want: &map[string]any{
				"name":    "test",
				"team_id": uuid.New(),
			},
		},
		{
			name: "name and project_id",
			args: args{
				Task: &models.Task{
					Name:      "test",
					ProjectID: uuid.New(),
				},
			},
			want: &map[string]any{
				"name":       "test",
				"project_id": uuid.New(),
			},
		},
		{
			name: "id and status",
			args: args{
				Task: &models.Task{
					ID:     uuid.New(),
					Status: models.TaskStatusDone,
				},
			},
			want: &map[string]any{
				"id":     uuid.New(),
				"status": models.TaskStatusDone,
			},
		},
		{
			name: "project_id and status",
			args: args{
				Task: &models.Task{
					ProjectID: uuid.New(),
					Status:    models.TaskStatusDone,
				},
			},
			want: &map[string]any{
				"name":   "test",
				"status": models.TaskStatusDone,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &DbTaskStore{}
			got := tr.TaskWhere(tt.args.Task)
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
						if tt.args.Task.Name == "" {
							t.Errorf("have key %v, but have %v", k, tt.args.Task.Name)
						}
					}
					if k == "id" {
						if tt.args.Task.ID == uuid.Nil {
							t.Errorf("have key %v, but have %v", k, tt.args.Task.ID)
						}
					}
					if k == "created_by_member_id" {
						if tt.args.Task.CreatedByMemberID == nil {
							t.Errorf("have key %v, but have %v", k, tt.args.Task.CreatedByMemberID)
						}
					}
					if k == "team_id" {
						if tt.args.Task.TeamID == uuid.Nil {
							t.Errorf("have key %v, but have %v", k, tt.args.Task.TeamID)
						}
					}
					if k == "status" {
						if tt.args.Task.Status == "" {
							t.Errorf("have key %v, but have %v", k, tt.args.Task.Status)
						}
					}
					if k == "project_id" {
						if tt.args.Task.ProjectID == uuid.Nil {
							t.Errorf("have key %v, but have %v", k, tt.args.Task.ProjectID)
						}
					}
				}
			}
		})
	}
}
