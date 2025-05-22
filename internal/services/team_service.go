package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
)

type TeamService interface {
	Store() TeamStore
	SetActiveTeam(ctx context.Context, userId uuid.UUID, teamId uuid.UUID) (*models.TeamMember, error)
	GetActiveTeam(ctx context.Context, userId uuid.UUID) (*models.TeamMember, error)
	FindTeamInfo(ctx context.Context, teamId, userId uuid.UUID) (*shared.TeamInfo, error)
	FindLatestTeamInfo(ctx context.Context, userId uuid.UUID) (*shared.TeamInfo, error)
	AddMember(ctx context.Context, teamId, userId uuid.UUID, role models.TeamMemberRole) (*models.TeamMember, error)
	RemoveMember(ctx context.Context, teamId, userId uuid.UUID) error
	CreateTeam(ctx context.Context, name string, slug string, userId uuid.UUID) (*shared.TeamInfo, error)
}

type TeamStore interface {
	CreateTeamWithOwnerMember(ctx context.Context, name string, slug string, userId uuid.UUID) (*shared.TeamInfo, error)
	CountTeamMembers(ctx context.Context, teamId uuid.UUID) (int64, error)
	FindTeamByStripeCustomerId(ctx context.Context, stripeCustomerId string) (*models.Team, error)
	FindTeamByID(ctx context.Context, teamId uuid.UUID) (*models.Team, error)
	CreateTeam(ctx context.Context, name string, slug string) (*models.Team, error)
	CheckTeamSlug(ctx context.Context, slug string) (bool, error)
	UpdateTeam(ctx context.Context, teamId uuid.UUID, name string) (*models.Team, error)
	DeleteTeam(ctx context.Context, teamId uuid.UUID) error
	FindTeamMembersByUserID(ctx context.Context, userId uuid.UUID) ([]*models.TeamMember, error)
	FindTeamMemberByTeamAndUserId(ctx context.Context, teamId uuid.UUID, userId uuid.UUID) (*models.TeamMember, error)
	FindLatestTeamMemberByUserID(ctx context.Context, userId uuid.UUID) (*models.TeamMember, error)
	CreateTeamMember(ctx context.Context, teamId, userId uuid.UUID, role models.TeamMemberRole, hasBillingAccess bool) (*models.TeamMember, error)
	DeleteTeamMember(ctx context.Context, teamId, userId uuid.UUID) error
	UpdateTeamMember(ctx context.Context, member *models.TeamMember) (*models.TeamMember, error)
	UpdateTeamMemberSelectedAt(ctx context.Context, teamId, userId uuid.UUID) error
}

type teamService struct {
	teamStore TeamStore
}

// CreateTeam implements TeamService.
func (t *teamService) CreateTeam(ctx context.Context, name string, slug string, userId uuid.UUID) (*shared.TeamInfo, error) {
	check, err := t.teamStore.CheckTeamSlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if !check {
		return nil, errors.New("team slug already exists")
	}
	teaminfo, err := t.teamStore.CreateTeamWithOwnerMember(ctx, name, slug, userId)
	if err != nil {
		return nil, err
	}
	if teaminfo == nil {
		return nil, errors.New("team not found")
	}
	return teaminfo, nil
}

// AddMember implements TeamService.
func (t *teamService) AddMember(ctx context.Context, teamId uuid.UUID, userId uuid.UUID, role models.TeamMemberRole) (*models.TeamMember, error) {
	member, err := t.teamStore.CreateTeamMember(ctx, teamId, userId, role, false)
	if err != nil {
		return nil, err
	}
	return member, nil
}

// RemoveMember implements TeamService.
func (t *teamService) RemoveMember(ctx context.Context, teamId uuid.UUID, userId uuid.UUID) error {
	err := t.teamStore.DeleteTeamMember(ctx, teamId, userId)
	if err != nil {
		return err
	}
	return nil
}

func (t *teamService) Store() TeamStore {
	return t.teamStore
}

// SetActiveTeam implements TeamService.
func (t *teamService) SetActiveTeam(ctx context.Context, teamId, userId uuid.UUID) (*models.TeamMember, error) {
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

func (t *teamService) GetActiveTeam(ctx context.Context, userId uuid.UUID) (*models.TeamMember, error) {
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
