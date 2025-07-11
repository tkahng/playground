package apis

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/middleware"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
)

func BindTeamsApi(api huma.API, appApi *Api) {
	teamInfoMiddleware := middleware.TeamInfoFromParam(api, appApi.App())
	teamInfoSlugMiddleware := middleware.TeamInfoFromTeamSlug(api, appApi.App())
	requireMember := middleware.RequireTeamMemberRolesMiddleware(api)
	requiredOwnerMember := middleware.RequireTeamMemberRolesMiddleware(api, models.TeamMemberRoleOwner)
	checkTeamDelete := middleware.TeamCanDelete(api, appApi.App())
	emailVerified := middleware.EmailVerifiedMiddleware(api)
	teamsGroup := huma.NewGroup(api)
	// get team members
	//  /api/team-members

	huma.Register(
		teamsGroup,
		huma.Operation{
			OperationID: "get-team-team-members",
			Method:      http.MethodGet,
			Path:        "/teams/{team-id}/members",
			Summary:     "get-team-team-members",
			Description: "get members of a team by team team ID",
			Tags:        []string{"Teams", "Team Members"},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Errors: []int{http.StatusInternalServerError, http.StatusBadRequest},
		},
		appApi.FindTeamTeamMembers,
	)

	// check team slug
	huma.Register(
		teamsGroup,
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
		appApi.CheckTeamSlug,
	)

	// get user teams
	huma.Register(
		teamsGroup,
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
		appApi.GetUserTeams,
	)

	// create team
	huma.Register(
		teamsGroup,
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
			Middlewares: huma.Middlewares{
				emailVerified,
			},
		},
		appApi.CreateTeam,
	)
	// get team
	huma.Register(
		teamsGroup,
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
			Middlewares: huma.Middlewares{
				teamInfoMiddleware,
				requireMember,
			},
		},
		appApi.GetTeam,
	)
	// get team by slug
	huma.Register(
		teamsGroup,
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
			Middlewares: huma.Middlewares{
				teamInfoSlugMiddleware,
			},
		},
		appApi.FindTeamInfoBySlug,
	)

	// update team
	huma.Register(
		teamsGroup,
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
			Middlewares: huma.Middlewares{
				teamInfoMiddleware,
				requiredOwnerMember,
			},
		},
		appApi.UpdateTeam,
	)

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
				teamInfoMiddleware,
				requiredOwnerMember,
				checkTeamDelete,
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
				teamInfoMiddleware,
				requiredOwnerMember,
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
				teamInfoMiddleware,
				requiredOwnerMember,
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
				teamInfoMiddleware,
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
	appApi.BindTeamMembersSseEvents(api)

	appApi.BindFindTeamMembersNotifications(teamsGroup)

	appApi.BindReadTeamMembersNotifications(teamsGroup)
}
