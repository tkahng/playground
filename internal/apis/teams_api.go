package apis

import (
	"context"
	"log/slog"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/contextstore"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/stores"
	"github.com/tkahng/authgo/internal/tools/mapper"
)

type TeamInfo struct {
	User   ApiUser    `json:"user"`
	Team   Team       `json:"team"`
	Member TeamMember `json:"member"`
}

type TeamMemberRole string

const (
	TeamMemberRoleOwner  TeamMemberRole = "owner"
	TeamMemberRoleMember TeamMemberRole = "member"
	TeamMemberRoleGuest  TeamMemberRole = "guest"
)

type TeamMember struct {
	_                struct{}       `db:"team_members" json:"-"`
	ID               uuid.UUID      `db:"id" json:"id"`
	TeamID           uuid.UUID      `db:"team_id" json:"team_id"`
	UserID           *uuid.UUID     `db:"user_id" json:"user_id"`
	Active           bool           `db:"active" json:"active"`
	Role             TeamMemberRole `db:"role" json:"role" enum:"owner,member,guest"`
	HasBillingAccess bool           `db:"has_billing_access" json:"has_billing_access"`
	LastSelectedAt   time.Time      `db:"last_selected_at" json:"last_selected_at"`
	CreatedAt        time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time      `db:"updated_at" json:"updated_at"`
	Team             *Team          `db:"team" src:"team_id" dest:"id" table:"team" json:"team,omitempty"`
	User             *ApiUser       `db:"user" src:"user_id" dest:"id" table:"users" json:"user,omitempty"`
}

type Team struct {
	_              struct{}        `db:"teams" json:"-"`
	ID             uuid.UUID       `db:"id" json:"id"`
	Name           string          `db:"name" json:"name"`
	Slug           string          `db:"slug" json:"slug"`
	CreatedAt      time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time       `db:"updated_at" json:"updated_at"`
	Members        []*TeamMember   `db:"members" src:"id" dest:"team_id" table:"team_members" json:"members,omitempty"`
	StripeCustomer *StripeCustomer `db:"stripe_customer" src:"id" dest:"team_id" table:"stripe_customers" json:"stripe_customer,omitempty" required:"false"`
}

func FromTeamModel(team *models.Team) *Team {
	if team == nil {
		return nil
	}
	return &Team{
		ID:        team.ID,
		Name:      team.Name,
		Slug:      team.Slug,
		CreatedAt: team.CreatedAt,
		UpdatedAt: team.UpdatedAt,
		Members:   mapper.Map(team.Members, FromTeamMemberModel),
	}
}
func FromTeamMemberModel(member *models.TeamMember) *TeamMember {
	if member == nil {
		return nil
	}
	return &TeamMember{
		ID:               member.ID,
		TeamID:           member.TeamID,
		UserID:           member.UserID,
		Active:           member.Active,
		Role:             TeamMemberRole(member.Role),
		HasBillingAccess: member.HasBillingAccess,
		LastSelectedAt:   member.LastSelectedAt,
		CreatedAt:        member.CreatedAt,
		UpdatedAt:        member.UpdatedAt,
		Team:             FromTeamModel(member.Team),
		User:             FromUserModel(member.User),
	}
}

type CreateTeamInput struct {
	Name string `json:"name" required:"true"`
	Slug string `json:"slug" required:"true"`
}

type TeamOutput struct {
	Body *Team `json:"body"`
}
type TeamInfoOutput struct {
	Body *TeamInfo `json:"body"`
}

func (api *Api) CreateTeam(
	ctx context.Context,
	input *struct {
		Body CreateTeamInput `json:"body" required:"true"`
	},
) (
	*TeamOutput,
	error,
) {
	info := contextstore.GetContextUserInfo(ctx)
	if info == nil {
		return nil, huma.Error401Unauthorized("unauthorized")
	}
	team, err := api.app.Team().CreateTeamWithOwner(
		ctx,
		input.Body.Name,
		input.Body.Slug,
		info.User.ID,
	)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, huma.Error500InternalServerError("team not found")
	}
	return &TeamOutput{
		Body: FromTeamModel(&team.Team),
	}, nil
}

func (api *Api) CheckTeamSlug(
	ctx context.Context,
	input *struct {
		Body struct {
			Slug string `json:"slug" required:"true"`
		} `json:"body" required:"true"`
	},
) (
	*struct {
		Body struct {
			Exists bool `json:"exists"`
		}
	},
	error,
) {
	exists, err := api.app.Adapter().TeamGroup().CheckTeamSlug(ctx, input.Body.Slug)
	if err != nil {
		return nil, err
	}
	return &struct {
		Body struct {
			Exists bool "json:\"exists\""
		}
	}{
		Body: struct {
			Exists bool `json:"exists"`
		}{
			Exists: exists,
		},
	}, nil
}

type TeamMemberListInput struct {
	PaginatedInput
	SortParams
}

