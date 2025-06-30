package stores

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/types"
)

type TeamInvitationFilter struct {
	PaginatedInput
	SortParams
	TeamIds   []uuid.UUID                    `query:"team_ids,omitempty" json:"team_ids,omitempty" format:"uuid" required:"false"`
	Statuses  []models.TeamInvitationStatus  `query:"statuses,omitempty" json:"statuses,omitempty" required:"false" minimum:"1" maximum:"100" enum:"pending,accepted,declined,canceled"`
	ExpiresAt types.OptionalParam[time.Time] `query:"expires_at,omitempty" json:"expires_at,omitempty" required:"false"`
}

type DbTeamInvitationStoreInterface interface { // size=16 (0x10)
	CreateInvitation(ctx context.Context, invitation *models.TeamInvitation) error
	FindInvitationByID(ctx context.Context, invitationId uuid.UUID) (*models.TeamInvitation, error)
	FindInvitationByToken(ctx context.Context, token string) (*models.TeamInvitation, error)
	FindPendingInvitation(ctx context.Context, teamId uuid.UUID, email string) (*models.TeamInvitation, error)
	FindTeamInvitations(ctx context.Context, params *TeamInvitationFilter) ([]*models.TeamInvitation, error)
	CountTeamInvitations(ctx context.Context, params *TeamInvitationFilter) (int64, error)
	GetInvitationByID(ctx context.Context, invitationId uuid.UUID) (*models.TeamInvitation, error)
	UpdateInvitation(ctx context.Context, invitation *models.TeamInvitation) error
}

type DbTeamInvitationStore struct {
	db database.Dbx
}

var _ DbTeamInvitationStoreInterface = (*DbTeamInvitationStore)(nil)

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

func (s *DbTeamInvitationStore) filter(params *TeamInvitationFilter) *map[string]any {
	where := map[string]any{}
	if len(params.TeamIds) > 0 {
		where[models.TeamInvitationTable.TeamID] = map[string]any{
			repository.In: params.TeamIds,
		}
	}
	if len(params.Statuses) > 0 {
		where[models.TeamInvitationTable.Status] = map[string]any{
			repository.In: params.Statuses,
		}
	}
	if params.ExpiresAt.IsSet {
		where[models.TeamInvitationTable.ExpiresAt] = map[string]any{
			repository.Gte: params.ExpiresAt.Value,
		}
	}
	return &where
}
func (s *DbTeamInvitationStore) CountTeamInvitations(ctx context.Context, params *TeamInvitationFilter) (int64, error) {
	where := s.filter(params)
	return repository.TeamInvitation.Count(
		ctx,
		s.db,
		where,
	)
}

func (s *DbTeamInvitationStore) FindTeamInvitations(ctx context.Context, params *TeamInvitationFilter) ([]*models.TeamInvitation, error) {
	limit, offset := params.LimitOffset()
	where := s.filter(params)
	invitations, err := repository.TeamInvitation.Get(
		ctx,
		s.db,
		where,
		&map[string]string{
			models.TeamInvitationTable.CreatedAt: "desc",
		},
		&limit,
		&offset,
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
