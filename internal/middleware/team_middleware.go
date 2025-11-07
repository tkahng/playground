package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"slices"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/contextstore"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
	appHttp "github.com/tkahng/playground/internal/tools/http"
	"github.com/tkahng/playground/internal/tools/types"
)

func TeamMemberFromParam(app core.App) HttpMiddelwareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// context
			// get team-member-id
			key := "team-member-id"
			teamMemberID := appHttp.GetParam(r, key)
			// if team-member-id is empty then move on
			if teamMemberID == "" {
				next.ServeHTTP(w, r)
				return
			}
			rawCtx := r.Context()
			// parse team-member-id to uuid
			parsedTeamMemberID, err := uuid.Parse(teamMemberID)
			if err != nil {
				slog.ErrorContext(
					rawCtx,
					"error while parsing teamMemberID in TeamMemberFromParam middleware",
					slog.Any("error", err),
					slog.String(key, teamMemberID),
				)
				_ = appHttp.WriteErr(w, r, http.StatusBadRequest, "error parsing team member id. invalid UUID format", err)
				return
			}
			// query teamMember from team-member-id.
			// do not filter by active, we will filter it later
			teamMember, err := app.Adapter().TeamMember().FindTeamMember(rawCtx, &stores.TeamMemberFilter{
				Ids: []uuid.UUID{parsedTeamMemberID},
			})
			if err != nil {
				slog.ErrorContext(
					rawCtx,
					"error while querying teamMember in TeamMemberFromParam middleware",
					slog.Any("error", err),
					slog.String(key, teamMemberID),
				)
				_ = appHttp.WriteErr(w, r, http.StatusInternalServerError, "error getting team member")
				return
			}
			// if teamMember id was not nil and valid uuid,
			// it is reasonable to assume that the teamMember exists,
			// assuming the uuid was provided within our system.
			// but this is not a security check, so we log it and move on
			if teamMember == nil {
				slog.WarnContext(
					rawCtx,
					"no teamMember with given team-member-id found in TeamMemberFromParam middleware",
					slog.String(key, teamMemberID),
				)
				next.ServeHTTP(w, r)
				return
			}
			// add teamMember to context
			newCtx := contextstore.SetContextTeamMember(rawCtx, teamMember)
			r = r.WithContext(newCtx)
			next.ServeHTTP(w, r)
		})
	}
}

func TeamFromParam(app core.App) HttpMiddelwareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// context
			// get team-id
			key := "team-id"
			teamId := appHttp.GetParam(r, key)
			// if team-id is empty then move on
			if teamId == "" {
				next.ServeHTTP(w, r)
				return
			}
			// parse team-id to uuid
			parsedTeamID, err := uuid.Parse(teamId)
			if err != nil {
				slog.ErrorContext(
					r.Context(),
					"error while parsing teamID in TeamFromParam middleware",
					slog.Any("error", err),
					slog.String(key, teamId),
				)
				_ = appHttp.WriteErr(w, r, http.StatusBadRequest, "error parsing team id. invalid UUID format")
				return
			}
			rawCtx := r.Context()
			// query team from team-id.
			// do not filter by active, we will filter it later
			team, err := app.Adapter().TeamGroup().FindTeamByID(rawCtx, parsedTeamID)
			if err != nil {
				slog.ErrorContext(
					rawCtx,
					"error while querying team in TeamFromParam middleware",
					slog.Any("error", err),
					slog.String(key, teamId),
				)
				_ = appHttp.WriteErr(w, r, http.StatusInternalServerError, "error while querying team")
				return
			}
			// if team id was not nil and valid uuid,
			// it is reasonable to assume that the team exists,
			// assuming the uuid was provided within our system.
			// however this is not a security check, so we will log it and move on.
			if team == nil {
				slog.WarnContext(
					rawCtx,
					"no team with given team-id found",
					slog.String(key, teamId),
				)
				next.ServeHTTP(w, r)
				return
			}
			// add team to context
			newCtx := contextstore.SetContextTeam(rawCtx, team)
			r = r.WithContext(newCtx)
			next.ServeHTTP(w, r)
		})
	}
}
func TeamFromParamSlug(app core.App) HttpMiddelwareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// context
			// get team-id
			key := "team-slug"
			teamSlug := appHttp.GetParam(r, key)
			// if team-id is empty then move on
			if teamSlug == "" {
				next.ServeHTTP(w, r)
				return
			}
			rawCtx := r.Context()
			// query team from team-slug.
			// do not filter by active, we will filter it later
			team, err := app.Adapter().TeamGroup().FindTeamBySlug(rawCtx, teamSlug)
			if err != nil {
				slog.ErrorContext(
					rawCtx,
					"error while querying team in TeamFromParam middleware",
					slog.Any("error", err),
					slog.String(key, teamSlug),
				)
				_ = appHttp.WriteErr(w, r, http.StatusInternalServerError, "error while querying team")
				return
			}
			// if team id was not nil and valid uuid,
			// it is reasonable to assume that the team exists,
			// assuming the uuid was provided within our system.
			// however this is not a security check, so we will log it and move on.
			if team == nil {
				slog.WarnContext(
					rawCtx,
					"no team with given team-slug found",
					slog.String(key, teamSlug),
				)
				next.ServeHTTP(w, r)
				return
			}
			// add team to context
			newCtx := contextstore.SetContextTeam(rawCtx, team)
			r = r.WithContext(newCtx)
			next.ServeHTTP(w, r)
		})
	}
}

