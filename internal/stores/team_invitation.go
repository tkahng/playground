package stores

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/shared"
)

type TeamInvitationParams struct {
	PaginatedInput
	SortParams
}

type DbTeamInvitationStoreInterface interface { // size=16 (0x10)
	CreateInvitation(ctx context.Context, invitation *models.TeamInvitation) error
	FindInvitationByID(ctx context.Context, invitationId uuid.UUID) (*models.TeamInvitation, error)
	FindInvitationByToken(ctx context.Context, token string) (*models.TeamInvitation, error)
	FindPendingInvitation(ctx context.Context, teamId uuid.UUID, email string) (*models.TeamInvitation, error)
	FindTeamInvitations(ctx context.Context, teamId uuid.UUID, params *TeamInvitationParams) ([]*models.TeamInvitation, error)
	GetInvitationByID(ctx context.Context, invitationId uuid.UUID) (*models.TeamInvitation, error)
	UpdateInvitation(ctx context.Context, invitation *models.TeamInvitation) error
}

type DbTeamInvitationStore struct {
	db database.Dbx
}

func NewDbTeamInvitationStore(db database.Dbx) *DbTeamInvitationStore {
	return &DbTeamInvitationStore{
		db: db,
	}
}
func (s *DbTeamInvitationStore) WithTx(db database.Dbx) *DbTeamInvitationStore {
	return &DbTeamInvitationStore{
		db: db,
	}
}

func (s *DbTeamInvitationStore) FindTeamInvitations(ctx context.Context, teamId uuid.UUID, params *TeamInvitationParams) ([]*models.TeamInvitation, error) {
	invitations, err := repository.TeamInvitation.Get(
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

func (s *DbTeamInvitationStore) FindInvitationByID(ctx context.Context, invitationId uuid.UUID) (*models.TeamInvitation, error) {
	invitation, err := repository.TeamInvitation.GetOne(
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
func (s *DbTeamInvitationStore) FindInvitationByToken(ctx context.Context, token string) (*models.TeamInvitation, error) {
	invitation, err := repository.TeamInvitation.GetOne(
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
func (s *DbTeamInvitationStore) CreateInvitation(ctx context.Context, invitation *models.TeamInvitation) error {
	_, err := repository.TeamInvitation.PostOne(
		ctx,
		s.db,
		invitation,
	)
	return err
}

// GetInvitationByID implements services.TeamInvitationStore.
func (s *DbTeamInvitationStore) GetInvitationByID(ctx context.Context, invitationId uuid.UUID) (*models.TeamInvitation, error) {
	invitation, err := repository.TeamInvitation.GetOne(
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
func (s *DbTeamInvitationStore) UpdateInvitation(ctx context.Context, invitation *models.TeamInvitation) error {
	_, err := repository.TeamInvitation.PutOne(
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
func (s *DbTeamInvitationStore) FindPendingInvitation(ctx context.Context, teamId uuid.UUID, email string) (*models.TeamInvitation, error) {
	invitation, err := repository.TeamInvitation.GetOne(
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
