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
	"github.com/tkahng/playground/internal/database/repository"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/tools/mapper"
	"github.com/tkahng/playground/internal/tools/types"
)

type TeamMemberListInput struct {
	PaginatedInput
	SortParams
}

type TeamMemberFilter struct {
	PaginatedInput
	SortParams
	Q                string                    `query:"q"`
	Ids              []uuid.UUID               `query:"ids"`
	Roles            []models.TeamMemberRole   `query:"roles"`
	UserIds          []uuid.UUID               `query:"user_ids"`
	TeamIds          []uuid.UUID               `query:"team_ids"`
	TeamNames        []string                  `query:"team_names,omitempty" required:"false" json:"team_names,omitempty"`
	UserEmails       []string                  `query:"user_emails,omitempty" required:"false" json:"user_emails,omitempty"`
	Active           types.OptionalParam[bool] `query:"active"`
	HasBillingAccess types.OptionalParam[bool] `query:"has_billing_access"`
}

type DbTeamMemberStoreInterface interface {
	LoadTeamMembersByUserAndTeamIds(ctx context.Context, userId uuid.UUID, teamIds ...uuid.UUID) ([]*models.TeamMember, error)
	LoadTeamMembersByIds(ctx context.Context, teamMemberIds ...uuid.UUID) ([]*models.TeamMember, error)
	FindTeamMembers(ctx context.Context, filter *TeamMemberFilter) ([]*models.TeamMember, error)
	CountTeamMembers(ctx context.Context, filter *TeamMemberFilter) (int64, error)
	CreateTeamFromUser(ctx context.Context, user *models.User) (*models.TeamMember, error)
	CreateTeamMember(ctx context.Context, model *models.TeamMember) (*models.TeamMember, error)
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
	qs := squirrel.Select("COUNT(org.team_members.*)").From("org.team_members")
	qs = s.setJoin(qs, filter)
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

func (s *DbTeamMemberStore) FindTeamMember(ctx context.Context, filter *TeamMemberFilter) (*models.TeamMember, error) {
	if filter != nil {
		filter.PaginatedInput = PaginatedInput{
			Page:    0,
			PerPage: 1,
		}
	}
	result, err := s.FindTeamMembers(ctx, filter)
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, nil
	}
	re := result[0]
	return re, nil
}

func (s *DbTeamMemberStore) FindTeamMembers(ctx context.Context, filter *TeamMemberFilter) ([]*models.TeamMember, error) {
	qs := squirrel.Select("org.team_members.*").From("org.team_members")
	qs = s.setJoin(qs, filter)
	qs = s.filterQuery(qs, filter)
	qs = s.sortQuery(qs, filter)
	qs = queryPagination(qs, filter)
	members, err := database.QueryWithBuilder[*models.TeamMember](
		ctx,
		s.db,
		qs.PlaceholderFormat(squirrel.Dollar),
	)
	if err != nil {
		slog.ErrorContext(
			ctx,
			"error while finding team members",
			slog.Any("error", err),
			slog.Any("filter", filter),
		)
		return nil, err
	}
	return members, nil
}

// enum:"team.name,team.slug,team.created_at,team.updated_at,user.email,user.name,user.created_at,user.updated_at"
var joinSortCols = []string{
	"team.name",
	"team.slug",
	"team.created_at",
	"team.updated_at",
	"user.email",
	"user.name",
	"user.created_at",
	"user.updated_at",
}

