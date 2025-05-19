package stores

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/crudrepo"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/services"
)

type PostgresInvitationStore struct {
	db database.Dbx
}

func (p *PostgresInvitationStore) FindTeamInvitations(ctx context.Context, teamId uuid.UUID) ([]*models.TeamInvitation, error) {
	invitations, err := crudrepo.TeamInvitation.Get(
		ctx,
		p.db,
		&map[string]any{
			"team_id": map[string]any{
				"_eq": teamId.String(),
			},
		},
		&map[string]string{
			"created_at": "desc",
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
func (p *PostgresInvitationStore) FindInvitationByID(ctx context.Context, invitationId uuid.UUID) (*models.TeamInvitation, error) {
	invitation, err := crudrepo.TeamInvitation.GetOne(
		ctx,
		p.db,
		&map[string]any{
			"id": map[string]any{
				"_eq": invitationId.String(),
			},
		},
	)
	if err != nil {
		return nil, err
	}
	return invitation, nil
}

// FindInvitationByToken implements services.TeamInvitationStore.
func (p *PostgresInvitationStore) FindInvitationByToken(ctx context.Context, token string) (*models.TeamInvitation, error) {
	invitation, err := crudrepo.TeamInvitation.GetOne(
		ctx,
		p.db,
		&map[string]any{
			"token": map[string]any{
				"_eq": token,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	return invitation, nil
}

// CreateInvitation implements services.TeamInvitationStore.
func (p *PostgresInvitationStore) CreateInvitation(ctx context.Context, invitation *models.TeamInvitation) error {
	_, err := crudrepo.TeamInvitation.PostOne(
		ctx,
		p.db,
		invitation,
	)
	return err
}

// GetInvitationByID implements services.TeamInvitationStore.
func (p *PostgresInvitationStore) GetInvitationByID(ctx context.Context, invitationId uuid.UUID) (*models.TeamInvitation, error) {
	invitation, err := crudrepo.TeamInvitation.GetOne(
		ctx,
		p.db,
		&map[string]any{
			"id": map[string]any{
				"_eq": invitationId.String(),
			},
		},
	)
	if err != nil {
		return nil, err
	}
	return invitation, nil
}

// UpdateInvitation implements services.TeamInvitationStore.
func (p *PostgresInvitationStore) UpdateInvitation(ctx context.Context, invitation *models.TeamInvitation) error {
	_, err := crudrepo.TeamInvitation.PutOne(
		ctx,
		p.db,
		invitation,
	)

	if err != nil {
		return err
	}
	return nil
}

func NewPostgresInvitationStore(db database.Dbx) *PostgresInvitationStore {
	return &PostgresInvitationStore{
		db: db,
	}
}

var _ services.TeamInvitationStore = (*PostgresInvitationStore)(nil)