// MemberIdBelongsToUser middleware ensures that the user is the member with id {team-member-id}
func MemberIdBelongsToUser(app core.App) HttpMiddelwareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawCtx := r.Context()
			userInfo := contextstore.GetContextTeamInfo(rawCtx)
			if userInfo == nil {
				_ = appHttp.WriteErr(w, r, http.StatusUnauthorized, "unauthorized at middleware", nil)
				return
			}
			teamMemberID := appHttp.GetParam(r, "team-member-id")
			if teamMemberID == "" {
				_ = appHttp.WriteErr(w, r, http.StatusBadRequest, "team slug is required", nil)
				return
			}
			parsedTeamMemberID, err := uuid.Parse(teamMemberID)
			if err != nil {
				_ = appHttp.WriteErr(w, r, http.StatusBadRequest, "error parsing team member id", err)
				return
			}
			if userInfo.Member.ID != parsedTeamMemberID {
				_ = appHttp.WriteErr(w, r, http.StatusUnauthorized, "unauthorized at middleware", nil)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// TeamInfoFromUserAndMemberID finds the team info from the userId and teamId of the member of {team-member-id}
func TeamInfoFromUserAndMemberID(app core.App) HttpMiddelwareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawCtx := r.Context()
			userInfo := contextstore.GetContextUserInfo(rawCtx)
			if userInfo == nil {
				_ = appHttp.WriteErr(w, r, http.StatusUnauthorized, "unauthorized at middleware", nil)
				return
			}
			teamMemberID := appHttp.GetParam(r, "team-member-id")
			if teamMemberID == "" {
				_ = appHttp.WriteErr(w, r, http.StatusBadRequest, "team slug is required", nil)
				return
			}
			parsedTeamMemberID, err := uuid.Parse(teamMemberID)
			if err != nil {
				_ = appHttp.WriteErr(w, r, http.StatusBadRequest, "error parsing team member id", err)
				return
			}
			teamMember, err := app.Adapter().TeamMember().FindTeamMember(rawCtx, &stores.TeamMemberFilter{
				Ids: []uuid.UUID{parsedTeamMemberID},
				Active: types.OptionalParam[bool]{
					Value: true, IsSet: true,
				},
			})
			if err != nil {
				_ = appHttp.WriteErr(w, r, http.StatusInternalServerError, "error getting team member", err)
				return
			}
			if teamMember == nil {
				_ = appHttp.WriteErr(w, r, http.StatusNotFound, "team member not found", nil)
				return
			}
			teamInfo, err := app.Team().FindTeamInfo(rawCtx, teamMember.TeamID, userInfo.User.ID)
			if err != nil {
				_ = appHttp.WriteErr(w, r, http.StatusInternalServerError, "error getting team info", err)
				return
			}
			if teamInfo == nil {
				_ = appHttp.WriteErr(w, r, http.StatusNotFound, "team info not found", nil)
				return
			}

			ctxx := contextstore.SetContextTeamInfo(rawCtx, teamInfo)
			r = r.WithContext(ctxx)
			next.ServeHTTP(w, r)
		})
	}
}

