package stores

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/crudrepo"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
)

func (s *DbTeamStore) FindTeamInvitations(ctx context.Context, teamId uuid.UUID) ([]*models.TeamInvitation, error) {
	invitations, err := crudrepo.TeamInvitation.Get(
		ctx,
		s.db,
		&map[string]any{
			models.TeamInvitationTable.TeamID: map[string]any{
				"_eq": teamId,
			},
			models.TeamInvitationTable.Status: map[string]any{
				"_eq": string(models.TeamInvitationStatusPending),
			},
			models.TeamInvitationTable.ExpiresAt: map[string]any{
				"_gt": time.Now(),
			},
		},
		&map[string]string{
			models.TeamInvitationTable.CreatedAt: "desc",
		},
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return invitations, nil
}

// FindInvitationByID implements services.TeamInvitationStore.
func (s *DbTeamStore) FindInvitationByID(ctx context.Context, invitationId uuid.UUID) (*models.TeamInvitation, error) {
	invitation, err := crudrepo.TeamInvitation.GetOne(
		ctx,
		s.db,
		&map[string]any{
			models.TeamInvitationTable.ID: map[string]any{
				"_eq": invitationId,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	if invitation == nil {
		return nil, nil
	}
	if invitation.ExpiresAt.Before(time.Now()) {
		fmt.Println("Invitation expired")
		fmt.Println(invitation.ExpiresAt)
		return nil, shared.ErrTokenExpired
	}
	return invitation, nil
}

// FindInvitationByToken implements services.TeamInvitationStore.
func (s *DbTeamStore) FindInvitationByToken(ctx context.Context, token string) (*models.TeamInvitation, error) {
	invitation, err := crudrepo.TeamInvitation.GetOne(
		ctx,
		s.db,
		&map[string]any{
			models.TeamInvitationTable.Token: map[string]any{
				"_eq": token,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	if invitation == nil {
		return nil, nil
	}
	if invitation.ExpiresAt.Before(time.Now()) {
		fmt.Println("Invitation expired")
		fmt.Println(invitation.ExpiresAt)
		return nil, shared.ErrTokenExpired
	}
	return invitation, nil
}

// CreateInvitation implements services.TeamInvitationStore.
func (s *DbTeamStore) CreateInvitation(ctx context.Context, invitation *models.TeamInvitation) error {
	_, err := crudrepo.TeamInvitation.PostOne(
		ctx,
		s.db,
		invitation,
	)
	return err
}

// GetInvitationByID implements services.TeamInvitationStore.
func (s *DbTeamStore) GetInvitationByID(ctx context.Context, invitationId uuid.UUID) (*models.TeamInvitation, error) {
	invitation, err := crudrepo.TeamInvitation.GetOne(
		ctx,
		s.db,
		&map[string]any{
			models.TeamInvitationTable.ID: map[string]any{
				"_eq": invitationId,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	if invitation == nil {
		return nil, nil
	}
	if invitation.ExpiresAt.Before(time.Now()) {
		return nil, shared.ErrTokenExpired
	}
	return invitation, nil
}

// UpdateInvitation implements services.TeamInvitationStore.
func (s *DbTeamStore) UpdateInvitation(ctx context.Context, invitation *models.TeamInvitation) error {
	_, err := crudrepo.TeamInvitation.PutOne(
		ctx,
		s.db,
		invitation,
	)

	if err != nil {
		return err
	}
	return nil
}

// FindPendingInvitation implements services.TeamInvitationStore.
func (s *DbTeamStore) FindPendingInvitation(ctx context.Context, teamId uuid.UUID, email string) (*models.TeamInvitation, error) {
	invitation, err := crudrepo.TeamInvitation.GetOne(
		ctx,
		s.db,
		&map[string]any{
			models.TeamInvitationTable.TeamID: map[string]any{
				"_eq": teamId,
			},
			models.TeamInvitationTable.Email: map[string]any{
				"_eq": email,
			},
			models.TeamInvitationTable.Status: map[string]any{
				"_eq": string(models.TeamInvitationStatusPending),
			},
			models.TeamInvitationTable.ExpiresAt: map[string]any{
				"_gt": time.Now(),
			},
		},
	)
	if err != nil {
		return nil, err
	}
	return invitation, nil
}
