package queries

import (
	"context"

	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/repository"
)

func CreateTeamFromUser(ctx context.Context, dbx database.Dbx, user *models.User) (*models.TeamMember, error) {
	team, err := func() (*models.Team, error) {
		teamModel := &models.Team{
			Name: user.Email,
			Slug: user.Email,
			// StripeCustomerID: nil,
		}
		team, err := repository.Team.PostOne(
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
		var userId = user.ID
		teamMember := &models.TeamMember{
			TeamID: team.ID,
			UserID: &userId,
			Role:   models.TeamMemberRoleOwner,
		}
		return repository.TeamMember.PostOne(
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
