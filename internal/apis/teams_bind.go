package apis

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/playground/internal/middleware"
	"github.com/tkahng/playground/internal/middleware/humamiddleware"
	"github.com/tkahng/playground/internal/models"
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

	appApi.FindTeamTeamMembersBind(teamsGroup)

	// check team slug
	appApi.CheckTeamSlugBind(teamsGroup)

	// create team
	appApi.CreateTeamBind(teamsGroup)

	// get team
	appApi.GetTeamBind(teamsGroup)
	// get team by slug
	appApi.FindTeamInfoBySlugBind(teamsGroup)

	// update team
	appApi.UpdateTeamBind(teamsGroup)

	// delete team
	appApi.DeleteTeamBind(teamsGroup)

	// update last selected team
	appApi.UpdateLastSelectedTeam(teamsGroup)

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
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.RequireTeamInfo(),
				middleware.RequireTeamMemberRolesMiddleware(models.TeamMemberRoleOwner),
			),
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
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.RequireTeamInfo(),
				middleware.RequireTeamMemberRolesMiddleware(models.TeamMemberRoleOwner),
			),
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
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.RequireTeamInfo(),
			),
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

	appApi.GetInvitationByTokenBind(teamsGroup)
	appApi.IssueSSETicketBind(teamsGroup)
	appApi.TeamMembersSseEventsBind(teamsGroup)

	appApi.FindTeamMembersNotificationsBind(teamsGroup)

	appApi.ReadTeamMembersNotificationsBind(teamsGroup)

	appApi.DeleteTeamMembersNotificationsBind(teamsGroup)

	appApi.UnreadNotificationsCountBind(teamsGroup)

	appApi.MarkAllNotificationsReadBind(teamsGroup)

	appApi.FindTeamMemberByIDBind(teamsGroup)

	appApi.GetUserTeamMembersBind(teamsGroup)

	appApi.UpdateTeamMemberBind(teamsGroup)

	appApi.DeactivateTeamMemberBind(teamsGroup)

	appApi.LeaveTeam(teamsGroup)

	appApi.ReassignBillingAccess(teamsGroup)
}
