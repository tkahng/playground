package stores

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/crudrepo"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/types"
)

type DbTeamMemberStore struct {
	db database.Dbx
}

func NewDbTeamMemberStore(db database.Dbx) *DbTeamMemberStore {
	return &DbTeamMemberStore{
		db: db,
	}
}

// WithTx returns a new DbTeamMemberStore with the given transaction.
func (s *DbTeamMemberStore) WithTx(tx database.Dbx) *DbTeamMemberStore {
	return &DbTeamMemberStore{
		db: tx,
	}
}

// FindUserByID implements services.TeamInvitationStore.
//
//	func (s *DbTeamStore) FindUserByID(ctx context.Context, userId uuid.UUID) (*models.User, error) {
//		user, err := crudrepo.User.GetOne(
//			ctx,
//			s.db,
//			&map[string]any{
//				models.UserTable.ID: map[string]any{
//					"_eq": userId,
//				},
//			},
//		)
//		if err != nil {
//			return nil, err
//		}
//		return user, nil
//	}
//
// FindTeamMember implements services.TeamStore.
func (s *DbTeamMemberStore) FindTeamMember(ctx context.Context, member *models.TeamMember) (*models.TeamMember, error) {
	if member == nil {
		return nil, nil
	}
	where := map[string]any{}
	if member.ID != uuid.Nil {
		where[models.TeamMemberTable.ID] = map[string]any{
			"_eq": member.ID,
		}
	}
	if member.TeamID != uuid.Nil {
		where[models.TeamMemberTable.TeamID] = map[string]any{
			"_eq": member.TeamID,
		}
	}
	if member.UserID != nil {
		where[models.TeamMemberTable.UserID] = map[string]any{
			"_eq": member.UserID,
		}
	}
	if member.Role != "" {
		where[models.TeamMemberTable.Role] = map[string]any{
			"_eq": member.Role,
		}
	}
	member, err := crudrepo.TeamMember.GetOne(
		ctx,
		s.db,
		&where,
	)
	if err != nil {
		return nil, err
	}
	return member, nil
}