// TeamCanDelete middleware checks whether the team can be deleted, i.e. it has no valid subscriptions
func TeamCanDelete(app core.App) HttpMiddelwareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawCtx := r.Context()
			teamInfo := contextstore.GetContextTeamInfo(rawCtx)
			if teamInfo == nil {
				_ = appHttp.WriteErr(w, r, http.StatusForbidden, "missing team membership", nil)
				return
			}
			can, err := app.Checker().TeamCannotHaveValidSubscription(rawCtx, teamInfo.Team.ID)
			if err != nil {
				_ = appHttp.WriteErr(w, r, http.StatusInternalServerError, "error checking if team can be deleted", err)
				return
			}
			if !can {
				_ = appHttp.WriteErr(w, r, http.StatusForbidden, "you are not allowed to delete this team", nil)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// TeamInfoFromTask captures the {task-id} path param to query its teamId, and along with the user info, queries the teamInfo membership.
// If the user has membership in the team of the task, that teamInfo is added to the context, and the request is forwarded to the next middleware,
// otherwise it returns an error
func TeamInfoFromTask(app core.App) HttpMiddelwareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawCtx := r.Context()
			userInfo := contextstore.GetContextUserInfo(rawCtx)
			if userInfo == nil {
				_ = appHttp.WriteErr(w, r, http.StatusUnauthorized, "unauthorized at middleware", nil)
				return
			}
			taskId := appHttp.GetParam(r, "task-id")
			if taskId == "" {
				_ = appHttp.WriteErr(w, r, http.StatusBadRequest, "task id is required", nil)
				return
			}
			parsedTaskId, err := uuid.Parse(taskId)
			if err != nil {
				_ = appHttp.WriteErr(w, r, http.StatusBadRequest, "error parsing task id", err)
				return
			}
			task, err := app.Adapter().Task().FindTaskByID(rawCtx, parsedTaskId)
			if err != nil {
				_ = appHttp.WriteErr(w, r, http.StatusInternalServerError, "error getting task", err)
				return
			}
			if task == nil {
				_ = appHttp.WriteErr(w, r, http.StatusNotFound, "task not found", nil)
				return
			}
			teamInfo, err := app.Team().FindTeamInfo(rawCtx, task.TeamID, userInfo.User.ID)
			if err != nil {
				_ = appHttp.WriteErr(w, r, http.StatusInternalServerError, "error getting team info", err)
				return
			}
			if teamInfo == nil {
				_ = appHttp.WriteErr(w, r, http.StatusUnauthorized, "you are not part of the task's team", nil)
				return
			}

			ctxx := contextstore.SetContextTeamInfo(rawCtx, teamInfo)
			r = r.WithContext(ctxx)
			next.ServeHTTP(w, r)
		})
	}
}

// TeamInfoFromTaskProject captures the {"task-project-id"} path param to query its teamId, and along with the user info, queries the teamInfo membership.
// If the user has membership in the team of the task project, that teamInfo is added to the context and the request is forwarded to the next middleware, otherwise it returns an error
func TeamInfoFromTaskProject(app core.App) HttpMiddelwareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawCtx := r.Context()
			userInfo := contextstore.GetContextUserInfo(rawCtx)
			if userInfo == nil {
				_ = appHttp.WriteErr(w, r, http.StatusUnauthorized, "unauthorized at middleware", nil)
				return
			}
			projectId := appHttp.GetParam(r, "task-project-id")
			if projectId == "" {
				next.ServeHTTP(w, r)
				return
			}
			parsedProjectID, err := uuid.Parse(projectId)
			if err != nil {
				_ = appHttp.WriteErr(w, r, http.StatusBadRequest, "error parsing project id", err)
				return
			}
			project, err := app.Adapter().Task().FindTaskProjectByID(rawCtx, parsedProjectID)
			if err != nil {
				_ = appHttp.WriteErr(w, r, http.StatusInternalServerError, "error getting project", err)
				return
			}
			if project == nil {
				_ = appHttp.WriteErr(w, r, http.StatusNotFound, "project not found", nil)
				return
			}
			teamInfo, err := app.Team().FindTeamInfo(rawCtx, project.TeamID, userInfo.User.ID)
			if err != nil {
				_ = appHttp.WriteErr(w, r, http.StatusInternalServerError, "error getting team info", err)
				return
			}
			if teamInfo == nil {
				_ = appHttp.WriteErr(w, r, http.StatusNotFound, "team not found", nil)
				return
			}

			ctxx := contextstore.SetContextTeamInfo(rawCtx, teamInfo)
			r = r.WithContext(ctxx)
			next.ServeHTTP(w, r)
		})
	}
}

