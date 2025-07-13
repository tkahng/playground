package stores_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/test"
	"github.com/tkahng/playground/internal/tools/types"
)

func TestGetUserTaskStats(t *testing.T) {
	test.Parallel(t)
	test.SkipIfShort(t)
	test.WithTx(t, func(ctx context.Context, dbxx database.Dbx) {
		adapter := stores.NewStorageAdapter(dbxx)
		taskStore := stores.NewDbTaskStore(dbxx)
		user, err := adapter.User().CreateUser(ctx, &models.User{
			Email: "tkahng@gmail.com",
		})

		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		member, err := adapter.TeamMember().CreateTeamFromUser(ctx, user)
		// member, err := adapter.TeamMember().CreateTeamMemberFromUserAndSlug(ctx, user, "TestTeam", models.TeamMemberRoleOwner)
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
		type args struct {
			ctx    context.Context
			db     database.Dbx
			userID uuid.UUID
		}
		tests := []struct {
			name    string
			args    args
			want    *models.TaskStats
			wantErr bool
		}{
			// TODO: Add test cases.
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := adapter.Task().GetTeamTaskStats(tt.args.ctx, tt.args.userID)
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
	})
}
func TestLoadTaskProjectsTasks(t *testing.T) {
	test.Parallel(t)
	test.SkipIfShort(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
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
		return test.ErrEndTest
	})
}
func TestFindTaskByID(t *testing.T) {
	test.Parallel(t)
	test.SkipIfShort(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
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
		return test.ErrEndTest
	})
}

func TestFindLastTaskOrder(t *testing.T) {
	test.Parallel(t)
	test.SkipIfShort(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
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
		return test.ErrEndTest
	})
}
func TestDeleteTask(t *testing.T) {
	test.Parallel(t)
	test.SkipIfShort(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
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
		return test.ErrEndTest
	})
}
func TestFindTaskProjectByID(t *testing.T) {
	test.Parallel(t)
	test.SkipIfShort(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
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
		return test.ErrEndTest
	})
}
func TestDeleteTaskProject(t *testing.T) {
	test.Parallel(t)
	test.SkipIfShort(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
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
		return test.ErrEndTest
	})
}
func TestListTasks(t *testing.T) {
	test.Parallel(t)
	test.SkipIfShort(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
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
		return test.ErrEndTest
	})
}
func TestCountTasks(t *testing.T) {
	test.Parallel(t)
	test.SkipIfShort(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
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
		return test.ErrEndTest
	})
}
func TestListTaskProjects(t *testing.T) {
	test.Parallel(t)
	test.SkipIfShort(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
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
		return test.ErrEndTest
	})
}
func TestCountTaskProjects(t *testing.T) {
	test.Parallel(t)
	test.SkipIfShort(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
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
		return test.ErrEndTest
	})
}
func TestCreateTaskProject(t *testing.T) {
	test.Parallel(t)
	test.SkipIfShort(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
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
					// if !reflect.DeepEqual(got.Name, tt.want.Name) {
					// 	t.Errorf("CreateTaskProject() Name = %v, want %v", got.Name, tt.want.Name)
					// }
					// if !reflect.DeepEqual(got.Description, tt.want.Description) {
					// 	t.Errorf("CreateTaskProject() Description = %v, want %v", got.Description, tt.want.Description)
					// }
					// if !reflect.DeepEqual(got.Status, tt.want.Status) {
					// 	t.Errorf("CreateTaskProject() Status = %v, want %v", got.Status, tt.want.Status)
					// }
					// if !reflect.DeepEqual(got.Rank, tt.want.Rank) {
					// 	t.Errorf("CreateTaskProject() Rank = %v, want %v", got.Rank, tt.want.Rank)
					// }
					// if !reflect.DeepEqual(got.UserID, tt.want.UserID) {
					// 	t.Errorf("CreateTaskProject() UserID = %v, want %v", got.UserID, tt.want.UserID)
					// }
				}
			})
		}
		return test.ErrEndTest
	})
}
func TestCreateTaskProjectWithTasks(t *testing.T) {
	test.Parallel(t)
	test.SkipIfShort(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
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
								Rank:        1000,
								Description: types.Pointer("Test Description 1"),
								Status:      models.TaskStatusDone,
							},
							{
								Name:        "Test Task 2",
								Rank:        2000,
								Description: types.Pointer("Test Description 2"),
								Status:      models.TaskStatusDone,
							},
						},
					},
				},
				want: &models.TaskProject{

					Name:        "Test Project",
					Description: types.Pointer("Test Description"),
					Status:      models.TaskProjectStatusDone,
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
					if !reflect.DeepEqual(got.Name, tt.want.Name) {
						t.Errorf("CreateTaskProjectWithTasks() Name = %v, want %v", got.Name, tt.want.Name)
					}
					if !reflect.DeepEqual(got.Description, tt.want.Description) {
						t.Errorf("CreateTaskProjectWithTasks() Description = %v, want %v", got.Description, tt.want.Description)
					}
					if !reflect.DeepEqual(got.Status, tt.want.Status) {
						t.Errorf("CreateTaskProjectWithTasks() Status = %v, want %v", got.Status, tt.want.Status)
					}
					// if !reflect.DeepEqual(got.UserID, tt.want.UserID) {
					// 	t.Errorf("CreateTaskProjectWithTasks() UserID = %v, want %v", got.UserID, tt.want.UserID)
					// }

					// Verify tasks were created
					tasks, err := taskStore.ListTasks(tt.args.ctx, &stores.TaskFilter{
						ProjectIds: []uuid.UUID{got.ID},
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
		return test.ErrEndTest
	})
}

func TestCreateTaskFromInput(t *testing.T) {
	test.Parallel(t)
	test.SkipIfShort(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
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
					TeamID:      member.TeamID,
					ProjectID:   taskProject.ID,
					Name:        "Test Task",
					Description: types.Pointer("Test Description"),
					Status:      models.TaskStatusDone,
					Rank:        1000,
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
					if !reflect.DeepEqual(got.Name, tt.want.Name) {
						t.Errorf("CreateTask() Name = %v, want %v", got.Name, tt.want.Name)
					}
					if !reflect.DeepEqual(got.Description, tt.want.Description) {
						t.Errorf("CreateTask() Description = %v, want %v", got.Description, tt.want.Description)
					}
					if !reflect.DeepEqual(got.Status, tt.want.Status) {
						t.Errorf("CreateTask() Status = %v, want %v", got.Status, tt.want.Status)
					}
					if !reflect.DeepEqual(got.Rank, tt.want.Rank) {
						t.Errorf("CreateTask() Rank = %v, want %v", got.Rank, tt.want.Rank)
					}
					// if !reflect.DeepEqual(got.UserID, tt.want.UserID) {
					// 	t.Errorf("CreateTask() UserID = %v, want %v", got.UserID, tt.want.UserID)
					// }
					if !reflect.DeepEqual(got.ProjectID, tt.want.ProjectID) {
						t.Errorf("CreateTask() ProjectID = %v, want %v", got.ProjectID, tt.want.ProjectID)
					}
				}
			})
		}
		return test.ErrEndTest
	})
}

// func TestUpdateTask(t *testing.T) {
// 	test.Short(t)
// 	ctx, dbx := test.DbSetup()
// 	_ = dbx.RunInTx( func(dbxx database.Dbx) error {
// 		userStore := stores.NewPostgresUserStore(dbxx)
// 		user, err := userStore.CreateUser(ctx, &models.User{
// 			Email: "tkahng@gmail.com",
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create user: %v", err)
// 		}
// 		member, err := queries.CreateTeamFromUser(ctx, dbxx, user)
// 		if err != nil {
// 			t.Fatalf("failed to create team from user: %v", err)
// 		}
// 		taskProject, err := queries.CreateTaskProject(ctx, dbxx, &stores.CreateTaskProjectDTO{
// 			Name:     "Test Project",
// 			Status:   models.TaskProjectStatusDone,
// 			TeamID:   member.TeamID,
// 			MemberID: member.ID,
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create task project: %v", err)
// 		}
// 		task, err := queries.CreateTask(ctx, dbxx, taskProject.ID, &shared.CreateTaskBaseDTO{
// 			Name:        "Test Task",
// 			Description: types.Pointer("Test Description"),
// 			Status:      shared.TaskStatusDone,
// 			Rank:       1000,
// 			TeamID:      member.TeamID,
// 			CreatedBy:   member.ID,
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create task: %v", err)
// 		}

// 		type args struct {
// 			ctx    context.Context
// 			db     database.Dbx
// 			taskID uuid.UUID
// 			input  *shared.UpdateTaskBaseDTO
// 		}
// 		tests := []struct {
// 			name    string
// 			args    args
// 			wantErr bool
// 		}{
// 			{
// 				name: "update task successfully",
// 				args: args{
// 					ctx:    ctx,
// 					db:     dbxx,
// 					taskID: task.ID,
// 					input: &shared.UpdateTaskBaseDTO{
// 						Name:        "Updated Task",
// 						Description: types.Pointer("Updated Description"),
// 						Status:      shared.TaskStatusInProgress,
// 						Rank:       2000,
// 						ParentID:    nil,
// 					},
// 				},
// 				wantErr: false,
// 			},
// 			{
// 				name: "update non-existing task",
// 				args: args{
// 					ctx:    ctx,
// 					db:     dbxx,
// 					taskID: uuid.New(),
// 					input: &shared.UpdateTaskBaseDTO{
// 						Name:   "Updated Task",
// 						Status: shared.TaskStatusInProgress,
// 					},
// 				},
// 				wantErr: true,
// 			},
// 		}
// 		for _, tt := range tests {
// 			t.Run(tt.name, func(t *testing.T) {
// 				err := queries.UpdateTask(tt.args.ctx, tt.args.db, tt.args.taskID, tt.args.input)
// 				if (err != nil) != tt.wantErr {
// 					t.Errorf("UpdateTask() error = %v, wantErr %v", err, tt.wantErr)
// 					return
// 				}

