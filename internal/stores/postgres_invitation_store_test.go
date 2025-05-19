package stores_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/stores"
	"github.com/tkahng/authgo/internal/test"
)

func TestPostgresInvitationStore_CRUD(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
		invStore := stores.NewPostgresInvitationStore(dbxx)
		teamStore := stores.NewPostgresTeamStore(dbxx)
		userStore := stores.NewPostgresUserStore(dbxx)

		// Create team and user
		team, err := teamStore.CreateTeam(ctx, "InviteTeam", "invite-team-slug", nil)
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}
		user, err := userStore.CreateUser(ctx, &models.User{Email: "inviter@example.com"})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}
		member, err := teamStore.CreateTeamMember(ctx, team.ID, user.ID, models.TeamMemberRole("owner"))
		if err != nil {
			t.Fatalf("CreateTeamMember() error = %v", err)
		}

		// Create invitation
		token := uuid.NewString()
		expiresAt := time.Now().Add(24 * time.Hour)
		invitation := &models.TeamInvitation{
			TeamID:    team.ID,
			InvitedBy: member.ID,
			Email:     "invitee@example.com",
			Role:      models.TeamMemberRoleMember,
			Token:     token,
			Status:    models.TeamInvitationStatusPending,
			ExpiresAt: expiresAt,
		}
		err = invStore.CreateInvitation(ctx, invitation)
		if err != nil {
			t.Fatalf("CreateInvitation() error = %v", err)
		}
		invitation, err = invStore.FindInvitationByToken(ctx, token)
		if err != nil {
			t.Fatalf("FindInvitationByToken() error = %v", err)
		}
		if invitation.ID == uuid.Nil {
			t.Errorf("Expected invitation ID to be set")
		}

		// Find by ID
		found, err := invStore.FindInvitationByID(ctx, invitation.ID)
		if err != nil || found == nil || found.ID != invitation.ID {
			t.Errorf("FindInvitationByID() = %v, err = %v", found, err)
		}

		// Find by Token
		foundByToken, err := invStore.FindInvitationByToken(ctx, token)
		if err != nil || foundByToken == nil || foundByToken.Token != token {
			t.Errorf("FindInvitationByToken() = %v, err = %v", foundByToken, err)
		}

		// Find all for team
		invs, err := invStore.FindTeamInvitations(ctx, team.ID)
		if err != nil || len(invs) == 0 {
			t.Errorf("FindTeamInvitations() = %v, err = %v", invs, err)
		}

		// GetInvitationByID
		got, err := invStore.GetInvitationByID(ctx, invitation.ID)
		if err != nil || got == nil || got.ID != invitation.ID {
			t.Errorf("GetInvitationByID() = %v, err = %v", got, err)
		}

		// Update invitation
		invitation.Status = models.TeamInvitationStatusAccepted
		err = invStore.UpdateInvitation(ctx, invitation)
		if err != nil {
			t.Errorf("UpdateInvitation() error = %v", err)
		}
		updated, err := invStore.FindInvitationByID(ctx, invitation.ID)
		if err != nil || updated.Status != models.TeamInvitationStatusAccepted {
			t.Errorf("UpdateInvitation() did not update status: %v, err = %v", updated, err)
		}

		return errors.New("rollback")
	})
}
