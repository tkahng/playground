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

		parent := repository.MustCreateOneCtx(t, ctx, repository.Task, db, &models.Task{
			TeamID:            team.ID,
			ProjectID:         project.ID,
			CreatedByMemberID: &ownerMember.ID,
			Name:              "Task Parent",
			Status:            models.TaskStatusTodo,
			AssigneeID:        &user1Member.ID,
			ReporterID:        &user1Member.ID,
		})

		task := repository.MustCreateOneCtx(t, ctx, repository.Task, db, &models.Task{
			TeamID:            team.ID,
			ProjectID:         project.ID,
			CreatedByMemberID: &ownerMember.ID,
			Name:              "Task 1",
			Status:            models.TaskStatusTodo,
			AssigneeID:        &user1Member.ID,
			ReporterID:        &user1Member.ID,
			ParentID:          &parent.ID,
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
		assert.NotNil(t, task.Parent)
		assert.Equal(t, parent.ID, task.Parent.ID)
		assert.Equal(t, 7, testPopulator.Recorder.Called())
	})
}
func TestPopulateMember(t *testing.T) {
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
		ownerMember.User = owner
		ownerMember.Team = team

		user1 := repository.MustCreateOneCtx(t, ctx, repository.User, db, &models.User{
			Email: "user1@example.com",
		})
		user1Member := repository.MustCreateOneCtx(t, ctx, repository.TeamMember, db, &models.TeamMember{
			UserID: &user1.ID,
			TeamID: team.ID,
			Role:   models.TeamMemberRoleMember,
			Active: true,
		})
		user1Member.User = user1
		user1Member.Team = team

		members := repository.MustFindWithOptionsCtx(t, ctx, repository.TeamMember, db)
		assert.Len(t, members, 2)
		adapter := stores.NewDbAdapterDecorators(db)
		testPopulator := NewTestPopulator(adapter)
		for _, member := range members {
			err := PopulateTeamMember(ctx, testPopulator, member)
			assert.NoError(t, err)
		}
		for _, member := range members {
			var existingMember *models.TeamMember
			switch member.ID {
			case ownerMember.ID:
				existingMember = ownerMember
			case user1Member.ID:
				existingMember = user1Member
			default:
				assert.Fail(t, "member not found")
			}
			assert.Equal(t, existingMember.ID, member.ID)
			assert.Equal(t, existingMember.UserID, member.UserID)
			assert.Equal(t, existingMember.TeamID, member.TeamID)
			assert.Equal(t, existingMember.Role, member.Role)
			assert.Equal(t, existingMember.Active, member.Active)
			assert.Equal(t, existingMember.HasBillingAccess, member.HasBillingAccess)
			assert.Equal(t, existingMember.User.ID, member.User.ID)
			assert.Equal(t, existingMember.User.Email, member.User.Email)
			assert.Equal(t, existingMember.User.EmailVerifiedAt, member.User.EmailVerifiedAt)
			assert.Equal(t, existingMember.User.CreatedAt, member.User.CreatedAt)
			assert.Equal(t, existingMember.Team.ID, member.Team.ID)
			assert.Equal(t, existingMember.Team.Name, member.Team.Name)
			assert.Equal(t, existingMember.Team.Slug, member.Team.Slug)
		}
		assert.Equal(t, 3, testPopulator.Recorder.called)
	})
}
func TestPopulateMember_Users_Loaded(t *testing.T) {
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
		ownerMember.User = owner
		ownerMember.Team = team

		user1 := repository.MustCreateOneCtx(t, ctx, repository.User, db, &models.User{
			Email: "user1@example.com",
		})
		user1Member := repository.MustCreateOneCtx(t, ctx, repository.TeamMember, db, &models.TeamMember{
			UserID: &user1.ID,
			TeamID: team.ID,
			Role:   models.TeamMemberRoleMember,
			Active: true,
		})
		user1Member.User = user1
		user1Member.Team = team

		members := repository.MustFindWithOptionsCtx(t, ctx, repository.TeamMember, db)
		assert.Len(t, members, 2)
		adapter := stores.NewDbAdapterDecorators(db)
		testPopulator := NewTestPopulator(adapter)
		if err := testPopulator.user.Load(ctx, owner.ID, user1.ID); err != nil {
			t.Fatal(err)
		}
		for _, member := range members {
			err := PopulateTeamMember(ctx, testPopulator, member)
			assert.NoError(t, err)
		}
		for _, member := range members {
			var existingMember *models.TeamMember
			switch member.ID {
			case ownerMember.ID:
				existingMember = ownerMember
			case user1Member.ID:
				existingMember = user1Member
			default:
				assert.Fail(t, "member not found")
			}
			assert.Equal(t, existingMember.ID, member.ID)
			assert.Equal(t, existingMember.UserID, member.UserID)
			assert.Equal(t, existingMember.TeamID, member.TeamID)
			assert.Equal(t, existingMember.Role, member.Role)
			assert.Equal(t, existingMember.Active, member.Active)
			assert.Equal(t, existingMember.HasBillingAccess, member.HasBillingAccess)
			assert.Equal(t, existingMember.User.ID, member.User.ID)
			assert.Equal(t, existingMember.User.Email, member.User.Email)
			assert.Equal(t, existingMember.User.EmailVerifiedAt, member.User.EmailVerifiedAt)
			assert.Equal(t, existingMember.User.CreatedAt, member.User.CreatedAt)
			assert.Equal(t, existingMember.Team.ID, member.Team.ID)
			assert.Equal(t, existingMember.Team.Name, member.Team.Name)
			assert.Equal(t, existingMember.Team.Slug, member.Team.Slug)
		}
		assert.Equal(t, 1, testPopulator.Recorder.called)
	})
}
