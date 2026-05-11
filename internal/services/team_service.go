package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/mapper"
	"github.com/tkahng/playground/internal/tools/slug"
	"github.com/tkahng/playground/internal/tools/types"
)

type TeamService interface {
	ProcessSlug(ctx context.Context, teamName string) (string, error)
	FindTeamMemberWithUserAndTeam(ctx context.Context, teamMemberID uuid.UUID) (*models.TeamMember, error)
	FindTeamInfo(ctx context.Context, teamId, userId uuid.UUID) (*models.TeamInfoModel, error)
	FindTeamInfoBySlug(ctx context.Context, slug string, userId uuid.UUID) (*models.TeamInfoModel, error)
	CreateTeamWithOwner(ctx context.Context, name string, userId uuid.UUID) (*models.TeamInfoModel, error)
	UpdateTeam(ctx context.Context, teamId uuid.UUID, name string) (*models.Team, error)
	DeleteTeam(ctx context.Context, teamId uuid.UUID, userId uuid.UUID) error
	FindTeamMembersByUserID(ctx context.Context, userId uuid.UUID, paginate *stores.TeamMemberListInput) ([]*models.TeamMember, error)
}

type TeamServiceImpl struct {
	adapter stores.StorageAdapterInterface
}

// ProcessSlug implements [TeamService].
// Generates a URL-safe slug from teamName. On conflict, appends a numeric
// suffix (e.g. "my-team-1", "my-team-2") following industry convention.
func (t *TeamServiceImpl) ProcessSlug(ctx context.Context, teamName string) (string, error) {
	base := slug.NewSlug(teamName)
	if base == "" {
		base = uuid.NewString()
	}

	existing, err := t.adapter.TeamGroup().FindTeamBySlug(ctx, base)
	if err != nil {
		return "", err
	}
	if existing == nil {
		return base, nil
	}

	for i := 1; i <= 99; i++ {
		candidate := fmt.Sprintf("%s-%d", base, i)
		existing, err := t.adapter.TeamGroup().FindTeamBySlug(ctx, candidate)
		if err != nil {
			return "", err
		}
		if existing == nil {
			return candidate, nil
		}
	}

	return "", errors.New("cannot generate unique team slug")
}

// FindTeamMemberWithUserAndTeam implements TeamService.
func (t *TeamServiceImpl) FindTeamMemberWithUserAndTeam(ctx context.Context, teamMemberID uuid.UUID) (*models.TeamMember, error) {
	member, err := t.adapter.TeamMember().FindTeamMember(ctx, &stores.TeamMemberFilter{
		Ids: []uuid.UUID{teamMemberID},
	})
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, nil
	}
	if member.UserID != nil {
		user, err := t.adapter.User().FindUserByID(ctx, *member.UserID)
		if err != nil {
			return nil, err
		}
		if user != nil {
			member.User = user
		}
	}
	team, err := t.adapter.TeamGroup().FindTeamByID(ctx, member.TeamID)
	if err != nil {
		return nil, err
	}
	if team != nil {
		member.Team = team
	}
	return member, nil
}

// FindTeamMembersByUserID implements TeamService.
func (t *TeamServiceImpl) FindTeamMembersByUserID(ctx context.Context, userId uuid.UUID, paginate *stores.TeamMemberListInput) ([]*models.TeamMember, error) {
	members, err := t.adapter.TeamMember().FindTeamMembersByUserID(
		ctx,
		userId,
		paginate,
	)
	if err != nil {
		return nil, err
	}
	if members == nil {
		return nil, nil
	}
	teamIds := mapper.Map(members, func(member *models.TeamMember) uuid.UUID {
		return member.TeamID
	})
	teams, err := t.adapter.TeamGroup().LoadTeamsByIds(ctx, teamIds...)
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

func NewTeamService(adapter stores.StorageAdapterInterface) TeamService {
	return &TeamServiceImpl{
		adapter: adapter,
	}
}

// DeleteTeam implements TeamService.
func (t *TeamServiceImpl) DeleteTeam(ctx context.Context, teamId uuid.UUID, userId uuid.UUID) error {
	teamInfo, err := t.FindTeamInfo(ctx, teamId, userId)
	if err != nil {
		return err
	}
	if teamInfo == nil {
		slog.ErrorContext(ctx, "team member not found")
		return errors.New("team member not found")
	}
	err = t.adapter.TeamGroup().DeleteTeam(ctx, teamId)
	// err = t.teamStore.DeleteTeam(ctx, teamId)
	if err != nil {
		slog.ErrorContext(ctx, "error deleting team", "teamId", teamId, "error", err)
		return err
	}
	return nil
}

// UpdateTeam implements TeamService.
func (t *TeamServiceImpl) UpdateTeam(ctx context.Context, teamId uuid.UUID, name string) (*models.Team, error) {
	// team, err := t.teamStore.UpdateTeam(ctx, teamId, name)
	team, err := t.adapter.TeamGroup().UpdateTeam(ctx, teamId, name)

	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, errors.New("team not found")
	}
	return team, nil
}

