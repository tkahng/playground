package stores

import (
	"context"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/repository"
	"github.com/tkahng/playground/internal/tools/mapper"
	"github.com/tkahng/playground/internal/tools/utils"
)

type TeamFilter struct {
	PaginatedInput
	SortParams
	Q           string      `query:"q"`
	Names       []string    `query:"names,omitempty" required:"false" json:"names,omitempty"`
	Slugs       []string    `query:"slugs,omitempty" required:"false" json:"slugs,omitempty"`
	Ids         []uuid.UUID `query:"ids,omitempty" required:"false" json:"ids,omitempty"`
	UserIds     []uuid.UUID `query:"user_ids,omitempty" required:"false" json:"user_ids,omitempty"`
	CustomerIds []string    `query:"customer_ids,omitempty" required:"false" json:"customer_ids,omitempty"`
}

type DbTeamGroupStoreInterface interface {
	CheckTeamSlug(ctx context.Context, slug string) (bool, error)
	CountTeams(ctx context.Context, params *TeamFilter) (int64, error)
	CreateTeam(ctx context.Context, name string, slug string) (*models.Team, error)
	DeleteTeam(ctx context.Context, teamId uuid.UUID) error
	FindTeam(ctx context.Context, team *TeamFilter) (*models.Team, error)
	FindTeamByID(ctx context.Context, teamId uuid.UUID) (*models.Team, error)
	FindTeamBySlug(ctx context.Context, slug string) (*models.Team, error)
	FindTeamByStripeCustomerId(ctx context.Context, stripeCustomerId string) (*models.Team, error)
	ListTeams(ctx context.Context, params *TeamFilter) ([]*models.Team, error)
	LoadTeamsByIds(ctx context.Context, teamIds ...uuid.UUID) ([]*models.Team, error)
	UpdateTeam(ctx context.Context, teamId uuid.UUID, name string) (*models.Team, error)
}

type DbTeamGroupStore struct {
	db database.Dbx
}

func NewDbTeamGroupStore(db database.Dbx) *DbTeamGroupStore {
	return &DbTeamGroupStore{
		db: db,
	}
}

func (s *DbTeamGroupStore) WithTx(tx database.Dbx) *DbTeamGroupStore {
	return &DbTeamGroupStore{
		db: tx,
	}
}

func (s *DbTeamGroupStore) FindTeam(ctx context.Context, filter *TeamFilter) (*models.Team, error) {
	where := s.filter(filter)
	team, err := repository.Team.GetOne(
		ctx,
		s.db,
		where,
	)
	if err != nil {
		return nil, err
	}
	return team, nil
}

