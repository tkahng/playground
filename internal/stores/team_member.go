package stores

import (
	"context"
	"errors"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/repository"
	"github.com/tkahng/playground/internal/tools/mapper"
	"github.com/tkahng/playground/internal/tools/types"
	"github.com/tkahng/playground/internal/tools/utils"
)

type TeamMemberListInput struct {
	PaginatedInput
	SortParams
}

type TeamMemberFilter struct {
	PaginatedInput
	SortParams
	Q       string                    `query:"q"`
	Ids     []uuid.UUID               `query:"ids"`
	Roles   []models.TeamMemberRole   `query:"roles"`
	UserIds []uuid.UUID               `query:"user_ids"`
	TeamIds []uuid.UUID               `query:"team_ids"`
	Active  types.OptionalParam[bool] `query:"active"`
}

type DbTeamMemberStoreInterface interface {
	LoadTeamMembersByUserAndTeamIds(ctx context.Context, userId uuid.UUID, teamIds ...uuid.UUID) ([]*models.TeamMember, error)
	LoadTeamMembersByIds(ctx context.Context, teamMemberIds ...uuid.UUID) ([]*models.TeamMember, error)
	FindTeamMembers(ctx context.Context, filter *TeamMemberFilter) ([]*models.TeamMember, error)
	CountTeamMembers(ctx context.Context, filter *TeamMemberFilter) (int64, error)
	CreateTeamFromUser(ctx context.Context, user *models.User) (*models.TeamMember, error)
	CreateTeamMember(ctx context.Context, teamId uuid.UUID, userId uuid.UUID, role models.TeamMemberRole, hasBillingAccess bool) (*models.TeamMember, error)
	DeleteTeamMember(ctx context.Context, teamId uuid.UUID, userId uuid.UUID) error
	FindLatestTeamMemberByUserID(ctx context.Context, userId uuid.UUID) (*models.TeamMember, error)
	FindTeamMember(ctx context.Context, member *TeamMemberFilter) (*models.TeamMember, error)
	FindTeamMembersByUserID(ctx context.Context, userId uuid.UUID, paginate *TeamMemberListInput) ([]*models.TeamMember, error)
	UpdateTeamMember(ctx context.Context, member *models.TeamMember) (*models.TeamMember, error)
	UpdateTeamMemberSelectedAt(ctx context.Context, teamId uuid.UUID, userId uuid.UUID) error
	CreateTeamMemberFromUserAndSlug(ctx context.Context, user *models.User, slug string, role models.TeamMemberRole) (*models.TeamMember, error)
	// VerifyAndUpdateTeamSubscriptionQuantity(ctx context.Context, adapter StorageAdapterInterface, teamId uuid.UUID) (int64, error)
}

type DbTeamMemberStore struct {
	db database.Dbx
}

// LoadTeamMembersByIds implements DbTeamMemberStoreInterface.
func (s *DbTeamMemberStore) LoadTeamMembersByIds(ctx context.Context, teamMemberIds ...uuid.UUID) ([]*models.TeamMember, error) {
	members, err := repository.TeamMember.Get(
		ctx,
		s.db,
		&map[string]any{
			models.TeamMemberTable.ID: map[string]any{
				"_in": teamMemberIds,
			},
		},
		nil,
		nil,
		nil,
	)

	if err != nil {
		return nil, err
	}
	memberMap := mapper.MapToPointer(members, teamMemberIds, func(m *models.TeamMember) uuid.UUID {
		return m.ID
	})
	return memberMap, nil
}

var _ DbTeamMemberStoreInterface = (*DbTeamMemberStore)(nil)

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