// TeamInfoFromTeamSlug captures the {team-slug} path param, and along with the user info, queries the teamInfo.
// If the user has membership in the team of {team-slug}, that teamInfo is added to the context, otherwise it returns an error
func TeamInfoFromTeamSlug(app core.App) HttpMiddelwareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawCtx := r.Context()
			userInfo := contextstore.GetContextUserInfo(rawCtx)
			if userInfo == nil {
				_ = appHttp.WriteErr(w, r, http.StatusUnauthorized, "unauthorized at middleware", nil)
				return
			}
			teamSlug := appHttp.GetParam(r, "team-slug")
			if teamSlug == "" {
				_ = appHttp.WriteErr(w, r, http.StatusBadRequest, "team slug is required", nil)
				return
			}
			teamInfo, err := app.Team().FindTeamInfoBySlug(rawCtx, teamSlug, userInfo.User.ID)
			if err != nil {
				_ = appHttp.WriteErr(w, r, http.StatusInternalServerError, "error getting team info", err)
				return
			}
			if teamInfo == nil {
				_ = appHttp.WriteErr(w, r, http.StatusNotFound, "team not found", nil)
				return
			}

			ctxx := contextstore.SetContextTeamInfo(rawCtx, teamInfo)
			r = r.WithContext(ctxx)
			next.ServeHTTP(w, r)
		})
	}
}

// TeamInfoFromContext creates a [models.TeamInfoModel] from values found in the context and adds it to the context.
//
//   - it calls [contextstore.GetContextUserInfo] for the user
//   - it calls [contextstore.GetContextTeam] for the team
//   - if the team is found, it queries the team member using the team.id and user.id.
//   - if the team not found, it calls [contextstore.GetContextTeamMember] for the team member, and queries the user's team member using the teamMember.TeamID and user.ID.
func TeamInfoFromContext(app core.App) HttpMiddelwareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawCtx := r.Context()
			// get the user. if not found, return an error
			userInfo := contextstore.GetContextUserInfo(rawCtx)
			if userInfo == nil {
				_ = appHttp.WriteErr(w, r, http.StatusUnauthorized, "you are not logged in")
				return
			}

			// get the team.
			team := contextstore.GetContextTeam(rawCtx)
			// if found, get the team member from the team.id and user.id
			if team != nil {
				teamMember, err := app.Adapter().TeamMember().FindTeamMember(rawCtx, &stores.TeamMemberFilter{
					TeamIds: []uuid.UUID{team.ID},
					UserIds: []uuid.UUID{userInfo.User.ID},
					Active: types.OptionalParam[bool]{
						Value: true, IsSet: true,
					},
				})
				if err != nil {
					slog.ErrorContext(
						rawCtx,
						"TeamInfoFromContext: error getting team member",
						slog.Any("error", err),
						slog.String("team_id", team.ID.String()),
						slog.String("user_id", userInfo.User.ID.String()),
					)
					_ = appHttp.WriteErr(w, r, http.StatusInternalServerError, "error getting team info", err)
					return
				}
				// if not found, user is not a member of this team.
				// but this is not a security check, therefore we just move on without setting the team info.
				// the next middleware will check the team info and return an error
				if teamMember == nil {
					next.ServeHTTP(w, r)
					return
				}
				teamMember.User = &userInfo.User
				teamMember.Team = team
				teamInfo := &models.TeamInfoModel{
					Team:   *team,
					User:   userInfo.User,
					Member: *teamMember,
				}
				ctxx := contextstore.SetContextTeamInfo(rawCtx, teamInfo)
				r = r.WithContext(ctxx)
				next.ServeHTTP(w, r)
				return
			}
			// if team is not found in context, check for team member in context
			ctxTeamMember := contextstore.GetContextTeamMember(rawCtx)
			// if not found, move on.
			if ctxTeamMember == nil {
				next.ServeHTTP(w, r)
				return
			}
			// if found, get the team info using its member.team_id and user.id
			teamInfo, err := app.Team().FindTeamInfo(rawCtx, ctxTeamMember.TeamID, userInfo.User.ID)
			if err != nil {
				slog.ErrorContext(
					rawCtx,
					"TeamInfoFromContext: error getting team info",
					slog.Any("error", err),
					slog.String("team_id", ctxTeamMember.TeamID.String()),
					slog.String("user_id", userInfo.User.ID.String()),
				)
				_ = appHttp.WriteErr(w, r, http.StatusInternalServerError, "error getting team info", err)
				return
			}
			// if not found, user is not a member of this team of the member found in context. but this is not a security check, therefore we just move on.
			if teamInfo == nil {
				next.ServeHTTP(w, r)
				return
			}
			// if found, add the team info to the context
			ctxx := contextstore.SetContextTeamInfo(rawCtx, teamInfo)
			r = r.WithContext(ctxx)
			next.ServeHTTP(w, r)
		})
	}
}

