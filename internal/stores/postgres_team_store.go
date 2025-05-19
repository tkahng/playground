package stores

import (
	"context"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/crudrepo"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/services"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/types"
)

type PostgresTeamStore struct {
	db database.Dbx
}

// FindUserByID implements services.TeamInvitationStore.
func (p *PostgresTeamStore) FindUserByID(ctx context.Context, userId uuid.UUID) (*models.User, error) {
	user, err := crudrepo.User.GetOne(
		ctx,
		p.db,
		&map[string]any{
			"id": map[string]any{
				"_eq": userId.String(),
			},
		},
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (p *PostgresTeamStore) FindTeamInvitations(ctx context.Context, teamId uuid.UUID) ([]*models.TeamInvitation, error) {
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
func (p *PostgresTeamStore) FindInvitationByID(ctx context.Context, invitationId uuid.UUID) (*models.TeamInvitation, error) {
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
	if !invitation.ExpiresAt.Before(time.Now()) {
		return nil, shared.ErrTokenExpired
	}
	return invitation, nil
}

// FindInvitationByToken implements services.TeamInvitationStore.
func (p *PostgresTeamStore) FindInvitationByToken(ctx context.Context, token string) (*models.TeamInvitation, error) {
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
	if !invitation.ExpiresAt.Before(time.Now()) {
		return nil, shared.ErrTokenExpired
	}
	return invitation, nil
}

// CreateInvitation implements services.TeamInvitationStore.
func (p *PostgresTeamStore) CreateInvitation(ctx context.Context, invitation *models.TeamInvitation) error {
	_, err := crudrepo.TeamInvitation.PostOne(
		ctx,
		p.db,
		invitation,
	)
	return err
}

// GetInvitationByID implements services.TeamInvitationStore.
func (p *PostgresTeamStore) GetInvitationByID(ctx context.Context, invitationId uuid.UUID) (*models.TeamInvitation, error) {
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
	if !invitation.ExpiresAt.Before(time.Now()) {
		return nil, shared.ErrTokenExpired
	}
	return invitation, nil
}

// UpdateInvitation implements services.TeamInvitationStore.
func (p *PostgresTeamStore) UpdateInvitation(ctx context.Context, invitation *models.TeamInvitation) error {
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

// DeleteTeamMember implements services.TeamStore.
func (s *PostgresTeamStore) DeleteTeamMember(ctx context.Context, teamId uuid.UUID, userId uuid.UUID) error {
	_, err := crudrepo.TeamMember.Delete(
		ctx,
		s.db,
		&map[string]any{
			"team_id": map[string]any{
				"_eq": teamId.String(),
			},
			"user_id": map[string]any{
				"_eq": userId.String(),
			},
		},
	)
	if err != nil {
		return err
	}
	return nil
}

// CheckTeamSlug implements services.TeamStore.
func (s *PostgresTeamStore) CheckTeamSlug(ctx context.Context, slug string) (bool, error) {
	team, err := crudrepo.Team.GetOne(
		ctx,
		s.db,
		&map[string]any{
			"slug": map[string]any{
				"_eq": slug,
			},
		},
	)
	if err != nil {
		return false, err
	}
	if team == nil {
		return true, nil
	}
	return false, nil
}

// UpdateTeamMember implements services.TeamStore.
func (s *PostgresTeamStore) UpdateTeamMember(ctx context.Context, member *models.TeamMember) (*models.TeamMember, error) {
	newMember, err := crudrepo.TeamMember.PutOne(
		ctx,
		s.db,
		member,
	)
	if err != nil {
		return nil, err
	}
	return newMember, nil
}

// CountTeamMembers implements services.TeamStore.
func (s *PostgresTeamStore) CountTeamMembers(ctx context.Context, teamId uuid.UUID) (int64, error) {
	c, err := crudrepo.TeamMember.Count(
		ctx,
		s.db,
		&map[string]any{
			"team_id": map[string]any{
				"_eq": teamId.String(),
			},
		},
	)
	if err != nil {
		return 0, err
	}
	return c, nil
}

func NewPostgresTeamStore(db database.Dbx) *PostgresTeamStore {
	return &PostgresTeamStore{
		db: db,
	}
}

// var _ services.TeamInvitationStore = &PostgresTeamStore{}
var _ services.TeamStore = &PostgresTeamStore{}
var _ services.TeamInvitationStore = &PostgresTeamStore{}

func (s *PostgresTeamStore) FindTeamByStripeCustomerId(ctx context.Context, stripeCustomerId string) (*models.Team, error) {
	data, err := crudrepo.Team.GetOne(
		ctx,
		s.db,
		&map[string]any{
			"stripe_customer_id": map[string]any{"_eq": stripeCustomerId},
		},
	)
	return database.OptionalRow(data, err)
}

func (q *PostgresTeamStore) FindTeamMemberByTeamAndUserId(ctx context.Context, teamId, userId uuid.UUID) (*models.TeamMember, error) {
	teamMember, err := crudrepo.TeamMember.GetOne(
		ctx,
		q.db,
		&map[string]any{
			"user_id": map[string]any{
				"_eq": userId.String(),
			},
			"team_id": map[string]any{
				"_eq": teamId.String(),
			},
		},
	)
	if err != nil {
		return nil, err
	}
	return teamMember, nil
}

// UpdateTeamMemberSelectedAt implements TeamQueryer.
func (q *PostgresTeamStore) UpdateTeamMemberSelectedAt(ctx context.Context, teamId, userId uuid.UUID) error {
	qquery := squirrel.Update("team_members").
		Where("team_id = ?", teamId).
		Where("user_id = ?", userId).
		Set("last_selected_at", time.Now())

	err := database.ExecWithBuilder(ctx, q.db, qquery.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return err
	}
	return nil
}

// FindLatestTeamMemberByUserID implements TeamQueryer.
func (q *PostgresTeamStore) FindLatestTeamMemberByUserID(ctx context.Context, userId uuid.UUID) (*models.TeamMember, error) {
	teamMember, err := crudrepo.TeamMember.Get(
		ctx,
		q.db,
		&map[string]any{
			"user_id": map[string]any{
				"_eq": userId.String(),
			},
		},
		&map[string]string{
			"last_selected_at": "DESC",
		},
		types.Pointer(1),
		nil,
	)
	if err != nil {
		return nil, err
	}
	if len(teamMember) == 0 {
		return nil, nil
	}
	return teamMember[0], nil
}

// DeleteTeam implements TeamQueryer.
func (q *PostgresTeamStore) DeleteTeam(ctx context.Context, teamId uuid.UUID) error {
	_, err := crudrepo.Team.Delete(
		ctx,
		q.db,
		&map[string]any{
			"id": map[string]any{
				"_eq": teamId.String(),
			},
		},
	)
	if err != nil {
		return err
	}
	return nil
}

// FindTeamByID implements TeamQueryer.
func (q *PostgresTeamStore) FindTeamByID(ctx context.Context, teamId uuid.UUID) (*models.Team, error) {
	return crudrepo.Team.GetOne(
		ctx,
		q.db,
		&map[string]any{
			"id": map[string]any{
				"_eq": teamId.String(),
			},
		},
	)
}

// FindTeamMembersByUserID implements TeamQueryer.
func (q *PostgresTeamStore) FindTeamMembersByUserID(ctx context.Context, userId uuid.UUID) ([]*models.TeamMember, error) {
	teamMembers, err := crudrepo.TeamMember.Get(
		ctx,
		q.db,
		&map[string]any{
			"user_id": map[string]any{
				"_eq": userId.String(),
			},
		},
		&map[string]string{
			"last_selected_at": "DESC",
		},
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return teamMembers, nil
}

// UpdateTeam implements TeamQueryer.
func (q *PostgresTeamStore) UpdateTeam(ctx context.Context, teamId uuid.UUID, name string, stripeCustomerId *string) (*models.Team, error) {
	team := &models.Team{
		ID:               teamId,
		Name:             name,
		StripeCustomerID: stripeCustomerId,
		UpdatedAt:        time.Now(),
	}
	_, err := crudrepo.Team.PutOne(
		ctx,
		q.db,
		team,
	)
	if err != nil {
		return nil, err
	}
	return team, nil
}

func (s *PostgresTeamStore) UpsertTeamCustomerStripeId(ctx context.Context, teamId uuid.UUID, stripeCustomerId *string) error {
	var dbx database.Dbx = s.db
	q := squirrel.Update("teams").
		Set("stripe_customer_id", stripeCustomerId).
		Where("id = ?", teamId)
	return database.ExecWithBuilder(ctx, dbx, q.PlaceholderFormat(squirrel.Dollar))
}

func (q *PostgresTeamStore) CreateTeam(ctx context.Context, name string, slug string, stripeCustomerId *string) (*models.Team, error) {
	teamModel := &models.Team{
		Name:             name,
		Slug:             slug,
		StripeCustomerID: stripeCustomerId,
	}
	team, err := crudrepo.Team.PostOne(
		ctx,
		q.db,
		teamModel,
	)
	if err != nil {
		return nil, err
	}
	return team, nil
}

func (q *PostgresTeamStore) CreateTeamMember(ctx context.Context, teamId, userId uuid.UUID, role models.TeamMemberRole) (*models.TeamMember, error) {
	teamMember := &models.TeamMember{
		TeamID: teamId,
		UserID: &userId,
		Role:   role,
		Active: true,
	}
	return crudrepo.TeamMember.PostOne(
		ctx,
		q.db,
		teamMember,
	)
}