func (s *DbTeamMemberStore) LoadTeamMembersByUserAndTeamIds(ctx context.Context, userId uuid.UUID, teamIds ...uuid.UUID) ([]*models.TeamMember, error) {
	if len(teamIds) == 0 {
		return nil, errors.New("teamIds cannot be empty")
	}
	where := &map[string]any{
		models.TeamMemberTable.UserID: map[string]any{
			"_eq": userId,
		},
		models.TeamMemberTable.TeamID: map[string]any{
			"_in": teamIds,
		},
		models.TeamMemberTable.Active: map[string]any{
			"_eq": true,
		},
	}
	members, err := repository.TeamMember.Get(
		ctx,
		s.db,
		where,
		nil,
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}
	memberMap := mapper.MapToPointer(members, teamIds, func(m *models.TeamMember) uuid.UUID {
		return m.TeamID
	})
	return memberMap, nil
}
func (s *DbTeamMemberStore) CountTeamMembers(ctx context.Context, filter *TeamMemberFilter) (int64, error) {
	qs := squirrel.Select("COUNT(team_members.*)").From("team_members")
	qs = s.filterQuery(qs, filter)
	c, err := database.QueryWithBuilder[database.CountOutput](
		ctx,
		s.db,
		qs.PlaceholderFormat(squirrel.Dollar),
	)
	// where := s.filter(filter)
	// c, err := repository.TeamMember.Count(
	// 	ctx,
	// 	s.db,
	// 	where,
	// )
	if err != nil {
		return 0, err
	}
	if len(c) == 0 {
		return 0, nil
	}
	return c[0].Count, nil
}
func (s *DbTeamMemberStore) FindTeamMembers(ctx context.Context, filter *TeamMemberFilter) ([]*models.TeamMember, error) {
	qs := squirrel.Select("team_members.*").From("team_members")
	qs = s.filterQuery(qs, filter)
	qs = s.sortQuery(qs, filter)
	qs = queryPagination(qs, filter)
	members, err := database.QueryWithBuilder[*models.TeamMember](
		ctx,
		s.db,
		qs.PlaceholderFormat(squirrel.Dollar),
	)
	// where := s.filter(filter)
	// sort := s.sort(filter)
	// limit, offset := filter.LimitOffset()
	// members, err := repository.TeamMember.Get(
	// 	ctx,
	// 	s.db,
	// 	where,
	// 	sort,
	// 	&limit,
	// 	&offset,
	// )
	if err != nil {
		return nil, err
	}
	return members, nil
}

func (s *DbTeamMemberStore) filter(filter *TeamMemberFilter) *map[string]any {
	if filter == nil {
		return nil
	}
	where := make(map[string]any)
	// if filter.Q != "" {

	// }
	if len(filter.Ids) > 0 {
		where[models.TeamMemberTable.ID] = map[string]any{
			"_in": filter.Ids,
		}
	}
	if len(filter.Roles) > 0 {
		where[models.TeamMemberTable.Role] = map[string]any{
			"_in": filter.Roles,
		}
	}
	if len(filter.TeamIds) > 0 {
		where[models.TeamMemberTable.TeamID] = map[string]any{
			"_in": filter.TeamIds,
		}
	}
	if len(filter.UserIds) > 0 {
		where[models.TeamMemberTable.UserID] = map[string]any{
			"_in": filter.UserIds,
		}
	}
	if filter.Active.IsSet {
		where[models.TeamMemberTable.Active] = map[string]any{
			"_eq": filter.Active.Value,
		}
	}
	return &where
}
func (s *DbTeamMemberStore) filterQuery(qs squirrel.SelectBuilder, filter *TeamMemberFilter) squirrel.SelectBuilder {
	if filter == nil {
		return qs
	}
	if filter.Q != "" {
		qs = qs.Join("teams on team_members.team_id = teams.id").Join("users on team_members.user_id = users.id")
		qs = qs.Where(squirrel.Or{
			squirrel.ILike{"teams.name": "%" + filter.Q + "%"},
			squirrel.ILike{"users.email": "%" + filter.Q + "%"},
		})
	}
	if len(filter.Ids) > 0 {
		qs = qs.Where(squirrel.Eq{models.TeamMemberTable.ID: filter.Ids})

	}
	if len(filter.Roles) > 0 {
		qs = qs.Where(
			squirrel.Eq{
				models.TeamMemberTable.Role: filter.Roles,
			},
		)

	}
	if len(filter.TeamIds) > 0 {
		qs = qs.Where(
			squirrel.Eq{
				models.TeamMemberTable.TeamID: filter.TeamIds,
			},
		)
	}
	if len(filter.UserIds) > 0 {
		qs = qs.Where(
			squirrel.Eq{
				models.TeamMemberTable.UserID: filter.UserIds,
			},
		)
	}
	if filter.Active.IsSet {
		qs = qs.Where(
			squirrel.Eq{
				models.TeamMemberTable.Active: filter.Active.Value,
			},
		)
	}
	return qs
}