// 				if !tt.wantErr {
// 					// Verify task was updated
// 					updatedTask, err := queries.FindTaskByID(tt.args.ctx, tt.args.db, tt.args.taskID)
// 					if err != nil {
// 						t.Errorf("Failed to get updated task: %v", err)
// 						return
// 					}
// 					if updatedTask.Name != tt.args.input.Name {
// 						t.Errorf("Task name not updated. got = %v, want %v", updatedTask.Name, tt.args.input.Name)
// 					}
// 					if *updatedTask.Description != *tt.args.input.Description {
// 						t.Errorf("Task description not updated. got = %v, want %v", *updatedTask.Description, *tt.args.input.Description)
// 					}
// 					if updatedTask.Status != models.TaskStatus(tt.args.input.Status) {
// 						t.Errorf("Task status not updated. got = %v, want %v", updatedTask.Status, tt.args.input.Status)
// 					}
// 					if updatedTask.Rank != tt.args.input.Rank {
// 						t.Errorf("Task order not updated. got = %v, want %v", updatedTask.Rank, tt.args.input.Rank)
// 					}
// 				}
// 			})
// 		}
// 		return test.EndTestErr
// 	})
// }
// func TestUpdateTaskProjectUpdateDate(t *testing.T) {
// 	test.Short(t)
// 	ctx, dbx := test.DbSetup()
// 	_ = dbx.RunInTx( func(dbxx database.Dbx) error {
// 		userStore := stores.NewPostgresUserStore(dbxx)
// 		user, err := userStore.CreateUser(ctx, &models.User{
// 			Email: "tkahng@gmail.com",
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create user: %v", err)
// 		}
// 		member, err := queries.CreateTeamFromUser(ctx, dbxx, user)
// 		if err != nil {
// 			t.Fatalf("failed to create team from user: %v", err)
// 		}
// 		taskProject, err := queries.CreateTaskProject(ctx, dbxx, &stores.CreateTaskProjectDTO{
// 			Name:     "Test Project",
// 			Status:   models.TaskProjectStatusDone,
// 			TeamID:   member.TeamID,
// 			MemberID: member.ID,
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create task project: %v", err)
// 		}

// 		type args struct {
// 			ctx           context.Context
// 			db            database.Dbx
// 			taskProjectID uuid.UUID
// 		}
// 		tests := []struct {
// 			name    string
// 			args    args
// 			wantErr bool
// 		}{
// 			{
// 				name: "update task project date successfully",
// 				args: args{
// 					ctx:           ctx,
// 					db:            dbxx,
// 					taskProjectID: taskProject.ID,
// 				},
// 				wantErr: false,
// 			},
// 			{
// 				name: "update non-existing task project date",
// 				args: args{
// 					ctx:           ctx,
// 					db:            dbxx,
// 					taskProjectID: uuid.New(),
// 				},
// 				wantErr: false,
// 			},
// 		}
// 		for _, tt := range tests {
// 			t.Run(tt.name, func(t *testing.T) {
// 				err := queries.UpdateTaskProjectUpdateDate(tt.args.ctx, tt.args.db, tt.args.taskProjectID)
// 				if (err != nil) != tt.wantErr {
// 					t.Errorf("UpdateTaskProjectUpdateDate() error = %v, wantErr %v", err, tt.wantErr)
// 					return
// 				}
// 			})
// 		}
// 		return test.EndTestErr
// 	})
// }
// func TestUpdateTaskProject(t *testing.T) {
// 	test.Short(t)
// 	ctx, dbx := test.DbSetup()
// 	_ = dbx.RunInTx( func(dbxx database.Dbx) error {
// 		userStore := stores.NewPostgresUserStore(dbxx)
// 		user, err := userStore.CreateUser(ctx, &models.User{
// 			Email: "tkahng@gmail.com",
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create user: %v", err)
// 		}
// 		member, err := queries.CreateTeamFromUser(ctx, dbxx, user)
// 		if err != nil {
// 			t.Fatalf("failed to create team from user: %v", err)
// 		}
// 		taskProject, err := queries.CreateTaskProject(ctx, dbxx, &stores.CreateTaskProjectDTO{
// 			Name:        "Test Project",
// 			Description: types.Pointer("Test Description"),
// 			Status:      models.TaskProjectStatusDone,
// 			Rank:       1000,
// 			TeamID:      member.TeamID,
// 			MemberID:    member.ID,
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create task project: %v", err)
// 		}

// 		type args struct {
// 			ctx           context.Context
// 			db            database.Dbx
// 			taskProjectID uuid.UUID
// 			input         *UpdateTaskProjectBaseDTO
// 		}
// 		tests := []struct {
// 			name    string
// 			args    args
// 			wantErr bool
// 		}{
// 			{
// 				name: "update task project successfully",
// 				args: args{
// 					ctx:           ctx,
// 					db:            dbxx,
// 					taskProjectID: taskProject.ID,
// 					input: &UpdateTaskProjectBaseDTO{
// 						Name:        "Updated Project",
// 						Description: types.Pointer("Updated Description"),
// 						Status:      shared.TaskProjectStatusInProgress,
// 						Rank:       2000,
// 					},
// 				},
// 				wantErr: false,
// 			},
// 			{
// 				name: "update non-existing task project",
// 				args: args{
// 					ctx:           ctx,
// 					db:            dbxx,
// 					taskProjectID: uuid.New(),
// 					input: &UpdateTaskProjectBaseDTO{
// 						Name:   "Updated Project",
// 						Status: shared.TaskProjectStatusInProgress,
// 					},
// 				},
// 				wantErr: true,
// 			},
// 		}
// 		for _, tt := range tests {
// 			t.Run(tt.name, func(t *testing.T) {
// 				err := queries.UpdateTaskProject(tt.args.ctx, tt.args.db, tt.args.taskProjectID, tt.args.input)
// 				if (err != nil) != tt.wantErr {
// 					t.Errorf("UpdateTaskProject() error = %v, wantErr %v", err, tt.wantErr)
// 					return
// 				}

// 				if !tt.wantErr {
// 					// Verify task project was updated
// 					updatedProject, err := queries.FindTaskProjectByID(tt.args.ctx, tt.args.db, tt.args.taskProjectID)
// 					if err != nil {
// 						t.Errorf("Failed to get updated task project: %v", err)
// 						return
// 					}
// 					if updatedProject.Name != tt.args.input.Name {
// 						t.Errorf("Task project name not updated. got = %v, want %v", updatedProject.Name, tt.args.input.Name)
// 					}
// 					if *updatedProject.Description != *tt.args.input.Description {
// 						t.Errorf("Task project description not updated. got = %v, want %v", *updatedProject.Description, *tt.args.input.Description)
// 					}
// 					if updatedProject.Status != models.TaskProjectStatus(tt.args.input.Status) {
// 						t.Errorf("Task project status not updated. got = %v, want %v", updatedProject.Status, tt.args.input.Status)
// 					}
// 					if updatedProject.Rank != tt.args.input.Rank {
// 						t.Errorf("Task project order not updated. got = %v, want %v", updatedProject.Rank, tt.args.input.Rank)
// 					}
// 				}
// 			})
// 		}
// 		return test.EndTestErr
// 	})
// }
// func TestUpdateTaskPositionStatus(t *testing.T) {
// 	test.Short(t)
// 	ctx, dbx := test.DbSetup()
// 	_ = dbx.RunInTx( func(dbxx database.Dbx) error {
// 		userStore := stores.NewPostgresUserStore(dbxx)
// 		user, err := userStore.CreateUser(ctx, &models.User{
// 			Email: "tkahng@gmail.com",
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create user: %v", err)
// 		}
// 		member, err := queries.CreateTeamFromUser(ctx, dbxx, user)
// 		if err != nil {
// 			t.Fatalf("failed to create team from user: %v", err)
// 		}
// 		taskProject, err := queries.CreateTaskProject(ctx, dbxx, &stores.CreateTaskProjectDTO{
// 			Name:     "Test Project",
// 			Status:   models.TaskProjectStatusDone,
// 			TeamID:   member.TeamID,
// 			MemberID: member.ID,
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create task project: %v", err)
// 		}

// 		task1, err := queries.CreateTask(ctx, dbxx, taskProject.ID, &shared.CreateTaskBaseDTO{
// 			Name:      "Task 1",
// 			Status:    shared.TaskStatusDone,
// 			Rank:     0,
// 			CreatedBy: member.ID,
// 			TeamID:    member.TeamID,
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create task: %v", err)
// 		}

// 		task2, err := queries.CreateTask(ctx, dbxx, taskProject.ID, &shared.CreateTaskBaseDTO{
// 			Name:      "Task 2",
// 			Status:    shared.TaskStatusDone,
// 			Rank:     1000,
// 			CreatedBy: member.ID,
// 			TeamID:    member.TeamID,
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create task: %v", err)
// 		}

