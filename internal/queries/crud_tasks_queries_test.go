package queries_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/db"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/test"
)

func TestLoadTaskProjectsTasks(t *testing.T) {
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
		tasks, err := queries.CreateTask(ctx, dbxx, user.ID, taskProject.ID, &shared.CreateTaskBaseDTO{
			Name:   "Test Task",
			Status: shared.TaskStatusDone,
		})
		if err != nil {
			t.Fatalf("failed to create task: %v", err)
		}
		type args struct {
			ctx        context.Context
			db         db.Dbx
			projectIds []uuid.UUID
		}
		tests := []struct {
			name    string
			args    args
			want    [][]*models.Task
			wantErr bool
		}{
			{
				name: "query tasks",
				args: args{
					ctx: ctx,
					db:  dbxx,
					projectIds: []uuid.UUID{
						taskProject.ID,
					},
				},
				want: [][]*models.Task{
					{
						{
							ID:        tasks.ID,
							Name:      tasks.Name,
							Status:    tasks.Status,
							ProjectID: tasks.ProjectID,
							UserID:    tasks.UserID,
							CreatedAt: tasks.CreatedAt,
							UpdatedAt: tasks.UpdatedAt,
						},
					},
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := queries.LoadTaskProjectsTasks(tt.args.ctx, tt.args.db, tt.args.projectIds...)
				if (err != nil) != tt.wantErr {
					t.Errorf("LoadTaskProjectsTasks() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !reflect.DeepEqual(got[0][0].Name, tt.want[0][0].Name) {
					t.Errorf("LoadTaskProjectsTasks() = %v, want %v", got[0][0].Name, tt.want[0][0].Name)
				}
				if !reflect.DeepEqual(got[0][0].Status, tt.want[0][0].Status) {
					t.Errorf("LoadTaskProjectsTasks() = %v, want %v", got[0][0].Status, tt.want[0][0].Status)
				}
				if !reflect.DeepEqual(got[0][0].ProjectID, tt.want[0][0].ProjectID) {
					t.Errorf("LoadTaskProjectsTasks() = %v, want %v", got[0][0].ProjectID, tt.want[0][0].ProjectID)
				}
			})
		}
		return test.EndTestErr
	})
}
func TestFindTaskByID(t *testing.T) {
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
		task, err := queries.CreateTask(ctx, dbxx, user.ID, taskProject.ID, &shared.CreateTaskBaseDTO{
			Name:   "Test Task",
			Status: shared.TaskStatusDone,
		})
		if err != nil {
			t.Fatalf("failed to create task: %v", err)
		}

		type args struct {
			ctx context.Context
			db  db.Dbx
			id  uuid.UUID
		}
		tests := []struct {
			name    string
			args    args
			want    *models.Task
			wantErr bool
		}{
			{
				name: "find existing task",
				args: args{
					ctx: ctx,
					db:  dbxx,
					id:  task.ID,
				},
				want:    task,
				wantErr: false,
			},
			{
				name: "find non-existing task",
				args: args{
					ctx: ctx,
					db:  dbxx,
					id:  uuid.New(),
				},
				want:    nil,
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := queries.FindTaskByID(tt.args.ctx, tt.args.db, tt.args.id)
				if (err != nil) != tt.wantErr {
					t.Errorf("FindTaskByID() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if tt.want == nil {
					if got != nil {
						t.Errorf("FindTaskByID() = %v, want nil", got)
					}
					return
				}
				if !reflect.DeepEqual(got.ID, tt.want.ID) {
					t.Errorf("FindTaskByID() = %v, want %v", got.ID, tt.want.ID)
				}
				if !reflect.DeepEqual(got.Name, tt.want.Name) {
					t.Errorf("FindTaskByID() = %v, want %v", got.Name, tt.want.Name)
				}
			})
		}
		return test.EndTestErr
	})
}
func TestFindLastTaskOrder(t *testing.T) {
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		user, err := queries.CreateUser(
			ctx,
			dbxx,
			&shared.AuthenticationInput{
				Email: "tkahng@gmail.com",
			},
		)
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		taskProject, err := queries.CreateTaskProject(
			ctx,
			dbxx,
			user.ID,
			&shared.CreateTaskProjectDTO{
				Name:   "Test Project",
				Status: shared.TaskProjectStatusDone,
			},
		)
		if err != nil {
			t.Fatalf("failed to create task project: %v", err)
		}

		_, err = queries.CreateTask(
			ctx,
			dbxx,
			user.ID,
			taskProject.ID,
			&shared.CreateTaskBaseDTO{
				Name:   "Test Task 1",
				Status: shared.TaskStatusDone,
				Order:  1000,
			},
		)
		if err != nil {
			t.Fatalf("failed to create task: %v", err)
		}

		type args struct {
			ctx           context.Context
			db            db.Dbx
			taskProjectID uuid.UUID
		}
		tests := []struct {
			name    string
			args    args
			want    float64
			wantErr bool
		}{
			{
				name: "find last order with existing tasks",
				args: args{
					ctx:           ctx,
					db:            dbxx,
					taskProjectID: taskProject.ID,
				},
				want:    2000,
				wantErr: false,
			},
			{
				name: "find last order with non-existing project",
				args: args{
					ctx:           ctx,
					db:            dbxx,
					taskProjectID: uuid.New(),
				},
				want:    0,
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := queries.FindLastTaskOrder(tt.args.ctx, tt.args.db, tt.args.taskProjectID)
				if (err != nil) != tt.wantErr {
					t.Errorf("FindLastTaskOrder() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != tt.want {
					t.Errorf("FindLastTaskOrder() = %v, want %v", got, tt.want)
				}
			})
		}
		return test.EndTestErr
	})
}
