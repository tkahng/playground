package queries_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/stores"
	"github.com/tkahng/authgo/internal/test"
)

func TestGetUserTaskStats(t *testing.T) {

	test.Short(t)
	ctx, dbx := test.DbSetup()

	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
		taskStore := stores.NewDbTaskStore(dbxx)
		user, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
			Email: "tkahng@gmail.com",
		})

		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		member, err := queries.CreateTeamFromUser(ctx, dbxx, user)
		if err != nil {
			t.Fatalf("failed to create team from user: %v", err)
		}
		taskProject, err := taskStore.CreateTaskProject(ctx, &shared.CreateTaskProjectDTO{
			Name:     "Test Project",
			Status:   shared.TaskProjectStatusDone,
			TeamID:   member.TeamID,
			MemberID: member.ID,
		})
		if err != nil {
			t.Fatalf("failed to create task project: %v", err)
		}
		_, err = taskStore.CreateTaskFromInput(ctx, member.TeamID, taskProject.ID, member.ID, &shared.CreateTaskProjectTaskDTO{
			Name:   "Test Task",
			Status: shared.TaskStatusDone,
		})
		if err != nil {
			t.Fatalf("failed to create task: %v", err)
		}
		type args struct {
			ctx    context.Context
			db     database.Dbx
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
				got, err := queries.GetTeamTaskStats(tt.args.ctx, tt.args.db, tt.args.userID)
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
