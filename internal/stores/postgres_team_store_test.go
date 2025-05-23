package stores_test

import (
	"context"
	"errors"
	"log/slog"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/crudrepo"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/stores"
	"github.com/tkahng/authgo/internal/test"
	"github.com/tkahng/authgo/internal/tools/types"
)

func TestCreateTeam(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
		teamStore := stores.NewPostgresTeamStore(dbxx)
		team, err := teamStore.CreateTeam(ctx, "Test Team", "test-team-slug")
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
		teamStore := stores.NewPostgresTeamStore(dbxx)
		team, err := teamStore.CreateTeam(ctx, "Old Name", "old-name-slug")
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}
		newName := "Updated Name"
		updated, err := teamStore.UpdateTeam(ctx, team.ID, newName)
		if err != nil {
			t.Fatalf("UpdateTeam() error = %v", err)
		}
		if updated.Name != newName {
			t.Errorf("UpdateTeam() = %v, want name %v", updated, newName)
		}
		return errors.New("rollback")
	})
}

func TestDeleteTeam(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
		teamStore := stores.NewPostgresTeamStore(dbxx)
		// Create a team to delete
		team, err := teamStore.CreateTeam(ctx, "ToDelete", "to-delete-slug")
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}
		err = teamStore.DeleteTeam(ctx, team.ID)
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
		teamStore := stores.NewPostgresTeamStore(dbxx)
		team, err := teamStore.CreateTeam(ctx, "FindMe", "find-me-slug")
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}
		found, err := teamStore.FindTeamByID(ctx, team.ID)
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
		teamStore := stores.NewPostgresTeamStore(dbxx)
		userStore := stores.NewPostgresUserStore(dbxx)
		team, err := teamStore.CreateTeam(ctx, "TeamWithMember", "team-with-member-slug")
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}
		user, err := userStore.CreateUser(ctx, &models.User{
			Email: "testuser@example.com",
		})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}
		userID := user.ID
		member, err := teamStore.CreateTeamMember(ctx, team.ID, userID, models.TeamMemberRoleMember, true)
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
		teamStore := stores.NewPostgresTeamStore(dbxx)
		userStore := stores.NewPostgresUserStore(dbxx)
		team, err := teamStore.CreateTeam(ctx, "TeamForMembers", "team-for-members-slug")
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}
		user, err := userStore.CreateUser(ctx, &models.User{
			Email: "tkahng@gmail.com",
		})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}
		userID := user.ID
		_, err = teamStore.CreateTeamMember(ctx, team.ID, userID, models.TeamMemberRoleMember, true)
		if err != nil {
			t.Fatalf("CreateTeamMember() error = %v", err)
		}
		members, err := teamStore.FindTeamMembersByUserID(ctx, userID, nil)
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
		teamStore := stores.NewPostgresTeamStore(dbxx)
		userStore := stores.NewPostgresUserStore(dbxx)
		team1, err := teamStore.CreateTeam(ctx, "team1", "team1-slug")
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}
		team2, err := teamStore.CreateTeam(ctx, "team2", "team2-slug")
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}
		user, err := userStore.CreateUser(ctx, &models.User{
			Email: "testuser@example.com",
		})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}
		userID := user.ID
		teamMember1, err := teamStore.CreateTeamMember(ctx, team1.ID, userID, models.TeamMemberRoleMember, true)
		if err != nil {
			t.Fatalf("CreateTeamMember() error = %v", err)
		}
		teamMember2, err := teamStore.CreateTeamMember(ctx, team2.ID, userID, models.TeamMemberRoleMember, true)
		if err != nil {
			t.Fatalf("CreateTeamMember() error = %v", err)
		}
		time.Sleep(time.Millisecond * 10)
		err = teamStore.UpdateTeamMemberSelectedAt(ctx, teamMember1.TeamID, userID)
		if err != nil {
			t.Fatalf("UpdateTeamMemberUpdatedAt() error = %v", err)
		}
		latest, err := teamStore.FindLatestTeamMemberByUserID(ctx, userID)
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
		err = teamStore.UpdateTeamMemberSelectedAt(ctx, teamMember1.TeamID, userID)
		if err != nil {
			t.Fatalf("UpdateTeamMemberUpdatedAt() error = %v", err)
		}
		latest, err = teamStore.FindLatestTeamMemberByUserID(ctx, userID)
		if err != nil {
			t.Fatalf("FindLatestTeamMemberByUserID() error = %v", err)
		}
		if latest == nil || latest.UserID == nil || *latest.UserID != userID {
			t.Errorf("FindLatestTeamMemberByUserID() = %v, want userID %v", latest, userID)
		}
		if latest.ID != teamMember1.ID {
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
		teamStore := stores.NewPostgresTeamStore(dbxx)
		userStore := stores.NewPostgresUserStore(dbxx)

		team, err := teamStore.CreateTeam(ctx, "UpdateMemberTeam", "update-member-team-slug")
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}
		user, err := userStore.CreateUser(ctx, &models.User{
			Email: "updatemember@example.com",
		})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}
		member, err := teamStore.CreateTeamMember(ctx, team.ID, user.ID, models.TeamMemberRoleMember, true)
		if err != nil {
			t.Fatalf("CreateTeamMember() error = %v", err)
		}
		// Capture the original updated_at
		original := member.CreatedAt

		// Sleep to ensure updated_at will be different
		time.Sleep(time.Second * 1)

		err = teamStore.UpdateTeamMemberSelectedAt(ctx, team.ID, user.ID)
		if err != nil {
			t.Fatalf("UpdateTeamMemberUpdatedAt() error = %v", err)
		}
		if err != nil {
			t.Fatalf("UpdateTeamMemberUpdatedAt() error = %v", err)
		}

		// Fetch the member again to check updated_at
		updated, err := teamStore.FindTeamMemberByTeamAndUserId(ctx, team.ID, user.ID)
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
func TestUpdateTeamMemberSelectedAt(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
		teamStore := stores.NewPostgresTeamStore(dbxx)
		userStore := stores.NewPostgresUserStore(dbxx)

		// Create team and user
		team, err := teamStore.CreateTeam(ctx, "SelectedAtTeam", "selected-at-team-slug")
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}
		user, err := userStore.CreateUser(ctx, &models.User{
			Email: "selectedat@example.com",
		})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}
		member, err := teamStore.CreateTeamMember(ctx, team.ID, user.ID, models.TeamMemberRoleMember, true)
		if err != nil {
			t.Fatalf("CreateTeamMember() error = %v", err)
		}
		original := member.LastSelectedAt
		// Call UpdateTeamMemberSelectedAt
		err = teamStore.UpdateTeamMemberSelectedAt(ctx, team.ID, user.ID)
		if err != nil {
			t.Fatalf("UpdateTeamMemberSelectedAt() error = %v", err)
		}

		// Fetch the member again and check last_selected_at
		updated, err := teamStore.FindTeamMemberByTeamAndUserId(ctx, team.ID, user.ID)
		if err != nil {
			t.Fatalf("FindTeamMemberByTeamAndUserId() error = %v", err)
		}
		if updated == nil {
			t.Fatalf("Updated member not found")
		}
		if updated.LastSelectedAt.IsZero() {
			t.Errorf("Expected LastSelectedAt to be set, got zero value")
		}
		// Should be within a reasonable time window (2s)
		if !updated.LastSelectedAt.After(original) {
			t.Errorf("LastSelectedAt not updated recently: %v", updated.LastSelectedAt)
		}
		return errors.New("rollback")
	})
}

