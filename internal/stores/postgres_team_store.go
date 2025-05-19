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

func NewPostgresTeamStore(db database.Dbx) *PostgresTeamStore {
	return &PostgresTeamStore{
		db: db,
	}
}

var _ services.TeamStore = &PostgresTeamStore{}

func (q *PostgresTeamStore) FindTeamMemberByUserAndTeamID(ctx context.Context, userId, teamId uuid.UUID) (*models.TeamMember, error) {
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

// UpdateTeamMemberUpdatedAt implements TeamQueryer.
func (q *PostgresTeamStore) UpdateTeamMemberUpdatedAt(ctx context.Context, teamMemberId uuid.UUID) error {
	qquery := squirrel.Update("team_members").
		Where("id = ?", teamMemberId).
		Set("updated_at", time.Now())

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
			"updated_at": "DESC",
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
			"updated_at": "DESC",
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

func (q *PostgresTeamStore) CreateTeamFromUser(ctx context.Context, user *models.User) (*models.Team, error) {
	team, err := q.CreateTeam(ctx, user.Email, nil)
	if err != nil {
		return nil, err
	}
	teamMember, err := q.CreateTeamMember(ctx, team.ID, user.ID, models.TeamMemberRoleAdmin)
	if err != nil {
		return nil, err
	}
	team.Members = []*models.TeamMember{teamMember}
	teamMember.Team = team
	return team, nil
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
	}
	return crudrepo.TeamMember.PostOne(
		ctx,
		q.db,
		teamMember,
	)
}
