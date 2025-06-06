package services

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
)

type TeamService interface {
	Store() TeamStore
	SetActiveTeamMember(ctx context.Context, userId uuid.UUID, teamId uuid.UUID) (*models.TeamMember, error)
	GetActiveTeamMember(ctx context.Context, userId uuid.UUID) (*models.TeamMember, error)
	FindTeamInfo(ctx context.Context, teamId, userId uuid.UUID) (*shared.TeamInfoModel, error)
	FindTeamInfoBySlug(ctx context.Context, slug string, userId uuid.UUID) (*shared.TeamInfoModel, error)
	FindLatestTeamInfo(ctx context.Context, userId uuid.UUID) (*shared.TeamInfoModel, error)
	AddMember(ctx context.Context, teamId, userId uuid.UUID, role models.TeamMemberRole, hasBillingAccess bool) (*models.TeamMember, error)
	RemoveMember(ctx context.Context, teamId, userId uuid.UUID) error
	LeaveTeam(ctx context.Context, teamId, userId uuid.UUID) error
	CreateTeam(ctx context.Context, name string, slug string, userId uuid.UUID) (*shared.TeamInfoModel, error)
	UpdateTeam(ctx context.Context, teamId uuid.UUID, name string) (*models.Team, error)
	DeleteTeam(ctx context.Context, teamId uuid.UUID, userId uuid.UUID) error
	FindTeamMembersByUserID(ctx context.Context, userId uuid.UUID, paginate *shared.TeamMemberListInput) ([]*models.TeamMember, error)
}

//	type TeamStripeStore interface {
//		FindActiveSubscriptionByCustomerId(ctx context.Context, customerId string) (*models.SubscriptionWithPrice, error)
//		// FindActiveSubscriptionWithPriceProductByCustomerId(ctx context.Context, customerId string) (*models.StripeSubscription, error)
//	}
type TeamStore interface {
	// find team
	ListTeams(ctx context.Context, params *shared.ListTeamsParams) ([]*models.Team, error)
	CountTeams(ctx context.Context, params *shared.ListTeamsParams) (int64, error)
	CreateTeamWithOwnerMember(ctx context.Context, name string, slug string, userId uuid.UUID) (*shared.TeamInfoModel, error)
	CountOwnerTeamMembers(ctx context.Context, teamId uuid.UUID) (int64, error)
	CountTeamMembers(ctx context.Context, teamId uuid.UUID) (int64, error)
	FindTeamByStripeCustomerId(ctx context.Context, stripeCustomerId string) (*models.Team, error)
	FindTeamByID(ctx context.Context, teamId uuid.UUID) (*models.Team, error)
	FindTeamBySlug(ctx context.Context, slug string) (*models.Team, error)

	LoadTeamsByIds(ctx context.Context, teamIds ...uuid.UUID) ([]*models.Team, error)
	CreateTeam(ctx context.Context, name string, slug string) (*models.Team, error)
	CheckTeamSlug(ctx context.Context, slug string) (bool, error)
	UpdateTeam(ctx context.Context, teamId uuid.UUID, name string) (*models.Team, error)
	DeleteTeam(ctx context.Context, teamId uuid.UUID) error
	CountTeamMembersByUserID(ctx context.Context, userId uuid.UUID) (int64, error)
	// find team members
	FindTeamMember(ctx context.Context, member *models.TeamMember) (*models.TeamMember, error)
	FindTeamMembersByUserID(ctx context.Context, userId uuid.UUID, paginate *shared.TeamMemberListInput) ([]*models.TeamMember, error)
	FindTeamMemberByTeamAndUserId(ctx context.Context, teamId uuid.UUID, userId uuid.UUID) (*models.TeamMember, error)
	FindLatestTeamMemberByUserID(ctx context.Context, userId uuid.UUID) (*models.TeamMember, error)

	CreateTeamMember(ctx context.Context, teamId, userId uuid.UUID, role models.TeamMemberRole, hasBillingAccess bool) (*models.TeamMember, error)
	DeleteTeamMember(ctx context.Context, teamId, userId uuid.UUID) error
	UpdateTeamMember(ctx context.Context, member *models.TeamMember) (*models.TeamMember, error)
	UpdateTeamMemberSelectedAt(ctx context.Context, teamId, userId uuid.UUID) error

	// misc methods
	FindUserByID(ctx context.Context, userId uuid.UUID) (*models.User, error)
}

type TeamServiceStore interface {
	TeamStore
	// TeamStripeStore
}

type teamService struct {
	teamStore TeamServiceStore
}

