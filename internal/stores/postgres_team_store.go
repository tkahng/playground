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
	"github.com/tkahng/authgo/internal/tools/types"
)

type PostgresTeamStore struct {
	db database.Dbx
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

var _ services.TeamStore = &PostgresTeamStore{}

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

func (q *PostgresTeamStore) CreateTeam(ctx context.Context, name string, stripeCustomerId *string) (*models.Team, error) {
	teamModel := &models.Team{
		Name:             name,
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