func (api *Api) GetUserTeamMembers(
	ctx context.Context,
	input *TeamMemberListInput,
) (
	*ApiPaginatedOutput[*TeamMember],
	error,
) {
	info := contextstore.GetContextUserInfo(ctx)
	if info == nil {
		return nil, huma.Error401Unauthorized("unauthorized")
	}
	filter := &stores.TeamMemberListInput{}
	filter.Page = input.Page
	filter.PerPage = input.PerPage
	filter.SortBy = input.SortBy
	filter.SortOrder = input.SortOrder
	teams, err := api.app.Team().FindTeamMembersByUserID(ctx, info.User.ID, filter)
	if err != nil {
		return nil, err
	}
	if len(teams) == 0 {
		return nil, huma.Error500InternalServerError("teams not found")
	}
	count, err := api.app.Adapter().TeamMember().CountTeamMembers(ctx, &stores.TeamMemberFilter{
		UserIds: []uuid.UUID{info.User.ID},
	})
	if err != nil {
		return nil, err
	}
	return &ApiPaginatedOutput[*TeamMember]{
		Body: ApiPaginatedResponse[*TeamMember]{
			Data: mapper.Map(teams, FromTeamMemberModel),
			Meta: ApiGenerateMeta(&input.PaginatedInput, count),
		},
	}, nil
}

type UserListTeamsParams struct {
	PaginatedInput
	SortParams
}

func (api *Api) GetUserTeams(
	ctx context.Context,
	input *UserListTeamsParams,
) (
	*ApiPaginatedOutput[*Team],
	error,
) {
	info := contextstore.GetContextUserInfo(ctx)
	if info == nil {
		return nil, huma.Error401Unauthorized("unauthorized")
	}
	params := &stores.TeamFilter{
		UserIds: []uuid.UUID{info.User.ID},
	}
	if input != nil {
		params.Page = input.Page
		params.PerPage = input.PerPage
		params.SortBy = input.SortBy
		params.SortOrder = input.SortOrder
	}

	teams, err := api.app.Adapter().TeamGroup().ListTeams(ctx, params)
	if err != nil {
		return nil, err
	}
	if len(teams) > 0 {
		teamIds := mapper.Map(teams, func(t *models.Team) uuid.UUID {
			return t.ID
		})
		members, err := api.app.Adapter().TeamMember().LoadTeamMembersByUserAndTeamIds(ctx, info.User.ID, teamIds...)
		if err != nil {
			return nil, err
		}
		for idx := range teamIds {
			team := teams[idx]
			member := members[idx]
			if team != nil && member != nil {
				team.Members = append(team.Members, member)
			}
		}
	}
	count, err := api.app.Adapter().TeamGroup().CountTeams(ctx, params)
	if err != nil {
		return nil, err
	}
	return &ApiPaginatedOutput[*Team]{
		Body: ApiPaginatedResponse[*Team]{
			Data: mapper.Map(teams, FromTeamModel),
			Meta: ApiGenerateMeta(&input.PaginatedInput, count),
		},
	}, nil
}

func (api *Api) FindTeamInfoBySlug(
	ctx context.Context,
	input *struct {
		Slug string `path:"team-slug" required:"true"`
	},
) (
	*TeamInfoOutput,
	error,
) {
	info := contextstore.GetContextTeamInfo(ctx)
	if info == nil {
		return nil, huma.Error401Unauthorized("unauthorized")
	}
	return &TeamInfoOutput{
		Body: &TeamInfo{
			Team:   *FromTeamModel(&info.Team),
			Member: *FromTeamMemberModel(&info.Member),
			User:   *FromUserModel(&info.User),
		},
	}, nil
}

func (api *Api) FindTeamMemberBySlug(
	ctx context.Context,
	input *struct {
		Slug string `path:"team-slug" required:"true"`
	},
) (
	*TeamMemberOutput,
	error,
) {
	info := contextstore.GetContextTeamInfo(ctx)
	if info == nil {
		return nil, huma.Error401Unauthorized("unauthorized")
	}
	return &TeamMemberOutput{
		Body: FromTeamMemberModel(&info.Member),
	}, nil
}

type TeamMemberOutput struct {
	Body *TeamMember `json:"body"`
}

func (api *Api) GetActiveTeamMember(
	ctx context.Context,
	input *struct{},
) (
	*TeamMemberOutput,
	error,
) {
	info := contextstore.GetContextUserInfo(ctx)
	if info == nil {
		return nil, huma.Error401Unauthorized("unauthorized")
	}
	team, err := api.app.Team().GetActiveTeamMember(ctx, info.User.ID)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, huma.Error404NotFound("team not found")
	}
	return &TeamMemberOutput{
		Body: FromTeamMemberModel(team),
	}, nil
}

type UpdateTeamInput struct {
	TeamID string `path:"team-id" required:"true"`
	Body   UpdateTeamDto
}

