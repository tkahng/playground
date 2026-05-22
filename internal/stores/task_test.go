//go:build integration

package stores_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/test"
	"github.com/tkahng/playground/internal/tools/types"
)

func TestSearchUserTasks(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewStorageAdapter(db)
		user := stores.CreateUser(adapter, ctx, "tkahng@gmail.com")
		team := stores.CreateTeam(adapter, ctx, "TestTeam")
		member := stores.CreateTeamMember(adapter, ctx, team, user, models.TeamMemberRoleOwner, true)
		project := stores.CreateTeamProject(adapter, ctx, member, "Test Project", "Test Project")
		task1 := &models.Task{
			ProjectID:         project.ID,
			Name:              "One",
			Status:            models.TaskStatusTodo,
			CreatedByMemberID: types.Pointer(member.ID),
			TeamID:            team.ID,
			Description:       types.Pointer("Uno"),
		}
		task2 := &models.Task{
			ProjectID:         project.ID,
			Name:              "Two",
			Status:            models.TaskStatusTodo,
			CreatedByMemberID: types.Pointer(member.ID),
			TeamID:            team.ID,
			Description:       types.Pointer("Dos"),
		}

		stores.CreateTask(adapter, ctx, task1)
		stores.CreateTask(adapter, ctx, task2)

		t.Run("search one", func(t *testing.T) {
			tasks, err := adapter.Task().ListTasks(ctx, &stores.TaskFilter{
				Q: "one",
			})
			if err != nil {
				t.Fatalf("failed to search tasks: %v", err)
			}
			if len(tasks) != 1 {
				t.Fatalf("expected 2 tasks, got %d", len(tasks))
			}
			if tasks[0].Name != "One" {
				t.Fatalf("expected task name to be One, got %s", tasks[0].Name)
			}
		})
		t.Run("search uno", func(t *testing.T) {
			tasks, err := adapter.Task().ListTasks(ctx, &stores.TaskFilter{
				Q: "uno",
			})
			if err != nil {
				t.Fatalf("failed to search tasks: %v", err)
			}
			if len(tasks) != 1 {
				t.Fatalf("expected 2 tasks, got %d", len(tasks))
			}
			if *tasks[0].Description != "Uno" {
				t.Fatalf("expected task description to be Uno, got %s", *tasks[0].Description)
			}
		})
		t.Run("search two", func(t *testing.T) {
			tasks, err := adapter.Task().ListTasks(ctx, &stores.TaskFilter{
				Q: "two",
			})
			if err != nil {
				t.Fatalf("failed to search tasks: %v", err)
			}
			if len(tasks) != 1 {
				t.Fatalf("expected 2 tasks, got %d", len(tasks))
			}
			if tasks[0].Name != "Two" {
				t.Fatalf("expected task name to be Two, got %s", tasks[0].Name)
			}
		})
		t.Run("search dos", func(t *testing.T) {
			tasks, err := adapter.Task().ListTasks(ctx, &stores.TaskFilter{
				Q: "dos",
			})
			if err != nil {
				t.Fatalf("failed to search tasks: %v", err)
			}
			if len(tasks) != 1 {
				t.Fatalf("expected 2 tasks, got %d", len(tasks))
			}
			if *tasks[0].Description != "Dos" {
				t.Fatalf("expected task description to be Dos, got %s", *tasks[0].Description)
			}
		})
	})
}

