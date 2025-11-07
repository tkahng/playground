package apis

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/playground/internal/middleware"
	"github.com/tkahng/playground/internal/middleware/humamiddleware"
	"github.com/tkahng/playground/internal/shared"
)

func bindTeamsApi(appApi *Api) {
	app := appApi.App()
	teamsGroup := huma.NewGroup(appApi.Api())
	teamsGroup.UseMiddleware(
		humamiddleware.HumaChiMiddlewares(
			middleware.TeamFromParam(app),
			middleware.TeamFromParamSlug(app),
			middleware.TeamMemberFromParam(app),
			middleware.TeamInfoFromContext(app),
		)...,
	)
	// get team members
	//  /api/team-members

	appApi.bindFindTeamTeamMembers(teamsGroup)

	// check team slug
	appApi.bindCheckTeamSlug(teamsGroup)

	// get user teams
	appApi.bindGetUserTeams(teamsGroup)

	// create team
	appApi.bindCreateTeam(teamsGroup)

	// get team
	appApi.bindGetTeam(teamsGroup)
	// get team by slug
	appApi.bindFindTeamInfoBySlug(teamsGroup)

	// update team
	appApi.bindUpdateTeam(teamsGroup)

	// delete team
	huma.Register(
		teamsGroup,
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
			Middlewares: huma.Middlewares{
				appApi.Middlewares().GetTeamInfoFromTeamIDParam(),
				appApi.Middlewares().GetTeamRequiredOwnerMember(),
				appApi.Middlewares().GetTeamCanDelete(),
			},
		},
		appApi.DeleteTeam,
	)

	// team invitations -----------------------------------------------------------------------------------------------------------

	// create team invitation
	huma.Register(
		teamsGroup,
		huma.Operation{
			OperationID: "create-team-invitation",
			Method:      http.MethodPost,
			Path:        "/teams/{team-id}/invitations",
			Summary:     "create-team-invitation",
			Description: "create a team invitation",
			Tags:        []string{"Teams"},
			Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: huma.Middlewares{
				appApi.Middlewares().GetTeamInfoFromTeamIDParam(),
				appApi.Middlewares().GetTeamRequiredOwnerMember(),
			},
		},
		appApi.CreateInvitation,
	)

	// cancel invitation
	huma.Register(
		teamsGroup,
		huma.Operation{
			OperationID: "cancel-invitation",
			Method:      http.MethodDelete,
			Path:        "/teams/{team-id}/invitations/{invitation-id}",
			Summary:     "cancel-invitation",
			Description: "cancel invitation",
			Tags:        []string{"Teams", "Team Invitations"},
			Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: huma.Middlewares{
				appApi.Middlewares().GetTeamInfoFromTeamIDParam(),
				appApi.Middlewares().GetTeamRequiredOwnerMember(),
			},
		},
		appApi.CencelInvitation,
	)

	// find team invitations

	huma.Register(
		teamsGroup,
		huma.Operation{
			OperationID: "find-team-invitations",
			Method:      http.MethodGet,
			Path:        "/teams/{team-id}/invitations",
			Summary:     "find-team-invitations",
			Description: "find team invitations",
			Tags:        []string{"Teams", "Team Invitations"},
			Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: huma.Middlewares{
				appApi.Middlewares().GetTeamInfoFromTeamIDParam(),
			},
		},
		appApi.FindInvitations,
	)

	// check valid invitation
	huma.Register(
		teamsGroup,
		huma.Operation{
			OperationID: "check-valid-invitation",
			Method:      http.MethodPost,
			Path:        "/team-invitations/check",
			Summary:     "check-valid-invitation",
			Description: "check valid invitation",
			Tags:        []string{"Team Invitations"},
			Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
			Security:    []map[string][]string{{
				// shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: huma.Middlewares{},
		},
		appApi.CheckValidInvitation,
	)

	// accept invitation
	huma.Register(
		teamsGroup,
		huma.Operation{
			OperationID: "accept-invitation",
			Method:      http.MethodPost,
			Path:        "/team-invitations/accept",
			Summary:     "accept-invitation",
			Description: "accept invitation",
			Tags:        []string{"Team Invitations"},
			Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: huma.Middlewares{},
		},
		appApi.AcceptInvitation,
	)

	// decline invitation
	huma.Register(
		teamsGroup,
		huma.Operation{
			OperationID: "decline-invitation",
			Method:      http.MethodPost,
			Path:        "/team-invitations/decline",
			Summary:     "decline-invitation",
			Description: "decline invitation",
			Tags:        []string{"Team Invitations"},
			Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: huma.Middlewares{},
		},
		appApi.DeclineInvitation,
	)

	// find user team invitations

	huma.Register(
		teamsGroup,
		huma.Operation{
			OperationID: "find-user-team-invitations",
			Method:      http.MethodGet,
			Path:        "/team-invitations",
			Summary:     "find-user-team-invitations",
			Description: "find user team invitations",
			Tags:        []string{"Team Invitations"},
			Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: huma.Middlewares{},
		},
		appApi.GetUserTeamInvitations,
	)

	// find user team invitation by token
	huma.Register(
		teamsGroup,
		huma.Operation{
			OperationID: "find-user-team-invitation-by-token",
			Method:      http.MethodGet,
			Path:        "/team-invitations/token/{token}",
			Summary:     "find-user-team-invitation-by-token",
			Description: "find user team invitation by token",
			Tags:        []string{"Team Invitations"},
			Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
			Security:    []map[string][]string{{
				// shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: huma.Middlewares{},
		},
		appApi.GetInvitationByToken,
	)
	appApi.bindTeamMembersSseEvents(teamsGroup)

	appApi.bindFindTeamMembersNotifications(teamsGroup)

	appApi.bindReadTeamMembersNotifications(teamsGroup)

	appApi.bindDeleteTeamMembersNotifications(teamsGroup)

	appApi.bindFindTeamMemberByID(teamsGroup)
}
