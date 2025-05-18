package teamservice

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
)

type TeamService interface {
	SelectMembershipByTeamID(ctx context.Context, userId uuid.UUID, teamId uuid.UUID) (*models.TeamMember, error)
	GetLastMembershipByUserID(ctx context.Context, userId uuid.UUID) (*models.TeamMember, error)
}

type teamService struct {
	teamStore TeamStore
}

// SelectMembershipByTeamID implements TeamService.
func (t *teamService) SelectMembershipByTeamID(ctx context.Context, userId uuid.UUID, teamId uuid.UUID) (*models.TeamMember, error) {
	team, err := t.teamStore.FindTeamMemberByUserAndTeamID(ctx, userId, teamId)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, nil
	}
	err = t.teamStore.UpdateTeamMemberUpdatedAt(ctx, team.ID)
	if err != nil {
		return nil, err
	}
	return team, nil
}

func (t *teamService) GetLastMembershipByUserID(ctx context.Context, userId uuid.UUID) (*models.TeamMember, error) {
	team, err := t.teamStore.FindLatestTeamMemberByUserID(ctx, userId)
	if err != nil {
		return nil, err
	}
	return team, nil
}

func NewTeamService(store TeamStore) TeamService {
	return &teamService{
		teamStore: store,
	}
}
