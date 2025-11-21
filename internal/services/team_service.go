package services

import (
	"context"
	"errors"
	"log/slog"
	"regexp"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/mapper"
	"github.com/tkahng/playground/internal/tools/slug"
	"github.com/tkahng/playground/internal/tools/types"
)

var (
	IsAlphaNumericAndDash *regexp.Regexp = regexp.MustCompile("^[A-Za-z0-9-]+$")
)

type TeamService interface {
	ProcessSlug(ctx context.Context, teamSlug string, teamName string) (string, error)
	FindTeamMemberWithUserAndTeam(ctx context.Context, teamMemberID uuid.UUID) (*models.TeamMember, error)
	FindTeamInfo(ctx context.Context, teamId, userId uuid.UUID) (*models.TeamInfoModel, error)
	FindTeamInfoBySlug(ctx context.Context, slug string, userId uuid.UUID) (*models.TeamInfoModel, error)
	CreateTeamWithOwner(ctx context.Context, name string, slug string, userId uuid.UUID) (*models.TeamInfoModel, error)
	UpdateTeam(ctx context.Context, teamId uuid.UUID, name string) (*models.Team, error)
	DeleteTeam(ctx context.Context, teamId uuid.UUID, userId uuid.UUID) error
	FindTeamMembersByUserID(ctx context.Context, userId uuid.UUID, paginate *stores.TeamMemberListInput) ([]*models.TeamMember, error)
}

type TeamServiceImpl struct {
	adapter stores.StorageAdapterInterface
}

// ProcessSlug implements [TeamService].
func (t *TeamServiceImpl) ProcessSlug(ctx context.Context, teamSlug string, teamName string) (string, error) {
	// if teamSlug is not empty
	if teamSlug != "" {
		// and if teamSlug is valid
		if IsAlphaNumericAndDash.MatchString(teamSlug) {
			existingTeam, err := t.adapter.TeamGroup().FindTeamBySlug(ctx, teamSlug)
			if err != nil {
				return "", err
			}
			// and if teamSlug is not taken return teamSlug
			if existingTeam != nil {
				return teamSlug, nil
			}
		}
	}
	// if teamName is valid
	if IsAlphaNumericAndDash.MatchString(teamName) {
		existingTeam, err := t.adapter.TeamGroup().FindTeamBySlug(ctx, teamName)
		if err != nil {
			return "", err
		}
		// and if teamName is not taken return teamName
		if existingTeam == nil {
			return teamName, nil
		}
	}
	// create new valid slug from teamName
	newSlug := slug.NewSlug(teamName)
	// check if newSlug is taken
	existingTeam, err := t.adapter.TeamGroup().FindTeamBySlug(ctx, newSlug)
	if err != nil {
		return "", err
	}
	// if not taken return it
	if existingTeam == nil {
		return newSlug, nil
	}
	// could not process slug
	return "", errors.New("cannot process team slug")
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
	if teamInfo.Member.Role != models.TeamMemberRoleOwner {
		return errors.New("only owner can delete team")
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
func (t *TeamServiceImpl) CreateTeamWithOwner(ctx context.Context, name string, slug string, userId uuid.UUID) (*models.TeamInfoModel, error) {
	user, err := t.adapter.User().FindUserByID(ctx, userId)
	if err != nil {
		return nil, err
	}
	// check, err := t.teamStore.CheckTeamSlug(ctx, slug)
	check, err := t.adapter.TeamGroup().CheckTeamSlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if !check {
		return nil, errors.New("team slug already exists")
	}
	team, err := t.adapter.TeamGroup().CreateTeam(ctx, name, slug)
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