// TeamInfoFromTeamIDParam captures the {team-id} path param, and along with the user info, queries the teamInfo.
// If the user has membership in the team of the task project, that teamInfo is added to the context, otherwise it returns an error
func TeamInfoFromTeamIDParam(app core.App) HttpMiddelwareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawCtx := r.Context()
			userInfo := contextstore.GetContextUserInfo(rawCtx)
			if userInfo == nil {
				_ = appHttp.WriteErr(w, r, http.StatusUnauthorized, "unauthorized at middleware", nil)
				return
			}
			teamId := appHttp.GetParam(r, "team-id")
			if teamId == "" {
				_ = appHttp.WriteErr(w, r, http.StatusBadRequest, "team id is required", nil)
				return
			}
			id, err := uuid.Parse(teamId)
			if err != nil {
				_ = appHttp.WriteErr(w, r, http.StatusBadRequest, "error parsing team id", err)
				return
			}
			teamInfo, err := app.Team().FindTeamInfo(rawCtx, id, userInfo.User.ID)
			if err != nil {
				_ = appHttp.WriteErr(w, r, http.StatusInternalServerError, "error getting team info", err)
				return
			}
			if teamInfo == nil {
				_ = appHttp.WriteErr(w, r, http.StatusNotFound, "team not found", nil)
				return
			}
			ctxx := contextstore.SetContextTeamInfo(rawCtx, teamInfo)
			r = r.WithContext(ctxx)
			next.ServeHTTP(w, r)
		})
	}
}

// RequireTeamMemberRolesMiddleware checks if the member has the required team member roles
func RequireTeamMemberRolesMiddleware(roles ...models.TeamMemberRole) HttpMiddelwareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawCtx := r.Context()
			if info := contextstore.GetContextTeamInfo(rawCtx); info != nil {
				if len(roles) == 0 {
					next.ServeHTTP(w, r)
					return
				}
				if slices.Contains(roles, info.Member.Role) {
					next.ServeHTTP(w, r)
					return
				}
				_ = appHttp.WriteErr(
					w,
					r,
					http.StatusForbidden,
					fmt.Sprintf("You do not have the required team member roles: %v", roles),
				)
			} else {
				_ = appHttp.WriteErr(
					w,
					r,
					http.StatusForbidden,
					fmt.Sprintf("You do not have the required team member roles: %v", roles),
				)
			}
		})
	}
}