func TestGetUserTaskStats(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, dbxx database.Dbx) {
		adapter := stores.NewStorageAdapter(dbxx)
		taskStore := stores.NewDbTaskStore(dbxx)
		user, err := adapter.User().CreateUser(ctx, &models.User{
			Email: "tkahng@gmail.com",
		})

		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		member, err := adapter.TeamMember().CreateTeamFromUser(ctx, user)
		if err != nil {
			t.Fatalf("failed to create team from user: %v", err)
		}
		taskProject, err := taskStore.CreateTaskProject(ctx, &stores.CreateTaskProjectDTO{
			Name:     "Test Project",
			Status:   models.TaskProjectStatusDone,
			TeamID:   member.TeamID,
			MemberID: member.ID,
		})
		if err != nil {
			t.Fatalf("failed to create task project: %v", err)
		}
		_, err = taskStore.CreateTaskFromInput(ctx, member.TeamID, taskProject.ID, member.ID, &stores.CreateTaskProjectTaskDTO{
			Name:   "Test Task",
			Status: models.TaskStatusDone,
		})
		if err != nil {
			t.Fatalf("failed to create task: %v", err)
		}
		// Setup: 1 project (done) + 1 task (done). Verify the CTE returns correct counts.
		stats, err := adapter.Task().GetTeamTaskStats(ctx, member.TeamID)
		if err != nil {
			t.Fatalf("GetTeamTaskStats() error = %v", err)
		}
		if stats == nil {
			t.Fatal("GetTeamTaskStats() returned nil")
		}
		if stats.TotalProjects != 1 {
			t.Errorf("TotalProjects = %d, want 1", stats.TotalProjects)
		}
		if stats.CompletedProjects != 1 {
			t.Errorf("CompletedProjects = %d, want 1", stats.CompletedProjects)
		}
		if stats.TotalTasks != 1 {
			t.Errorf("TotalTasks = %d, want 1", stats.TotalTasks)
		}
		if stats.CompletedTasks != 1 {
			t.Errorf("CompletedTasks = %d, want 1", stats.CompletedTasks)
		}

		// Add a second project (todo) + 2 more tasks (1 done, 1 todo). Verify incremental counts.
		taskProject2, err := taskStore.CreateTaskProject(ctx, &stores.CreateTaskProjectDTO{
			Name:     "Second Project",
			Status:   models.TaskProjectStatusTodo,
			TeamID:   member.TeamID,
			MemberID: member.ID,
		})
		if err != nil {
			t.Fatalf("failed to create second task project: %v", err)
		}
		_, err = taskStore.CreateTaskFromInput(ctx, member.TeamID, taskProject2.ID, member.ID, &stores.CreateTaskProjectTaskDTO{
			Name:   "Todo Task",
			Status: models.TaskStatusTodo,
		})
		if err != nil {
			t.Fatalf("failed to create todo task: %v", err)
		}
		_, err = taskStore.CreateTaskFromInput(ctx, member.TeamID, taskProject2.ID, member.ID, &stores.CreateTaskProjectTaskDTO{
			Name:   "Done Task 2",
			Status: models.TaskStatusDone,
		})
		if err != nil {
			t.Fatalf("failed to create second done task: %v", err)
		}

		stats2, err := adapter.Task().GetTeamTaskStats(ctx, member.TeamID)
		if err != nil {
			t.Fatalf("GetTeamTaskStats() second check error = %v", err)
		}
		if stats2.TotalProjects != 2 {
			t.Errorf("TotalProjects = %d, want 2", stats2.TotalProjects)
		}
		if stats2.CompletedProjects != 1 {
			t.Errorf("CompletedProjects = %d, want 1", stats2.CompletedProjects)
		}
		if stats2.TotalTasks != 3 {
			t.Errorf("TotalTasks = %d, want 3", stats2.TotalTasks)
		}
		if stats2.CompletedTasks != 2 {
			t.Errorf("CompletedTasks = %d, want 2", stats2.CompletedTasks)
		}
	})
}
func TestLoadTaskProjectsTasks(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, dbxx database.Dbx) {
		adapter := stores.NewStorageAdapter(dbxx)

		user, err := adapter.User().CreateUser(ctx, &models.User{
			Email: "tkahng@gmail.com",
		})
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		member, err := adapter.TeamMember().CreateTeamMemberFromUserAndSlug(ctx, user, "TestTeam", models.TeamMemberRoleOwner)
		if err != nil {
			t.Fatalf("failed to create team from user: %v", err)
		}
		taskProject, err := adapter.Task().CreateTaskProject(ctx, &stores.CreateTaskProjectDTO{
			Name:     "Test Project",
			Status:   models.TaskProjectStatusDone,
			TeamID:   member.TeamID,
			MemberID: member.ID,
		})
		if err != nil {
			t.Fatalf("failed to create task project: %v", err)
		}
		tasks, err := adapter.Task().CreateTask(ctx, &models.Task{
			Name:              "Test Task",
			Status:            models.TaskStatusDone,
			CreatedByMemberID: types.Pointer(member.ID),
			ProjectID:         taskProject.ID,
			TeamID:            member.TeamID,
		})
		if err != nil {
			t.Fatalf("failed to create task: %v", err)
		}
		type args struct {
			ctx        context.Context
			db         database.Dbx
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
							ID:                tasks.ID,
							Name:              tasks.Name,
							Status:            tasks.Status,
							ProjectID:         tasks.ProjectID,
							CreatedByMemberID: tasks.CreatedByMemberID,
							TeamID:            tasks.TeamID,
							CreatedAt:         tasks.CreatedAt,
							UpdatedAt:         tasks.UpdatedAt,
						},
					},
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := adapter.Task().LoadTaskProjectsTasks(tt.args.ctx, tt.args.projectIds...)
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
	})
}
func TestFindTaskByID(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, dbxx database.Dbx) {
		adapter := stores.NewStorageAdapter(dbxx)
		teamstore := adapter.TeamMember()
		taskStore := adapter.Task()
		user, err := adapter.User().CreateUser(ctx, &models.User{
			Email: "tkahng@gmail.com",
		})
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		member, err := teamstore.CreateTeamMemberFromUserAndSlug(ctx, user, "TestTeam", models.TeamMemberRoleOwner)
		if err != nil {
			t.Fatalf("failed to create team from user: %v", err)
		}
		if member == nil {
			t.Fatalf("failed to create team member")
		}
		taskProject, err := taskStore.CreateTaskProject(ctx, &stores.CreateTaskProjectDTO{
			Name:     "Test Project",
			Status:   models.TaskProjectStatusDone,
			TeamID:   member.TeamID,
			MemberID: member.ID,
		})
		if err != nil {
			t.Fatalf("failed to create task project: %v", err)
		}
		task, err := taskStore.CreateTask(ctx, &models.Task{
			Name:              "Test Task",
			Status:            models.TaskStatusDone,
			CreatedByMemberID: types.Pointer(member.ID),
			TeamID:            member.TeamID,
			ProjectID:         taskProject.ID,
		})
		if err != nil {
			t.Fatalf("failed to create task: %v", err)
		}

		type args struct {
			ctx context.Context
			db  database.Dbx
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
				got, err := taskStore.FindTaskByID(tt.args.ctx, tt.args.id)
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
	})
}