func (s *DbTeamMemberStore) FindTeamMember(ctx context.Context, filter *TeamMemberFilter) (*models.TeamMember, error) {
	where := s.filter(filter)
	member, err := repository.TeamMember.GetOne(
		ctx,
		s.db,
		where,
	)
	if err != nil {
		return nil, err
	}
	return member, nil
}

func (s *DbTeamMemberStore) sortQuery(qs squirrel.SelectBuilder, filter Sortable) squirrel.SelectBuilder {
	if filter == nil {
		return qs // return original query if no filter is provided
	}

	sortBy, sortOrder := filter.Sort()
	if sortBy != "" && slices.Contains(repository.TeamMemberBuilder.ColumnNames(), utils.Quote(sortBy)) {
		qs = qs.OrderBy(utils.Quote(sortBy) + " " + strings.ToUpper(sortOrder))
	} else if sortBy == "team.name" {
		qs = qs.OrderBy("teams.name " + strings.ToUpper(sortOrder))
	} else if sortBy == "user.email" {
		qs = qs.OrderBy("users.email " + strings.ToUpper(sortOrder))
	} else {
		slog.Info("sort by field not found in repository columns", "sortBy", sortBy, "sortOrder", sortOrder, "columns", repository.TeamMemberBuilder.ColumnNames())
	}
	return qs
}

// func (s *DbTeamMemberStore) sort(filter Sortable) *map[string]string {
// 	if filter == nil {
// 		return nil // return nil if no filter is provided
// 	}

// 	sortBy, sortOrder := filter.Sort()
// 	if sortBy != "" && slices.Contains(repository.TeamMemberBuilder.ColumnNames(), utils.Quote(sortBy)) {
// 		return &map[string]string{
// 			sortBy: sortOrder,
// 		}
// 	} else {
// 		slog.Info("sort by field not found in repository columns", "sortBy", sortBy, "sortOrder", sortOrder, "columns", repository.UserBuilder.ColumnNames())
// 	}

// 	return nil // default no sorting
// }

