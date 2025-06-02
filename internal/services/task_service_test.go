package services_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/services"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/stores"
	"github.com/tkahng/authgo/internal/test"
	"github.com/tkahng/authgo/internal/tools/types"
)

func TestDefineTaskOrderNumberByStatus(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
		userStore := stores.NewPostgresUserStore(dbxx)
		teamstore := stores.NewPostgresTeamStore(dbxx)
		taskStore := stores.NewTaskStore(dbxx)
		taskService := services.NewTaskService(taskStore)
		user, err := userStore.CreateUser(ctx, &models.User{
			Email: "tkahng@gmail.com",
		})
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		member, err := teamstore.CreateTeamFromUser(ctx, user)
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

		task1, err := taskStore.CreateTask(ctx, &models.Task{
			Name:              "Task 1",
			Status:            models.TaskStatusDone,
			Rank:              1000,
			TeamID:            member.TeamID,
			ProjectID:         taskProject.ID,
			CreatedByMemberID: types.Pointer(member.ID),
		})
		if err != nil {
			t.Fatalf("failed to create task: %v", err)
		}
		task2, err := taskStore.CreateTask(ctx, &models.Task{
			Name:              "Task 2",
			Status:            models.TaskStatusDone,
			Rank:              2000,
			TeamID:            member.TeamID,
			ProjectID:         taskProject.ID,
			CreatedByMemberID: types.Pointer(member.ID),
		})
		if err != nil {
			t.Fatalf("failed to create task: %v", err)
		}
		task3, err := taskStore.CreateTask(ctx, &models.Task{
			Name:              "Task 3",
			Status:            models.TaskStatusDone,
			Rank:              3000,
			ProjectID:         taskProject.ID,
			TeamID:            member.TeamID,
			CreatedByMemberID: types.Pointer(member.ID),
		})
		if err != nil || task3 == nil {
			t.Fatalf("failed to create task: %v", err)
		}
		type args struct {
			ctx           context.Context
			db            database.Dbx
			taskId        uuid.UUID
			taskProjectId uuid.UUID
			status        models.TaskStatus
			currentOrder  float64
			position      int64
		}
		tests := []struct {
			name    string
			args    args
			want    float64
			wantErr bool
		}{
			{
				name: "get order for first position",
				args: args{
					ctx:           ctx,
					db:            dbxx,
					taskId:        task1.ID,
					taskProjectId: taskProject.ID,
					status:        models.TaskStatusDone,
					currentOrder:  1000,
					position:      0,
				},
				want:    1000,
				wantErr: false,
			},
			{
				name: "move second to first position",
				args: args{
					ctx:           ctx,
					db:            dbxx,
					taskId:        task2.ID,
					taskProjectId: taskProject.ID,
					status:        models.TaskStatusDone,
					currentOrder:  1000,
					position:      0,
				},
				want:    0,
				wantErr: false,
			},
			{
				name: "move first to last position",
				args: args{
					ctx:           ctx,
					db:            dbxx,
					taskId:        task1.ID,
					taskProjectId: taskProject.ID,
					status:        models.TaskStatusDone,
					currentOrder:  0,
					position:      2,
				},
				want:    4000,
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := taskService.CalculateNewPosition(
					tt.args.ctx,
					tt.args.taskProjectId,
					tt.args.status,
					tt.args.position,
					tt.args.taskId,
				)
				if (err != nil) != tt.wantErr {
					t.Errorf("DefineTaskOrderNumberByStatus() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != tt.want {
					t.Errorf("DefineTaskOrderNumberByStatus() = %v, want %v", got, tt.want)
				}
			})
		}
		return test.EndTestErr
	})
}