func TestPostgresTeamStore_CheckTeamSlug(t *testing.T) {
	type fields struct {
		db database.Dbx
	}
	type args struct {
		ctx  context.Context
		slug string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := stores.NewPostgresTeamStore(tt.fields.db)
			got, err := s.CheckTeamSlug(tt.args.ctx, tt.args.slug)
			if (err != nil) != tt.wantErr {
				t.Errorf("PostgresTeamStore.CheckTeamSlug() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("PostgresTeamStore.CheckTeamSlug() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPostgresTeamStore_UpdateTeamMember(t *testing.T) {
	type fields struct {
		db database.Dbx
	}
	type args struct {
		ctx    context.Context
		member *models.TeamMember
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *models.TeamMember
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := stores.NewPostgresTeamStore(tt.fields.db)
			got, err := s.UpdateTeamMember(tt.args.ctx, tt.args.member)
			if (err != nil) != tt.wantErr {
				t.Errorf("PostgresTeamStore.UpdateTeamMember() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PostgresTeamStore.UpdateTeamMember() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPostgresTeamStore_CountTeamMembers(t *testing.T) {
	type fields struct {
		db database.Dbx
	}
	type args struct {
		ctx    context.Context
		teamId uuid.UUID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int64
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := stores.NewPostgresTeamStore(tt.fields.db)
			got, err := s.CountTeamMembers(tt.args.ctx, tt.args.teamId)
			if (err != nil) != tt.wantErr {
				t.Errorf("PostgresTeamStore.CountTeamMembers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("PostgresTeamStore.CountTeamMembers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPostgresInvitationStore_CRUD(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
		teamStore := stores.NewPostgresTeamStore(dbxx)
		userStore := stores.NewPostgresUserStore(dbxx)

		// Create team and user
		team, err := teamStore.CreateTeam(ctx, "InviteTeam", "invite-team-slug")
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}
		user, err := userStore.CreateUser(ctx, &models.User{Email: "inviter@example.com"})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}
		member, err := teamStore.CreateTeamMember(ctx, team.ID, user.ID, models.TeamMemberRole("owner"), true)
		if err != nil {
			t.Fatalf("CreateTeamMember() error = %v", err)
		}

		// Create invitation
		token := uuid.NewString()
		expiresAt := time.Now().Add(24 * time.Hour)
		invitation := &models.TeamInvitation{
			TeamID:          team.ID,
			InviterMemberID: member.ID,
			Email:           "invitee@example.com",
			Role:            models.TeamMemberRoleMember,
			Token:           token,
			Status:          models.TeamInvitationStatusPending,
			ExpiresAt:       expiresAt,
		}
		err = teamStore.CreateInvitation(ctx, invitation)
		if err != nil {
			t.Fatalf("CreateInvitation() error = %v", err)
		}
		invitation, err = teamStore.FindInvitationByToken(ctx, token)
		if err != nil {
			t.Fatalf("FindInvitationByToken() error = %v", err)
		}
		if invitation.ID == uuid.Nil {
			t.Errorf("Expected invitation ID to be set")
		}

		// Find by ID
		found, err := teamStore.FindInvitationByID(ctx, invitation.ID)
		if err != nil || found == nil || found.ID != invitation.ID {
			t.Errorf("FindInvitationByID() = %v, err = %v", found, err)
		}

		// Find by Token
		foundByToken, err := teamStore.FindInvitationByToken(ctx, token)
		if err != nil || foundByToken == nil || foundByToken.Token != token {
			t.Errorf("FindInvitationByToken() = %v, err = %v", foundByToken, err)
		}

		// Find all for team
		invs, err := teamStore.FindTeamInvitations(ctx, team.ID)
		if err != nil || len(invs) == 0 {
			t.Errorf("FindTeamInvitations() = %v, err = %v", invs, err)
		}

		// GetInvitationByID
		got, err := teamStore.GetInvitationByID(ctx, invitation.ID)
		if err != nil || got == nil || got.ID != invitation.ID {
			t.Errorf("GetInvitationByID() = %v, err = %v", got, err)
		}

		// Update invitation
		invitation.Status = models.TeamInvitationStatusAccepted
		err = teamStore.UpdateInvitation(ctx, invitation)
		if err != nil {
			t.Errorf("UpdateInvitation() error = %v", err)
		}
		updated, err := teamStore.FindInvitationByID(ctx, invitation.ID)
		if err != nil || updated.Status != models.TeamInvitationStatusAccepted {
			t.Errorf("UpdateInvitation() did not update status: %v, err = %v", updated, err)
		}

		return errors.New("rollback")
	})
}

func TestPostgresTeamStore_FindTeamByStripeCustomerId(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
		teamStore := stores.NewPostgresTeamStore(dbxx)
		stripeID := "cus_test_123"
		team, err := teamStore.CreateTeam(ctx, "StripeTeam", "stripe-team-slug")
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}
		customerStore := stores.NewPostgresStripeStore(dbxx)
		customer, err := customerStore.CreateCustomer(ctx, &models.StripeCustomer{
			ID:           stripeID,
			TeamID:       types.Pointer(team.ID),
			CustomerType: models.StripeCustomerTypeTeam,
		})
		if err != nil {
			t.Fatalf("CreateCustomer() error = %v", err)
		}
		if customer == nil || customer.ID != stripeID {
			t.Errorf("CreateCustomer() = %v, want ID %v", customer, stripeID)
		}
		found, err := teamStore.FindTeamByStripeCustomerId(ctx, stripeID)
		if err != nil {
			t.Fatalf("FindTeamByStripeCustomerId() error = %v", err)
		}
		if found == nil || found.ID != team.ID {
			t.Errorf("FindTeamByStripeCustomerId() = %v, want %v", found, team.ID)
		}
		return errors.New("rollback")
	})
}
func TestPostgresTeamStore_FindPendingInvitation(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
		teamStore := stores.NewPostgresTeamStore(dbxx)
		userStore := stores.NewPostgresUserStore(dbxx)

		// Create team and user
		team, err := teamStore.CreateTeam(ctx, "PendingInviteTeam", "pending-invite-team-slug")
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}
		user, err := userStore.CreateUser(ctx, &models.User{
			Email: "pendinginvite@example.com",
		})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}
		member, err := teamStore.CreateTeamMember(ctx, team.ID, user.ID, models.TeamMemberRoleOwner, true)
		if err != nil {
			t.Fatalf("CreateTeamMember() error = %v", err)
		}

		// Create a pending invitation
		token := uuid.NewString()
		expiresAt := time.Now().Add(1 * time.Hour)
		invitation := &models.TeamInvitation{
			TeamID:          team.ID,
			InviterMemberID: member.ID,
			Email:           "invitee-pending@example.com",
			Role:            models.TeamMemberRoleMember,
			Token:           token,
			Status:          models.TeamInvitationStatusPending,
			ExpiresAt:       expiresAt,
		}
		err = teamStore.CreateInvitation(ctx, invitation)
		if err != nil {
			t.Fatalf("CreateInvitation() error = %v", err)
		}

		// Should find the pending invitation
		found, err := teamStore.FindPendingInvitation(ctx, team.ID, "invitee-pending@example.com")
		if err != nil {
			t.Fatalf("FindPendingInvitation() error = %v", err)
		}
		if found == nil || found.Email != "invitee-pending@example.com" {
			t.Errorf("FindPendingInvitation() = %v, want email %v", found, "invitee-pending@example.com")
		}

		// Create an expired invitation
		expiredInvitation := &models.TeamInvitation{
			TeamID:          team.ID,
			InviterMemberID: member.ID,
			Email:           "expired@example.com",
			Role:            models.TeamMemberRoleMember,
			Token:           uuid.NewString(),
			Status:          models.TeamInvitationStatusPending,
			ExpiresAt:       time.Now().Add(-1 * time.Hour),
		}
		err = teamStore.CreateInvitation(ctx, expiredInvitation)
		if err != nil {
			t.Fatalf("CreateInvitation() error = %v", err)
		}

		// Should not find the expired invitation
		expired, err := teamStore.FindPendingInvitation(ctx, team.ID, "expired@example.com")
		if err != nil {
			t.Fatalf("FindPendingInvitation() error = %v", err)
		}
		if expired != nil {
			t.Errorf("Expected no pending invitation for expired, got %v", expired)
		}

		// Should not find for wrong email
		notFound, err := teamStore.FindPendingInvitation(ctx, team.ID, "notfound@example.com")
		if err != nil {
			t.Fatalf("FindPendingInvitation() error = %v", err)
		}
		if notFound != nil {
			t.Errorf("Expected nil for not found email, got %v", notFound)
		}

		// Should not find for wrong team
		otherTeam, err := teamStore.CreateTeam(ctx, "OtherTeam", "other-team-slug")
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}
		other, err := teamStore.FindPendingInvitation(ctx, otherTeam.ID, "invitee-pending@example.com")
		if err != nil {
			t.Fatalf("FindPendingInvitation() error = %v", err)
		}
		if other != nil {
			t.Errorf("Expected nil for other team, got %v", other)
		}

		return errors.New("rollback")
	})
}