func (s *DbTeamMemberStore) CreateTeamFromUser(ctx context.Context, user *models.User) (*models.TeamMember, error) {
	team, err := repository.Team.PostOne(
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
	// Create a team member for the user
	teamMember, err := repository.TeamMember.PostOne(
		ctx,
		s.db,
		&models.TeamMember{
			TeamID:           team.ID,
			UserID:           types.Pointer(user.ID),
			Role:             models.TeamMemberRoleOwner,
			HasBillingAccess: true,
			Active:           true,
		},
	)
	if err != nil {
		return nil, err
	}
	if teamMember == nil {
		return nil, errors.New("team member not found")
	}
	return teamMember, nil
}

func (s *DbTeamMemberStore) CreateTeamMemberFromUserAndSlug(ctx context.Context, user *models.User, slug string, role models.TeamMemberRole) (*models.TeamMember, error) {
	team, err := repository.Team.GetOne(
		ctx,
		s.db,
		&map[string]any{
			models.TeamTable.Slug: map[string]any{
				"_eq": slug,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	if team == nil {
		// If team already exists, just create a team member
		team, err = repository.Team.PostOne(
			ctx,
			s.db,
			&models.Team{
				Name: slug,
				Slug: slug,
			},
		)
		if err != nil {
			return nil, err
		}
		if team == nil {
			return nil, errors.New("team not found")
		}
	}
	billingAccess := role == models.TeamMemberRoleOwner

	teamMember, err := repository.TeamMember.PostOne(
		ctx,
		s.db,
		&models.TeamMember{
			TeamID:           team.ID,
			UserID:           types.Pointer(user.ID),
			Role:             role,
			HasBillingAccess: billingAccess,
			Active:           true,
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
	_, err := repository.TeamMember.Delete(
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
	newMember, err := repository.TeamMember.PutOne(
		ctx,
		s.db,
		member,
	)
	if err != nil {
		return nil, err
	}
	return newMember, nil
}

// CountTeamMembers implements services.TeamStore.

// UpdateTeamMemberSelectedAt implements TeamQueryer.
func (s *DbTeamMemberStore) UpdateTeamMemberSelectedAt(ctx context.Context, teamId, userId uuid.UUID) error {
	qquery := squirrel.Update("team_members").
		Where(squirrel.Eq{models.TeamMemberTable.TeamID: teamId}).
		Where(squirrel.Eq{models.TeamMemberTable.UserID: userId}).
		Set(models.TeamMemberTable.LastSelectedAt, time.Now())

	_, err := database.ExecWithBuilder(ctx, s.db, qquery.PlaceholderFormat(squirrel.Dollar))
	return err
}

// FindLatestTeamMemberByUserID implements TeamQueryer.
func (s *DbTeamMemberStore) FindLatestTeamMemberByUserID(ctx context.Context, userId uuid.UUID) (*models.TeamMember, error) {
	teamMember, err := repository.TeamMember.Get(
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
func (s *DbTeamMemberStore) FindTeamMembersByUserID(ctx context.Context, userId uuid.UUID, paginate *TeamMemberListInput) ([]*models.TeamMember, error) {
	limit, offset := pagination(&paginate.PaginatedInput)
	orderby := make(map[string]string)
	if paginate.SortBy != "" && paginate.SortOrder != "" && slices.Contains(repository.TeamMemberBuilder.ColumnNames(), utils.Quote(paginate.SortBy)) {
		orderby[paginate.SortBy] = paginate.SortOrder
	} else {
		orderby["last_selected_at"] = "DESC"
	}
	qs := squirrel.Select("team_members.*").From("team_members")
	qs = qs.Where(squirrel.Eq{"user_id": userId})
	qs = qs.Where(squirrel.Eq{"active": true})
	if paginate.SortBy == "team.name" {
		qs = qs.Join("teams on team_members.team_id = teams.id").OrderBy("teams.name " + strings.ToUpper(paginate.SortOrder))
	} else if slices.Contains(repository.TeamMemberBuilder.ColumnNames(), utils.Quote(paginate.SortBy)) {
		qs = qs.OrderBy(paginate.SortBy + " " + strings.ToUpper(paginate.SortOrder))
	} else {
		qs = qs.OrderBy("last_selected_at DESC")
	}
	qs = qs.Limit(uint64(limit)).Offset(uint64(offset))
	teamMembers, err := database.QueryWithBuilder[*models.TeamMember](ctx, s.db, qs.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return nil, err
	}

	return teamMembers, nil
}

func (s *DbTeamMemberStore) CreateTeamMember(ctx context.Context, teamId, userId uuid.UUID, role models.TeamMemberRole, hasBillingAccess bool) (*models.TeamMember, error) {
	teamMember := &models.TeamMember{
		TeamID:           teamId,
		UserID:           &userId,
		Role:             role,
		Active:           true,
		HasBillingAccess: hasBillingAccess,
	}
	return repository.TeamMember.PostOne(
		ctx,
		s.db,
		teamMember,
	)
}