// 		type args struct {
// 			ctx      context.Context
// 			db       database.Dbx
// 			taskID   uuid.UUID
// 			position int64
// 			status   models.TaskStatus
// 		}
// 		tests := []struct {
// 			name    string
// 			args    args
// 			wantErr bool
// 		}{
// 			{
// 				name: "update task position status successfully",
// 				args: args{
// 					ctx:      ctx,
// 					db:       dbxx,
// 					taskID:   task1.ID,
// 					position: 1,
// 					status:   models.TaskStatusDone,
// 				},
// 				wantErr: false,
// 			},
// 			{
// 				name: "update non-existing task position status",
// 				args: args{
// 					ctx:      ctx,
// 					db:       dbxx,
// 					taskID:   uuid.New(),
// 					position: 0,
// 					status:   models.TaskStatusDone,
// 				},
// 				wantErr: true,
// 			},
// 			{
// 				name: "move task to first position",
// 				args: args{
// 					ctx:      ctx,
// 					db:       dbxx,
// 					taskID:   task2.ID,
// 					position: 0,
// 					status:   models.TaskStatusDone,
// 				},
// 				wantErr: false,
// 			},
// 		}
// 		for _, tt := range tests {
// 			t.Run(tt.name, func(t *testing.T) {
// 				err := queries.UpdateTaskPositionStatus(tt.args.ctx, tt.args.db, tt.args.taskID, tt.args.position, tt.args.status)
// 				if (err != nil) != tt.wantErr {
// 					t.Errorf("UpdateTaskPositionStatus() error = %v, wantErr %v", err, tt.wantErr)
// 					return
// 				}

// 				if !tt.wantErr {
// 					// Verify task was updated
// 					updatedTask, err := queries.FindTaskByID(tt.args.ctx, tt.args.db, tt.args.taskID)
// 					if err != nil {
// 						t.Errorf("Failed to get updated task: %v", err)
// 						return
// 					}

// 					if updatedTask.Status != tt.args.status {
// 						t.Errorf("Task status not updated. got = %v, want %v", updatedTask.Status, tt.args.status)
// 					}

// 					// Get task project to verify update date
// 					taskProject, err := queries.FindTaskProjectByID(tt.args.ctx, tt.args.db, updatedTask.ProjectID)
// 					if err != nil {
// 						t.Errorf("Failed to get task project: %v", err)
// 						return
// 					}

// 					if taskProject.UpdatedAt.IsZero() {
// 						t.Error("Task project update date not updated")
// 					}
// 				}
// 			})
// 		}
// 		return test.EndTestErr
// 	})
// }

// func TestLoadTaskProjectsTasks(t *testing.T) {
// 	test.Short(t)
// 	ctx, dbx := test.DbSetup()
// 	_ = dbx.RunInTx( func(dbxx database.Dbx) error {
// 		user, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
// 			Email: "tkahng@gmail.com",
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create user: %v", err)
// 		}

// 		member, err := queries.CreateTeamFromUser(ctx, dbxx, user)
// 		if err != nil {
// 			t.Fatalf("failed to create team from user: %v", err)
// 		}
// 		taskProject, err := queries.CreateTaskProject(ctx, dbxx, &stores.CreateTaskProjectDTO{
// 			Name:     "Test Project",
// 			Status:   models.TaskProjectStatusDone,
// 			TeamID:   member.TeamID,
// 			MemberID: member.ID,
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create task project: %v", err)
// 		}
// 		tasks, err := queries.CreateTask(ctx, dbxx, taskProject.ID, &shared.CreateTaskBaseDTO{
// 			Name:      "Test Task",
// 			Status:    shared.TaskStatusDone,
// 			CreatedBy: member.ID,
// 			TeamID:    member.TeamID,
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create task: %v", err)
// 		}
// 		type args struct {
// 			ctx        context.Context
// 			db         database.Dbx
// 			projectIds []uuid.UUID
// 		}
// 		tests := []struct {
// 			name    string
// 			args    args
// 			want    [][]*models.Task
// 			wantErr bool
// 		}{
// 			{
// 				name: "query tasks",
// 				args: args{
// 					ctx: ctx,
// 					db:  dbxx,
// 					projectIds: []uuid.UUID{
// 						taskProject.ID,
// 					},
// 				},
// 				want: [][]*models.Task{
// 					{
// 						{
// 							ID:        tasks.ID,
// 							Name:      tasks.Name,
// 							Status:    tasks.Status,
// 							ProjectID: tasks.ProjectID,
// 							CreatedBy: tasks.CreatedBy,
// 							TeamID:    tasks.TeamID,
// 							CreatedAt: tasks.CreatedAt,
// 							UpdatedAt: tasks.UpdatedAt,
// 						},
// 					},
// 				},
// 				wantErr: false,
// 			},
// 		}
// 		for _, tt := range tests {
// 			t.Run(tt.name, func(t *testing.T) {
// 				got, err := queries.LoadTaskProjectsTasks(tt.args.ctx, tt.args.db, tt.args.projectIds...)
// 				if (err != nil) != tt.wantErr {
// 					t.Errorf("LoadTaskProjectsTasks() error = %v, wantErr %v", err, tt.wantErr)
// 					return
// 				}
// 				if !reflect.DeepEqual(got[0][0].Name, tt.want[0][0].Name) {
// 					t.Errorf("LoadTaskProjectsTasks() = %v, want %v", got[0][0].Name, tt.want[0][0].Name)
// 				}
// 				if !reflect.DeepEqual(got[0][0].Status, tt.want[0][0].Status) {
// 					t.Errorf("LoadTaskProjectsTasks() = %v, want %v", got[0][0].Status, tt.want[0][0].Status)
// 				}
// 				if !reflect.DeepEqual(got[0][0].ProjectID, tt.want[0][0].ProjectID) {
// 					t.Errorf("LoadTaskProjectsTasks() = %v, want %v", got[0][0].ProjectID, tt.want[0][0].ProjectID)
// 				}
// 			})
// 		}
// 		return test.EndTestErr
// 	})
// }
// func TestFindTaskByID(t *testing.T) {
// 	test.Short(t)
// 	ctx, dbx := test.DbSetup()
// 	_ = dbx.RunInTx( func(dbxx database.Dbx) error {
// 		user, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
// 			Email: "tkahng@gmail.com",
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create user: %v", err)
// 		}
// 		member, err := queries.CreateTeamFromUser(ctx, dbxx, user)
// 		if err != nil {
// 			t.Fatalf("failed to create team from user: %v", err)
// 		}
// 		if member == nil {
// 			t.Fatalf("failed to create team member")
// 		}
// 		taskProject, err := queries.CreateTaskProject(ctx, dbxx, &stores.CreateTaskProjectDTO{
// 			Name:     "Test Project",
// 			Status:   models.TaskProjectStatusDone,
// 			TeamID:   member.TeamID,
// 			MemberID: member.ID,
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create task project: %v", err)
// 		}
// 		task, err := queries.CreateTask(ctx, dbxx, taskProject.ID, &shared.CreateTaskBaseDTO{
// 			Name:      "Test Task",
// 			Status:    shared.TaskStatusDone,
// 			CreatedBy: member.ID,
// 			TeamID:    member.TeamID,
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create task: %v", err)
// 		}

// 		type args struct {
// 			ctx context.Context
// 			db  database.Dbx
// 			id  uuid.UUID
// 		}
// 		tests := []struct {
// 			name    string
// 			args    args
// 			want    *models.Task
// 			wantErr bool
// 		}{
// 			{
// 				name: "find existing task",
// 				args: args{
// 					ctx: ctx,
// 					db:  dbxx,
// 					id:  task.ID,
// 				},
// 				want:    task,
// 				wantErr: false,
// 			},
// 			{
// 				name: "find non-existing task",
// 				args: args{
// 					ctx: ctx,
// 					db:  dbxx,
// 					id:  uuid.New(),
// 				},
// 				want:    nil,
// 				wantErr: false,
// 			},
// 		}
// 		for _, tt := range tests {
// 			t.Run(tt.name, func(t *testing.T) {
// 				got, err := queries.FindTaskByID(tt.args.ctx, tt.args.db, tt.args.id)
// 				if (err != nil) != tt.wantErr {
// 					t.Errorf("FindTaskByID() error = %v, wantErr %v", err, tt.wantErr)
// 					return
// 				}
// 				if tt.want == nil {
// 					if got != nil {
// 						t.Errorf("FindTaskByID() = %v, want nil", got)
// 					}
// 					return
// 				}
// 				if !reflect.DeepEqual(got.ID, tt.want.ID) {
// 					t.Errorf("FindTaskByID() = %v, want %v", got.ID, tt.want.ID)
// 				}
// 				if !reflect.DeepEqual(got.Name, tt.want.Name) {
// 					t.Errorf("FindTaskByID() = %v, want %v", got.Name, tt.want.Name)
// 				}
// 			})
// 		}
// 		return test.EndTestErr
// 	})
// }
// func TestFindLastTaskOrder(t *testing.T) {
// 	test.Short(t)
// 	ctx, dbx := test.DbSetup()
// 	_ = dbx.RunInTx( func(dbxx database.Dbx) error {
// 		user, err := queries.CreateUser(
// 			ctx,
// 			dbxx,
// 			&shared.AuthenticationInput{
// 				Email: "tkahng@gmail.com",
// 			},
// 		)
// 		if err != nil {
// 			t.Fatalf("failed to create user: %v", err)
// 		}
// 		member, err := queries.CreateTeamFromUser(ctx, dbxx, user)
// 		if err != nil {
// 			t.Fatalf("failed to create team from user: %v", err)
// 		}
// 		taskProject, err := queries.CreateTaskProject(
// 			ctx,
// 			dbxx,
// 			&stores.CreateTaskProjectDTO{
// 				Name:     "Test Project",
// 				Status:   models.TaskProjectStatusDone,
// 				TeamID:   member.TeamID,
// 				MemberID: member.ID,
// 			},
// 		)
// 		if err != nil {
// 			t.Fatalf("failed to create task project: %v", err)
// 		}