// FindTeamMembersByUserID implements TeamService.
func (t *teamService) FindTeamMembersByUserID(ctx context.Context, userId uuid.UUID, paginate *shared.TeamMemberListInput) ([]*models.TeamMember, error) {
	members, err := t.teamStore.FindTeamMembersByUserID(ctx, userId, paginate)
	if err != nil {
		return nil, err
	}
	if members == nil {
		return nil, nil
	}
	teamIds := mapper.Map(members, func(member *models.TeamMember) uuid.UUID {
		return member.TeamID
	})
	teams, err := t.teamStore.LoadTeamsByIds(ctx, teamIds...)
	if err != nil {
		return nil, err
	}
	for idx, member := range members {
		team := teams[idx]
		if team != nil {
			member.Team = team
		}
	}
	return members, nil
}

func NewTeamService(store TeamServiceStore) TeamService {
	return &teamService{
		teamStore: store,
	}
}

// LeaveTeam implements TeamService.
func (t *teamService) LeaveTeam(ctx context.Context, teamId uuid.UUID, userId uuid.UUID) error {
	teamInfo, err := t.FindTeamInfo(ctx, teamId, userId)
	if err != nil {
		return err
	}
	if teamInfo == nil {
		return errors.New("team member not found")
	}
	if teamInfo.Member.Role == models.TeamMemberRoleOwner {
		count, err := t.teamStore.CountOwnerTeamMembers(ctx, teamId)
		if err != nil {
			return err
		}
		if count == 1 {
			return errors.New("owner cannot leave team")
		}
	}
	err = t.teamStore.DeleteTeamMember(ctx, teamId, userId)
	if err != nil {
		return err
	}
	return nil
}

// DeleteTeam implements TeamService.
func (t *teamService) DeleteTeam(ctx context.Context, teamId uuid.UUID, userId uuid.UUID) error {
	teamInfo, err := t.FindTeamInfo(ctx, teamId, userId)
	if err != nil {
		return err
	}
	if teamInfo == nil {
		slog.ErrorContext(ctx, "team member not found")
		return errors.New("team member not found")
	}
	if teamInfo.Member.Role != models.TeamMemberRoleOwner {
		return errors.New("only owner can delete team")
	}

	return nil
}

// UpdateTeam implements TeamService.
func (t *teamService) UpdateTeam(ctx context.Context, teamId uuid.UUID, name string) (*models.Team, error) {
	team, err := t.teamStore.UpdateTeam(ctx, teamId, name)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, errors.New("team not found")
	}
	return team, nil
}

// CreateTeam implements TeamService.
func (t *teamService) CreateTeam(ctx context.Context, name string, slug string, userId uuid.UUID) (*shared.TeamInfoModel, error) {
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
func (t *teamService) AddMember(ctx context.Context, teamId uuid.UUID, userId uuid.UUID, role models.TeamMemberRole, hasBillingAccess bool) (*models.TeamMember, error) {
	member, err := t.teamStore.CreateTeamMember(ctx, teamId, userId, role, hasBillingAccess)
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

// SetActiveTeamMember impleements TeamService.
func (t *teamService) SetActiveTeamMember(ctx context.Context, teamId, userId uuid.UUID) (*models.TeamMember, error) {
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

func (t *teamService) GetActiveTeamMember(ctx context.Context, userId uuid.UUID) (*models.TeamMember, error) {
	team, err := t.teamStore.FindLatestTeamMemberByUserID(ctx, userId)
	if err != nil {
		return nil, err
	}
	return team, nil
}
func (t *teamService) FindTeamInfo(ctx context.Context, teamId, userId uuid.UUID) (*shared.TeamInfoModel, error) {
	slog.InfoContext(ctx, "FindTeamInfo", "teamId", teamId, "userId", userId)
	user, err := t.teamStore.FindUserByID(ctx, userId)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
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

	return &shared.TeamInfoModel{
		Team:   *team,
		Member: *member,
		User:   *user,
	}, nil
}

func (t *teamService) FindTeamInfoBySlug(ctx context.Context, slug string, userId uuid.UUID) (*shared.TeamInfoModel, error) {
	slog.InfoContext(ctx, "FindTeamInfoBySlug", "slug", slug, "userId", userId)
	user, err := t.teamStore.FindUserByID(ctx, userId)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	team, err := t.teamStore.FindTeamBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, nil
	}
	member, err := t.teamStore.FindTeamMemberByTeamAndUserId(ctx, team.ID, userId)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, nil
	}
	member.Team = team // Ensure the team is set in the member
	member.User = user // Ensure the user is set in the member
	return &shared.TeamInfoModel{
		Team:   *team,
		Member: *member,
		User:   *user,
	}, nil
}

func (t *teamService) FindLatestTeamInfo(ctx context.Context, userId uuid.UUID) (*shared.TeamInfoModel, error) {
	user, err := t.teamStore.FindUserByID(ctx, userId)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
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
	return &shared.TeamInfoModel{
		Team:   *team,
		Member: *member,
		User:   *user,
	}, nil
}
