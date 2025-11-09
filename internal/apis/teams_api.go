package apis

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/contextstore"
	"github.com/tkahng/playground/internal/middleware"
	"github.com/tkahng/playground/internal/middleware/humamiddleware"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/mapper"
)

type TeamInfo struct {
	User   ApiUser    `json:"user"`
	Team   Team       `json:"team"`
	Member TeamMember `json:"member"`
}

// TeamMemberRole
// enum:"owner,member,guest"
type TeamMemberRole string

func (role TeamMemberRole) String() string {
	return string(role)
}

const (
	TeamMemberRoleOwner  TeamMemberRole = "owner"
	TeamMemberRoleMember TeamMemberRole = "member"
	TeamMemberRoleGuest  TeamMemberRole = "guest"
)

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
type TeamWithMember struct {
	Team
	Member *TeamMember `json:"member,omitempty"`
}

func fromTeamModel(team *models.Team) *Team {
	if team == nil {
		return nil
	}
	return &Team{
		ID:        team.ID,
		Name:      team.Name,
		Slug:      team.Slug,
		CreatedAt: team.CreatedAt,
		UpdatedAt: team.UpdatedAt,
		Members:   mapper.Map(team.Members, fromTeamMemberModel),
	}
}

type CreateTeamInput struct {
	Name string `json:"name" required:"true"`
	Slug string `json:"slug" required:"true"`
}

type TeamOutput struct {
	Body *Team `json:"body"`
}
type TeamWithMemberOutput struct {
	Body *TeamWithMember `json:"body"`
}
type TeamInfoOutput struct {
	Body *TeamInfo `json:"body"`
}

func (api *Api) bindCreateTeam(humaApi huma.API) {
	huma.Register(
		humaApi,
		huma.Operation{
			OperationID: "create-team",
			Method:      http.MethodPost,
			Path:        "/teams",
			Summary:     "create-team",
			Description: "create a new team",
			Tags:        []string{"Teams"},
			Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.EmailVerifiedMiddleware(),
			),
		},
		func(ctx context.Context, input *struct {
			Body CreateTeamInput `json:"body" required:"true"`
		}) (*TeamWithMemberOutput, error) {
			if ok, err := api.App().Adapter().TeamGroup().CheckTeamSlug(ctx, input.Body.Slug); !ok {
				if err != nil {
					slog.ErrorContext(
						ctx,
						"error ocurred while checking slug",
					)
					return nil, fmt.Errorf("error ocurred while checking slug")
				}
				return nil, huma.Error400BadRequest("slug already exists")
			}
			info := contextstore.GetContextUserInfo(ctx)
			if info == nil {
				return nil, huma.Error401Unauthorized("unauthorized")
			}
			user := &info.User
			var teamInfo *models.TeamInfoModel
			runInTxErr := api.App().Adapter().RunInTxCtx(ctx, func(txCtx context.Context) error {
				teamInfoTx, err := api.App().Team().CreateTeamWithOwner(
					txCtx,
					input.Body.Name,
					input.Body.Slug,
					user.ID,
				)
				if err != nil {
					return err
				}
				if teamInfoTx == nil {
					return huma.Error500InternalServerError("team not found")
				}
				team := &teamInfoTx.Team

				_, err = api.App().Payment().CreateTeamCustomer(
					txCtx,
					team,
					user,
				)
				if err != nil {
					return err
				}

				teamInfo = teamInfoTx
				return nil
			})
			if runInTxErr != nil {
				return nil, runInTxErr
			}
			return &TeamWithMemberOutput{
				Body: &TeamWithMember{
					Team:   *fromTeamModel(&teamInfo.Team),
					Member: fromTeamMemberModel(&teamInfo.Member),
				},
			}, nil
		},
	)
}

