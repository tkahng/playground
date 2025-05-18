package queries

import (
	"context"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/crudrepo"
	"github.com/tkahng/authgo/internal/db"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/tools/types"
)

type TeamQueryer interface {
	CreateTeamFromUser(ctx context.Context, dbx db.Dbx, user *models.User) (*models.Team, error)
	FindTeamByID(ctx context.Context, dbx db.Dbx, teamId uuid.UUID) (*models.Team, error)
	CreateTeam(ctx context.Context, dbx db.Dbx, name string, stripeCustomerId *string) (*models.Team, error)
	UpdateTeam(ctx context.Context, dbx db.Dbx, teamId uuid.UUID, name string, stripeCustomerId *string) (*models.Team, error)
	DeleteTeam(ctx context.Context, dbx db.Dbx, teamId uuid.UUID) error
	FindTeamMembersByUserID(ctx context.Context, dbx db.Dbx, userId uuid.UUID) ([]*models.TeamMember, error)
	FindLatestTeamMemberByUserID(ctx context.Context, dbx db.Dbx, userId uuid.UUID) (*models.TeamMember, error)
	CreateTeamMember(ctx context.Context, dbx db.Dbx, teamId, userId uuid.UUID, role models.TeamMemberRole) (*models.TeamMember, error)
	UpdateTeamMemberUpdatedAt(ctx context.Context, dbx db.Dbx, teamMemberId uuid.UUID) error
}

type TeamQueries struct {
}

var _ TeamQueryer = &TeamQueries{}

// UpdateTeamMemberUpdatedAt implements TeamQueryer.
func (q *TeamQueries) UpdateTeamMemberUpdatedAt(ctx context.Context, dbx db.Dbx, teamMemberId uuid.UUID) error {
	qquery := squirrel.Update("team_members").
		Where("id = ?", teamMemberId).
		Set("updated_at", time.Now())

	err := ExecWithBuilder(ctx, dbx, qquery.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return err
	}
	// member, err := crudrepo.TeamMember.GetOne(
	// 	ctx,
	// 	dbx,
	// 	&map[string]any{
	// 		"id": map[string]any{
	// 			"_eq": teamMemberId.String(),
	// 		},
	// 	},
	// )
	// if err != nil {
	// 	return err
	// }
	// if member == nil {
	// 	return errors.New("team member not found")
	// }

	// member.UpdatedAt = time.Now()

	// _, err = crudrepo.TeamMember.PutOne(ctx, dbx, member)
	// if err != nil {
	// 	return err
	// }
	return nil
}

// FindLatestTeamMemberByUserID implements TeamQueryer.
func (q *TeamQueries) FindLatestTeamMemberByUserID(ctx context.Context, dbx db.Dbx, userId uuid.UUID) (*models.TeamMember, error) {
	teamMember, err := crudrepo.TeamMember.Get(
		ctx,
		dbx,
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
func (q *TeamQueries) DeleteTeam(ctx context.Context, dbx db.Dbx, teamId uuid.UUID) error {
	_, err := crudrepo.Team.Delete(
		ctx,
		dbx,
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
func (q *TeamQueries) FindTeamByID(ctx context.Context, dbx db.Dbx, teamId uuid.UUID) (*models.Team, error) {
	return FindTeamByID(ctx, dbx, teamId)
}

func FindTeamByID(ctx context.Context, dbx db.Dbx, teamId uuid.UUID) (*models.Team, error) {
	return crudrepo.Team.GetOne(
		ctx,
		dbx,
		&map[string]any{
			"id": map[string]any{
				"_eq": teamId.String(),
			},
		},
	)
}

// FindTeamMembersByUserID implements TeamQueryer.
func (q *TeamQueries) FindTeamMembersByUserID(ctx context.Context, dbx db.Dbx, userId uuid.UUID) ([]*models.TeamMember, error) {
	teamMembers, err := crudrepo.TeamMember.Get(
		ctx,
		dbx,
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
func (q *TeamQueries) UpdateTeam(ctx context.Context, dbx db.Dbx, teamId uuid.UUID, name string, stripeCustomerId *string) (*models.Team, error) {
	team := &models.Team{
		ID:               teamId,
		Name:             name,
		StripeCustomerID: stripeCustomerId,
		UpdatedAt:        time.Now(),
	}
	_, err := crudrepo.Team.PutOne(
		ctx,
		dbx,
		team,
	)
	if err != nil {
		return nil, err
	}
	return team, nil
}
func CreateTeamFromUser(ctx context.Context, dbx db.Dbx, user *models.User) (*models.TeamMember, error) {
	team, err := func() (*models.Team, error) {
		teamModel := &models.Team{
			Name:             user.Email,
			StripeCustomerID: nil,
		}
		team, err := crudrepo.Team.PostOne(
			ctx,
			dbx,
			teamModel,
		)
		if err != nil {
			return nil, err
		}
		return team, nil
	}()
	if err != nil {
		return nil, err
	}
	teamMember, err := func() (*models.TeamMember, error) {
		var userId uuid.UUID = user.ID
		teamMember := &models.TeamMember{
			TeamID: team.ID,
			UserID: &userId,
			Role:   models.TeamMemberRoleAdmin,
		}
		return crudrepo.TeamMember.PostOne(
			ctx,
			dbx,
			teamMember,
		)
	}()
	if err != nil {
		return nil, err
	}
	teamMember.Team = team
	return teamMember, nil
}

func (q *TeamQueries) CreateTeamFromUser(ctx context.Context, dbx db.Dbx, user *models.User) (*models.Team, error) {
	team, err := q.CreateTeam(ctx, dbx, user.Email, nil)
	if err != nil {
		return nil, err
	}
	teamMember, err := q.CreateTeamMember(ctx, dbx, team.ID, user.ID, models.TeamMemberRoleAdmin)
	if err != nil {
		return nil, err
	}
	team.Members = []*models.TeamMember{teamMember}
	teamMember.Team = team
	return team, nil
}

func (*TeamQueries) CreateTeam(ctx context.Context, dbx db.Dbx, name string, stripeCustomerId *string) (*models.Team, error) {
	teamModel := &models.Team{
		Name:             name,
		StripeCustomerID: stripeCustomerId,
	}
	team, err := crudrepo.Team.PostOne(
		ctx,
		dbx,
		teamModel,
	)
	if err != nil {
		return nil, err
	}
	return team, nil
}

func (q *TeamQueries) CreateTeamMember(ctx context.Context, dbx db.Dbx, teamId, userId uuid.UUID, role models.TeamMemberRole) (*models.TeamMember, error) {
	teamMember := &models.TeamMember{
		TeamID: teamId,
		UserID: &userId,
		Role:   role,
	}
	return crudrepo.TeamMember.PostOne(
		ctx,
		dbx,
		teamMember,
	)
}