// 		_, err = queries.CreateTask(
// 			ctx,
// 			dbxx,
// 			taskProject.ID,
// 			&shared.CreateTaskBaseDTO{
// 				Name:   "Test Task 1",
// 				Status: shared.TaskStatusDone,
// 				Rank:  1000,
// 			},
// 		)
// 		if err != nil {
// 			t.Fatalf("failed to create task: %v", err)
// 		}

// 		type args struct {
// 			ctx           context.Context
// 			db            database.Dbx
// 			taskProjectID uuid.UUID
// 		}
// 		tests := []struct {
// 			name    string
// 			args    args
// 			want    float64
// 			wantErr bool
// 		}{
// 			{
// 				name: "find last order with existing tasks",
// 				args: args{
// 					ctx:           ctx,
// 					db:            dbxx,
// 					taskProjectID: taskProject.ID,
// 				},
// 				want:    2000,
// 				wantErr: false,
// 			},
// 			{
// 				name: "find last order with non-existing project",
// 				args: args{
// 					ctx:           ctx,
// 					db:            dbxx,
// 					taskProjectID: uuid.New(),
// 				},
// 				want:    0,
// 				wantErr: false,
// 			},
// 		}
// 		for _, tt := range tests {
// 			t.Run(tt.name, func(t *testing.T) {
// 				got, err := queries.FindLastTaskOrder(tt.args.ctx, tt.args.db, tt.args.taskProjectID)
// 				if (err != nil) != tt.wantErr {
// 					t.Errorf("FindLastTaskOrder() error = %v, wantErr %v", err, tt.wantErr)
// 					return
// 				}
// 				if got != tt.want {
// 					t.Errorf("FindLastTaskOrder() = %v, want %v", got, tt.want)
// 				}
// 			})
// 		}
// 		return test.EndTestErr
// 	})
// }
// func TestDeleteTask(t *testing.T) {
// 	test.Short(t)
// 	ctx, dbx := test.DbSetup()
// 	_ = dbx.RunInTx( func(dbxx database.Dbx) error {
// 		user, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
// 			Email: "tkahng@gmail.com",
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create user: %v", err)
// 		}
// 		member, err := queries.CreateTeamFromUser(ctx, dbxx, user)
// 		if err != nil {
// 			t.Fatalf("failed to create team from user: %v", err)
// 		}
// 		taskProject, err := queries.CreateTaskProject(ctx, dbxx, &stores.CreateTaskProjectDTO{
// 			Name:     "Test Project",
// 			Status:   models.TaskProjectStatusDone,
// 			TeamID:   member.TeamID,
// 			MemberID: member.ID,
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create task project: %v", err)
// 		}
// 		task, err := queries.CreateTask(ctx, dbxx, taskProject.ID, &shared.CreateTaskBaseDTO{
// 			Name:   "Test Task",
// 			Status: shared.TaskStatusDone,
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create task: %v", err)
// 		}

// 		type args struct {
// 			ctx    context.Context
// 			db     database.Dbx
// 			taskID uuid.UUID
// 		}
// 		tests := []struct {
// 			name    string
// 			args    args
// 			wantErr bool
// 		}{
// 			{
// 				name: "delete existing task",
// 				args: args{
// 					ctx:    ctx,
// 					db:     dbxx,
// 					taskID: task.ID,
// 				},
// 				wantErr: false,
// 			},
// 			{
// 				name: "delete non-existing task",
// 				args: args{
// 					ctx:    ctx,
// 					db:     dbxx,
// 					taskID: uuid.New(),
// 				},
// 				wantErr: false,
// 			},
// 		}
// 		for _, tt := range tests {
// 			t.Run(tt.name, func(t *testing.T) {
// 				if err := queries.DeleteTask(tt.args.ctx, tt.args.db, tt.args.taskID); (err != nil) != tt.wantErr {
// 					t.Errorf("DeleteTask() error = %v, wantErr %v", err, tt.wantErr)
// 				}
// 			})
// 		}
// 		return test.EndTestErr
// 	})
// }
// func TestFindTaskProjectByID(t *testing.T) {
// 	test.Short(t)
// 	ctx, dbx := test.DbSetup()
// 	_ = dbx.RunInTx( func(dbxx database.Dbx) error {
// 		user, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
// 			Email: "tkahng@gmail.com",
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create user: %v", err)
// 		}
// 		member, err := queries.CreateTeamFromUser(ctx, dbxx, user)
// 		if err != nil {
// 			t.Fatalf("failed to create team from user: %v", err)
// 		}
// 		taskProject, err := queries.CreateTaskProject(ctx, dbxx, &stores.CreateTaskProjectDTO{
// 			Name:     "Test Project",
// 			Status:   models.TaskProjectStatusDone,
// 			TeamID:   member.TeamID,
// 			MemberID: member.ID,
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create task project: %v", err)
// 		}

// 		type args struct {
// 			ctx context.Context
// 			db  database.Dbx
// 			id  uuid.UUID
// 		}
// 		tests := []struct {
// 			name    string
// 			args    args
// 			want    *models.TaskProject
// 			wantErr bool
// 		}{
// 			{
// 				name: "find existing task project",
// 				args: args{
// 					ctx: ctx,
// 					db:  dbxx,
// 					id:  taskProject.ID,
// 				},
// 				want:    taskProject,
// 				wantErr: false,
// 			},
// 			{
// 				name: "find non-existing task project",
// 				args: args{
// 					ctx: ctx,
// 					db:  dbxx,
// 					id:  uuid.New(),
// 				},
// 				want:    nil,
// 				wantErr: false,
// 			},
// 		}
// 		for _, tt := range tests {
// 			t.Run(tt.name, func(t *testing.T) {
// 				got, err := queries.FindTaskProjectByID(tt.args.ctx, tt.args.db, tt.args.id)
// 				if (err != nil) != tt.wantErr {
// 					t.Errorf("FindTaskProjectByID() error = %v, wantErr %v", err, tt.wantErr)
// 					return
// 				}
// 				if tt.want == nil {
// 					if got != nil {
// 						t.Errorf("FindTaskProjectByID() = %v, want nil", got)
// 					}
// 					return
// 				}
// 				if !reflect.DeepEqual(got.ID, tt.want.ID) {
// 					t.Errorf("FindTaskProjectByID() = %v, want %v", got.ID, tt.want.ID)
// 				}
// 				if !reflect.DeepEqual(got.Name, tt.want.Name) {
// 					t.Errorf("FindTaskProjectByID() = %v, want %v", got.Name, tt.want.Name)
// 				}
// 			})
// 		}
// 		return test.EndTestErr
// 	})
// }
// func TestDeleteTaskProject(t *testing.T) {
// 	test.Short(t)
// 	ctx, dbx := test.DbSetup()
// 	_ = dbx.RunInTx( func(dbxx database.Dbx) error {
// 		user, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
// 			Email: "tkahng@gmail.com",
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create user: %v", err)
// 		}
// 		member, err := queries.CreateTeamFromUser(ctx, dbxx, user)
// 		if err != nil {
// 			t.Fatalf("failed to create team from user: %v", err)
// 		}
// 		taskProject, err := queries.CreateTaskProject(ctx, dbxx, &stores.CreateTaskProjectDTO{
// 			Name:     "Test Project",
// 			Status:   models.TaskProjectStatusDone,
// 			TeamID:   member.TeamID,
// 			MemberID: member.ID,
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create task project: %v", err)
// 		}

// 		type args struct {
// 			ctx           context.Context
// 			db            database.Dbx
// 			taskProjectID uuid.UUID
// 		}
// 		tests := []struct {
// 			name    string
// 			args    args
// 			wantErr bool
// 		}{
// 			{
// 				name: "delete existing task project",
// 				args: args{
// 					ctx:           ctx,
// 					db:            dbxx,
// 					taskProjectID: taskProject.ID,
// 				},
// 				wantErr: false,
// 			},
// 			{
// 				name: "delete non-existing task project",
// 				args: args{
// 					ctx:           ctx,
// 					db:            dbxx,
// 					taskProjectID: uuid.New(),
// 				},
// 				wantErr: false,
// 			},
// 		}
// 		for _, tt := range tests {
// 			t.Run(tt.name, func(t *testing.T) {
// 				if err := queries.DeleteTaskProject(tt.args.ctx, tt.args.db, tt.args.taskProjectID); (err != nil) != tt.wantErr {
// 					t.Errorf("DeleteTaskProject() error = %v, wantErr %v", err, tt.wantErr)
// 				}
// 			})
// 		}
// 		return test.EndTestErr
// 	})
// }
// func TestListTasks(t *testing.T) {
// 	test.Short(t)
// 	ctx, dbx := test.DbSetup()
// 	_ = dbx.RunInTx( func(dbxx database.Dbx) error {
// 		user, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
// 			Email: "tkahng@gmail.com",
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create user: %v", err)
// 		}
// 		member, err := queries.CreateTeamFromUser(ctx, dbxx, user)
// 		if err != nil {
// 			t.Fatalf("failed to create team from user: %v", err)
// 		}
// 		taskProject, err := queries.CreateTaskProject(ctx, dbxx, &stores.CreateTaskProjectDTO{
// 			Name:     "Test Project",
// 			Status:   models.TaskProjectStatusDone,
// 			TeamID:   member.TeamID,
// 			MemberID: member.ID,
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create task project: %v", err)
// 		}
// 		task, err := queries.CreateTask(ctx, dbxx, taskProject.ID, &shared.CreateTaskBaseDTO{
// 			Name:        "Test Task",
// 			Description: nil,
// 			Status:      shared.TaskStatusDone,
// 			TeamID:      member.TeamID,
// 			CreatedBy:   member.ID,
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create task: %v", err)
// 		}

