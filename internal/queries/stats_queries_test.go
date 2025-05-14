package queries_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/db"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/test"
)

func TestGetUserTaskStats(t *testing.T) {

	test.Short(t)
ctx, dbx := test.DbSetup()

	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		user, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
			Email: "tkahng@gmail.com",
		})
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		taskProject, err := queries.CreateTaskProject(ctx, dbxx, user.ID, &shared.CreateTaskProjectDTO{
			Name:   "Test Project",
			Status: shared.TaskProjectStatusDone,
		})
		if err != nil {
			t.Fatalf("failed to create task project: %v", err)
		}
		_, err = queries.CreateTask(ctx, dbxx, user.ID, taskProject.ID, &shared.CreateTaskBaseDTO{
			Name:   "Test Task",
			Status: shared.TaskStatusDone,
		})
		if err != nil {
			t.Fatalf("failed to create task: %v", err)
		}
		type args struct {
			ctx    context.Context
			db     db.Dbx
			userID uuid.UUID
		}
		tests := []struct {
			name    string
			args    args
			want    *shared.TaskStats
			wantErr bool
		}{
			// TODO: Add test cases.
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := queries.GetUserTaskStats(tt.args.ctx, tt.args.db, tt.args.userID)
				if (err != nil) != tt.wantErr {
					t.Errorf("GetUserTaskStats() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got.TotalTasks, tt.want.TotalTasks) {
					t.Errorf("GetUserTaskStats() = %v, want %v", got.TotalTasks, tt.want.TotalTasks)
				}
				if !reflect.DeepEqual(got.CompletedTasks, tt.want.CompletedTasks) {
					t.Errorf("GetUserTaskStats() = %v, want %v", got.CompletedTasks, tt.want.CompletedTasks)
				}
				if !reflect.DeepEqual(got.TotalProjects, tt.want.TotalProjects) {
					t.Errorf("GetUserTaskStats() = %v, want %v", got.TotalProjects, tt.want.TotalProjects)
				}
				if !reflect.DeepEqual(got.CompletedProjects, tt.want.CompletedProjects) {
					t.Errorf("GetUserTaskStats() = %v, want %v", got.CompletedProjects, tt.want.CompletedProjects)
				}
			})
		}
		return errors.New("rollback")
	})
}
