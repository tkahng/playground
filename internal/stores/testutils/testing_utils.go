package testutils

import (
	"context"

	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
)

func CreateUser(adapter stores.StorageAdapterInterface, ctx context.Context, email string) *models.User {
	user, err := adapter.User().CreateUser(ctx, &models.User{
		Email: email,
	})
	if err != nil {
		panic(err)
	}
	return user
}

func CreateTeam(adapter stores.StorageAdapterInterface, ctx context.Context, slug string) *models.Team {
	team, err := adapter.TeamGroup().CreateTeam(ctx, slug, slug)
	if err != nil {
		panic(err)
	}
	return team
}

func CreateTeamMember(adapter stores.StorageAdapterInterface, ctx context.Context, team *models.Team, user *models.User, role models.TeamMemberRole, billingAccess bool) *models.TeamMember {
	member, err := adapter.TeamMember().CreateTeamMember(ctx, team.ID, user.ID, role, billingAccess)
	if err != nil {
		panic(err)
	}
	return member
}

func CreateTeamProject(adapter stores.StorageAdapterInterface, ctx context.Context, member *models.TeamMember, name string, description string) *models.TaskProject {
	taskProject, err := adapter.Task().CreateTaskProject(ctx, &stores.CreateTaskProjectDTO{
		Name:        name,
		Status:      models.TaskProjectStatusDone,
		TeamID:      member.TeamID,
		MemberID:    member.ID,
		Description: &description,
	})
	if err != nil {
		panic(err)
	}
	return taskProject
}

func CreateTask(adapter stores.StorageAdapterInterface, ctx context.Context, task *models.Task) *models.Task {
	task, err := adapter.Task().CreateTask(ctx, task)
	if err != nil {
		panic(err)
	}
	return task
}