// 		type args struct {
// 			ctx   context.Context
// 			db    database.Dbx
// 			input *stores.TaskFilter
// 		}
// 		tests := []struct {
// 			name      string
// 			args      args
// 			wantCount int
// 			wantErr   bool
// 		}{
// 			{
// 				name: "list tasks with filter",
// 				args: args{
// 					ctx: ctx,
// 					db:  dbxx,
// 					input: &stores.TaskFilter{
// 						TaskListFilter: shared.TaskListFilter{
// 							ProjectID: taskProject.ID.String(),
// 							Status: []shared.TaskStatus{
// 								shared.TaskStatusDone,
// 							},
// 						},
// 						PaginatedInput: shared.PaginatedInput{
// 							Page:    0,
// 							PerPage: 10,
// 						},
// 					},
// 				},
// 				wantCount: 1,
// 				wantErr:   false,
// 			},
// 			{
// 				name: "list tasks without filter",
// 				args: args{
// 					ctx: ctx,
// 					db:  dbxx,
// 					input: &stores.TaskFilter{
// 						PaginatedInput: shared.PaginatedInput{
// 							Page:    0,
// 							PerPage: 10,
// 						},
// 					},
// 				},
// 				wantCount: 1,
// 				wantErr:   false,
// 			},
// 		}
// 		for _, tt := range tests {
// 			t.Run(tt.name, func(t *testing.T) {
// 				got, err := queries.ListTasks(
// 					tt.args.ctx,
// 					tt.args.db,
// 					tt.args.input,
// 				)
// 				if (err != nil) != tt.wantErr {
// 					t.Errorf(
// 						"ListTasks() error = %v, wantErr %v",
// 						err,
// 						tt.wantErr,
// 					)
// 					return
// 				}
// 				if len(got) != tt.wantCount {
// 					t.Errorf("ListTasks() got length = %v, want length %v", len(got), tt.wantCount)
// 					return
// 				}
// 				if len(got) > 0 {
// 					if !reflect.DeepEqual(got[0].ID, task.ID) {
// 						t.Errorf("ListTasks() = %v, want %v", got[0].ID, task.ID)
// 					}
// 					if !reflect.DeepEqual(got[0].Name, task.Name) {
// 						t.Errorf("ListTasks() = %v, want %v", got[0].Name, task.Name)
// 					}
// 				}
// 			})
// 		}
// 		return test.EndTestErr
// 	})
// }
// func TestCountTasks(t *testing.T) {
// 	test.Short(t)
// 	ctx, dbx := test.DbSetup()
// 	_ = dbx.RunInTx( func(dbxx database.Dbx) error {
// 		user, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
// 			Email: "tkahng@gmail.com",
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create user: %v", err)
// 		}
// 		member, err := queries.CreateTeamFromUser(ctx, dbxx, user)
// 		if err != nil {
// 			t.Fatalf("failed to create team from user: %v", err)
// 		}
// 		taskProject, err := queries.CreateTaskProject(ctx, dbxx, &stores.CreateTaskProjectDTO{
// 			Name:     "Test Project",
// 			Status:   models.TaskProjectStatusDone,
// 			TeamID:   member.TeamID,
// 			MemberID: member.ID,
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create task project: %v", err)
// 		}
// 		_, err = queries.CreateTask(ctx, dbxx, taskProject.ID, &shared.CreateTaskBaseDTO{
// 			Name:      "Test Task",
// 			Status:    shared.TaskStatusDone,
// 			CreatedBy: member.ID,
// 			TeamID:    member.TeamID,
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create task: %v", err)
// 		}

// 		type args struct {
// 			ctx    context.Context
// 			db     database.Dbx
// 			filter *shared.TaskListFilter
// 		}
// 		tests := []struct {
// 			name    string
// 			args    args
// 			want    int64
// 			wantErr bool
// 		}{
// 			{
// 				name: "count tasks with filter",
// 				args: args{
// 					ctx: ctx,
// 					db:  dbxx,
// 					filter: &shared.TaskListFilter{
// 						ProjectID: taskProject.ID.String(),
// 						Status: []shared.TaskStatus{
// 							shared.TaskStatusDone,
// 						},
// 					},
// 				},
// 				want:    1,
// 				wantErr: false,
// 			},
// 			{
// 				name: "count tasks without filter",
// 				args: args{
// 					ctx:    ctx,
// 					db:     dbxx,
// 					filter: nil,
// 				},
// 				want:    1,
// 				wantErr: false,
// 			},
// 		}
// 		for _, tt := range tests {
// 			t.Run(tt.name, func(t *testing.T) {
// 				got, err := queries.CountTasks(tt.args.ctx, tt.args.db, tt.args.filter)
// 				if (err != nil) != tt.wantErr {
// 					t.Errorf("CountTasks() error = %v, wantErr %v", err, tt.wantErr)
// 					return
// 				}
// 				if got != tt.want {
// 					t.Errorf("CountTasks() = %v, want %v", got, tt.want)
// 				}
// 			})
// 		}
// 		return test.EndTestErr
// 	})
// }
// func TestListTaskProjects(t *testing.T) {
// 	test.Short(t)
// 	ctx, dbx := test.DbSetup()
// 	_ = dbx.RunInTx( func(dbxx database.Dbx) error {
// 		user, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
// 			Email: "tkahng@gmail.com",
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create user: %v", err)
// 		}
// 		taskProject, err := queries.CreateTaskProject(ctx, dbxx, &stores.CreateTaskProjectDTO{
// 			Name:   "Test Project",
// 			Status: models.TaskProjectStatusDone,
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create task project: %v", err)
// 		}

// 		type args struct {
// 			ctx   context.Context
// 			db    database.Dbx
// 			input *shared.TaskProjectsListParams
// 		}
// 		tests := []struct {
// 			name      string
// 			args      args
// 			wantCount int
// 			wantErr   bool
// 		}{
// 			{
// 				name: "list task projects with filter",
// 				args: args{
// 					ctx: ctx,
// 					db:  dbxx,
// 					input: &shared.TaskProjectsListParams{
// 						TaskProjectsListFilter: shared.TaskProjectsListFilter{
// 							UserID: user.ID.String(),
// 						},
// 						PaginatedInput: shared.PaginatedInput{
// 							Page:    0,
// 							PerPage: 10,
// 						},
// 					},
// 				},
// 				wantCount: 1,
// 				wantErr:   false,
// 			},
// 			{
// 				name: "list task projects without filter",
// 				args: args{
// 					ctx: ctx,
// 					db:  dbxx,
// 					input: &shared.TaskProjectsListParams{
// 						PaginatedInput: shared.PaginatedInput{
// 							Page:    0,
// 							PerPage: 10,
// 						},
// 					},
// 				},
// 				wantCount: 1,
// 				wantErr:   false,
// 			},
// 		}
// 		for _, tt := range tests {
// 			t.Run(tt.name, func(t *testing.T) {
// 				got, err := queries.ListTaskProjects(
// 					tt.args.ctx,
// 					tt.args.db,
// 					tt.args.input,
// 				)
// 				if (err != nil) != tt.wantErr {
// 					t.Errorf(
// 						"ListTaskProjects() error = %v, wantErr %v",
// 						err,
// 						tt.wantErr,
// 					)
// 					return
// 				}
// 				if len(got) != tt.wantCount {
// 					t.Errorf("ListTaskProjects() got length = %v, want length %v", len(got), tt.wantCount)
// 					return
// 				}
// 				if len(got) > 0 {
// 					if !reflect.DeepEqual(got[0].ID, taskProject.ID) {
// 						t.Errorf("ListTaskProjects() = %v, want %v", got[0].ID, taskProject.ID)
// 					}
// 					if !reflect.DeepEqual(got[0].Name, taskProject.Name) {
// 						t.Errorf("ListTaskProjects() = %v, want %v", got[0].Name, taskProject.Name)
// 					}
// 				}
// 			})
// 		}
// 		return test.EndTestErr
// 	})
// }
// func TestCountTaskProjects(t *testing.T) {
// 	test.Short(t)
// 	ctx, dbx := test.DbSetup()
// 	_ = dbx.RunInTx( func(dbxx database.Dbx) error {
// 		user, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
// 			Email: "tkahng@gmail.com",
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create user: %v", err)
// 		}
// 		_, err = queries.CreateTaskProject(ctx, dbxx, &stores.CreateTaskProjectDTO{
// 			Name:   "Test Project",
// 			Status: models.TaskProjectStatusDone,
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create task project: %v", err)
// 		}

