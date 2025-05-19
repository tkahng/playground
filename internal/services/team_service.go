package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
)

type TeamService interface {
	SelectMembershipByTeamID(ctx context.Context, userId uuid.UUID, teamId uuid.UUID) (*models.TeamMember, error)
	GetLastMembershipByUserID(ctx context.Context, userId uuid.UUID) (*models.TeamMember, error)
}

type TeamStore interface {
	CreateTeamFromUser(ctx context.Context, user *models.User) (*models.Team, error)
	FindTeamByID(ctx context.Context, teamId uuid.UUID) (*models.Team, error)
	CreateTeam(ctx context.Context, name string, stripeCustomerId *string) (*models.Team, error)
	UpdateTeam(ctx context.Context, teamId uuid.UUID, name string, stripeCustomerId *string) (*models.Team, error)
	DeleteTeam(ctx context.Context, teamId uuid.UUID) error
	FindTeamMembersByUserID(ctx context.Context, userId uuid.UUID) ([]*models.TeamMember, error)
	FindTeamMemberByUserAndTeamID(ctx context.Context, teamId uuid.UUID, userId uuid.UUID) (*models.TeamMember, error)
	FindLatestTeamMemberByUserID(ctx context.Context, userId uuid.UUID) (*models.TeamMember, error)
	CreateTeamMember(ctx context.Context, teamId, userId uuid.UUID, role models.TeamMemberRole) (*models.TeamMember, error)
	UpdateTeamMemberUpdatedAt(ctx context.Context, teamId, userId uuid.UUID) error
}

type teamService struct {
	teamStore TeamStore
}

// SelectMembershipByTeamID implements TeamService.
func (t *teamService) SelectMembershipByTeamID(ctx context.Context, teamId, userId uuid.UUID) (*models.TeamMember, error) {
	member, err := t.teamStore.FindTeamMemberByUserAndTeamID(ctx, teamId, userId)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, nil
	}
	err = t.teamStore.UpdateTeamMemberUpdatedAt(ctx, teamId, member.ID)
	if err != nil {
		return nil, err
	}
	return member, nil
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
