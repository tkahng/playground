package teams_test

import (
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/tkahng/authgo/internal/crudrepo"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/services/teams"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/test"
)

func TestCreateTeamFromUser(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
		user, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
			Email: "test@example.com",
		})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}
		teamStore := teams.NewPostgresTeamStore(dbxx)
		team, err := teamStore.CreateTeamFromUser(ctx, user)
		if err != nil {
			t.Fatalf("CreateTeamFromUser() error = %v", err)
		}
		if team == nil || team.Name != user.Email {
			t.Errorf("CreateTeamFromUser() = %v, want team with name %v", team, user.Email)
		}
		if len(team.Members) != 1 || team.Members[0].UserID == nil || *team.Members[0].UserID != user.ID {
			t.Errorf("CreateTeamFromUser() did not create correct team member")
		}
		return errors.New("rollback")
	})
}

func TestCreateTeam(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
		teamQueries := teams.NewPostgresTeamStore(dbxx)
		team, err := teamQueries.CreateTeam(ctx, "Test Team", nil)
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}
		if team == nil || team.Name != "Test Team" {
			t.Errorf("CreateTeam() = %v, want name 'Test Team'", team)
		}
		return errors.New("rollback")
	})
}

func TestUpdateTeam(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
		teamQueries := teams.NewPostgresTeamStore(dbxx)
		team, err := teamQueries.CreateTeam(ctx, "Old Name", nil)
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}
		newName := "Updated Name"
		stripeID := "cus_123"
		updated, err := teamQueries.UpdateTeam(ctx, team.ID, newName, &stripeID)
		if err != nil {
			t.Fatalf("UpdateTeam() error = %v", err)
		}
		if updated.Name != newName || updated.StripeCustomerID == nil || *updated.StripeCustomerID != stripeID {
			t.Errorf("UpdateTeam() = %v, want name %v and stripeID %v", updated, newName, stripeID)
		}
		return errors.New("rollback")
	})
}

func TestDeleteTeam(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
		teamQueries := teams.NewPostgresTeamStore(dbxx)
		// Create a team to delete
		team, err := teamQueries.CreateTeam(ctx, "ToDelete", nil)
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}
		err = teamQueries.DeleteTeam(ctx, team.ID)
		if err != nil {
			t.Errorf("DeleteTeam() error = %v", err)
		}
		return errors.New("rollback")
	})
}

func TestFindTeamByID(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
		teamQueries := teams.NewPostgresTeamStore(dbxx)
		team, err := teamQueries.CreateTeam(ctx, "FindMe", nil)
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}
		found, err := teamQueries.FindTeamByID(ctx, team.ID)
		if err != nil {
			t.Fatalf("FindTeamByID() error = %v", err)
		}
		if found == nil || found.ID != team.ID {
			t.Errorf("FindTeamByID() = %v, want %v", found, team.ID)
		}
		return errors.New("rollback")
	})
}

func TestCreateTeamMember(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
		teamQueries := teams.NewPostgresTeamStore(dbxx)
		team, err := teamQueries.CreateTeam(ctx, "TeamWithMember", nil)
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}
		user, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
			Email: "testuser@example.com",
		})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}
		userID := user.ID
		member, err := teamQueries.CreateTeamMember(ctx, team.ID, userID, models.TeamMemberRoleMember)
		if err != nil {
			t.Fatalf("CreateTeamMember() error = %v", err)
		}
		if member.TeamID != team.ID || member.UserID == nil || *member.UserID != userID {
			t.Errorf("CreateTeamMember() = %v, want teamID %v and userID %v", member, team.ID, userID)
		}
		return errors.New("rollback")
	})
}

func TestFindTeamMembersByUserID(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
		teamQueries := teams.NewPostgresTeamStore(dbxx)
		team, err := teamQueries.CreateTeam(ctx, "TeamForMembers", nil)
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}
		user, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
			Email: "tkahng@gmail.com",
		})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}
		userID := user.ID
		_, err = teamQueries.CreateTeamMember(ctx, team.ID, userID, models.TeamMemberRoleMember)
		if err != nil {
			t.Fatalf("CreateTeamMember() error = %v", err)
		}
		members, err := teamQueries.FindTeamMembersByUserID(ctx, userID)
		if err != nil {
			t.Fatalf("FindTeamMembersByUserID() error = %v", err)
		}
		if len(members) == 0 || *members[0].UserID != userID {
			t.Errorf("FindTeamMembersByUserID() = %v, want userID %v", members, userID)
		}
		return errors.New("rollback")
	})
}