func (api *Api) bindCheckTeamSlug(
	humaApi huma.API,
) {
	huma.Register(
		humaApi,
		huma.Operation{
			OperationID: "check-team-slug",
			Method:      http.MethodPost,
			Path:        "/teams/check-slug",
			Summary:     "check-team-slug",
			Description: "check if a team slug is available",
			Tags:        []string{"Teams"},
			Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
		},
		api.CheckTeamSlug,
	)
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
	exists, err := api.App().Adapter().TeamGroup().CheckTeamSlug(ctx, input.Body.Slug)
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

type UserListTeamsParams struct {
	PaginatedInput
	SortParams
}

func (api *Api) bindGetUserTeams(humaApi huma.API) {
	huma.Register(
		humaApi,
		huma.Operation{
			OperationID: "get-user-teams",
			Method:      http.MethodGet,
			Path:        "/teams",
			Summary:     "get-user-teams",
			Description: "get all teams for a user",
			Tags:        []string{"Teams"},
			Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
		},
		api.GetUserTeams,
	)
}

func (api *Api) GetUserTeams(
	ctx context.Context,
	input *UserListTeamsParams,
) (
	*ApiPaginatedOutput[*TeamWithMember],
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

	teams, err := api.App().Adapter().TeamGroup().ListTeams(ctx, params)
	if err != nil {
		return nil, err
	}
	var teamsWithMember []*TeamWithMember
	if len(teams) > 0 {
		teamIds := mapper.Map(teams, func(t *models.Team) uuid.UUID {
			return t.ID
		})
		members, err := api.App().Adapter().TeamMember().LoadTeamMembersByUserAndTeamIds(ctx, info.User.ID, teamIds...)
		if err != nil {
			return nil, err
		}
		for idx := range teamIds {
			team := teams[idx]
			member := members[idx]
			member.User = &info.User
			teamWithMember := &TeamWithMember{
				Team:   *fromTeamModel(team),
				Member: fromTeamMemberModel(member),
			}
			teamsWithMember = append(teamsWithMember, teamWithMember)
		}
	}
	count, err := api.App().Adapter().TeamGroup().CountTeams(ctx, params)
	if err != nil {
		return nil, err
	}
	return &ApiPaginatedOutput[*TeamWithMember]{
		Body: ApiPaginatedResponse[*TeamWithMember]{
			Data: teamsWithMember,
			Meta: ApiGenerateMeta(&input.PaginatedInput, count),
		},
	}, nil
}

func (api *Api) bindFindTeamInfoBySlug(humaApi huma.API) {
	huma.Register(
		humaApi,
		huma.Operation{
			OperationID: "get-team-by-slug",
			Method:      http.MethodGet,
			Path:        "/teams/slug/{team-slug}",
			Summary:     "get-team-info-by-slug",
			Description: "get a team by slug",
			Tags:        []string{"Teams"},
			Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.RequireTeamInfo(),
			),
		},
		func(ctx context.Context, input *struct {
			Slug string `path:"team-slug" required:"true"`
		}) (*TeamInfoOutput, error) {
			info := contextstore.GetContextTeamInfo(ctx)
			if info == nil {
				return nil, huma.Error401Unauthorized("unauthorized")
			}
			return &TeamInfoOutput{
				Body: &TeamInfo{
					Team:   *fromTeamModel(&info.Team),
					Member: *fromTeamMemberModel(&info.Member),
					User:   *fromUserModel(&info.User),
				},
			}, nil
		},
	)
}

func (api *Api) bindUpdateTeam(humaApi huma.API) {
	huma.Register(
		humaApi,
		huma.Operation{
			OperationID: "update-team",
			Method:      http.MethodPut,
			Path:        "/teams/{team-id}",
			Summary:     "update-team",
			Description: "update a team by ID",
			Tags:        []string{"Teams"},
			Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.RequireTeamInfo(),
				middleware.RequireTeamMemberRolesMiddleware(models.TeamMemberRoleOwner),
			),
		},
		api.UpdateTeam,
	)
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
	team, err := api.App().Team().UpdateTeam(ctx, info.Team.ID, input.Body.Name)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, huma.Error500InternalServerError("team not found")
	}
	return &TeamOutput{
		Body: fromTeamModel(team),
	}, nil
}

func (api *Api) bindDeleteTeam(humaApi huma.API) {
	huma.Register(
		humaApi,
		huma.Operation{
			OperationID: "delete-team",
			Method:      http.MethodDelete,
			Path:        "/teams/{team-id}",
			Summary:     "delete-team",
			Description: "delete a team by ID",
			Tags:        []string{"Teams"},
			Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.RequireTeamInfo(),
				middleware.RequireTeamMemberRolesMiddleware(models.TeamMemberRoleOwner),
				middleware.TeamCanDelete(api.App()),
			),
		},
		func(ctx context.Context, input *struct {
			TeamID string `path:"team-id" required:"true"`
		}) (*TeamOutput, error) {
			info := contextstore.GetContextTeamInfo(ctx)
			if info == nil {
				return nil, huma.Error401Unauthorized("unauthorized")
			}
			err := api.App().Team().DeleteTeam(ctx, info.Team.ID, info.User.ID)
			if err != nil {
				slog.ErrorContext(ctx, "error deleting team", "teamId", info.Team.ID.String(), "userId", info.User.ID.String(), "error", err)
				return nil, err
			}
			return nil, nil
		},
	)
}

func (api *Api) bindGetTeam(humaApi huma.API) {
	huma.Register(
		humaApi,
		huma.Operation{
			OperationID: "get-team",
			Method:      http.MethodGet,
			Path:        "/teams/{team-id}",
			Summary:     "get-team",
			Description: "get a team by ID",
			Tags:        []string{"Teams"},
			Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: huma.Middlewares{},
		},
		func(ctx context.Context, input *struct {
			TeamID string `path:"team-id" required:"true"`
		}) (*TeamOutput, error) {
			team := contextstore.GetContextTeam(ctx)
			if team == nil {
				return nil, huma.Error404NotFound("team not found")
			}
			return &TeamOutput{
				Body: fromTeamModel(team),
			}, nil
		},
	)
}
