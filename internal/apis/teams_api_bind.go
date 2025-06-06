package apis

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/middleware"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
)

func BindTeamsApi(api huma.API, appApi *Api) {
	teamInfoMiddleware := middleware.TeamInfoFromParam(api, appApi.app)
	teamInfoSlugMiddleware := middleware.TeamInfoFromTeamSlug(api, appApi.app)
	requireMember := middleware.RequireTeamMemberRolesMiddleware(api)
	requiredOwnerMember := middleware.RequireTeamMemberRolesMiddleware(api, models.TeamMemberRoleOwner)
	checkTeamDelete := middleware.TeamCanDelete(api, appApi.app)
	emailVerified := middleware.EmailVerifiedMiddleware(api)
	teamsGroup := huma.NewGroup(api)
	// get team members
	//  /api/team-members
	huma.Register(
		teamsGroup,
		huma.Operation{
			OperationID: "get-team-members",
			Method:      http.MethodGet,
			Path:        "/team-members",
			Summary:     "get-team-members",
			Description: "get all team members",
			Tags:        []string{"Teams"},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Errors: []int{http.StatusInternalServerError, http.StatusBadRequest},
		},
		appApi.GetUserTeamMembers,
	)

	huma.Register(
		teamsGroup,
		huma.Operation{
			OperationID: "get-active-team-member",
			Method:      http.MethodGet,
			Path:        "/team-members/active",
			Summary:     "get-active-team-member",
			Description: "get the active team member by team ID and user ID",
			Tags:        []string{"Teams"},
			Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
		},
		appApi.GetActiveTeamMember,
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
			Path:        "/teams/slug/{slug}",
			Summary:     "get-team-by-slug",
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
	// get teammember by slug
	huma.Register(
		teamsGroup,
		huma.Operation{
			OperationID: "get-team-member-by-slug",
			Method:      http.MethodGet,
			Path:        "/team/slug/{slug}/member",
			Summary:     "get-team-member-by-slug",
			Description: "get a team member by team slug",
			Tags:        []string{"Teams"},
			Errors:      []int{http.StatusInternalServerError, http.StatusBadRequest},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: huma.Middlewares{
				teamInfoSlugMiddleware,
			},
		},
		appApi.FindTeamMemberBySlug,
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

}
