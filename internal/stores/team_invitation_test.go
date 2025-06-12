package stores_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/stores"
	"github.com/tkahng/authgo/internal/test"
)

func TestTeamStore_InvitationCRUD(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
		adapter := stores.NewStorageAdapter(dbxx)
		store := stores.NewDbTeamInvitationStore(dbxx)
		user, err := adapter.User().CreateUser(
			ctx,
			&models.User{
				Email: "newuser@example.com",
			},
		) // Should be nil, just to check method exists
		if err != nil || user == nil {
			t.Errorf("CreateUser() error = %v, user = %v", err, user)
		}

		// Create team and member
		team, err := adapter.TeamGroup().CreateTeam(ctx, "InviteTeam", "invite-team-slug")
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}
		userID := user.ID
		member, err := adapter.TeamMember().CreateTeamMember(ctx, team.ID, userID, "owner", false)
		if err != nil {
			t.Fatalf("CreateTeamMember() error = %v", err)
		}

		// Create a valid invitation (not expired)
		token := uuid.NewString()
		now := time.Now()
		notExpired := now.Add(24 * time.Hour)
		inv := &models.TeamInvitation{
			TeamID:          team.ID,
			InviterMemberID: member.ID,
			Email:           "invitee@example.com",
			Role:            models.TeamMemberRoleMember,
			Token:           token,
			Status:          models.TeamInvitationStatusPending,
			ExpiresAt:       notExpired,
		}

		err = store.CreateInvitation(ctx, inv)
		if err != nil {
			t.Fatalf("CreateInvitation() error = %v", err)
		}
		inv, err = store.FindInvitationByToken(ctx, token)
		if err != nil || inv == nil {
			t.Errorf("FindInvitationByToken() = %v, err = %v", inv, err)
		}
		// FindTeamInvitations should return the invitation
		invs, err := store.FindTeamInvitations(ctx, team.ID)
		if err != nil || len(invs) == 0 {
			t.Errorf("FindTeamInvitations() = %v, err = %v", invs, err)
		}
		// original := inv.ExpiresAt

		inv.ExpiresAt = now
		err = store.UpdateInvitation(ctx, inv)
		if err != nil {
			t.Errorf("UpdateInvitation() error = %v", err)
		}
		time.Sleep(1 * time.Second)
		// FindInvitationByID should return ErrTokenExpired (not expired)
		newinv, err := store.FindInvitationByToken(ctx, token)
		if err == nil || newinv != nil {
			t.Fatalf("FindInvitationByID() = %v, err = %v", newinv, err)
		}
		if err != shared.ErrTokenExpired {
			t.Fatalf("FindInvitationByID() expected ErrTokenExpired, got %v", err)
		}

		// Update invitation status
		inv.Status = models.TeamInvitationStatusAccepted
		err = store.UpdateInvitation(ctx, inv)
		if err != nil {
			t.Fatalf("UpdateInvitation() error = %v", err)
		}

		// Create an expired invitation
		expiredToken := uuid.NewString()
		expiredTime := time.Now()
		expiredInv := &models.TeamInvitation{
			TeamID:          team.ID,
			InviterMemberID: member.ID,
			Email:           "expired@example.com",
			Role:            models.TeamMemberRoleMember,
			Token:           expiredToken,
			Status:          models.TeamInvitationStatusPending,
			ExpiresAt:       expiredTime,
		}
		err = store.CreateInvitation(ctx, expiredInv)
		if err != nil {
			t.Fatalf("CreateInvitation() error = %v", err)
		}

		// FindInvitationByID should succeed for expired invitation
		found, err := store.FindInvitationByToken(ctx, expiredToken)
		if err == nil || found != nil {
			t.Errorf("FindInvitationByID (expired) = %v, err = %v", found, err)
		}

		return errors.New("rollback")
	})
}

func TestInvitationStore_CRUD(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
		adapter := stores.NewStorageAdapter(dbxx)
		teamStore := stores.NewDbTeamInvitationStore(dbxx)

		// Create team and user
		team, err := adapter.TeamGroup().CreateTeam(ctx, "InviteTeam", "invite-team-slug")
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}
		user, err := adapter.User().CreateUser(ctx, &models.User{Email: "inviter@example.com"})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}
		member, err := adapter.TeamMember().CreateTeamMember(ctx, team.ID, user.ID, models.TeamMemberRole("owner"), true)
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
func TestTeamStore_FindPendingInvitation(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	_ = dbx.RunInTx(func(dbxx database.Dbx) error {
		adapter := stores.NewStorageAdapter(dbxx)
		invitationStore := stores.NewDbTeamInvitationStore(dbxx)

		// Create team and user
		team, err := adapter.TeamGroup().CreateTeam(ctx, "PendingInviteTeam", "pending-invite-team-slug")
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}
		user, err := adapter.User().CreateUser(ctx, &models.User{
			Email: "pendinginvite@example.com",
		})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}
		member, err := adapter.TeamMember().CreateTeamMember(ctx, team.ID, user.ID, models.TeamMemberRoleOwner, true)
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
		err = invitationStore.CreateInvitation(ctx, invitation)
		if err != nil {
			t.Fatalf("CreateInvitation() error = %v", err)
		}

		// Should find the pending invitation
		found, err := invitationStore.FindPendingInvitation(ctx, team.ID, "invitee-pending@example.com")
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
		err = invitationStore.CreateInvitation(ctx, expiredInvitation)
		if err != nil {
			t.Fatalf("CreateInvitation() error = %v", err)
		}

		// Should not find the expired invitation
		expired, err := invitationStore.FindPendingInvitation(ctx, team.ID, "expired@example.com")
		if err != nil {
			t.Fatalf("FindPendingInvitation() error = %v", err)
		}
		if expired != nil {
			t.Errorf("Expected no pending invitation for expired, got %v", expired)
		}

		// Should not find for wrong email
		notFound, err := invitationStore.FindPendingInvitation(ctx, team.ID, "notfound@example.com")
		if err != nil {
			t.Fatalf("FindPendingInvitation() error = %v", err)
		}
		if notFound != nil {
			t.Errorf("Expected nil for not found email, got %v", notFound)
		}

		// Should not find for wrong team
		otherTeam, err := adapter.TeamGroup().CreateTeam(ctx, "OtherTeam", "other-team-slug")
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}
		other, err := invitationStore.FindPendingInvitation(ctx, otherTeam.ID, "invitee-pending@example.com")
		if err != nil {
			t.Fatalf("FindPendingInvitation() error = %v", err)
		}
		if other != nil {
			t.Errorf("Expected nil for other team, got %v", other)
		}

		return errors.New("rollback")
	})
}
