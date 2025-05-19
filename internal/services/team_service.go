package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
)

type TeamService interface {
	Store() TeamStore
	SetLastSelectedTeamMember(ctx context.Context, userId uuid.UUID, teamId uuid.UUID) (*models.TeamMember, error)
	GetLastSelectedTeamMember(ctx context.Context, userId uuid.UUID) (*models.TeamMember, error)
	FindTeamInfo(ctx context.Context, teamId, userId uuid.UUID) (*shared.TeamInfo, error)
	FindLatestTeamInfo(ctx context.Context, userId uuid.UUID) (*shared.TeamInfo, error)
}

type TeamStore interface {
	CountTeamMembers(ctx context.Context, teamId uuid.UUID) (int64, error)

	FindTeamByStripeCustomerId(ctx context.Context, stripeCustomerId string) (*models.Team, error)
	// CreateTeamFromUser(ctx context.Context, user *models.User) (*models.Team, error)
	FindTeamByID(ctx context.Context, teamId uuid.UUID) (*models.Team, error)
	CreateTeam(ctx context.Context, name string, stripeCustomerId *string) (*models.Team, error)
	UpdateTeam(ctx context.Context, teamId uuid.UUID, name string, stripeCustomerId *string) (*models.Team, error)
	DeleteTeam(ctx context.Context, teamId uuid.UUID) error
	FindTeamMembersByUserID(ctx context.Context, userId uuid.UUID) ([]*models.TeamMember, error)
	FindTeamMemberByTeamAndUserId(ctx context.Context, teamId uuid.UUID, userId uuid.UUID) (*models.TeamMember, error)
	FindLatestTeamMemberByUserID(ctx context.Context, userId uuid.UUID) (*models.TeamMember, error)
	CreateTeamMember(ctx context.Context, teamId, userId uuid.UUID, role models.TeamMemberRole) (*models.TeamMember, error)
	UpdateTeamMember(ctx context.Context, member *models.TeamMember) (*models.TeamMember, error)
	UpdateTeamMemberSelectedAt(ctx context.Context, teamId, userId uuid.UUID) error
}

type teamService struct {
	teamStore TeamStore
}

func (t *teamService) Store() TeamStore {
	return t.teamStore
}

// SetLastSelectedTeamMember implements TeamService.
func (t *teamService) SetLastSelectedTeamMember(ctx context.Context, teamId, userId uuid.UUID) (*models.TeamMember, error) {
	member, err := t.teamStore.FindTeamMemberByTeamAndUserId(ctx, teamId, userId)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, nil
	}
	err = t.teamStore.UpdateTeamMemberSelectedAt(ctx, teamId, member.ID)
	if err != nil {
		return nil, err
	}
	return member, nil
}

func (t *teamService) GetLastSelectedTeamMember(ctx context.Context, userId uuid.UUID) (*models.TeamMember, error) {
	team, err := t.teamStore.FindLatestTeamMemberByUserID(ctx, userId)
	if err != nil {
		return nil, err
	}
	return team, nil
}
func (t *teamService) FindTeamInfo(ctx context.Context, teamId, userId uuid.UUID) (*shared.TeamInfo, error) {
	team, err := t.teamStore.FindTeamByID(ctx, teamId)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, nil
	}
	member, err := t.teamStore.FindTeamMemberByTeamAndUserId(ctx, teamId, userId)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, nil
	}
	return &shared.TeamInfo{
		Team:   *team,
		Member: *member,
	}, nil
}

func (t *teamService) FindLatestTeamInfo(ctx context.Context, userId uuid.UUID) (*shared.TeamInfo, error) {
	member, err := t.teamStore.FindLatestTeamMemberByUserID(ctx, userId)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, nil
	}
	team, err := t.teamStore.FindTeamByID(ctx, member.TeamID)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, nil
	}
	return &shared.TeamInfo{
		Team:   *team,
		Member: *member,
	}, nil
}

func NewTeamService(store TeamStore) TeamService {
	return &teamService{
		teamStore: store,
	}
}