func needsJoin(filter *TeamMemberFilter) bool {
	if filter == nil {
		return false
	}
	// check filters
	if filter.Q != "" {
		return true
	}
	if len(filter.TeamNames) > 0 {
		return true
	}
	if len(filter.UserEmails) > 0 {
		return true
	}
	// check sorts
	if filter.SortBy != "" {
		if slices.Contains(joinSortCols, filter.SortBy) {
			return true
		}
	}

	return false
}
func (s *DbTeamMemberStore) setJoin(qs squirrel.SelectBuilder, filter *TeamMemberFilter) squirrel.SelectBuilder {
	if needsJoin(filter) {
		qs = qs.Join("org.teams on org.team_members.team_id = org.teams.id").Join("auth.users on org.team_members.user_id = auth.users.id")
	}
	return qs
}
func (s *DbTeamMemberStore) filterQuery(qs squirrel.SelectBuilder, filter *TeamMemberFilter) squirrel.SelectBuilder {
	if filter == nil {
		return qs
	}
	if filter.Q != "" {
		qs = qs.Where(squirrel.Or{
			squirrel.ILike{"org.teams.name": "%" + filter.Q + "%"},
			squirrel.ILike{"auth.users.email": "%" + filter.Q + "%"},
		})
	}
	if len(filter.Ids) > 0 {
		qs = qs.Where(squirrel.Eq{"org.team_members.id": filter.Ids})

	}
	if len(filter.Roles) > 0 {
		qs = qs.Where(
			squirrel.Eq{
				"org.team_members.role": filter.Roles,
			},
		)
	}
	if len(filter.UserIds) > 0 {
		qs = qs.Where(
			squirrel.Eq{
				"org.team_members.user_id": filter.UserIds,
			},
		)
	}
	if len(filter.TeamIds) > 0 {
		qs = qs.Where(
			squirrel.Eq{
				"org.team_members.team_id": filter.TeamIds,
			},
		)
	}
	if len(filter.TeamNames) > 0 {
		qs = qs.Where(
			squirrel.Eq{
				"org.teams.name": filter.TeamNames,
			},
		)
	}
	if len(filter.UserEmails) > 0 {
		qs = qs.Where(
			squirrel.Eq{
				"auth.users.email": filter.UserEmails,
			},
		)
	}
	if filter.Active.IsSet {
		qs = qs.Where(
			squirrel.Eq{
				"org.team_members.active": filter.Active.Value,
			},
		)
	}
	if filter.HasBillingAccess.IsSet {
		qs = qs.Where(
			squirrel.Eq{
				"org.team_members.has_billing_access": filter.HasBillingAccess.Value,
			},
		)
	}

	return qs
}

func (s *DbTeamMemberStore) sortQuery(qs squirrel.SelectBuilder, filter Sortable) squirrel.SelectBuilder {
	if filter == nil {
		return qs // return original query if no filter is provided
	}
	sortBy, sortOrder := filter.Sort()
	if sortBy == "" {
		return qs
	}
	if sortOrder == "" {
		sortOrder = "ASC"
	}
	// if sortBy is in the registered fieldnames, it is a scalar field. direct sort.
	if slices.Contains(repository.TeamMemberBuilder.FieldNames(), sortBy) {
		qs = qs.OrderBy(sortBy + " " + strings.ToUpper(sortOrder))

	} else if slices.Contains(joinSortCols, sortBy) {
		// if the sort by field is a joined field, we need to prefix it with the table name.
		// if `team.` is a prefix for team fields, `user.` is a prefix for user fields.
		if strings.HasPrefix(sortBy, "team.") {
			sortBy = "org." + strings.ReplaceAll(sortBy, "team.", "teams.")
		} else if strings.HasPrefix(sortBy, "user.") {
			sortBy = "auth." + strings.ReplaceAll(sortBy, "user.", "users.")
		}
		qs = qs.OrderBy(sortBy + " " + strings.ToUpper(sortOrder))
	} else {
		slog.Warn("sort by field not found in repository columns", "sortBy", sortBy, "sortOrder", sortOrder)
	}
	return qs
}

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

// CreateTeamMemberFromUserAndSlug creates a team member from a user and a team slug.
// If the team does not exist, it will be created.
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
	qquery := squirrel.Update("org.team_members").
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
			models.TeamMemberTable.Active: map[string]any{
				"_eq": true,
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
	if paginate.SortBy != "" && paginate.SortOrder != "" && slices.Contains(repository.TeamMemberBuilder.FieldNames(), paginate.SortBy) {
		orderby[paginate.SortBy] = paginate.SortOrder
	} else {
		orderby["last_selected_at"] = "DESC"
	}
	qs := squirrel.Select("org.team_members.*").From("org.team_members")
	qs = qs.Where(squirrel.Eq{"org.team_members.user_id": userId})
	qs = qs.Where(squirrel.Eq{"org.team_members.active": true})
	if paginate.SortBy == "team.name" {
		qs = qs.Join("org.teams on org.team_members.team_id = org.teams.id").OrderBy("org.teams.name " + strings.ToUpper(paginate.SortOrder))
	} else if slices.Contains(repository.TeamMemberBuilder.FieldNames(), paginate.SortBy) {
		qs = qs.OrderBy(paginate.SortBy + " " + strings.ToUpper(paginate.SortOrder))
	} else {
		qs = qs.OrderBy("org.team_members.last_selected_at DESC")
	}
	qs = qs.Limit(uint64(limit)).Offset(uint64(offset))
	teamMembers, err := database.QueryWithBuilder[*models.TeamMember](ctx, s.db, qs.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return nil, err
	}

	return teamMembers, nil
}
func (s *DbTeamMemberStore) CreateTeamMember(ctx context.Context, model *models.TeamMember) (*models.TeamMember, error) {
	return repository.TeamMember.PostOne(
		ctx,
		s.db,
		model,
	)
}
