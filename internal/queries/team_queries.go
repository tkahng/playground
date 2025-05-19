package queries

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/crudrepo"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
)

func CreateTeamFromUser(ctx context.Context, dbx database.Dbx, user *models.User) (*models.TeamMember, error) {
	team, err := func() (*models.Team, error) {
		teamModel := &models.Team{
			Name:             user.Email,
			Slug:             user.Email,
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
