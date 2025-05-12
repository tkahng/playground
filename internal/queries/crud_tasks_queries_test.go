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
	"github.com/tkahng/authgo/internal/tools/types"
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
func TestDeleteTask(t *testing.T) {
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
			ctx    context.Context
			db     db.Dbx
			taskID uuid.UUID
		}
		tests := []struct {
			name    string
			args    args
			wantErr bool
		}{
			{
				name: "delete existing task",
				args: args{
					ctx:    ctx,
					db:     dbxx,
					taskID: task.ID,
				},
				wantErr: false,
			},
			{
				name: "delete non-existing task",
				args: args{
					ctx:    ctx,
					db:     dbxx,
					taskID: uuid.New(),
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if err := queries.DeleteTask(tt.args.ctx, tt.args.db, tt.args.taskID); (err != nil) != tt.wantErr {
					t.Errorf("DeleteTask() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
		return test.EndTestErr
	})
}
func TestFindTaskProjectByID(t *testing.T) {
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

		type args struct {
			ctx context.Context
			db  db.Dbx
			id  uuid.UUID
		}
		tests := []struct {
			name    string
			args    args
			want    *models.TaskProject
			wantErr bool
		}{
			{
				name: "find existing task project",
				args: args{
					ctx: ctx,
					db:  dbxx,
					id:  taskProject.ID,
				},
				want:    taskProject,
				wantErr: false,
			},
			{
				name: "find non-existing task project",
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
				got, err := queries.FindTaskProjectByID(tt.args.ctx, tt.args.db, tt.args.id)
				if (err != nil) != tt.wantErr {
					t.Errorf("FindTaskProjectByID() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if tt.want == nil {
					if got != nil {
						t.Errorf("FindTaskProjectByID() = %v, want nil", got)
					}
					return
				}
				if !reflect.DeepEqual(got.ID, tt.want.ID) {
					t.Errorf("FindTaskProjectByID() = %v, want %v", got.ID, tt.want.ID)
				}
				if !reflect.DeepEqual(got.Name, tt.want.Name) {
					t.Errorf("FindTaskProjectByID() = %v, want %v", got.Name, tt.want.Name)
				}
			})
		}
		return test.EndTestErr
	})
}
func TestDeleteTaskProject(t *testing.T) {
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

		type args struct {
			ctx           context.Context
			db            db.Dbx
			taskProjectID uuid.UUID
		}
		tests := []struct {
			name    string
			args    args
			wantErr bool
		}{
			{
				name: "delete existing task project",
				args: args{
					ctx:           ctx,
					db:            dbxx,
					taskProjectID: taskProject.ID,
				},
				wantErr: false,
			},
			{
				name: "delete non-existing task project",
				args: args{
					ctx:           ctx,
					db:            dbxx,
					taskProjectID: uuid.New(),
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if err := queries.DeleteTaskProject(tt.args.ctx, tt.args.db, tt.args.taskProjectID); (err != nil) != tt.wantErr {
					t.Errorf("DeleteTaskProject() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
		return test.EndTestErr
	})
}
func TestListTasks(t *testing.T) {
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
			ctx   context.Context
			db    db.Dbx
			input *shared.TaskListParams
		}
		tests := []struct {
			name      string
			args      args
			wantCount int
			wantErr   bool
		}{
			{
				name: "list tasks with filter",
				args: args{
					ctx: ctx,
					db:  dbxx,
					input: &shared.TaskListParams{
						TaskListFilter: shared.TaskListFilter{
							ProjectID: taskProject.ID.String(),
							Status: []shared.TaskStatus{
								shared.TaskStatusDone,
							},
						},
						PaginatedInput: shared.PaginatedInput{
							Page:    0,
							PerPage: 10,
						},
					},
				},
				wantCount: 1,
				wantErr:   false,
			},
			{
				name: "list tasks without filter",
				args: args{
					ctx: ctx,
					db:  dbxx,
					input: &shared.TaskListParams{
						PaginatedInput: shared.PaginatedInput{
							Page:    0,
							PerPage: 10,
						},
					},
				},
				wantCount: 1,
				wantErr:   false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := queries.ListTasks(
					tt.args.ctx,
					tt.args.db,
					tt.args.input,
				)
				if (err != nil) != tt.wantErr {
					t.Errorf(
						"ListTasks() error = %v, wantErr %v",
						err,
						tt.wantErr,
					)
					return
				}
				if len(got) != tt.wantCount {
					t.Errorf("ListTasks() got length = %v, want length %v", len(got), tt.wantCount)
					return
				}
				if len(got) > 0 {
					if !reflect.DeepEqual(got[0].ID, task.ID) {
						t.Errorf("ListTasks() = %v, want %v", got[0].ID, task.ID)
					}
					if !reflect.DeepEqual(got[0].Name, task.Name) {
						t.Errorf("ListTasks() = %v, want %v", got[0].Name, task.Name)
					}
				}
			})
		}
		return test.EndTestErr
	})
}
func TestCountTasks(t *testing.T) {
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
			filter *shared.TaskListFilter
		}
		tests := []struct {
			name    string
			args    args
			want    int64
			wantErr bool
		}{
			{
				name: "count tasks with filter",
				args: args{
					ctx: ctx,
					db:  dbxx,
					filter: &shared.TaskListFilter{
						ProjectID: taskProject.ID.String(),
						Status: []shared.TaskStatus{
							shared.TaskStatusDone,
						},
					},
				},
				want:    1,
				wantErr: false,
			},
			{
				name: "count tasks without filter",
				args: args{
					ctx:    ctx,
					db:     dbxx,
					filter: nil,
				},
				want:    1,
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := queries.CountTasks(tt.args.ctx, tt.args.db, tt.args.filter)
				if (err != nil) != tt.wantErr {
					t.Errorf("CountTasks() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != tt.want {
					t.Errorf("CountTasks() = %v, want %v", got, tt.want)
				}
			})
		}
		return test.EndTestErr
	})
}
func TestListTaskProjects(t *testing.T) {
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

		type args struct {
			ctx   context.Context
			db    db.Dbx
			input *shared.TaskProjectsListParams
		}
		tests := []struct {
			name      string
			args      args
			wantCount int
			wantErr   bool
		}{
			{
				name: "list task projects with filter",
				args: args{
					ctx: ctx,
					db:  dbxx,
					input: &shared.TaskProjectsListParams{
						TaskProjectsListFilter: shared.TaskProjectsListFilter{
							UserID: user.ID.String(),
						},
						PaginatedInput: shared.PaginatedInput{
							Page:    0,
							PerPage: 10,
						},
					},
				},
				wantCount: 1,
				wantErr:   false,
			},
			{
				name: "list task projects without filter",
				args: args{
					ctx: ctx,
					db:  dbxx,
					input: &shared.TaskProjectsListParams{
						PaginatedInput: shared.PaginatedInput{
							Page:    0,
							PerPage: 10,
						},
					},
				},
				wantCount: 1,
				wantErr:   false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := queries.ListTaskProjects(
					tt.args.ctx,
					tt.args.db,
					tt.args.input,
				)
				if (err != nil) != tt.wantErr {
					t.Errorf(
						"ListTaskProjects() error = %v, wantErr %v",
						err,
						tt.wantErr,
					)
					return
				}
				if len(got) != tt.wantCount {
					t.Errorf("ListTaskProjects() got length = %v, want length %v", len(got), tt.wantCount)
					return
				}
				if len(got) > 0 {
					if !reflect.DeepEqual(got[0].ID, taskProject.ID) {
						t.Errorf("ListTaskProjects() = %v, want %v", got[0].ID, taskProject.ID)
					}
					if !reflect.DeepEqual(got[0].Name, taskProject.Name) {
						t.Errorf("ListTaskProjects() = %v, want %v", got[0].Name, taskProject.Name)
					}
				}
			})
		}
		return test.EndTestErr
	})
}
func TestCountTaskProjects(t *testing.T) {
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		user, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
			Email: "tkahng@gmail.com",
		})
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		_, err = queries.CreateTaskProject(ctx, dbxx, user.ID, &shared.CreateTaskProjectDTO{
			Name:   "Test Project",
			Status: shared.TaskProjectStatusDone,
		})
		if err != nil {
			t.Fatalf("failed to create task project: %v", err)
		}

		type args struct {
			ctx    context.Context
			db     db.Dbx
			filter *shared.TaskProjectsListFilter
		}
		tests := []struct {
			name    string
			args    args
			want    int64
			wantErr bool
		}{
			{
				name: "count task projects with filter",
				args: args{
					ctx: ctx,
					db:  dbxx,
					filter: &shared.TaskProjectsListFilter{
						UserID: user.ID.String(),
					},
				},
				want:    1,
				wantErr: false,
			},
			{
				name: "count task projects without filter",
				args: args{
					ctx:    ctx,
					db:     dbxx,
					filter: nil,
				},
				want:    1,
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := queries.CountTaskProjects(tt.args.ctx, tt.args.db, tt.args.filter)
				if (err != nil) != tt.wantErr {
					t.Errorf("CountTaskProjects() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != tt.want {
					t.Errorf("CountTaskProjects() = %v, want %v", got, tt.want)
				}
			})
		}
		return test.EndTestErr
	})
}
func TestCreateTaskProject(t *testing.T) {
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		user, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
			Email: "tkahng@gmail.com",
		})
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		type args struct {
			ctx    context.Context
			db     db.Dbx
			userID uuid.UUID
			input  *shared.CreateTaskProjectDTO
		}
		tests := []struct {
			name    string
			args    args
			want    *models.TaskProject
			wantErr bool
		}{
			{
				name: "create task project successfully",
				args: args{
					ctx:    ctx,
					db:     dbxx,
					userID: user.ID,
					input: &shared.CreateTaskProjectDTO{
						Name:        "Test Project",
						Description: types.Pointer("Test Description"),
						Status:      shared.TaskProjectStatusDone,
						Order:       1000,
					},
				},
				want: &models.TaskProject{
					UserID:      user.ID,
					Name:        "Test Project",
					Description: types.Pointer("Test Description"),
					Status:      models.TaskProjectStatusDone,
					Order:       1000,
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := queries.CreateTaskProject(tt.args.ctx, tt.args.db, tt.args.userID, tt.args.input)
				if (err != nil) != tt.wantErr {
					t.Errorf("CreateTaskProject() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if tt.want != nil {
					if !reflect.DeepEqual(got.Name, tt.want.Name) {
						t.Errorf("CreateTaskProject() Name = %v, want %v", got.Name, tt.want.Name)
					}
					if !reflect.DeepEqual(got.Description, tt.want.Description) {
						t.Errorf("CreateTaskProject() Description = %v, want %v", got.Description, tt.want.Description)
					}
					if !reflect.DeepEqual(got.Status, tt.want.Status) {
						t.Errorf("CreateTaskProject() Status = %v, want %v", got.Status, tt.want.Status)
					}
					if !reflect.DeepEqual(got.Order, tt.want.Order) {
						t.Errorf("CreateTaskProject() Order = %v, want %v", got.Order, tt.want.Order)
					}
					if !reflect.DeepEqual(got.UserID, tt.want.UserID) {
						t.Errorf("CreateTaskProject() UserID = %v, want %v", got.UserID, tt.want.UserID)
					}
				}
			})
		}
		return test.EndTestErr
	})
}
func TestCreateTaskProjectWithTasks(t *testing.T) {
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx db.Dbx) error {
		user, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
			Email: "tkahng@gmail.com",
		})
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		type args struct {
			ctx    context.Context
			db     db.Dbx
			userID uuid.UUID
			input  *shared.CreateTaskProjectWithTasksDTO
		}
		tests := []struct {
			name    string
			args    args
			want    *models.TaskProject
			wantErr bool
		}{
			{
				name: "create task project with tasks successfully",
				args: args{
					ctx:    ctx,
					db:     dbxx,
					userID: user.ID,
					input: &shared.CreateTaskProjectWithTasksDTO{
						CreateTaskProjectDTO: shared.CreateTaskProjectDTO{
							Name:        "Test Project",
							Description: types.Pointer("Test Description"),
							Status:      shared.TaskProjectStatusDone,
						},
						Tasks: []shared.CreateTaskBaseDTO{
							{
								Name:        "Test Task 1",
								Description: types.Pointer("Test Description 1"),
								Status:      shared.TaskStatusDone,
							},
							{
								Name:        "Test Task 2",
								Description: types.Pointer("Test Description 2"),
								Status:      shared.TaskStatusDone,
							},
						},
					},
				},
				want: &models.TaskProject{
					UserID:      user.ID,
					Name:        "Test Project",
					Description: types.Pointer("Test Description"),
					Status:      models.TaskProjectStatusDone,
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := queries.CreateTaskProjectWithTasks(tt.args.ctx, tt.args.db, tt.args.userID, tt.args.input)
				if (err != nil) != tt.wantErr {
					t.Errorf("CreateTaskProjectWithTasks() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if tt.want != nil {
					if !reflect.DeepEqual(got.Name, tt.want.Name) {
						t.Errorf("CreateTaskProjectWithTasks() Name = %v, want %v", got.Name, tt.want.Name)
					}
					if !reflect.DeepEqual(got.Description, tt.want.Description) {
						t.Errorf("CreateTaskProjectWithTasks() Description = %v, want %v", got.Description, tt.want.Description)
					}
					if !reflect.DeepEqual(got.Status, tt.want.Status) {
						t.Errorf("CreateTaskProjectWithTasks() Status = %v, want %v", got.Status, tt.want.Status)
					}
					if !reflect.DeepEqual(got.UserID, tt.want.UserID) {
						t.Errorf("CreateTaskProjectWithTasks() UserID = %v, want %v", got.UserID, tt.want.UserID)
					}

					// Verify tasks were created
					tasks, err := queries.ListTasks(tt.args.ctx, tt.args.db, &shared.TaskListParams{
						TaskListFilter: shared.TaskListFilter{
							ProjectID: got.ID.String(),
						},
					})
					if err != nil {
						t.Errorf("Failed to list tasks: %v", err)
					}
					if len(tasks) != len(tt.args.input.Tasks) {
						t.Errorf("Expected %d tasks, got %d", len(tt.args.input.Tasks), len(tasks))
					}
				}
			})
		}
		return test.EndTestErr
	})
}