func TestFindLatestTeamMemberByUserID(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
		teamQueries := teams.NewPostgresTeamStore(dbxx)
		team1, err := teamQueries.CreateTeam(ctx, "team1", nil)
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}
		team2, err := teamQueries.CreateTeam(ctx, "team2", nil)
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}
		user, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
			Email: "testuser@example.com",
		})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}
		userID := user.ID
		teamMember1, err := teamQueries.CreateTeamMember(ctx, team1.ID, userID, models.TeamMemberRoleMember)
		if err != nil {
			t.Fatalf("CreateTeamMember() error = %v", err)
		}
		teamMember2, err := teamQueries.CreateTeamMember(ctx, team2.ID, userID, models.TeamMemberRoleMember)
		if err != nil {
			t.Fatalf("CreateTeamMember() error = %v", err)
		}
		time.Sleep(time.Millisecond * 10)
		err = teamQueries.UpdateTeamMemberUpdatedAt(ctx, teamMember1.ID)
		if err != nil {
			t.Fatalf("UpdateTeamMemberUpdatedAt() error = %v", err)
		}
		latest, err := teamQueries.FindLatestTeamMemberByUserID(ctx, userID)
		if err != nil {
			t.Fatalf("FindLatestTeamMemberByUserID() error = %v", err)
		}
		if latest == nil || latest.UserID == nil || *latest.UserID != userID {
			t.Errorf("FindLatestTeamMemberByUserID() = %v, want userID %v", latest, userID)
		}
		if latest.ID != teamMember1.ID {
			t.Errorf("FindLatestTeamMemberByUserID() = %v, want teamMember1 ID %v", latest.ID, teamMember1.ID)
		}
		time.Sleep(time.Millisecond * 10)
		err = teamQueries.UpdateTeamMemberUpdatedAt(ctx, teamMember2.ID)
		if err != nil {
			t.Fatalf("UpdateTeamMemberUpdatedAt() error = %v", err)
		}
		latest, err = teamQueries.FindLatestTeamMemberByUserID(ctx, userID)
		if err != nil {
			t.Fatalf("FindLatestTeamMemberByUserID() error = %v", err)
		}
		if latest == nil || latest.UserID == nil || *latest.UserID != userID {
			t.Errorf("FindLatestTeamMemberByUserID() = %v, want userID %v", latest, userID)
		}
		if latest.ID != teamMember2.ID {
			t.Errorf("FindLatestTeamMemberByUserID() = %v, want teamMember2 ID %v", latest.ID, teamMember2.ID)
		}
		return errors.New("rollback")
	})
}

func TestUpdateTeamMemberUpdatedAt(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	t.Cleanup(func() {
		_, err := crudrepo.TeamMember.Delete(ctx, dbx, nil)
		if err != nil {
			slog.ErrorContext(ctx, "Error deleting team members", slog.Any("error", err))
		}
		_, err = crudrepo.Team.Delete(ctx, dbx, nil)
		if err != nil {
			slog.ErrorContext(ctx, "Error deleting teams", slog.Any("error", err))
		}
		_, err = crudrepo.User.Delete(ctx, dbx, nil)
		if err != nil {
			slog.ErrorContext(ctx, "Error deleting users", slog.Any("error", err))
		}
	})

	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
		teamQueries := teams.NewPostgresTeamStore(dbxx)
		team, err := teamQueries.CreateTeam(ctx, "UpdateMemberTeam", nil)
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}
		user, err := queries.CreateUser(ctx, dbxx, &shared.AuthenticationInput{
			Email: "updatemember@example.com",
		})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}
		member, err := teamQueries.CreateTeamMember(ctx, team.ID, user.ID, models.TeamMemberRoleMember)
		if err != nil {
			t.Fatalf("CreateTeamMember() error = %v", err)
		}
		// Capture the original updated_at
		original := member.CreatedAt

		// Sleep to ensure updated_at will be different
		time.Sleep(time.Second * 1)

		err = teamQueries.UpdateTeamMemberUpdatedAt(ctx, member.ID)
		if err != nil {
			t.Fatalf("UpdateTeamMemberUpdatedAt() error = %v", err)
		}
		if err != nil {
			t.Fatalf("UpdateTeamMemberUpdatedAt() error = %v", err)
		}

		// Fetch the member again to check updated_at
		updated, err := crudrepo.TeamMember.GetOne(
			ctx,
			dbxx,
			&map[string]any{
				"id": map[string]any{
					"_eq": member.ID.String(),
				},
			},
		)
		if err != nil {
			t.Fatalf("GetOne() error = %v", err)
		}
		if updated == nil {
			t.Fatalf("Updated member not found")
		}
		if !updated.UpdatedAt.After(original) {
			t.Errorf(
				"UpdateTeamMemberUpdatedAt() did not update updated_at: before=%v after=%v",
				original,
				updated.UpdatedAt,
			)
		}
		return errors.New("rollback")
	})
}