// 		type args struct {
// 			ctx    context.Context
// 			db     database.Dbx
// 			filter *shared.TaskProjectsListFilter
// 		}
// 		tests := []struct {
// 			name    string
// 			args    args
// 			want    int64
// 			wantErr bool
// 		}{
// 			{
// 				name: "count task projects with filter",
// 				args: args{
// 					ctx: ctx,
// 					db:  dbxx,
// 					filter: &shared.TaskProjectsListFilter{
// 						UserID: user.ID.String(),
// 					},
// 				},
// 				want:    1,
// 				wantErr: false,
// 			},
// 			{
// 				name: "count task projects without filter",
// 				args: args{
// 					ctx:    ctx,
// 					db:     dbxx,
// 					filter: nil,
// 				},
// 				want:    1,
// 				wantErr: false,
// 			},
// 		}
// 		for _, tt := range tests {
// 			t.Run(tt.name, func(t *testing.T) {
// 				got, err := queries.CountTaskProjects(tt.args.ctx, tt.args.db, tt.args.filter)
// 				if (err != nil) != tt.wantErr {
// 					t.Errorf("CountTaskProjects() error = %v, wantErr %v", err, tt.wantErr)
// 					return
// 				}
// 				if got != tt.want {
// 					t.Errorf("CountTaskProjects() = %v, want %v", got, tt.want)
// 				}
// 			})
// 		}
// 		return test.EndTestErr
// 	})
// }
// func TestCreateTaskProject(t *testing.T) {
// 	test.Short(t)
// 	ctx, dbx := test.DbSetup()
// 	_ = dbx.RunInTx( func(dbxx database.Dbx) error {
// 		user, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
// 			Email: "tkahng@gmail.com",
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create user: %v", err)
// 		}
// 		member, err := queries.CreateTeamFromUser(ctx, dbxx, user)
// 		if err != nil {
// 			t.Fatalf("failed to create team from user: %v", err)
// 		}
// 		type args struct {
// 			ctx    context.Context
// 			db     database.Dbx
// 			userID uuid.UUID
// 			input  *stores.CreateTaskProjectDTO
// 		}
// 		tests := []struct {
// 			name    string
// 			args    args
// 			want    *models.TaskProject
// 			wantErr bool
// 		}{
// 			{
// 				name: "create task project successfully",
// 				args: args{
// 					ctx:    ctx,
// 					db:     dbxx,
// 					userID: user.ID,
// 					input: &stores.CreateTaskProjectDTO{
// 						Name:        "Test Project",
// 						Description: types.Pointer("Test Description"),
// 						Status:      models.TaskProjectStatusDone,
// 						Rank:       1000,
// 						TeamID:      member.TeamID,
// 						MemberID:    member.ID,
// 					},
// 				},
// 				want: &models.TaskProject{
// 					// UserID:      user.ID,
// 					Name:        "Test Project",
// 					Description: types.Pointer("Test Description"),
// 					Status:      models.TaskProjectStatusDone,
// 					Rank:       1000,
// 					TeamID:      member.TeamID,
// 					CreatedBy:   member.ID,
// 				},
// 				wantErr: false,
// 			},
// 		}
// 		for _, tt := range tests {
// 			t.Run(tt.name, func(t *testing.T) {
// 				got, err := queries.CreateTaskProject(tt.args.ctx, tt.args.db, tt.args.input)
// 				if (err != nil) != tt.wantErr {
// 					t.Errorf("CreateTaskProject() error = %v, wantErr %v", err, tt.wantErr)
// 					return
// 				}
// 				if tt.want != nil {
// 					if !reflect.DeepEqual(got.Name, tt.want.Name) {
// 						t.Errorf("CreateTaskProject() Name = %v, want %v", got.Name, tt.want.Name)
// 					}
// 					if !reflect.DeepEqual(got.Description, tt.want.Description) {
// 						t.Errorf("CreateTaskProject() Description = %v, want %v", got.Description, tt.want.Description)
// 					}
// 					if !reflect.DeepEqual(got.Status, tt.want.Status) {
// 						t.Errorf("CreateTaskProject() Status = %v, want %v", got.Status, tt.want.Status)
// 					}
// 					if !reflect.DeepEqual(got.Rank, tt.want.Rank) {
// 						t.Errorf("CreateTaskProject() Rank = %v, want %v", got.Rank, tt.want.Rank)
// 					}
// 					// if !reflect.DeepEqual(got.UserID, tt.want.UserID) {
// 					// 	t.Errorf("CreateTaskProject() UserID = %v, want %v", got.UserID, tt.want.UserID)
// 					// }
// 				}
// 			})
// 		}
// 		return test.EndTestErr
// 	})
// }
// func TestCreateTaskProjectWithTasks(t *testing.T) {
// 	test.Short(t)
// 	ctx, dbx := test.DbSetup()
// 	_ = dbx.RunInTx( func(dbxx database.Dbx) error {
// 		user, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
// 			Email: "tkahng@gmail.com",
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create user: %v", err)
// 		}

// 		type args struct {
// 			ctx    context.Context
// 			db     database.Dbx
// 			userID uuid.UUID
// 			input  *shared.CreateTaskProjectWithTasksDTO
// 		}
// 		tests := []struct {
// 			name    string
// 			args    args
// 			want    *models.TaskProject
// 			wantErr bool
// 		}{
// 			{
// 				name: "create task project with tasks successfully",
// 				args: args{
// 					ctx:    ctx,
// 					db:     dbxx,
// 					userID: user.ID,
// 					input: &shared.CreateTaskProjectWithTasksDTO{
// 						CreateTaskProjectDTO: stores.CreateTaskProjectDTO{
// 							Name:        "Test Project",
// 							Description: types.Pointer("Test Description"),
// 							Status:      models.TaskProjectStatusDone,
// 						},
// 						Tasks: []shared.CreateTaskBaseDTO{
// 							{
// 								Name:        "Test Task 1",
// 								Description: types.Pointer("Test Description 1"),
// 								Status:      shared.TaskStatusDone,
// 							},
// 							{
// 								Name:        "Test Task 2",
// 								Description: types.Pointer("Test Description 2"),
// 								Status:      shared.TaskStatusDone,
// 							},
// 						},
// 					},
// 				},
// 				want: &models.TaskProject{

// 					Name:        "Test Project",
// 					Description: types.Pointer("Test Description"),
// 					Status:      models.TaskProjectStatusDone,
// 				},
// 				wantErr: false,
// 			},
// 		}
// 		for _, tt := range tests {
// 			t.Run(tt.name, func(t *testing.T) {
// 				got, err := queries.CreateTaskProjectWithTasks(tt.args.ctx, tt.args.db, tt.args.input)
// 				if (err != nil) != tt.wantErr {
// 					t.Errorf("CreateTaskProjectWithTasks() error = %v, wantErr %v", err, tt.wantErr)
// 					return
// 				}
// 				if tt.want != nil {
// 					if !reflect.DeepEqual(got.Name, tt.want.Name) {
// 						t.Errorf("CreateTaskProjectWithTasks() Name = %v, want %v", got.Name, tt.want.Name)
// 					}
// 					if !reflect.DeepEqual(got.Description, tt.want.Description) {
// 						t.Errorf("CreateTaskProjectWithTasks() Description = %v, want %v", got.Description, tt.want.Description)
// 					}
// 					if !reflect.DeepEqual(got.Status, tt.want.Status) {
// 						t.Errorf("CreateTaskProjectWithTasks() Status = %v, want %v", got.Status, tt.want.Status)
// 					}
// 					// if !reflect.DeepEqual(got.UserID, tt.want.UserID) {
// 					// 	t.Errorf("CreateTaskProjectWithTasks() UserID = %v, want %v", got.UserID, tt.want.UserID)
// 					// }

// 					// Verify tasks were created
// 					tasks, err := queries.ListTasks(tt.args.ctx, tt.args.db, &stores.TaskFilter{
// 						TaskListFilter: shared.TaskListFilter{
// 							ProjectID: got.ID.String(),
// 						},
// 					})
// 					if err != nil {
// 						t.Errorf("Failed to list tasks: %v", err)
// 					}
// 					if len(tasks) != len(tt.args.input.Tasks) {
// 						t.Errorf("Expected %d tasks, got %d", len(tt.args.input.Tasks), len(tasks))
// 					}
// 				}
// 			})
// 		}
// 		return test.EndTestErr
// 	})
// }
// func TestCreateTask(t *testing.T) {
// 	test.Short(t)
// 	ctx, dbx := test.DbSetup()
// 	_ = dbx.RunInTx( func(dbxx database.Dbx) error {
// 		user, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
// 			Email: "tkahng@gmail.com",
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create user: %v", err)
// 		}
// 		member, err := queries.CreateTeamFromUser(ctx, dbxx, user)
// 		if err != nil {
// 			t.Fatalf("failed to create team from user: %v", err)
// 		}
// 		taskProject, err := queries.CreateTaskProject(ctx, dbxx, &stores.CreateTaskProjectDTO{
// 			Name:     "Test Project",
// 			Status:   models.TaskProjectStatusDone,
// 			TeamID:   member.TeamID,
// 			MemberID: member.ID,
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create task project: %v", err)
// 		}