// CreateTeamWithOwnerMember implements services.TeamStore.
func (s *DbTeamStore) CreateTeamWithOwnerMember(ctx context.Context, name string, slug string, userId uuid.UUID) (*shared.TeamInfoModel, error) {
	var teamInfo *shared.TeamInfoModel
	err := s.Transact(
		ctx,
		func(store *DbTeamStore) error {
			team, err := store.CreateTeam(ctx, name, slug)
			if err != nil {
				return err
			}
			if team == nil {
				return fmt.Errorf("team not found")
			}
			teamMember, err := store.CreateTeamMember(ctx, team.ID, userId, models.TeamMemberRoleOwner, true)
			if err != nil {
				return err
			}
			if teamMember == nil {
				return fmt.Errorf("team member not found")
			}
			teamInfo = &shared.TeamInfoModel{
				Team:   *team,
				Member: *teamMember,
			}
			return nil
		},
	)
	if err != nil {
		return nil, err
	}
	return teamInfo, nil
}
func (s *DbTeamStore) CreateTeamWithOwnerMember2(ctx context.Context, name string, slug string, userId uuid.UUID) (*shared.TeamInfoModel, error) {
	var teamInfo *shared.TeamInfoModel
	err := s.Transact(ctx, func(store *DbTeamStore) error {
		user, err := store.FindUserByID(ctx, userId)
		if err != nil {
			return err
		}
		if user == nil {
			return fmt.Errorf("user not found")
		}
		team, err := store.CreateTeam(ctx, name, slug)
		if err != nil {
			return err
		}
		if team == nil {
			return fmt.Errorf("team not found")
		}
		teamMember, err := store.CreateTeamMember(ctx, team.ID, userId, models.TeamMemberRoleOwner, true)
		if err != nil {
			return err
		}
		if teamMember == nil {
			return fmt.Errorf("team member not found")
		}
		teamMember.Team = team
		teamMember.User = user
		teamInfo = &shared.TeamInfoModel{
			Team:   *team,
			Member: *teamMember,
			User:   *user,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return teamInfo, nil
}

func (s *DbTeamMemberStore) CreateTeamFromUser(ctx context.Context, user *models.User) (*models.TeamMember, error) {
	team, err := crudrepo.Team.PostOne(
		ctx,
		s.db,
		&models.Team{
			Name: user.Email,
			Slug: user.Email,
		},
	)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, errors.New("team not found")
	}
	teamMember, err := crudrepo.TeamMember.PostOne(
		ctx,
		s.db,
		&models.TeamMember{
			TeamID:           team.ID,
			UserID:           types.Pointer(user.ID),
			Role:             models.TeamMemberRoleOwner,
			HasBillingAccess: true,
		},
	)
	if err != nil {
		return nil, err
	}
	if teamMember == nil {
		return nil, errors.New("team member not found")
	}
	teamMember.Team = team
	teamMember.User = user
	return teamMember, nil
}

// DeleteTeamMember implements services.TeamStore.
func (s *DbTeamMemberStore) DeleteTeamMember(ctx context.Context, teamId uuid.UUID, userId uuid.UUID) error {
	_, err := crudrepo.TeamMember.Delete(
		ctx,
		s.db,
		&map[string]any{
			models.TeamMemberTable.TeamID: map[string]any{
				"_eq": teamId,
			},
			models.TeamMemberTable.UserID: map[string]any{
				"_eq": userId,
			},
		},
	)
	if err != nil {
		return err
	}
	return nil
}

// UpdateTeamMember implements services.TeamStore.
func (s *DbTeamMemberStore) UpdateTeamMember(ctx context.Context, member *models.TeamMember) (*models.TeamMember, error) {
	newMember, err := crudrepo.TeamMember.PutOne(
		ctx,
		s.db,
		member,
	)
	if err != nil {
		return nil, err
	}
	return newMember, nil
}

// CountOwnerTeamMembers implements services.TeamStore.
func (s *DbTeamMemberStore) CountOwnerTeamMembers(ctx context.Context, teamId uuid.UUID) (int64, error) {
	c, err := crudrepo.TeamMember.Count(
		ctx,
		s.db,
		&map[string]any{
			models.TeamMemberTable.TeamID: map[string]any{
				"_eq": teamId,
			},
			models.TeamMemberTable.Role: map[string]any{
				"_eq": string(models.TeamMemberRoleOwner),
			},
		},
	)
	if err != nil {
		return 0, err
	}
	return c, nil
}

// CountTeamMembers implements services.TeamStore.
func (s *DbTeamMemberStore) CountTeamMembers(ctx context.Context, teamId uuid.UUID) (int64, error) {
	c, err := crudrepo.TeamMember.Count(
		ctx,
		s.db,
		&map[string]any{
			models.TeamMemberTable.TeamID: map[string]any{
				"_eq": teamId,
			},
		},
	)
	if err != nil {
		return 0, err
	}
	return c, nil
}
func (s *DbTeamMemberStore) FindTeamMemberByTeamAndUserId(ctx context.Context, teamId, userId uuid.UUID) (*models.TeamMember, error) {
	teamMember, err := crudrepo.TeamMember.GetOne(
		ctx,
		s.db,
		&map[string]any{
			models.TeamMemberTable.UserID: map[string]any{
				"_eq": userId,
			},
			models.TeamMemberTable.TeamID: map[string]any{
				"_eq": teamId,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	return teamMember, nil
}

// UpdateTeamMemberSelectedAt implements TeamQueryer.
func (s *DbTeamMemberStore) UpdateTeamMemberSelectedAt(ctx context.Context, teamId, userId uuid.UUID) error {
	qquery := squirrel.Update("team_members").
		Where(squirrel.Eq{models.TeamMemberTable.TeamID: teamId}).
		Where(squirrel.Eq{models.TeamMemberTable.UserID: userId}).
		Set(models.TeamMemberTable.LastSelectedAt, time.Now())

	err := database.ExecWithBuilder(ctx, s.db, qquery.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return err
	}
	return nil
}

// FindLatestTeamMemberByUserID implements TeamQueryer.
func (s *DbTeamMemberStore) FindLatestTeamMemberByUserID(ctx context.Context, userId uuid.UUID) (*models.TeamMember, error) {
	teamMember, err := crudrepo.TeamMember.Get(
		ctx,
		s.db,
		&map[string]any{
			models.TeamMemberTable.UserID: map[string]any{
				"_eq": userId,
			},
		},
		&map[string]string{
			models.TeamMemberTable.LastSelectedAt: "DESC",
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

// FindTeamMembersByUserID implements TeamQueryer.
func (s *DbTeamMemberStore) FindTeamMembersByUserID(ctx context.Context, userId uuid.UUID, paginate *shared.TeamMemberListInput) ([]*models.TeamMember, error) {
	limit, offset := database.PaginateRepo(&paginate.PaginatedInput)
	orderby := make(map[string]string)
	if paginate.SortBy != "" && paginate.SortOrder != "" && slices.Contains(crudrepo.TeamMemberBuilder.ColumnNames(), paginate.SortBy) {
		orderby[paginate.SortBy] = paginate.SortOrder
	} else {
		orderby["last_selected_at"] = "DESC"
	}
	qs := squirrel.Select("team_members.*").From("team_members")
	qs = qs.Where(squirrel.Eq{"user_id": userId})
	qs = qs.Where(squirrel.Eq{"active": true})
	if paginate.SortBy == "team.name" {
		qs = qs.Join("teams on team_members.team_id = teams.id").OrderBy("teams.name " + strings.ToUpper(paginate.SortOrder))
	} else if slices.Contains(crudrepo.TeamMemberBuilder.ColumnNames(), paginate.SortBy) {
		qs = qs.OrderBy(paginate.SortBy + " " + strings.ToUpper(paginate.SortOrder))
	} else {
		qs = qs.OrderBy("last_selected_at DESC")
	}
	qs = qs.Limit(uint64(*limit)).Offset(uint64(*offset))
	teamMembers, err := database.QueryWithBuilder[*models.TeamMember](ctx, s.db, qs.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return nil, err
	}

	return teamMembers, nil
}

func (s *DbTeamMemberStore) CountTeamMembersByUserID(ctx context.Context, userId uuid.UUID) (int64, error) {
	c, err := crudrepo.TeamMember.Count(
		ctx,
		s.db,
		&map[string]any{
			models.TeamMemberTable.UserID: map[string]any{
				"_eq": userId,
			},
			models.TeamMemberTable.Active: map[string]any{
				"_eq": true,
			},
		},
	)
	if err != nil {
		return 0, err
	}
	return c, nil
}
func (s *DbTeamMemberStore) CreateTeamMember(ctx context.Context, teamId, userId uuid.UUID, role models.TeamMemberRole, hasBillingAccess bool) (*models.TeamMember, error) {
	teamMember := &models.TeamMember{
		TeamID:           teamId,
		UserID:           &userId,
		Role:             role,
		Active:           true,
		HasBillingAccess: hasBillingAccess,
	}
	return crudrepo.TeamMember.PostOne(
		ctx,
		s.db,
		teamMember,
	)
}
