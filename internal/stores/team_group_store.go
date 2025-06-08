package stores

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/database"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
	"github.com/tkahng/authgo/internal/tools/utils"
)

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

// FindTeam implements services.TeamStore.
func (s *DbTeamGroupStore) FindTeam(ctx context.Context, team *models.Team) (*models.Team, error) {
	if team == nil {
		return nil, nil
	}
	where := map[string]any{}
	if team.ID != uuid.Nil {
		where[models.TeamTable.ID] = map[string]any{
			"_eq": team.ID,
		}
	}

	if team.Slug != "" {
		where[models.TeamTable.Slug] = map[string]any{
			"_eq": team.Slug,
		}
	}
	if team.Name != "" {
		where[models.TeamTable.Name] = map[string]any{
			"_eq": team.Name,
		}
	}
	if team.StripeCustomer != nil {
		if team.StripeCustomer.ID != "" {
			where[models.TeamTable.StripeCustomer] = map[string]any{
				"id": map[string]any{
					"_eq": team.StripeCustomer.ID,
				},
			}
		}
	}
	team, err := repository.Team.GetOne(
		ctx,
		s.db,
		&where,
	)
	if err != nil {
		return nil, err
	}
	return team, nil
}

// LoadTeamsByIds implements services.TeamStore.
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

func (s *DbTeamGroupStore) ListTeams(ctx context.Context, params *shared.ListTeamsParams) ([]*models.Team, error) {
	// Build the query
	if params == nil {
		params = &shared.ListTeamsParams{}
	}
	if params.UserID != "" && params.SortBy == "team_members.last_selected_at" {
		return nil, fmt.Errorf("cannot sort by team_members.last_selected_at without filtering by user_id")
	}
	limit, offset := database.PaginateRepo(&params.PaginatedInput)
	qs := squirrel.Select("teams.*").From("teams")
	qs = listTeamsFilter(qs, params)
	qs = listTeamsOrderBy(qs, params)
	qs = qs.Limit(uint64(*limit)).Offset(uint64(*offset))
	teams, err := database.QueryWithBuilder[*models.Team](ctx, s.db, qs.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return nil, err
	}
	return teams, nil
}

func (s *DbTeamGroupStore) CountTeams(ctx context.Context, params *shared.ListTeamsParams) (int64, error) {
	qs := squirrel.Select("COUNT(teams.*)").From("teams")
	qs = listTeamsFilter(qs, params)
	count, err := database.QueryWithBuilder[database.CountOutput](ctx, s.db, qs.PlaceholderFormat(squirrel.Dollar))
	if err != nil {
		return 0, err
	}
	if len(count) == 0 {
		return 0, nil
	}
	return count[0].Count, nil
}

// CheckTeamSlug implements services.TeamStore.
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

func listTeamsFilter(qs squirrel.SelectBuilder, params *shared.ListTeamsParams) squirrel.SelectBuilder {
	if params == nil {
		return qs
	}
	if params.Q != "" {
		qs = qs.Where(
			squirrel.Or{
				squirrel.ILike{models.TeamTable.Name: "%" + params.Q + "%"},
				squirrel.ILike{models.TeamTable.Slug: "%" + params.Q + "%"},
			},
		)
	}
	if params.UserID != "" {
		qs = qs.Join("team_members ON teams.id = team_members.team_id").
			Where(squirrel.Eq{"team_members.user_id": params.UserID})
	}
	return qs
}

func listTeamsOrderBy(qs squirrel.SelectBuilder, params *shared.ListTeamsParams) squirrel.SelectBuilder {
	fmt.Println("sortby", params.SortBy, "sortorder", params.SortOrder)
	if params.SortParams.SortBy != "" && params.SortParams.SortOrder != "" {
		if params.SortParams.SortBy == "team_members.last_selected_at" {
			qs = qs.OrderBy("team_members.last_selected_at " + strings.ToUpper(params.SortParams.SortOrder))
		} else if slices.Contains(repository.TeamBuilder.ColumnNames(), utils.Quote(params.SortParams.SortBy)) {
			qs = qs.OrderBy(params.SortParams.SortBy + " " + strings.ToUpper(params.SortParams.SortOrder))
		}
	} else {
		qs = qs.OrderBy("created_at DESC")
	}
	return qs
}
