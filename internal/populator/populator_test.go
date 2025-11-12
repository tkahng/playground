package populator

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/database/repository"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/types"
)

func TestPopulateTask(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		team := repository.MustCreateOneCtx(t, ctx, repository.Team, db, &models.Team{
			Name: "team1",
			Slug: "team1",
		})
		owner := repository.MustCreateOneCtx(t, ctx, repository.User, db, &models.User{
			Email:           "owner@example.com",
			EmailVerifiedAt: types.Pointer(time.Now()),
		})
		ownerMember := repository.MustCreateOneCtx(t, ctx, repository.TeamMember, db, &models.TeamMember{
			UserID:           &owner.ID,
			TeamID:           team.ID,
			Role:             models.TeamMemberRoleOwner,
			Active:           true,
			HasBillingAccess: true,
		})

		user1 := repository.MustCreateOneCtx(t, ctx, repository.User, db, &models.User{
			Email: "user1@example.com",
		})
		user1Member := repository.MustCreateOneCtx(t, ctx, repository.TeamMember, db, &models.TeamMember{
			UserID: &user1.ID,
			TeamID: team.ID,
			Role:   models.TeamMemberRoleMember,
			Active: true,
		})

		project := repository.MustCreateOneCtx(t, ctx, repository.TaskProject, db, &models.TaskProject{
			TeamID:            team.ID,
			CreatedByMemberID: &ownerMember.ID,
			Name:              "Project 1",
			Status:            models.TaskProjectStatusTodo,
		})

		task := repository.MustCreateOneCtx(t, ctx, repository.Task, db, &models.Task{
			TeamID:            team.ID,
			ProjectID:         project.ID,
			CreatedByMemberID: &ownerMember.ID,
			Name:              "Task 1",
			Status:            models.TaskStatusTodo,
			AssigneeID:        &user1Member.ID,
			ReporterID:        &user1Member.ID,
		})
		assert.NotNil(t, task)
		adapter := stores.NewDbAdapterDecorators(db)
		testPopulator := NewTestPopulator(adapter)
		err := PopulateTask(ctx, testPopulator, task)
		assert.NoError(t, err)
		assert.NotNil(t, task.Team)
		assert.Equal(t, team.ID, task.Team.ID)
		assert.NotNil(t, task.Project)
		assert.Equal(t, project.ID, task.Project.ID)
		assert.NotNil(t, task.CreatedByMember)
		assert.Equal(t, ownerMember.ID, task.CreatedByMember.ID)
		assert.NotNil(t, task.Assignee)
		assert.Equal(t, user1Member.ID, task.Assignee.ID)
		assert.NotNil(t, task.Reporter)
		assert.Equal(t, user1Member.ID, task.Reporter.ID)
		assert.Equal(t, 6, testPopulator.Recorder.Called())
	})

}