// CreateTeamWithOwner implements TeamService.
func (t *TeamServiceImpl) CreateTeamWithOwner(ctx context.Context, name string, userId uuid.UUID) (*models.TeamInfoModel, error) {
	user, err := t.adapter.User().FindUserByID(ctx, userId)
	if err != nil {
		return nil, err
	}
	newSlug, err := t.ProcessSlug(ctx, name)
	if err != nil {
		return nil, err
	}
	if newSlug == "" {
		return nil, errors.New("error processing team slug.")
	}
	team, err := t.adapter.TeamGroup().CreateTeam(ctx, name, newSlug)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, errors.New("team not found")
	}
	teamMember, err := t.adapter.TeamMember().CreateTeamMember(ctx, &models.TeamMember{
		TeamID:           team.ID,
		UserID:           types.Pointer(userId),
		Role:             models.TeamMemberRoleOwner,
		HasBillingAccess: true,
		Active:           true,
	})
	if err != nil {
		return nil, err
	}
	if teamMember == nil {
		return nil, errors.New("team member not found")
	}
	teamMember.User = user
	teamInfo := &models.TeamInfoModel{
		Team:   *team,
		Member: *teamMember,
		User:   *user,
	}
	return teamInfo, nil
}

// SetActiveTeamMember impleements TeamService.
func (t *TeamServiceImpl) SetActiveTeamMember(ctx context.Context, teamId, userId uuid.UUID) (*models.TeamMember, error) {
	// member, err := t.teamStore.FindTeamMemberByTeamAndUserId(ctx, teamId, userId)
	member, err := t.adapter.TeamMember().FindTeamMember(ctx, &stores.TeamMemberFilter{
		TeamIds: []uuid.UUID{teamId},
		UserIds: []uuid.UUID{userId},
		Active: types.OptionalParam[bool]{
			Value: true, IsSet: true,
		},
	})
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, errors.New("team member not found")
	}
	err = t.adapter.TeamMember().UpdateTeamMemberSelectedAt(ctx, teamId, userId)
	// err = t.teamStore.UpdateTeamMemberSelectedAt(ctx, teamId, member.ID)
	if err != nil {
		return nil, err
	}
	return member, nil
}

func (t *TeamServiceImpl) GetActiveTeamMember(ctx context.Context, userId uuid.UUID) (*models.TeamMember, error) {
	// team, err := t.teamStore.FindLatestTeamMemberByUserID(ctx, userId)
	team, err := t.adapter.TeamMember().FindLatestTeamMemberByUserID(ctx, userId)
	if err != nil {
		return nil, err
	}
	return team, nil
}
func (t *TeamServiceImpl) FindTeamInfo(ctx context.Context, teamId, userId uuid.UUID) (*models.TeamInfoModel, error) {
	user, err := t.adapter.User().FindUserByID(ctx, userId)
	// user, err := t.teamStore.FindUserByID(ctx, userId)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	team, err := t.adapter.TeamGroup().FindTeamByID(ctx, teamId)
	// team, err := t.teamStore.FindTeamByID(ctx, teamId)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, nil
	}
	member, err := t.adapter.TeamMember().FindTeamMember(ctx,
		&stores.TeamMemberFilter{
			TeamIds: []uuid.UUID{teamId},
			UserIds: []uuid.UUID{userId},
			Active: types.OptionalParam[bool]{
				Value: true, IsSet: true,
			},
		})
	// member, err := t.teamStore.FindTeamMemberByTeamAndUserId(ctx, teamId, userId)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, nil
	}
	member.User = user
	return &models.TeamInfoModel{
		Team:   *team,
		Member: *member,
		User:   *user,
	}, nil
}

func (t *TeamServiceImpl) FindTeamInfoBySlug(ctx context.Context, slug string, userId uuid.UUID) (*models.TeamInfoModel, error) {
	user, err := t.adapter.User().FindUserByID(ctx, userId)
	// user, err := t.teamStore.FindUserByID(ctx, userId)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	team, err := t.adapter.TeamGroup().FindTeamBySlug(ctx, slug)
	// team, err := t.teamStore.FindTeamBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, nil
	}
	member, err := t.adapter.TeamMember().FindTeamMember(ctx, &stores.TeamMemberFilter{
		TeamIds: []uuid.UUID{team.ID},
		UserIds: []uuid.UUID{userId},
		Active: types.OptionalParam[bool]{
			Value: true, IsSet: true,
		},
	})
	// member, err := t.teamStore.FindTeamMemberByTeamAndUserId(ctx, team.ID, userId)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, nil
	}
	member.Team = team // Ensure the team is set in the member
	member.User = user // Ensure the user is set in the member
	return &models.TeamInfoModel{
		Team:   *team,
		Member: *member,
		User:   *user,
	}, nil
}

func (t *TeamServiceImpl) FindLatestTeamInfo(ctx context.Context, userId uuid.UUID) (*models.TeamInfoModel, error) {
	// user, err := t.teamStore.FindUserByID(ctx, userId)
	user, err := t.adapter.User().FindUserByID(ctx, userId)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// member, err := t.teamStore.FindLatestTeamMemberByUserID(ctx, userId)
	member, err := t.adapter.TeamMember().FindLatestTeamMemberByUserID(ctx, userId)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, nil
	}
	// team, err := t.teamStore.FindTeamByID(ctx, member.TeamID)
	team, err := t.adapter.TeamGroup().FindTeamByID(ctx, member.TeamID)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, nil
	}
	return &models.TeamInfoModel{
		Team:   *team,
		Member: *member,
		User:   *user,
	}, nil
}