// 		type args struct {
// 			ctx       context.Context
// 			db        database.Dbx
// 			userID    uuid.UUID
// 			projectID uuid.UUID
// 			input     *shared.CreateTaskBaseDTO
// 		}
// 		tests := []struct {
// 			name    string
// 			args    args
// 			want    *models.Task
// 			wantErr bool
// 		}{
// 			{
// 				name: "create task successfully",
// 				args: args{
// 					ctx:       ctx,
// 					db:        dbxx,
// 					userID:    user.ID,
// 					projectID: taskProject.ID,
// 					input: &shared.CreateTaskBaseDTO{
// 						Name:        "Test Task",
// 						Description: types.Pointer("Test Description"),
// 						Status:      shared.TaskStatusDone,
// 						Rank:       1000,
// 					},
// 				},
// 				want: &models.Task{
// 					ProjectID:   taskProject.ID,
// 					Name:        "Test Task",
// 					Description: types.Pointer("Test Description"),
// 					Status:      models.TaskStatusDone,
// 					Rank:       1000,
// 				},
// 				wantErr: false,
// 			},
// 		}
// 		for _, tt := range tests {
// 			t.Run(tt.name, func(t *testing.T) {
// 				got, err := queries.CreateTask(tt.args.ctx, tt.args.db, tt.args.projectID, tt.args.input)
// 				if (err != nil) != tt.wantErr {
// 					t.Errorf("CreateTask() error = %v, wantErr %v", err, tt.wantErr)
// 					return
// 				}
// 				if tt.want != nil {
// 					if !reflect.DeepEqual(got.Name, tt.want.Name) {
// 						t.Errorf("CreateTask() Name = %v, want %v", got.Name, tt.want.Name)
// 					}
// 					if !reflect.DeepEqual(got.Description, tt.want.Description) {
// 						t.Errorf("CreateTask() Description = %v, want %v", got.Description, tt.want.Description)
// 					}
// 					if !reflect.DeepEqual(got.Status, tt.want.Status) {
// 						t.Errorf("CreateTask() Status = %v, want %v", got.Status, tt.want.Status)
// 					}
// 					if !reflect.DeepEqual(got.Rank, tt.want.Rank) {
// 						t.Errorf("CreateTask() Rank = %v, want %v", got.Rank, tt.want.Rank)
// 					}
// 					// if !reflect.DeepEqual(got.UserID, tt.want.UserID) {
// 					// 	t.Errorf("CreateTask() UserID = %v, want %v", got.UserID, tt.want.UserID)
// 					// }
// 					if !reflect.DeepEqual(got.ProjectID, tt.want.ProjectID) {
// 						t.Errorf("CreateTask() ProjectID = %v, want %v", got.ProjectID, tt.want.ProjectID)
// 					}
// 				}
// 			})
// 		}
// 		return test.EndTestErr
// 	})
// }
// func TestDefineTaskOrderNumberByStatus(t *testing.T) {
// 	test.Short(t)
// 	ctx, dbx := test.DbSetup()
// 	_ = dbx.RunInTx( func(dbxx database.Dbx) error {
// 		user, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
// 			Email: "tkahng@gmail.com",
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create user: %v", err)
// 		}
// 		member, err := queries.CreateTeamFromUser(ctx, dbxx, user)
// 		if err != nil {
// 			t.Fatalf("failed to create team from user: %v", err)
// 		}
// 		taskProject, err := queries.CreateTaskProject(ctx, dbxx, &stores.CreateTaskProjectDTO{
// 			Name:     "Test Project",
// 			Status:   models.TaskProjectStatusDone,
// 			TeamID:   member.TeamID,
// 			MemberID: member.ID,
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create task project: %v", err)
// 		}

// 		task1, err := queries.CreateTask(ctx, dbxx, taskProject.ID, &shared.CreateTaskBaseDTO{
// 			Name:      "Task 1",
// 			Status:    shared.TaskStatusDone,
// 			Rank:     0,
// 			TeamID:    member.TeamID,
// 			CreatedBy: member.ID,
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create task: %v", err)
// 		}
// 		utils.PrettyPrintJSON(task1)
// 		task2, err := queries.CreateTask(ctx, dbxx, taskProject.ID, &shared.CreateTaskBaseDTO{
// 			Name:      "Task 2",
// 			Status:    shared.TaskStatusDone,
// 			Rank:     1000,
// 			TeamID:    member.TeamID,
// 			CreatedBy: member.ID,
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create task: %v", err)
// 		}
// 		utils.PrettyPrintJSON(task2)
// 		task3, err := queries.CreateTask(ctx, dbxx, taskProject.ID, &shared.CreateTaskBaseDTO{
// 			Name:      "Task 3",
// 			Status:    shared.TaskStatusDone,
// 			Rank:     2000,
// 			TeamID:    member.TeamID,
// 			CreatedBy: member.ID,
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create task: %v", err)
// 		}
// 		utils.PrettyPrintJSON(task3)
// 		type args struct {
// 			ctx           context.Context
// 			db            database.Dbx
// 			taskId        uuid.UUID
// 			taskProjectId uuid.UUID
// 			status        models.TaskStatus
// 			currentOrder  float64
// 			position      int64
// 		}
// 		tests := []struct {
// 			name    string
// 			args    args
// 			want    float64
// 			wantErr bool
// 		}{
// 			{
// 				name: "get order for first position",
// 				args: args{
// 					ctx:           ctx,
// 					db:            dbxx,
// 					taskId:        task1.ID,
// 					taskProjectId: taskProject.ID,
// 					status:        models.TaskStatusDone,
// 					currentOrder:  0,
// 					position:      0,
// 				},
// 				want:    0,
// 				wantErr: false,
// 			},
// 			{
// 				name: "move second to first position",
// 				args: args{
// 					ctx:           ctx,
// 					db:            dbxx,
// 					taskId:        task2.ID,
// 					taskProjectId: taskProject.ID,
// 					status:        models.TaskStatusDone,
// 					currentOrder:  1000,
// 					position:      0,
// 				},
// 				want:    -1000,
// 				wantErr: false,
// 			},
// 			{
// 				name: "move first to last position",
// 				args: args{
// 					ctx:           ctx,
// 					db:            dbxx,
// 					taskId:        task1.ID,
// 					taskProjectId: taskProject.ID,
// 					status:        models.TaskStatusDone,
// 					currentOrder:  0,
// 					position:      2,
// 				},
// 				want:    3000,
// 				wantErr: false,
// 			},
// 		}
// 		for _, tt := range tests {
// 			t.Run(tt.name, func(t *testing.T) {
// 				got, err := queries.DefineTaskOrderNumberByStatus(tt.args.ctx, tt.args.db, tt.args.taskId, tt.args.taskProjectId, tt.args.status, tt.args.currentOrder, tt.args.position)
// 				if (err != nil) != tt.wantErr {
// 					t.Errorf("DefineTaskOrderNumberByStatus() error = %v, wantErr %v", err, tt.wantErr)
// 					return
// 				}
// 				if got != tt.want {
// 					t.Errorf("DefineTaskOrderNumberByStatus() = %v, want %v", got, tt.want)
// 				}
// 			})
// 		}
// 		return test.EndTestErr
// 	})
// }
// func TestUpdateTask(t *testing.T) {
// 	test.Short(t)
// 	ctx, dbx := test.DbSetup()
// 	_ = dbx.RunInTx( func(dbxx database.Dbx) error {
// 		user, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
// 			Email: "tkahng@gmail.com",
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create user: %v", err)
// 		}
// 		member, err := queries.CreateTeamFromUser(ctx, dbxx, user)
// 		if err != nil {
// 			t.Fatalf("failed to create team from user: %v", err)
// 		}
// 		taskProject, err := queries.CreateTaskProject(ctx, dbxx, &stores.CreateTaskProjectDTO{
// 			Name:     "Test Project",
// 			Status:   models.TaskProjectStatusDone,
// 			TeamID:   member.TeamID,
// 			MemberID: member.ID,
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create task project: %v", err)
// 		}
// 		task, err := queries.CreateTask(ctx, dbxx, taskProject.ID, &shared.CreateTaskBaseDTO{
// 			Name:        "Test Task",
// 			Description: types.Pointer("Test Description"),
// 			Status:      shared.TaskStatusDone,
// 			Rank:       1000,
// 			TeamID:      member.TeamID,
// 			CreatedBy:   member.ID,
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create task: %v", err)
// 		}

// 		type args struct {
// 			ctx    context.Context
// 			db     database.Dbx
// 			taskID uuid.UUID
// 			input  *shared.UpdateTaskBaseDTO
// 		}
// 		tests := []struct {
// 			name    string
// 			args    args
// 			wantErr bool
// 		}{
// 			{
// 				name: "update task successfully",
// 				args: args{
// 					ctx:    ctx,
// 					db:     dbxx,
// 					taskID: task.ID,
// 					input: &shared.UpdateTaskBaseDTO{
// 						Name:        "Updated Task",
// 						Description: types.Pointer("Updated Description"),
// 						Status:      shared.TaskStatusInProgress,
// 						Rank:       2000,
// 						ParentID:    nil,
// 					},
// 				},
// 				wantErr: false,
// 			},
// 			{
// 				name: "update non-existing task",
// 				args: args{
// 					ctx:    ctx,
// 					db:     dbxx,
// 					taskID: uuid.New(),
// 					input: &shared.UpdateTaskBaseDTO{
// 						Name:   "Updated Task",
// 						Status: shared.TaskStatusInProgress,
// 					},
// 				},
// 				wantErr: true,
// 			},
// 		}
// 		for _, tt := range tests {
// 			t.Run(tt.name, func(t *testing.T) {
// 				err := queries.UpdateTask(tt.args.ctx, tt.args.db, tt.args.taskID, tt.args.input)
// 				if (err != nil) != tt.wantErr {
// 					t.Errorf("UpdateTask() error = %v, wantErr %v", err, tt.wantErr)
// 					return
// 				}

// 				if !tt.wantErr {
// 					// Verify task was updated
// 					updatedTask, err := queries.FindTaskByID(tt.args.ctx, tt.args.db, tt.args.taskID)
// 					if err != nil {
// 						t.Errorf("Failed to get updated task: %v", err)
// 						return
// 					}
// 					if updatedTask.Name != tt.args.input.Name {
// 						t.Errorf("Task name not updated. got = %v, want %v", updatedTask.Name, tt.args.input.Name)
// 					}
// 					if *updatedTask.Description != *tt.args.input.Description {
// 						t.Errorf("Task description not updated. got = %v, want %v", *updatedTask.Description, *tt.args.input.Description)
// 					}
// 					if updatedTask.Status != models.TaskStatus(tt.args.input.Status) {
// 						t.Errorf("Task status not updated. got = %v, want %v", updatedTask.Status, tt.args.input.Status)
// 					}
// 					if updatedTask.Rank != tt.args.input.Rank {
// 						t.Errorf("Task order not updated. got = %v, want %v", updatedTask.Rank, tt.args.input.Rank)
// 					}
// 				}
// 			})
// 		}
// 		return test.EndTestErr
// 	})
// }
// func TestUpdateTaskProjectUpdateDate(t *testing.T) {
// 	test.Short(t)
// 	ctx, dbx := test.DbSetup()
// 	_ = dbx.RunInTx( func(dbxx database.Dbx) error {
// 		user, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
// 			Email: "tkahng@gmail.com",
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create user: %v", err)
// 		}
// 		member, err := queries.CreateTeamFromUser(ctx, dbxx, user)
// 		if err != nil {
// 			t.Fatalf("failed to create team from user: %v", err)
// 		}
// 		taskProject, err := queries.CreateTaskProject(ctx, dbxx, &stores.CreateTaskProjectDTO{
// 			Name:     "Test Project",
// 			Status:   models.TaskProjectStatusDone,
// 			TeamID:   member.TeamID,
// 			MemberID: member.ID,
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create task project: %v", err)
// 		}

