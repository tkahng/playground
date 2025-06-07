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
	"github.com/tkahng/authgo/internal/services"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
	"github.com/tkahng/authgo/internal/tools/types"
)

type PostgresTeamServiceStore struct {
	*DbTeamStore
	*DbStripeStore
}

func NewPostgresTeamServiceStore(db database.Dbx) *PostgresTeamServiceStore {
	return &PostgresTeamServiceStore{
		DbTeamStore:   NewDbTeamStore(db),
		DbStripeStore: NewDbStripeStore(db),
	}
}

func (p *PostgresTeamServiceStore) WithTx(tx database.Dbx) *PostgresTeamServiceStore {
	return &PostgresTeamServiceStore{
		DbTeamStore:   p.DbTeamStore.WithTx(tx),
		DbStripeStore: p.DbStripeStore.WithTx(tx),
	}
}

type DbTeamStore struct {
	db database.Dbx
}

// FindTeam implements services.TeamStore.
func (s *DbTeamStore) FindTeam(ctx context.Context, team *models.Team) (*models.Team, error) {
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
	team, err := crudrepo.Team.GetOne(
		ctx,
		s.db,
		&where,
	)
	if err != nil {
		return nil, err
	}
	return team, nil
}

// FindTeamMember implements services.TeamStore.
func (s *DbTeamStore) FindTeamMember(ctx context.Context, member *models.TeamMember) (*models.TeamMember, error) {
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

// LoadTeamsByIds implements services.TeamStore.
func (s *DbTeamStore) LoadTeamsByIds(ctx context.Context, teamIds ...uuid.UUID) ([]*models.Team, error) {
	var ids []string
	for _, id := range teamIds {
		ids = append(ids, id.String())
	}
	teams, err := crudrepo.Team.Get(
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

// FindPendingInvitation implements services.TeamInvitationStore.
func (s *DbTeamStore) FindPendingInvitation(ctx context.Context, teamId uuid.UUID, email string) (*models.TeamInvitation, error) {
	invitation, err := crudrepo.TeamInvitation.GetOne(
		ctx,
		s.db,
		&map[string]any{
			models.TeamInvitationTable.TeamID: map[string]any{
				"_eq": teamId,
			},
			models.TeamInvitationTable.Email: map[string]any{
				"_eq": email,
			},
			models.TeamInvitationTable.Status: map[string]any{
				"_eq": string(models.TeamInvitationStatusPending),
			},
			models.TeamInvitationTable.ExpiresAt: map[string]any{
				"_gt": time.Now(),
			},
		},
	)
	if err != nil {
		return nil, err
	}
	return invitation, nil
}

func NewDbTeamStore(db database.Dbx) *DbTeamStore {
	return &DbTeamStore{
		db: db,
	}
}

func (s *DbTeamStore) WithTx(tx database.Dbx) *DbTeamStore {
	return &DbTeamStore{
		db: tx,
	}
}

// CreateTeamWithOwnerMember implements services.TeamStore.
func (s *DbTeamStore) CreateTeamWithOwnerMember(ctx context.Context, name string, slug string, userId uuid.UUID) (*shared.TeamInfoModel, error) {
	var teamInfo *shared.TeamInfoModel
	err := s.db.RunInTransaction(
		ctx,
		func(d database.Dbx) error {
			store := s.WithTx(d)
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

func (s *DbTeamStore) CreateTeamFromUser(ctx context.Context, user *models.User) (*models.TeamMember, error) {
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

// FindUserByID implements services.TeamInvitationStore.
func (s *DbTeamStore) FindUserByID(ctx context.Context, userId uuid.UUID) (*models.User, error) {
	user, err := crudrepo.User.GetOne(
		ctx,
		s.db,
		&map[string]any{
			models.UserTable.ID: map[string]any{
				"_eq": userId,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *DbTeamStore) FindTeamInvitations(ctx context.Context, teamId uuid.UUID) ([]*models.TeamInvitation, error) {
	invitations, err := crudrepo.TeamInvitation.Get(
		ctx,
		s.db,
		&map[string]any{
			models.TeamInvitationTable.TeamID: map[string]any{
				"_eq": teamId,
			},
			models.TeamInvitationTable.Status: map[string]any{
				"_eq": string(models.TeamInvitationStatusPending),
			},
			models.TeamInvitationTable.ExpiresAt: map[string]any{
				"_gt": time.Now(),
			},
		},
		&map[string]string{
			models.TeamInvitationTable.CreatedAt: "desc",
		},
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return invitations, nil
}

// FindInvitationByID implements services.TeamInvitationStore.
func (s *DbTeamStore) FindInvitationByID(ctx context.Context, invitationId uuid.UUID) (*models.TeamInvitation, error) {
	invitation, err := crudrepo.TeamInvitation.GetOne(
		ctx,
		s.db,
		&map[string]any{
			models.TeamInvitationTable.ID: map[string]any{
				"_eq": invitationId,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	if invitation == nil {
		return nil, nil
	}
	if invitation.ExpiresAt.Before(time.Now()) {
		fmt.Println("Invitation expired")
		fmt.Println(invitation.ExpiresAt)
		return nil, shared.ErrTokenExpired
	}
	return invitation, nil
}

// FindInvitationByToken implements services.TeamInvitationStore.
func (s *DbTeamStore) FindInvitationByToken(ctx context.Context, token string) (*models.TeamInvitation, error) {
	invitation, err := crudrepo.TeamInvitation.GetOne(
		ctx,
		s.db,
		&map[string]any{
			models.TeamInvitationTable.Token: map[string]any{
				"_eq": token,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	if invitation == nil {
		return nil, nil
	}
	if invitation.ExpiresAt.Before(time.Now()) {
		fmt.Println("Invitation expired")
		fmt.Println(invitation.ExpiresAt)
		return nil, shared.ErrTokenExpired
	}
	return invitation, nil
}

// CreateInvitation implements services.TeamInvitationStore.
func (s *DbTeamStore) CreateInvitation(ctx context.Context, invitation *models.TeamInvitation) error {
	_, err := crudrepo.TeamInvitation.PostOne(
		ctx,
		s.db,
		invitation,
	)
	return err
}

// GetInvitationByID implements services.TeamInvitationStore.
func (s *DbTeamStore) GetInvitationByID(ctx context.Context, invitationId uuid.UUID) (*models.TeamInvitation, error) {
	invitation, err := crudrepo.TeamInvitation.GetOne(
		ctx,
		s.db,
		&map[string]any{
			models.TeamInvitationTable.ID: map[string]any{
				"_eq": invitationId,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	if invitation == nil {
		return nil, nil
	}
	if invitation.ExpiresAt.Before(time.Now()) {
		return nil, shared.ErrTokenExpired
	}
	return invitation, nil
}

// UpdateInvitation implements services.TeamInvitationStore.
func (s *DbTeamStore) UpdateInvitation(ctx context.Context, invitation *models.TeamInvitation) error {
	_, err := crudrepo.TeamInvitation.PutOne(
		ctx,
		s.db,
		invitation,
	)

	if err != nil {
		return err
	}
	return nil
}

// DeleteTeamMember implements services.TeamStore.
func (s *DbTeamStore) DeleteTeamMember(ctx context.Context, teamId uuid.UUID, userId uuid.UUID) error {
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

// CheckTeamSlug implements services.TeamStore.
func (s *DbTeamStore) CheckTeamSlug(ctx context.Context, slug string) (bool, error) {
	team, err := crudrepo.Team.GetOne(
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

// UpdateTeamMember implements services.TeamStore.
func (s *DbTeamStore) UpdateTeamMember(ctx context.Context, member *models.TeamMember) (*models.TeamMember, error) {
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
func (s *DbTeamStore) CountOwnerTeamMembers(ctx context.Context, teamId uuid.UUID) (int64, error) {
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
func (s *DbTeamStore) CountTeamMembers(ctx context.Context, teamId uuid.UUID) (int64, error) {
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

// var _ services.TeamInvitationStore = &PostgresTeamStore{}
var _ services.TeamStore = &DbTeamStore{}
var _ services.TeamInvitationStore = &DbTeamStore{}

func (s *DbTeamStore) FindTeamByStripeCustomerId(ctx context.Context, stripeCustomerId string) (*models.Team, error) {
	data, err := crudrepo.Team.GetOne(
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

func (s *DbTeamStore) FindTeamMemberByTeamAndUserId(ctx context.Context, teamId, userId uuid.UUID) (*models.TeamMember, error) {
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
func (s *DbTeamStore) UpdateTeamMemberSelectedAt(ctx context.Context, teamId, userId uuid.UUID) error {
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
func (s *DbTeamStore) FindLatestTeamMemberByUserID(ctx context.Context, userId uuid.UUID) (*models.TeamMember, error) {
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

// DeleteTeam implements TeamQueryer.
func (s *DbTeamStore) DeleteTeam(ctx context.Context, teamId uuid.UUID) error {
	_, err := crudrepo.Team.Delete(
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
func (s *DbTeamStore) FindTeamByID(ctx context.Context, teamId uuid.UUID) (*models.Team, error) {
	return crudrepo.Team.GetOne(
		ctx,
		s.db,
		&map[string]any{
			models.TeamTable.ID: map[string]any{
				"_eq": teamId,
			},
		},
	)
}

func (s *DbTeamStore) FindTeamBySlug(ctx context.Context, slug string) (*models.Team, error) {
	return crudrepo.Team.GetOne(
		ctx,
		s.db,
		&map[string]any{
			models.TeamTable.Slug: map[string]any{
				"_eq": slug,
			},
		},
	)
}

// FindTeamMembersByUserID implements TeamQueryer.
func (s *DbTeamStore) FindTeamMembersByUserID(ctx context.Context, userId uuid.UUID, paginate *shared.TeamMemberListInput) ([]*models.TeamMember, error) {
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

func (s *DbTeamStore) CountTeamMembersByUserID(ctx context.Context, userId uuid.UUID) (int64, error) {
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

// UpdateTeam implements TeamQueryer.
func (s *DbTeamStore) UpdateTeam(ctx context.Context, teamId uuid.UUID, name string) (*models.Team, error) {
	team := &models.Team{
		ID:   teamId,
		Name: name,
		// StripeCustomerID: stripeCustomerId,
		UpdatedAt: time.Now(),
	}
	_, err := crudrepo.Team.PutOne(
		ctx,
		s.db,
		team,
	)
	if err != nil {
		return nil, err
	}
	return team, nil
}

func (s *DbTeamStore) CreateTeam(ctx context.Context, name string, slug string) (*models.Team, error) {
	teamModel := &models.Team{
		Name: name,
		Slug: slug,
	}
	team, err := crudrepo.Team.PostOne(
		ctx,
		s.db,
		teamModel,
	)
	if err != nil {
		return nil, err
	}
	return team, nil
}

func (s *DbTeamStore) CreateTeamMember(ctx context.Context, teamId, userId uuid.UUID, role models.TeamMemberRole, hasBillingAccess bool) (*models.TeamMember, error) {
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

func (s *DbTeamStore) ListTeams(ctx context.Context, params *shared.ListTeamsParams) ([]*models.Team, error) {
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

func (s *DbTeamStore) CountTeams(ctx context.Context, params *shared.ListTeamsParams) (int64, error) {
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
	if params.SortParams.SortBy != "" && params.SortParams.SortOrder != "" {
		if params.SortParams.SortBy == "team_members.last_selected_at" {
			qs = qs.OrderBy("team_members.last_selected_at " + strings.ToUpper(params.SortParams.SortOrder))
		} else if slices.Contains(crudrepo.TeamBuilder.ColumnNames(), params.SortParams.SortBy) {
			qs = qs.OrderBy(params.SortParams.SortBy + " " + strings.ToUpper(params.SortParams.SortOrder))
		}
	} else {
		qs = qs.OrderBy("created_at DESC")
	}
	return qs
}