func (s *DbTeamGroupStore) LoadTeamsByIds(ctx context.Context, teamIds ...uuid.UUID) ([]*models.Team, error) {
	var ids []string
	for _, id := range teamIds {
		ids = append(ids, id.String())
	}
	teams, err := repository.Team.Get(
		ctx,
		s.db,
		&map[string]any{
			models.TeamTable.ID: map[string]any{
				"_in": ids,
			},
		},
		nil,
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return mapper.MapTo(mapper.Map(teams, func(t *models.Team) models.Team {
		return *t
	}), teamIds, func(t models.Team) uuid.UUID {
		return t.ID
	}), nil
}

func (s *DbTeamGroupStore) FindTeamByStripeCustomerId(ctx context.Context, stripeCustomerId string) (*models.Team, error) {
	data, err := repository.Team.GetOne(
		ctx,
		s.db,
		&map[string]any{
			models.TeamTable.StripeCustomer: map[string]any{
				models.StripeCustomerTable.ID: map[string]any{
					"_eq": stripeCustomerId,
				},
			},
		},
	)
	return database.OptionalRow(data, err)
}

// DeleteTeam implements TeamQueryer.
func (s *DbTeamGroupStore) DeleteTeam(ctx context.Context, teamId uuid.UUID) error {
	_, err := repository.Team.Delete(
		ctx,
		s.db,
		&map[string]any{
			models.TeamTable.ID: map[string]any{
				"_eq": teamId,
			},
		},
	)
	if err != nil {
		return err
	}
	return nil
}

// FindTeamByID implements TeamQueryer.
func (s *DbTeamGroupStore) FindTeamByID(ctx context.Context, teamId uuid.UUID) (*models.Team, error) {
	return repository.Team.GetOne(
		ctx,
		s.db,
		&map[string]any{
			models.TeamTable.ID: map[string]any{
				"_eq": teamId,
			},
		},
	)
}

func (s *DbTeamGroupStore) FindTeamBySlug(ctx context.Context, slug string) (*models.Team, error) {
	return repository.Team.GetOne(
		ctx,
		s.db,
		&map[string]any{
			models.TeamTable.Slug: map[string]any{
				"_eq": slug,
			},
		},
	)
}

// UpdateTeam implements TeamQueryer.
func (s *DbTeamGroupStore) UpdateTeam(ctx context.Context, teamId uuid.UUID, name string) (*models.Team, error) {
	team := &models.Team{
		ID:   teamId,
		Name: name,
		// StripeCustomerID: stripeCustomerId,
		UpdatedAt: time.Now(),
	}
	_, err := repository.Team.PutOne(
		ctx,
		s.db,
		team,
	)
	if err != nil {
		return nil, err
	}
	return team, nil
}

func (s *DbTeamGroupStore) CreateTeam(ctx context.Context, name string, slug string) (*models.Team, error) {
	teamModel := &models.Team{
		Name: name,
		Slug: slug,
	}
	team, err := repository.Team.PostOne(
		ctx,
		s.db,
		teamModel,
	)
	if err != nil {
		return nil, err
	}
	return team, nil
}

func (s *DbTeamGroupStore) ListTeams(ctx context.Context, params *TeamFilter) ([]*models.Team, error) {
	where := s.filter(params)
	limit, offset := params.LimitOffset()
	sort := s.sort(params)
	teams, err := repository.Team.Get(
		ctx,
		s.db,
		where,
		sort,
		&limit,
		&offset,
	)
	if err != nil {
		return nil, err
	}
	return teams, nil
}

func (s *DbTeamGroupStore) sort(params *TeamFilter) *map[string]string {
	if params == nil {
		return nil
	}
	order := make(map[string]string)
	if params.SortBy != "" {
		if slices.Contains(repository.TeamBuilder.ColumnNames(), utils.Quote(params.SortBy)) {
			order[params.SortBy] = params.SortOrder
		}
	}
	return &order
}

func (s *DbTeamGroupStore) filter(params *TeamFilter) *map[string]any {
	if params == nil {
		return nil
	}
	where := make(map[string]any)
	if params.Q != "" {
		where["_or"] = []map[string]any{
			{
				"_and": []map[string]any{
					{
						models.TeamTable.Name: map[string]any{"_ilike": "%" + params.Q + "%"},
					},
				},
			},
			{
				"_and": []map[string]any{
					{
						models.TeamTable.Slug: map[string]any{"_ilike": "%" + params.Q + "%"},
					},
				},
			},
		}
	}
	if len(params.UserIds) > 0 {
		where[models.TeamTable.Members] = map[string]any{
			models.TeamMemberTable.UserID: map[string]any{
				"_in": params.UserIds,
			},
		}
	}
	if len(params.Ids) > 0 {
		where[models.TeamTable.ID] = map[string]any{
			"_in": params.Ids,
		}
	}
	if len(params.Names) > 0 {
		where[models.TeamTable.Name] = map[string]any{
			"_in": params.Names,
		}
	}
	if len(params.Slugs) > 0 {
		where[models.TeamTable.Slug] = map[string]any{
			"_in": params.Slugs,
		}
	}
	if len(params.CustomerIds) > 0 {
		where[models.TeamTable.StripeCustomer] = map[string]any{
			models.StripeCustomerTable.ID: map[string]any{
				"_in": params.CustomerIds,
			},
		}
	}

	return &where
}

func (s *DbTeamGroupStore) CountTeams(ctx context.Context, params *TeamFilter) (int64, error) {
	where := s.filter(params)
	count, err := repository.Team.Count(
		ctx,
		s.db,
		where,
	)
	// qs := squirrel.Select("COUNT(teams.*)").From("teams")
	// qs = listTeamsFilter(qs, params)
	// count, err := database.QueryWithBuilder[database.CountOutput](ctx, s.db, qs.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return 0, err
	}
	return count, nil
	// if len(count) == 0 {
	// 	return 0, nil
	// }
	// return count[0].Count, nil
}

func (s *DbTeamGroupStore) CheckTeamSlug(ctx context.Context, slug string) (bool, error) {
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
		return false, err
	}
	if team == nil {
		return true, nil
	}
	return false, nil
}

// func listTeamsFilter(qs squirrel.SelectBuilder, params *TeamFilter) squirrel.SelectBuilder {
// 	if params == nil {
// 		return qs
// 	}
// 	if params.Q != "" {
// 		qs = qs.Where(
// 			squirrel.Or{
// 				squirrel.ILike{models.TeamTable.Name: "%" + params.Q + "%"},
// 				squirrel.ILike{models.TeamTable.Slug: "%" + params.Q + "%"},
// 			},
// 		)
// 	}
// 	if len(params.UserIds) > 0 {
// 		qs = qs.Join("team_members ON teams.id = team_members.team_id").
// 			Where(squirrel.Eq{"team_members.user_id": params.UserIds})
// 	}
// 	if len(params.Ids) > 0 {
// 		qs = qs.Where(squirrel.Eq{models.TeamTable.ID: params.Ids})
// 	}
// 	if len(params.Names) > 0 {
// 		qs = qs.Where(squirrel.Eq{models.TeamTable.Name: params.Names})
// 	}
// 	if len(params.Slugs) > 0 {
// 		qs = qs.Where(squirrel.Eq{models.TeamTable.Slug: params.Slugs})
// 	}

// 	return qs
// }

// func listTeamsOrderBy(qs squirrel.SelectBuilder, params *TeamFilter) squirrel.SelectBuilder {
// 	fmt.Println("sortby", params.SortBy, "sortorder", params.SortOrder)
// 	if params.SortBy != "" && params.SortOrder != "" {
// 		if params.SortBy == "team_members.last_selected_at" {
// 			qs = qs.OrderBy("team_members.last_selected_at " + strings.ToUpper(params.SortOrder))
// 		} else if slices.Contains(repository.TeamBuilder.ColumnNames(), utils.Quote(params.SortBy)) {
// 			qs = qs.OrderBy(params.SortBy + " " + strings.ToUpper(params.SortOrder))
// 		}
// 	} else {
// 		qs = qs.OrderBy("created_at DESC")
// 	}
// 	return qs
// }