type UpdateTeamDto struct {
	Name string `json:"name" required:"true"`
	Slug string `json:"slug" required:"true"`
}

func (api *Api) UpdateTeam(
	ctx context.Context,
	input *UpdateTeamInput,
) (
	*TeamOutput,
	error,
) {
	info := contextstore.GetContextTeamInfo(ctx)
	if info == nil {
		return nil, huma.Error401Unauthorized("unauthorized")
	}
	team, err := api.app.Team().UpdateTeam(ctx, info.Team.ID, input.Body.Name)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, huma.Error500InternalServerError("team not found")
	}
	return &TeamOutput{
		Body: FromTeamModel(team),
	}, nil
}

func (api *Api) DeleteTeam(
	ctx context.Context,
	input *struct {
		TeamID string `path:"team-id" required:"true"`
	},
) (
	*TeamOutput,
	error,
) {
	info := contextstore.GetContextTeamInfo(ctx)
	if info == nil {
		return nil, huma.Error401Unauthorized("unauthorized")
	}
	slog.InfoContext(ctx, "Deleting team", slog.String("team_id", info.Team.ID.String()), slog.String("user_id", info.User.ID.String()))
	err := api.app.Team().DeleteTeam(ctx, info.Team.ID, info.User.ID)
	if err != nil {
		slog.ErrorContext(ctx, "error deleting team", "teamId", info.Team.ID.String(), "userId", info.User.ID.String(), "error", err)
		return nil, err
	}
	slog.InfoContext(ctx, "Team deleted successfully", slog.String("team_id", info.Team.ID.String()), slog.String("user_id", info.User.ID.String()))
	return nil, nil
}

func (api *Api) GetTeam(
	ctx context.Context,
	input *struct {
		TeamID string `path:"team-id" required:"true"`
	},
) (
	*TeamOutput,
	error,
) {
	info := contextstore.GetContextUserInfo(ctx)
	if info == nil {
		return nil, huma.Error401Unauthorized("unauthorized")
	}
	uid, err := uuid.Parse(input.TeamID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid team ID")
	}
	team, err := api.app.Adapter().TeamGroup().FindTeamByID(ctx, uid)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, huma.Error404NotFound("team not found")
	}
	return &TeamOutput{
		Body: FromTeamModel(team),
	}, nil
}

type SetCurrentTeamInput struct {
	TeamID string `json:"team_id" required:"true"`
}

func (api *Api) SetCurrentTeam(
	ctx context.Context,
	input *struct {
		Body SetCurrentTeamInput `json:"body" required:"true"`
	},
) (
	*struct {
		Body struct {
			Success bool `json:"success"`
		}
	},
	error,
) {
	info := contextstore.GetContextUserInfo(ctx)
	if info == nil {
		return nil, huma.Error401Unauthorized("unauthorized")
	}
	passedTeamID, err := uuid.Parse(input.Body.TeamID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid team ID")
	}
	teamMember, err := api.app.Team().SetActiveTeamMember(ctx, info.User.ID, passedTeamID)
	if err != nil {
		return nil, err
	}
	if teamMember == nil {
		return nil, huma.Error404NotFound("team not found")
	}
	return &struct {
		Body struct {
			Success bool `json:"success"`
		}
	}{
		Body: struct {
			Success bool `json:"success"`
		}{
			Success: true,
		},
	}, nil
}

type FindTeamTeamMembersInput struct {
	PaginatedInput
	SortParams
	TeamID string `path:"team-id" required:"true" format:"uuid"`
}

func (api *Api) FindTeamTeamMembers(
	ctx context.Context,
	input *FindTeamTeamMembersInput,
) (
	*ApiPaginatedOutput[*TeamMember],
	error,
) {
	teamID, err := uuid.Parse(input.TeamID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid team ID")
	}
	info := contextstore.GetContextUserInfo(ctx)
	if info == nil {
		return nil, huma.Error401Unauthorized("unauthorized")
	}
	filter := &stores.TeamMemberFilter{}
	filter.Page = input.Page
	filter.PerPage = input.PerPage
	filter.SortBy = input.SortBy
	filter.SortOrder = input.SortOrder
	filter.TeamIds = []uuid.UUID{teamID}
	teams, err := api.app.Adapter().TeamMember().FindTeamMembers(ctx, filter)
	if err != nil {
		return nil, err
	}
	if len(teams) == 0 {
		return nil, huma.Error500InternalServerError("teams not found")
	}
	count, err := api.app.Adapter().TeamMember().CountTeamMembers(ctx, filter)
	if err != nil {
		return nil, err
	}
	return &ApiPaginatedOutput[*TeamMember]{
		Body: ApiPaginatedResponse[*TeamMember]{
			Data: mapper.Map(teams, FromTeamMemberModel),
			Meta: ApiGenerateMeta(&input.PaginatedInput, count),
		},
	}, nil
}
