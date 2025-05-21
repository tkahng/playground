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

func TestPostgresTeamStore_InvitationCRUD(t *testing.T) {
	test.Short(t)
	ctx, dbx := test.DbSetup()
	dbx.RunInTransaction(ctx, func(dbxx database.Dbx) error {
		store := stores.NewPostgresTeamStore(dbxx)
		userStore := stores.NewPostgresUserStore(dbxx)
		user, err := userStore.CreateUser(
			ctx,
			&models.User{
				Email: "newuser@example.com",
			},
		) // Should be nil, just to check method exists
		if err != nil || user == nil {
			t.Errorf("CreateUser() error = %v, user = %v", err, user)
		}

		// Create team and member
		team, err := store.CreateTeam(ctx, "InviteTeam", "invite-team-slug")
		if err != nil {
			t.Fatalf("CreateTeam() error = %v", err)
		}
		userID := user.ID
		member, err := store.CreateTeamMember(ctx, team.ID, userID, models.TeamMemberRole("owner"))
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