func TestFindLastTaskOrder(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, dbxx database.Dbx) {
		adapter := stores.NewStorageAdapter(dbxx)

		user, err := adapter.User().CreateUser(
			ctx,
			&models.User{
				Email: "tkahng@gmail.com",
			},
		)
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		member, err := adapter.TeamMember().CreateTeamMemberFromUserAndSlug(ctx, user, "TestTeam", models.TeamMemberRoleOwner)
		if err != nil {
			t.Fatalf("failed to create team from user: %v", err)
		}
		taskProject, err := adapter.Task().CreateTaskProject(
			ctx,
			&stores.CreateTaskProjectDTO{
				Name:     "Test Project",
				Status:   models.TaskProjectStatusDone,
				TeamID:   member.TeamID,
				MemberID: member.ID,
			},
		)
		if err != nil {
			t.Fatalf("failed to create task project: %v", err)
		}

		_, err = adapter.Task().CreateTask(
			ctx,
			&models.Task{
				Name:              "Test Task 1",
				Status:            models.TaskStatusDone,
				Rank:              1000,
				ProjectID:         taskProject.ID,
				TeamID:            member.TeamID,
				CreatedByMemberID: types.Pointer(member.ID),
			},
		)
		if err != nil {
			t.Fatalf("failed to create task: %v", err)
		}

		type args struct {
			ctx           context.Context
			db            database.Dbx
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
				got, err := adapter.Task().FindLastTaskRank(tt.args.ctx, tt.args.taskProjectID)
				if (err != nil) != tt.wantErr {
					t.Errorf("FindLastTaskOrder() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != tt.want {
					t.Errorf("FindLastTaskOrder() = %v, want %v", got, tt.want)
				}
			})
		}
	})
}
func TestDeleteTask(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, dbxx database.Dbx) {
		adapter := stores.NewStorageAdapter(dbxx)
		user, err := adapter.User().CreateUser(ctx, &models.User{
			Email: "tkahng@gmail.com",
		})
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		member, err := adapter.TeamMember().CreateTeamMemberFromUserAndSlug(ctx, user, "TestTeam", models.TeamMemberRoleOwner)
		if err != nil {
			t.Fatalf("failed to create team from user: %v", err)
		}
		taskProject, err := adapter.Task().CreateTaskProject(ctx, &stores.CreateTaskProjectDTO{
			Name:     "Test Project",
			Status:   models.TaskProjectStatusDone,
			TeamID:   member.TeamID,
			MemberID: member.ID,
		})
		if err != nil {
			t.Fatalf("failed to create task project: %v", err)
		}

		task, err := adapter.Task().CreateTask(ctx, &models.Task{
			Name:              "Test Task 1",
			Status:            models.TaskStatusDone,
			Rank:              1000,
			ProjectID:         taskProject.ID,
			CreatedByMemberID: types.Pointer(member.ID),
			TeamID:            member.TeamID,
		})
		if err != nil {
			t.Fatalf("failed to create task: %v", err)
		}

		type args struct {
			ctx    context.Context
			db     database.Dbx
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
				if err := adapter.Task().DeleteTask(tt.args.ctx, tt.args.taskID); (err != nil) != tt.wantErr {
					t.Errorf("DeleteTask() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})
}
func TestFindTaskProjectByID(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, dbxx database.Dbx) {
		adapter := stores.NewStorageAdapter(dbxx)
		userStore := adapter.User()
		teamstore := adapter.TeamMember()
		taskStore := adapter.Task()
		user, err := userStore.CreateUser(ctx, &models.User{
			Email: "tkahng@gmail.com",
		})
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		member, err := teamstore.CreateTeamMemberFromUserAndSlug(ctx, user, "TestTeam", models.TeamMemberRoleOwner)
		if err != nil {
			t.Fatalf("failed to create team from user: %v", err)
		}
		taskProject, err := taskStore.CreateTaskProject(ctx, &stores.CreateTaskProjectDTO{
			Name:     "Test Project",
			Status:   models.TaskProjectStatusDone,
			TeamID:   member.TeamID,
			MemberID: member.ID,
		})
		if err != nil {
			t.Fatalf("failed to create task project: %v", err)
		}

		type args struct {
			ctx context.Context
			db  database.Dbx
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
				got, err := taskStore.FindTaskProjectByID(tt.args.ctx, tt.args.id)
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
	})
}
func TestDeleteTaskProject(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, dbxx database.Dbx) {
		adapter := stores.NewStorageAdapter(dbxx)
		userStore := adapter.User()
		teamstore := adapter.TeamMember()
		taskStore := adapter.Task()
		user, err := userStore.CreateUser(ctx, &models.User{
			Email: "tkahng@gmail.com",
		})
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		member, err := teamstore.CreateTeamMemberFromUserAndSlug(ctx, user, "TestTeam", models.TeamMemberRoleOwner)
		if err != nil {
			t.Fatalf("failed to create team from user: %v", err)
		}
		taskProject, err := taskStore.CreateTaskProject(ctx, &stores.CreateTaskProjectDTO{
			Name:     "Test Project",
			Status:   models.TaskProjectStatusDone,
			TeamID:   member.TeamID,
			MemberID: member.ID,
		})
		if err != nil {
			t.Fatalf("failed to create task project: %v", err)
		}

		type args struct {
			ctx           context.Context
			db            database.Dbx
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
				if err := taskStore.DeleteTaskProject(tt.args.ctx, tt.args.taskProjectID); (err != nil) != tt.wantErr {
					t.Errorf("DeleteTaskProject() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})
}
func TestListTasks(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, dbxx database.Dbx) {
		adapter := stores.NewStorageAdapter(dbxx)
		userStore := adapter.User()
		teamstore := adapter.TeamMember()
		taskStore := adapter.Task()
		user, err := userStore.CreateUser(ctx, &models.User{
			Email: "tkahng@gmail.com",
		})
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		member, err := teamstore.CreateTeamMemberFromUserAndSlug(ctx, user, "TestTeam", models.TeamMemberRoleOwner)
		if err != nil {
			t.Fatalf("failed to create team from user: %v", err)
		}
		taskProject, err := taskStore.CreateTaskProject(ctx, &stores.CreateTaskProjectDTO{
			Name:     "Test Project",
			Status:   models.TaskProjectStatusDone,
			TeamID:   member.TeamID,
			MemberID: member.ID,
		})
		if err != nil {
			t.Fatalf("failed to create task project: %v", err)
		}
		task, err := taskStore.CreateTask(ctx, &models.Task{
			Name:              "Test Task",
			Description:       nil,
			Status:            models.TaskStatusDone,
			ProjectID:         taskProject.ID,
			TeamID:            member.TeamID,
			CreatedByMemberID: types.Pointer(member.ID),
		})
		if err != nil {
			t.Fatalf("failed to create task: %v", err)
		}

		type args struct {
			ctx   context.Context
			db    database.Dbx
			input *stores.TaskFilter
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
					input: &stores.TaskFilter{
						ProjectIds: []uuid.UUID{taskProject.ID},
						Statuses:   []models.TaskStatus{models.TaskStatusDone},
						PaginatedInput: stores.PaginatedInput{
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
					input: &stores.TaskFilter{
						PaginatedInput: stores.PaginatedInput{
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
				got, err := taskStore.ListTasks(tt.args.ctx, tt.args.input)
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
	})
}
func TestCountTasks(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, dbxx database.Dbx) {
		adapter := stores.NewStorageAdapter(dbxx)
		userStore := adapter.User()
		teamstore := adapter.TeamMember()
		taskStore := adapter.Task()
		user, err := userStore.CreateUser(ctx, &models.User{
			Email: "tkahng@gmail.com",
		})
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		member, err := teamstore.CreateTeamMemberFromUserAndSlug(ctx, user, "TestTeam", models.TeamMemberRoleOwner)
		if err != nil {
			t.Fatalf("failed to create team from user: %v", err)
		}
		taskProject, err := taskStore.CreateTaskProject(ctx, &stores.CreateTaskProjectDTO{
			Name:     "Test Project",
			Status:   models.TaskProjectStatusDone,
			TeamID:   member.TeamID,
			MemberID: member.ID,
		})
		if err != nil {
			t.Fatalf("failed to create task project: %v", err)
		}
		_, err = taskStore.CreateTask(ctx, &models.Task{
			Name:              "Test Task",
			Description:       nil,
			Status:            models.TaskStatusDone,
			ProjectID:         taskProject.ID,
			TeamID:            member.TeamID,
			CreatedByMemberID: types.Pointer(member.ID),
		})
		if err != nil {
			t.Fatalf("failed to create task: %v", err)
		}

		type args struct {
			ctx    context.Context
			db     database.Dbx
			filter *stores.TaskFilter
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
					filter: &stores.TaskFilter{
						ProjectIds: []uuid.UUID{taskProject.ID},
						Statuses:   []models.TaskStatus{models.TaskStatusDone},
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
				got, err := taskStore.CountTasks(tt.args.ctx, tt.args.filter)
				if (err != nil) != tt.wantErr {
					t.Errorf("CountTasks() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != tt.want {
					t.Errorf("CountTasks() = %v, want %v", got, tt.want)
				}
			})
		}
	})
}
func TestListTaskProjects(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, dbxx database.Dbx) {
		adapter := stores.NewStorageAdapter(dbxx)
		userStore := adapter.User()
		teamstore := adapter.TeamMember()
		taskStore := adapter.Task()
		user, err := userStore.CreateUser(ctx, &models.User{
			Email: "tkahng@gmail.com",
		})
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		member, err := teamstore.CreateTeamMemberFromUserAndSlug(ctx, user, "TestTeam", models.TeamMemberRoleOwner)
		if err != nil {
			t.Fatalf("failed to create team from user: %v", err)
		}
		taskProject, err := taskStore.CreateTaskProject(ctx, &stores.CreateTaskProjectDTO{
			Name:     "Test Project",
			Status:   models.TaskProjectStatusDone,
			TeamID:   member.TeamID,
			MemberID: member.ID,
		})
		if err != nil {
			t.Fatalf("failed to create task project: %v", err)
		}

		type args struct {
			ctx   context.Context
			db    database.Dbx
			input *stores.TaskProjectsFilter
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
					input: &stores.TaskProjectsFilter{
						TeamIds: []uuid.UUID{member.TeamID},
						PaginatedInput: stores.PaginatedInput{
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
					input: &stores.TaskProjectsFilter{
						PaginatedInput: stores.PaginatedInput{
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
				got, err := taskStore.ListTaskProjects(tt.args.ctx, tt.args.input)
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
	})
}
func TestCountTaskProjects(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, dbxx database.Dbx) {
		adapter := stores.NewStorageAdapter(dbxx)
		userStore := adapter.User()
		teamstore := adapter.TeamMember()
		taskStore := adapter.Task()
		user, err := userStore.CreateUser(ctx, &models.User{
			Email: "tkahng@gmail.com",
		})
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		member, err := teamstore.CreateTeamMemberFromUserAndSlug(ctx, user, "TestTeam", models.TeamMemberRoleOwner)
		if err != nil {
			t.Fatalf("failed to create team from user: %v", err)
		}
		_, err = taskStore.CreateTaskProject(ctx, &stores.CreateTaskProjectDTO{
			Name:     "Test Project",
			Status:   models.TaskProjectStatusDone,
			TeamID:   member.TeamID,
			MemberID: member.ID,
		})
		if err != nil {
			t.Fatalf("failed to create task project: %v", err)
		}

		type args struct {
			ctx    context.Context
			db     database.Dbx
			filter *stores.TaskProjectsFilter
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
					filter: &stores.TaskProjectsFilter{
						TeamIds: []uuid.UUID{member.TeamID},
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
				got, err := taskStore.CountTaskProjects(tt.args.ctx, tt.args.filter)
				if (err != nil) != tt.wantErr {
					t.Errorf("CountTaskProjects() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != tt.want {
					t.Errorf("CountTaskProjects() = %v, want %v", got, tt.want)
				}
			})
		}
	})
}
func TestCreateTaskProject(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, dbxx database.Dbx) {
		adapter := stores.NewStorageAdapter(dbxx)
		userStore := adapter.User()
		teamstore := adapter.TeamMember()
		taskStore := adapter.Task()
		user, err := userStore.CreateUser(ctx, &models.User{
			Email: "tkahng@gmail.com",
		})
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		member, err := teamstore.CreateTeamMemberFromUserAndSlug(ctx, user, "TestTeam", models.TeamMemberRoleOwner)
		if err != nil {
			t.Fatalf("failed to create team from user: %v", err)
		}
		type args struct {
			ctx    context.Context
			db     database.Dbx
			userID uuid.UUID
			input  *stores.CreateTaskProjectDTO
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
					input: &stores.CreateTaskProjectDTO{
						Name:        "Test Project",
						Description: types.Pointer("Test Description"),
						Status:      models.TaskProjectStatusDone,
						Rank:        1000,
						TeamID:      member.TeamID,
						MemberID:    member.ID,
					},
				},
				want: &models.TaskProject{
					Name:              "Test Project",
					Description:       types.Pointer("Test Description"),
					Status:            models.TaskProjectStatusDone,
					Rank:              1000,
					TeamID:            member.TeamID,
					CreatedByMemberID: types.Pointer(member.ID),
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := taskStore.CreateTaskProject(tt.args.ctx, tt.args.input)
				if (err != nil) != tt.wantErr {
					t.Errorf("CreateTaskProject() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if tt.want != nil {
					if got == nil {
						t.Errorf("CreateTaskProject() got = nil, want %v", tt.want)
					}
					assert.Equal(t, got.Name, tt.want.Name)
					assert.Equal(t, got.Description, tt.want.Description)
					assert.Equal(t, got.Status, tt.want.Status)
					assert.Equal(t, got.Rank, tt.want.Rank)
					assert.Equal(t, got.TeamID, tt.want.TeamID)
					assert.Equal(t, got.CreatedByMemberID, tt.want.CreatedByMemberID)
				}
			})
		}
	})
}
func TestCreateTaskProjectWithTasks(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, dbxx database.Dbx) {
		adapter := stores.NewStorageAdapter(dbxx)
		userStore := adapter.User()
		teamstore := adapter.TeamMember()
		taskStore := adapter.Task()
		user, err := userStore.CreateUser(ctx, &models.User{
			Email: "tkahng@gmail.com",
		})
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		member, err := teamstore.CreateTeamMemberFromUserAndSlug(ctx, user, "TestTeam", models.TeamMemberRoleOwner)
		if err != nil {
			t.Fatalf("failed to create team from user: %v", err)
		}

		type args struct {
			ctx    context.Context
			db     database.Dbx
			userID uuid.UUID
			input  *stores.CreateTaskProjectWithTasksDTO
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
					input: &stores.CreateTaskProjectWithTasksDTO{
						CreateTaskProjectDTO: stores.CreateTaskProjectDTO{
							Name:        "Test Project",
							TeamID:      member.TeamID,
							MemberID:    member.ID,
							Description: types.Pointer("Test Description"),
							Status:      models.TaskProjectStatusDone,
						},
						Tasks: []stores.CreateTaskProjectTaskDTO{
							{
								Name:        "Test Task 1",
								Rank:        0,
								Description: types.Pointer("Test Description 1"),
								Status:      models.TaskStatusDone,
							},
							{
								Name:        "Test Task 2",
								Rank:        1000,
								Description: types.Pointer("Test Description 2"),
								Status:      models.TaskStatusDone,
							},
						},
					},
				},
				want: &models.TaskProject{
					TeamID:            member.TeamID,
					Name:              "Test Project",
					Description:       types.Pointer("Test Description"),
					Status:            models.TaskProjectStatusDone,
					CreatedByMemberID: types.Pointer(member.ID),
					Tasks: []*models.Task{
						{
							Name:              "Test Task 1",
							Rank:              0,
							TeamID:            member.TeamID,
							Description:       types.Pointer("Test Description 1"),
							Status:            models.TaskStatusDone,
							CreatedByMemberID: types.Pointer(member.ID),
						}, {
							Name:              "Test Task 2",
							Rank:              1000,
							TeamID:            member.TeamID,
							Description:       types.Pointer("Test Description 2"),
							Status:            models.TaskStatusDone,
							CreatedByMemberID: types.Pointer(member.ID),
						},
					},
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := taskStore.CreateTaskProjectWithTasks(tt.args.ctx, tt.args.input)
				if (err != nil) != tt.wantErr {
					t.Errorf("CreateTaskProjectWithTasks() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if tt.want != nil {
					assert.Equal(t, tt.want.Name, got.Name)
					assert.Equal(t, tt.want.Description, got.Description)
					assert.Equal(t, tt.want.Status, got.Status)
					assert.Equal(t, tt.want.TeamID, got.TeamID)
					assert.Equal(t, tt.want.CreatedByMemberID, got.CreatedByMemberID)
					require.NotNil(t, got.WorkflowID)
					require.NotNil(t, got.WorkflowStatusID)

					tasks, err := taskStore.ListTasks(tt.args.ctx, &stores.TaskFilter{
						ProjectIds: []uuid.UUID{got.ID},
					})
					if err != nil {
						t.Errorf("Failed to list tasks: %v", err)
					}
					assert.Equal(t, len(tasks), len(tt.want.Tasks))
					for i, task := range tasks {
						assert.Equal(t, tt.want.Tasks[i].Name, task.Name)
						assert.Equal(t, tt.want.Tasks[i].Rank, task.Rank)
						assert.Equal(t, tt.want.Tasks[i].Description, task.Description)
						assert.Equal(t, tt.want.Tasks[i].Status, task.Status)
						assert.Equal(t, tt.want.Tasks[i].TeamID, task.TeamID)
						assert.Equal(t, tt.want.Tasks[i].CreatedByMemberID, task.CreatedByMemberID)
						require.NotNil(t, task.WorkflowStatusID)
					}
				}
			})
		}
	})
}

func TestCreateTaskFromInput(t *testing.T) {
	t.Parallel()
	test.SkipIfShort(t)
	database.WithNewTestTx(t, func(ctx context.Context, dbxx database.Dbx) {
		adapter := stores.NewStorageAdapter(dbxx)
		userStore := adapter.User()
		teamstore := adapter.TeamMember()
		taskStore := adapter.Task()
		user, err := userStore.CreateUser(ctx, &models.User{
			Email: "tkahng@gmail.com",
		})
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		member, err := teamstore.CreateTeamMemberFromUserAndSlug(ctx, user, "TestTeam", models.TeamMemberRoleOwner)
		if err != nil {
			t.Fatalf("failed to create team from user: %v", err)
		}

		taskProject, err := taskStore.CreateTaskProject(ctx, &stores.CreateTaskProjectDTO{
			Name:     "Test Project",
			Status:   models.TaskProjectStatusDone,
			TeamID:   member.TeamID,
			MemberID: member.ID,
		})
		if err != nil {
			t.Fatalf("failed to create task project: %v", err)
		}

		type args struct {
			ctx       context.Context
			db        database.Dbx
			teamID    uuid.UUID
			projectID uuid.UUID
			memberID  uuid.UUID
			input     *stores.CreateTaskProjectTaskDTO
		}
		tests := []struct {
			name    string
			args    args
			want    *models.Task
			wantErr bool
		}{
			{
				name: "create task successfully",
				args: args{
					ctx:       ctx,
					db:        dbxx,
					teamID:    member.TeamID,
					projectID: taskProject.ID,
					memberID:  member.ID,
					input: &stores.CreateTaskProjectTaskDTO{
						Name:        "Test Task",
						Description: types.Pointer("Test Description"),
						Status:      models.TaskStatusDone,
						Rank:        1000,
					},
				},
				want: &models.Task{
					TeamID:            member.TeamID,
					ProjectID:         taskProject.ID,
					CreatedByMemberID: types.Pointer(member.ID),
					Name:              "Test Task",
					Description:       types.Pointer("Test Description"),
					Status:            models.TaskStatusDone,
					Rank:              1000,
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := taskStore.CreateTaskFromInput(tt.args.ctx, tt.args.teamID, tt.args.projectID, tt.args.memberID, tt.args.input)
				if (err != nil) != tt.wantErr {
					t.Errorf("CreateTask() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if tt.want != nil {
					assert.Equal(t, tt.want.TeamID, got.TeamID)
					assert.Equal(t, tt.want.ProjectID, got.ProjectID)
					assert.Equal(t, tt.want.CreatedByMemberID, got.CreatedByMemberID)
					assert.Equal(t, tt.want.Name, got.Name)
					assert.Equal(t, tt.want.Description, got.Description)
					assert.Equal(t, tt.want.Status, got.Status)
					assert.Equal(t, tt.want.Rank, got.Rank)
					require.NotNil(t, got.WorkflowStatusID)
				}
			})
		}
	})
}

func TestFindAndUpdateTask(t *testing.T) {
	t.Parallel()
	database.WithNewTestTx(t, func(ctx context.Context, dbxx database.Dbx) {
		adapter := stores.NewStorageAdapter(dbxx)
		userStore := adapter.User()
		queries := adapter.Task()
		user, err := userStore.CreateUser(ctx, &models.User{
			Email: "tkahng@gmail.com",
		})
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		member, err := adapter.TeamMember().CreateTeamFromUser(ctx, user)
		if err != nil {
			t.Fatalf("failed to create team from user: %v", err)
		}
		taskProject, err := queries.CreateTaskProject(ctx, &stores.CreateTaskProjectDTO{
			Name:     "Test Project",
			Status:   models.TaskProjectStatusDone,
			TeamID:   member.TeamID,
			MemberID: member.ID,
		})
		if err != nil {
			t.Fatalf("failed to create task project: %v", err)
		}
		task, err := queries.CreateTask(ctx, &models.Task{
			ProjectID:         taskProject.ID,
			Name:              "Test Task",
			Description:       types.Pointer("Test Description"),
			Status:            models.TaskStatusDone,
			Rank:              0,
			TeamID:            member.TeamID,
			CreatedByMemberID: types.Pointer(member.ID),
		})
		if err != nil {
			t.Fatalf("failed to create task: %v", err)
		}

		type args struct {
			ctx    context.Context
			db     database.Dbx
			taskID uuid.UUID
			input  *stores.UpdateTaskDto
		}
		tests := []struct {
			name    string
			args    args
			wantErr bool
		}{
			{
				name: "update task successfully",
				args: args{
					ctx:    ctx,
					db:     dbxx,
					taskID: task.ID,
					input: &stores.UpdateTaskDto{
						Name:        "Updated Task",
						Description: types.Pointer("Updated Description"),
						Status:      models.TaskStatusInProgress,
						ParentID:    nil,
					},
				},
				wantErr: false,
			},
			{
				name: "update non-existing task",
				args: args{
					ctx:    ctx,
					db:     dbxx,
					taskID: uuid.New(),
					input: &stores.UpdateTaskDto{
						Name:   "Updated Task",
						Status: models.TaskStatusInProgress,
					},
				},
				wantErr: true,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := queries.FindAndUpdateTask(tt.args.ctx, tt.args.taskID, tt.args.input)
				if (err != nil) != tt.wantErr {
					t.Errorf("UpdateTask() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				if !tt.wantErr {
					// Verify task was updated
					updatedTask, err := queries.FindTaskByID(tt.args.ctx, tt.args.taskID)
					if err != nil {
						t.Errorf("Failed to get updated task: %v", err)
						return
					}
					assert.Equal(t, tt.args.input.Name, updatedTask.Name)
					assert.Equal(t, tt.args.input.Description, updatedTask.Description)
					assert.Equal(t, tt.args.input.Status, updatedTask.Status)
					assert.Equal(t, tt.args.input.ParentID, updatedTask.ParentID)
					assert.Equal(t, tt.args.input.StartAt, updatedTask.StartAt)
					assert.Equal(t, tt.args.input.EndAt, updatedTask.EndAt)
					assert.Equal(t, tt.args.input.AssigneeID, updatedTask.AssigneeID)
					assert.Equal(t, tt.args.input.ReporterID, updatedTask.ReporterID)
				}
			})
		}
	})
}

func TestUpdateTaskProject(t *testing.T) {
	t.Parallel()
	database.WithNewTestTx(t, func(ctx context.Context, dbxx database.Dbx) {
		adapter := stores.NewStorageAdapter(dbxx)
		userStore := adapter.User()
		queries := adapter.Task()
		user, err := userStore.CreateUser(ctx, &models.User{
			Email: "tkahng@gmail.com",
		})
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		member, err := adapter.TeamMember().CreateTeamFromUser(ctx, user)
		if err != nil {
			t.Fatalf("failed to create team from user: %v", err)
		}
		taskProject, err := queries.CreateTaskProject(ctx, &stores.CreateTaskProjectDTO{
			Name:        "Test Project",
			Description: types.Pointer("Test Description"),
			Status:      models.TaskProjectStatusDone,
			Rank:        1000,
			TeamID:      member.TeamID,
			MemberID:    member.ID,
		})
		if err != nil {
			t.Fatalf("failed to create task project: %v", err)
		}

		type args struct {
			ctx           context.Context
			db            database.Dbx
			taskProjectID uuid.UUID
			input         *stores.UpdateTaskProjectBaseDTO
		}
		tests := []struct {
			name    string
			args    args
			wantErr bool
		}{
			{
				name: "update task project successfully",
				args: args{
					ctx:           ctx,
					db:            dbxx,
					taskProjectID: taskProject.ID,
					input: &stores.UpdateTaskProjectBaseDTO{
						Name:        "Updated Project",
						Description: types.Pointer("Updated Description"),
						Status:      models.TaskProjectStatusInProgress,
						Rank:        2000,
					},
				},
				wantErr: false,
			},
			{
				name: "update non-existing task project",
				args: args{
					ctx:           ctx,
					db:            dbxx,
					taskProjectID: uuid.New(),
					input: &stores.UpdateTaskProjectBaseDTO{
						Name:   "Updated Project",
						Status: models.TaskProjectStatusInProgress,
					},
				},
				wantErr: true,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := queries.UpdateTaskProject(tt.args.ctx, tt.args.taskProjectID, tt.args.input)
				if (err != nil) != tt.wantErr {
					t.Errorf("UpdateTaskProject() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				if !tt.wantErr {
					// Verify task project was updated
					updatedProject, err := queries.FindTaskProjectByID(tt.args.ctx, tt.args.taskProjectID)
					if err != nil {
						t.Errorf("Failed to get updated task project: %v", err)
						return
					}
					if updatedProject.Name != tt.args.input.Name {
						t.Errorf("Task project name not updated. got = %v, want %v", updatedProject.Name, tt.args.input.Name)
					}
					if *updatedProject.Description != *tt.args.input.Description {
						t.Errorf("Task project description not updated. got = %v, want %v", *updatedProject.Description, *tt.args.input.Description)
					}
					if updatedProject.Status != models.TaskProjectStatus(tt.args.input.Status) {
						t.Errorf("Task project status not updated. got = %v, want %v", updatedProject.Status, tt.args.input.Status)
					}
					if updatedProject.Rank != tt.args.input.Rank {
						t.Errorf("Task project order not updated. got = %v, want %v", updatedProject.Rank, tt.args.input.Rank)
					}
					if updatedProject.WorkflowStatusID == nil {
						t.Errorf("Task project workflow status was not set")
					}
				}
			})
		}
	})
}

func TestUpdateTaskPositionStatus(t *testing.T) {
	t.Parallel()
	database.WithNewTestTx(t, func(ctx context.Context, dbxx database.Dbx) {
		adapter := stores.NewStorageAdapter(dbxx)
		taskStore := adapter.Task()
		userStore := adapter.User()
		user, err := userStore.CreateUser(ctx, &models.User{
			Email: "tkahng@gmail.com",
		})
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		member, err := adapter.TeamMember().CreateTeamMemberFromUserAndSlug(ctx, user, "TestTeam", models.TeamMemberRoleOwner)
		if err != nil {
			t.Fatalf("failed to create team from user: %v", err)
		}
		taskProject, err := taskStore.CreateTaskProject(ctx, &stores.CreateTaskProjectDTO{
			Name:     "Test Project",
			Status:   models.TaskProjectStatusDone,
			TeamID:   member.TeamID,
			MemberID: member.ID,
		})
		if err != nil {
			t.Fatalf("failed to create task project: %v", err)
		}

		task1, err := taskStore.CreateTask(ctx, &models.Task{
			Name:              "Task 1",
			Status:            models.TaskStatusDone,
			ProjectID:         taskProject.ID,
			Rank:              0,
			CreatedByMemberID: &member.ID,
			TeamID:            member.TeamID,
			CreatedByMember:   member,
		})
		if err != nil {
			t.Fatalf("failed to create task: %v", err)
		}

		task2, err := taskStore.CreateTask(ctx, &models.Task{
			Name:              "Task 2",
			Status:            models.TaskStatusDone,
			ProjectID:         taskProject.ID,
			Rank:              1000,
			CreatedByMemberID: &member.ID,
			TeamID:            member.TeamID,
			CreatedByMember:   member,
		})
		if err != nil {
			t.Fatalf("failed to create task: %v", err)
		}

		task3, err := taskStore.CreateTask(ctx, &models.Task{
			Name:              "Task 3",
			Status:            models.TaskStatusDone,
			ProjectID:         taskProject.ID,
			Rank:              2000,
			CreatedByMemberID: &member.ID,
			TeamID:            member.TeamID,
			CreatedByMember:   member,
		})
		if err != nil {
			t.Fatalf("failed to create task: %v", err)
		}

		type args struct {
			ctx      context.Context
			db       database.Dbx
			taskID   uuid.UUID
			position int64
			status   models.TaskStatus
		}
		tests := []struct {
			name    string
			args    args
			wantErr bool
		}{
			{
				name: "update task1 from position 0 to 1: task2, task1, task3",
				args: args{
					ctx:      ctx,
					db:       dbxx,
					taskID:   task1.ID,
					position: 1,
					status:   models.TaskStatusDone,
				},
				wantErr: false,
			},
			{
				name: "update non-existing task position status",
				args: args{
					ctx:      ctx,
					db:       dbxx,
					taskID:   uuid.New(),
					position: 0,
					status:   models.TaskStatusDone,
				},
				wantErr: true,
			},
			{
				name: "update task3 from position 2 to 0:  task3, task2, task1",
				args: args{
					ctx:      ctx,
					db:       dbxx,
					taskID:   task3.ID,
					position: 0,
					status:   models.TaskStatusDone,
				},
				wantErr: false,
			},
			{
				name: "update task2 from position 1 to 2:  task3, task1, task2",
				args: args{
					ctx:      ctx,
					db:       dbxx,
					taskID:   task2.ID,
					position: 2,
					status:   models.TaskStatusDone,
				},
				wantErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := taskStore.UpdateTaskRankStatus(tt.args.ctx, tt.args.taskID, tt.args.position, tt.args.status, nil)
				if (err != nil) != tt.wantErr {
					t.Errorf("UpdateTaskPositionStatus() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				if !tt.wantErr {
					// Verify task was updated
					updatedTask, err := taskStore.FindTaskByID(tt.args.ctx, tt.args.taskID)
					if err != nil {
						t.Errorf("Failed to get updated task: %v", err)
						return
					}
					if updatedTask.Status != tt.args.status {
						t.Errorf("Task status not updated. got = %v, want %v", updatedTask.Status, tt.args.status)
					}
					if updatedTask.WorkflowStatusID == nil {
						t.Errorf("Task workflow status was not set")
					}
					// find position of updated task in project by ordering by rank.
					updatedTasks, listTaskErr := taskStore.ListTasks(tt.args.ctx, &stores.TaskFilter{
						SortParams: stores.SortParams{
							SortBy:    "rank",
							SortOrder: "asc",
						},
						ProjectIds: []uuid.UUID{taskProject.ID},
					})
					if listTaskErr != nil {
						t.Errorf("Failed to get updated task: %v", err)
						return
					}
					taskNames := make([]string, len(updatedTasks))
					taskRank := make([]float64, len(updatedTasks))
					for i, task := range updatedTasks {
						taskNames[i] = task.Name
						taskRank[i] = task.Rank
					}
					t.Logf("updated tasks: %v", taskNames)
					t.Logf("updated tasks rank: %v", taskRank)
					// Get task project to verify update date
					newTaskProject, err := taskStore.FindTaskProjectByID(tt.args.ctx, updatedTask.ProjectID)
					if err != nil {
						t.Errorf("Failed to get task project: %v", err)
						return
					}
					for i, task := range updatedTasks {
						if task.ID == tt.args.taskID {
							if i != int(tt.args.position) {
								t.Errorf("Task position not updated. got = %v, want %v", i, tt.args.position)
							}
						}
					}
					// Get task project to verify update date

					if !newTaskProject.UpdatedAt.After(taskProject.UpdatedAt) {
						t.Errorf("Task project update date not updated. original = %v, updated %v", taskProject.UpdatedAt, newTaskProject.UpdatedAt)
					}
				}
			})
		}

		// Verify passing an explicit workflowStatusID resolves correctly.
		// Uses a fresh task to avoid interfering with the position tests above.
		t.Run("update task with explicit workflow status id", func(t *testing.T) {
			if taskProject.WorkflowID == nil {
				t.Fatal("taskProject.WorkflowID is nil")
			}
			statuses, loadErr := taskStore.LoadWorkflowStatuses(ctx, *taskProject.WorkflowID)
			if loadErr != nil || len(statuses) == 0 {
				t.Fatalf("failed to load workflow statuses: %v", loadErr)
			}
			var inProgressStatusID *uuid.UUID
			for _, s := range statuses[0] {
				if s.Category == string(models.TaskStatusInProgress) {
					id := s.ID
					inProgressStatusID = &id
					break
				}
			}
			if inProgressStatusID == nil {
				t.Fatal("in_progress workflow status not found")
			}
			freshTask, freshErr := taskStore.CreateTask(ctx, &models.Task{
				Name:      "Fresh task for explicit status test",
				Status:    models.TaskStatusDone,
				ProjectID: taskProject.ID,
				Rank:      9999,
				TeamID:    task1.TeamID,
			})
			if freshErr != nil {
				t.Fatalf("failed to create fresh task: %v", freshErr)
			}
			if err := taskStore.UpdateTaskRankStatus(ctx, freshTask.ID, 0, models.TaskStatusInProgress, inProgressStatusID); err != nil {
				t.Fatalf("UpdateTaskRankStatus with explicit workflowStatusID failed: %v", err)
			}
			updated, findErr := taskStore.FindTaskByID(ctx, freshTask.ID)
			if findErr != nil {
				t.Fatalf("FindTaskByID failed: %v", findErr)
			}
			if updated.Status != models.TaskStatusInProgress {
				t.Errorf("Status = %v, want in_progress", updated.Status)
			}
			if updated.WorkflowStatusID == nil || *updated.WorkflowStatusID != *inProgressStatusID {
				t.Errorf("WorkflowStatusID = %v, want %v", updated.WorkflowStatusID, inProgressStatusID)
			}
		})
	})
}