// 		type args struct {
// 			ctx           context.Context
// 			db            database.Dbx
// 			taskProjectID uuid.UUID
// 		}
// 		tests := []struct {
// 			name    string
// 			args    args
// 			wantErr bool
// 		}{
// 			{
// 				name: "update task project date successfully",
// 				args: args{
// 					ctx:           ctx,
// 					db:            dbxx,
// 					taskProjectID: taskProject.ID,
// 				},
// 				wantErr: false,
// 			},
// 			{
// 				name: "update non-existing task project date",
// 				args: args{
// 					ctx:           ctx,
// 					db:            dbxx,
// 					taskProjectID: uuid.New(),
// 				},
// 				wantErr: false,
// 			},
// 		}
// 		for _, tt := range tests {
// 			t.Run(tt.name, func(t *testing.T) {
// 				err := queries.UpdateTaskProjectUpdateDate(tt.args.ctx, tt.args.db, tt.args.taskProjectID)
// 				if (err != nil) != tt.wantErr {
// 					t.Errorf("UpdateTaskProjectUpdateDate() error = %v, wantErr %v", err, tt.wantErr)
// 					return
// 				}
// 			})
// 		}
// 		return test.EndTestErr
// 	})
// }
// func TestUpdateTaskProject(t *testing.T) {
// 	test.Short(t)
// 	ctx, dbx := test.DbSetup()
// 	_ = dbx.RunInTx( func(dbxx database.Dbx) error {
// 		user, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
// 			Email: "tkahng@gmail.com",
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create user: %v", err)
// 		}
// 		member, err := queries.CreateTeamFromUser(ctx, dbxx, user)
// 		if err != nil {
// 			t.Fatalf("failed to create team from user: %v", err)
// 		}
// 		taskProject, err := queries.CreateTaskProject(ctx, dbxx, &stores.CreateTaskProjectDTO{
// 			Name:        "Test Project",
// 			Description: types.Pointer("Test Description"),
// 			Status:      models.TaskProjectStatusDone,
// 			Rank:       1000,
// 			TeamID:      member.TeamID,
// 			MemberID:    member.ID,
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create task project: %v", err)
// 		}

// 		type args struct {
// 			ctx           context.Context
// 			db            database.Dbx
// 			taskProjectID uuid.UUID
// 			input         *UpdateTaskProjectBaseDTO
// 		}
// 		tests := []struct {
// 			name    string
// 			args    args
// 			wantErr bool
// 		}{
// 			{
// 				name: "update task project successfully",
// 				args: args{
// 					ctx:           ctx,
// 					db:            dbxx,
// 					taskProjectID: taskProject.ID,
// 					input: &UpdateTaskProjectBaseDTO{
// 						Name:        "Updated Project",
// 						Description: types.Pointer("Updated Description"),
// 						Status:      shared.TaskProjectStatusInProgress,
// 						Rank:       2000,
// 					},
// 				},
// 				wantErr: false,
// 			},
// 			{
// 				name: "update non-existing task project",
// 				args: args{
// 					ctx:           ctx,
// 					db:            dbxx,
// 					taskProjectID: uuid.New(),
// 					input: &UpdateTaskProjectBaseDTO{
// 						Name:   "Updated Project",
// 						Status: shared.TaskProjectStatusInProgress,
// 					},
// 				},
// 				wantErr: true,
// 			},
// 		}
// 		for _, tt := range tests {
// 			t.Run(tt.name, func(t *testing.T) {
// 				err := queries.UpdateTaskProject(tt.args.ctx, tt.args.db, tt.args.taskProjectID, tt.args.input)
// 				if (err != nil) != tt.wantErr {
// 					t.Errorf("UpdateTaskProject() error = %v, wantErr %v", err, tt.wantErr)
// 					return
// 				}

// 				if !tt.wantErr {
// 					// Verify task project was updated
// 					updatedProject, err := queries.FindTaskProjectByID(tt.args.ctx, tt.args.db, tt.args.taskProjectID)
// 					if err != nil {
// 						t.Errorf("Failed to get updated task project: %v", err)
// 						return
// 					}
// 					if updatedProject.Name != tt.args.input.Name {
// 						t.Errorf("Task project name not updated. got = %v, want %v", updatedProject.Name, tt.args.input.Name)
// 					}
// 					if *updatedProject.Description != *tt.args.input.Description {
// 						t.Errorf("Task project description not updated. got = %v, want %v", *updatedProject.Description, *tt.args.input.Description)
// 					}
// 					if updatedProject.Status != models.TaskProjectStatus(tt.args.input.Status) {
// 						t.Errorf("Task project status not updated. got = %v, want %v", updatedProject.Status, tt.args.input.Status)
// 					}
// 					if updatedProject.Rank != tt.args.input.Rank {
// 						t.Errorf("Task project order not updated. got = %v, want %v", updatedProject.Rank, tt.args.input.Rank)
// 					}
// 				}
// 			})
// 		}
// 		return test.EndTestErr
// 	})
// }
// func TestUpdateTaskPositionStatus(t *testing.T) {
// 	test.Short(t)
// 	ctx, dbx := test.DbSetup()
// 	_ = dbx.RunInTx( func(dbxx database.Dbx) error {
// 		user, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
// 			Email: "tkahng@gmail.com",
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create user: %v", err)
// 		}
// 		member, err := queries.CreateTeamFromUser(ctx, dbxx, user)
// 		if err != nil {
// 			t.Fatalf("failed to create team from user: %v", err)
// 		}
// 		taskProject, err := queries.CreateTaskProject(ctx, dbxx, &stores.CreateTaskProjectDTO{
// 			Name:     "Test Project",
// 			Status:   models.TaskProjectStatusDone,
// 			TeamID:   member.TeamID,
// 			MemberID: member.ID,
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create task project: %v", err)
// 		}

// 		task1, err := queries.CreateTask(ctx, dbxx, taskProject.ID, &shared.CreateTaskBaseDTO{
// 			Name:      "Task 1",
// 			Status:    shared.TaskStatusDone,
// 			Rank:     0,
// 			CreatedBy: member.ID,
// 			TeamID:    member.TeamID,
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create task: %v", err)
// 		}

// 		task2, err := queries.CreateTask(ctx, dbxx, taskProject.ID, &shared.CreateTaskBaseDTO{
// 			Name:      "Task 2",
// 			Status:    shared.TaskStatusDone,
// 			Rank:     1000,
// 			CreatedBy: member.ID,
// 			TeamID:    member.TeamID,
// 		})
// 		if err != nil {
// 			t.Fatalf("failed to create task: %v", err)
// 		}

// 		type args struct {
// 			ctx      context.Context
// 			db       database.Dbx
// 			taskID   uuid.UUID
// 			position int64
// 			status   models.TaskStatus
// 		}
// 		tests := []struct {
// 			name    string
// 			args    args
// 			wantErr bool
// 		}{
// 			{
// 				name: "update task position status successfully",
// 				args: args{
// 					ctx:      ctx,
// 					db:       dbxx,
// 					taskID:   task1.ID,
// 					position: 1,
// 					status:   models.TaskStatusDone,
// 				},
// 				wantErr: false,
// 			},
// 			{
// 				name: "update non-existing task position status",
// 				args: args{
// 					ctx:      ctx,
// 					db:       dbxx,
// 					taskID:   uuid.New(),
// 					position: 0,
// 					status:   models.TaskStatusDone,
// 				},
// 				wantErr: true,
// 			},
// 			{
// 				name: "move task to first position",
// 				args: args{
// 					ctx:      ctx,
// 					db:       dbxx,
// 					taskID:   task2.ID,
// 					position: 0,
// 					status:   models.TaskStatusDone,
// 				},
// 				wantErr: false,
// 			},
// 		}
// 		for _, tt := range tests {
// 			t.Run(tt.name, func(t *testing.T) {
// 				err := queries.UpdateTaskPositionStatus(tt.args.ctx, tt.args.db, tt.args.taskID, tt.args.position, tt.args.status)
// 				if (err != nil) != tt.wantErr {
// 					t.Errorf("UpdateTaskPositionStatus() error = %v, wantErr %v", err, tt.wantErr)
// 					return
// 				}

// 				if !tt.wantErr {
// 					// Verify task was updated
// 					updatedTask, err := queries.FindTaskByID(tt.args.ctx, tt.args.db, tt.args.taskID)
// 					if err != nil {
// 						t.Errorf("Failed to get updated task: %v", err)
// 						return
// 					}

// 					if updatedTask.Status != tt.args.status {
// 						t.Errorf("Task status not updated. got = %v, want %v", updatedTask.Status, tt.args.status)
// 					}

// 					// Get task project to verify update date
// 					taskProject, err := queries.FindTaskProjectByID(tt.args.ctx, tt.args.db, updatedTask.ProjectID)
// 					if err != nil {
// 						t.Errorf("Failed to get task project: %v", err)
// 						return
// 					}

// 					if taskProject.UpdatedAt.IsZero() {
// 						t.Error("Task project update date not updated")
// 					}
// 				}
// 			})
// 		}
// 		return test.EndTestErr
// 	})
// }
