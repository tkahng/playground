package stores

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/repository"
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/internal/tools/types"
)

type TeamInvitationFilter struct {
	PaginatedInput
	SortParams
	TeamIds   []uuid.UUID                    `query:"team_ids,omitempty" json:"team_ids,omitempty" format:"uuid" required:"false"`
	Emails    []string                       `query:"emails,omitempty" json:"emails,omitempty" required:"false" minimum:"1" maximum:"100" format:"email"`
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
	AcceptInvitation(
		ctx context.Context,
		adapter StorageAdapterInterface,
		userId uuid.UUID,
		invitationToken string,
		out *models.TeamMember,
	) error
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
	if len(params.Emails) > 0 {
		where[models.TeamInvitationTable.Email] = map[string]any{
			repository.In: params.Emails,
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

var invitationStatuses = []models.TeamInvitationStatus{
	models.TeamInvitationStatusPending,
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
			models.TeamInvitationTable.Status: map[string]any{
				"_in": invitationStatuses,
			},
		},
	)
	if err != nil {
		return nil, err
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
		},
	)
	if err != nil {
		return nil, err
	}
	if invitation == nil {
		return nil, nil
	}
	if invitation.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("invitation expired")
	}
	return invitation, nil
}

func (i *DbTeamInvitationStore) AcceptInvitation(
	ctx context.Context,
	adapter StorageAdapterInterface,
	userId uuid.UUID,
	invitationToken string,
	out *models.TeamMember,
) error {
	invite, err := adapter.TeamInvitation().FindInvitationByToken(ctx, invitationToken)
	if err != nil {
		return err
	}
	if invite == nil {
		return fmt.Errorf("invitation not found")
	}
	user, err := adapter.User().FindUserByID(ctx, userId)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}
	if invite.Email != user.Email {
		return fmt.Errorf("user does not match invitation")
	}
	if invite.Status != models.TeamInvitationStatusPending {
		return fmt.Errorf("invitation is not pending")
	}
	invite.Status = models.TeamInvitationStatusAccepted
	teamMember, err := adapter.TeamMember().CreateTeamMember(ctx, invite.TeamID, user.ID, invite.Role, false)
	if err != nil {
		return err
	}
	if teamMember == nil {
		return fmt.Errorf("team member not created")
	}
	err = adapter.TeamInvitation().UpdateInvitation(ctx, invite)
	if err != nil {
		return err
	}
	*out = *teamMember
	return nil
}
