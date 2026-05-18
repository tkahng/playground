package middleware

import (
	"context"
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
)

// SseTicketAuth authenticates SSE connections via a short-lived ticket issued by
// POST /team-members/{id}/sse/ticket. It runs before TeamInfoFromContext and sets
// UserInfo in context so that TeamInfoFromContext can resolve the team membership
// normally. It is a no-op when the user is already authenticated via JWT.
func SseTicketAuth(app core.App) HTTPMiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawCtx := r.Context()

			// JWT auth already ran — skip ticket logic.
			if contextstore.GetContextUserInfo(rawCtx) != nil {
				next.ServeHTTP(w, r)
				return
			}

			t := appHttp.GetQuery(r, "ticket")
			if t == "" {
				next.ServeHTTP(w, r)
				return
			}

			userID, teamMemberID, ok := app.SseTickets().Validate(t)
			if !ok {
				appHttp.WriteErr(w, r, http.StatusUnauthorized, "invalid or expired SSE ticket")
				return
			}

			// The {team-member-id} path param must match the ticket's team member.
			if pathID := appHttp.GetParam(r, "team-member-id"); pathID != "" {
				parsed, err := uuid.Parse(pathID)
				if err != nil || parsed != teamMemberID {
					appHttp.WriteErr(w, r, http.StatusUnauthorized, "SSE ticket does not match requested team member")
					return
				}
			}

			user, err := app.Adapter().User().FindUserByID(rawCtx, userID)
			if err != nil || user == nil {
				appHttp.WriteErr(w, r, http.StatusUnauthorized, "user not found")
				return
			}
			userInfo, err := app.Adapter().User().GetUserInfo(rawCtx, user.Email)
			if err != nil || userInfo == nil {
				appHttp.WriteErr(w, r, http.StatusUnauthorized, "user info not found")
				return
			}

			r = r.WithContext(contextstore.SetContextUserInfo(rawCtx, userInfo))
			next.ServeHTTP(w, r)
		})
	}
}

// RequireTeamInfo checks if the request has team info
func RequireTeamInfo() HTTPMiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawCtx := r.Context()
			teamInfo := contextstore.GetContextTeamInfo(rawCtx)
			if teamInfo == nil {
				appHttp.WriteErr(w, r, http.StatusForbidden, "team info not found. you are not a member of the team related to this request")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireTeamMemberRolesMiddleware checks if the member has the required team member roles
func RequireTeamMemberRolesMiddleware(roles ...models.TeamMemberRole) HTTPMiddlewareFunc {
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
				appHttp.WriteErr(
					w,
					r,
					http.StatusForbidden,
					fmt.Sprintf("You do not have the required team member roles: %v", roles),
				)
			} else {
				appHttp.WriteErr(
					w,
					r,
					http.StatusForbidden,
					"You do not have the required team info",
				)
			}
		})
	}
}

func RequireTeamPermission(app core.App, permissionName string) HTTPMiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawCtx := r.Context()
			info := contextstore.GetContextTeamInfo(rawCtx)
			if info == nil {
				appHttp.WriteErr(
					w,
					r,
					http.StatusForbidden,
					"You do not have the required team info",
				)
				return
			}
			allowed, err := app.Adapter().Rbac().HasTeamRolePermission(rawCtx, info.Member.Role, permissionName)
			if err != nil {
				slog.ErrorContext(
					rawCtx,
					"RequireTeamPermission: error checking team permission",
					slog.Any("error", err),
					slog.String("role", string(info.Member.Role)),
					slog.String("permission", permissionName),
				)
				appHttp.WriteErr(w, r, http.StatusInternalServerError, "error checking team permission", err)
				return
			}
			if !allowed {
				appHttp.WriteErr(
					w,
					r,
					http.StatusForbidden,
					"Forbidden",
				)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireTeamMemberRolesMiddleware checks if the member has the required team member roles
func RequireTeamMemberBillingAccessMiddleware() HTTPMiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawCtx := r.Context()
			if info := contextstore.GetContextTeamInfo(rawCtx); info != nil {
				if info.Member.HasBillingAccess {
					next.ServeHTTP(w, r)
					return
				}
				appHttp.WriteErr(
					w,
					r,
					http.StatusForbidden,
					"You do not have the required billing access",
				)
			} else {
				appHttp.WriteErr(
					w,
					r,
					http.StatusForbidden,
					"You do not have the required team info",
				)
			}
		})
	}
}

type getTeamIDFunc func(ctx context.Context) (uuid.UUID, bool)

var getTeamIDFuncs []getTeamIDFunc = []getTeamIDFunc{
	func(ctx context.Context) (uuid.UUID, bool) {
		team := contextstore.GetContextTeam(ctx)
		if team != nil {
			return team.ID, true
		}
		return uuid.Nil, false
	},
	func(ctx context.Context) (uuid.UUID, bool) {
		teamMember := contextstore.GetContextTeamMember(ctx)
		if teamMember != nil {
			return teamMember.TeamID, true
		}
		return uuid.Nil, false
	},
	func(ctx context.Context) (uuid.UUID, bool) {
		task := contextstore.GetContextTask(ctx)
		if task != nil {
			return task.TeamID, true
		}
		return uuid.Nil, false
	},
	func(ctx context.Context) (uuid.UUID, bool) {
		project := contextstore.GetContextTaskProject(ctx)
		if project != nil {
			return project.TeamID, true
		}
		return uuid.Nil, false
	},
}

func GetTeamIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	for _, fn := range getTeamIDFuncs {
		if teamId, ok := fn(ctx); ok {
			return teamId, true
		}
	}
	return uuid.Nil, false
}

// TeamInfoFromContext creates a [models.TeamInfoModel] from values found in the context and adds it to the context.
//
//   - it calls [contextstore.GetContextUserInfo] for the user. if the user is not found, it moves on.
//   - it calls [contextstore.GetContextTeam] for the team
//   - if the team is found, it queries the team member using the team.id and user.id.
//   - if the team not found, it calls [contextstore.GetContextTeamMember] for the team member, and queries the user's team member using the teamMember.TeamID and user.ID.
//   - if the team member is not found, it moves on without setting the team info
func TeamInfoFromContext(app core.App) HTTPMiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawCtx := r.Context()
			// get the user. if not found, move on. could be public route.
			userInfo := contextstore.GetContextUserInfo(rawCtx)
			if userInfo == nil {
				next.ServeHTTP(w, r)
				return
			}

			// get team id from various context values
			teamId, ok := GetTeamIDFromContext(rawCtx)
			if !ok {
				next.ServeHTTP(w, r)
				return
			}
			// if found, get the team info using its member.team_id and user.id
			teamInfo, err := app.Team().FindTeamInfo(rawCtx, teamId, userInfo.User.ID)
			if err != nil {
				slog.ErrorContext(
					rawCtx,
					"TeamInfoFromContext: error getting team info",
					slog.Any("error", err),
					slog.String("team_id", teamId.String()),
					slog.String("user_id", userInfo.User.ID.String()),
				)
				appHttp.WriteErr(w, r, http.StatusInternalServerError, "error getting team info", err)
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

// TeamMemberFromParam captures the {team-member-id} path param, and if found, stores the teamMember in the context, otherwise it simply moves on
func TeamMemberFromParam(app core.App) HTTPMiddlewareFunc {
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
				appHttp.WriteErr(w, r, http.StatusBadRequest, "error parsing team member id. invalid UUID format", err)
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
				appHttp.WriteErr(w, r, http.StatusInternalServerError, "error getting team member")
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

// TeamFromParam captures the {team-id} path param, and if found, stores the team in the context, otherwise it simply moves on
func TeamFromParam(app core.App) HTTPMiddlewareFunc {
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
				appHttp.WriteErr(w, r, http.StatusBadRequest, "error parsing team id. invalid UUID format")
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
				appHttp.WriteErr(w, r, http.StatusInternalServerError, "error while querying team")
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

// TeamFromParamSlug captures the {team-slug} path param, and if found, stores the team in the context, otherwise it simply moves on
func TeamFromParamSlug(app core.App) HTTPMiddlewareFunc {
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
				appHttp.WriteErr(w, r, http.StatusInternalServerError, "error while querying team")
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

// MemberIDBelongsToUser middleware ensures that the user is the member with id {team-member-id}
func MemberIDBelongsToUser() HTTPMiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawCtx := r.Context()
			userInfo := contextstore.GetContextTeamInfo(rawCtx)
			if userInfo == nil {
				appHttp.WriteErr(w, r, http.StatusUnauthorized, "unauthorized at middleware", nil)
				return
			}
			teamMemberID := appHttp.GetParam(r, "team-member-id")
			if teamMemberID == "" {
				appHttp.WriteErr(w, r, http.StatusBadRequest, "team slug is required", nil)
				return
			}
			parsedTeamMemberID, err := uuid.Parse(teamMemberID)
			if err != nil {
				appHttp.WriteErr(w, r, http.StatusBadRequest, "error parsing team member id", err)
				return
			}
			if userInfo.Member.ID != parsedTeamMemberID {
				appHttp.WriteErr(w, r, http.StatusUnauthorized, "unauthorized at middleware", nil)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// TeamCanDelete middleware checks whether the team can be deleted, i.e. it has no valid subscriptions
func TeamCanDelete(app core.App) HTTPMiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawCtx := r.Context()
			teamInfo := contextstore.GetContextTeamInfo(rawCtx)
			if teamInfo == nil {
				appHttp.WriteErr(w, r, http.StatusForbidden, "missing team membership", nil)
				return
			}
			can, err := app.Checker().TeamCannotHaveValidSubscription(rawCtx, teamInfo.Team.ID)
			if err != nil {
				appHttp.WriteErr(w, r, http.StatusInternalServerError, "error checking if team can be deleted", err)
				return
			}
			if !can {
				appHttp.WriteErr(w, r, http.StatusForbidden, "you are not allowed to delete this team", nil)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// TeamInfoFromTask captures the {task-id} path param to query its teamId, and along with the user info, queries the teamInfo membership.
// If the user has membership in the team of the task, that teamInfo is added to the context, and the request is forwarded to the next middleware,
// otherwise it returns an error
func TeamInfoFromTask(app core.App) HTTPMiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawCtx := r.Context()
			userInfo := contextstore.GetContextUserInfo(rawCtx)
			if userInfo == nil {
				appHttp.WriteErr(w, r, http.StatusUnauthorized, "unauthorized at middleware", nil)
				return
			}
			taskId := appHttp.GetParam(r, "task-id")
			if taskId == "" {
				appHttp.WriteErr(w, r, http.StatusBadRequest, "task id is required", nil)
				return
			}
			parsedTaskId, err := uuid.Parse(taskId)
			if err != nil {
				appHttp.WriteErr(w, r, http.StatusBadRequest, "error parsing task id", err)
				return
			}
			task, err := app.Adapter().Task().FindTaskByID(rawCtx, parsedTaskId)
			if err != nil {
				appHttp.WriteErr(w, r, http.StatusInternalServerError, "error getting task", err)
				return
			}
			if task == nil {
				appHttp.WriteErr(w, r, http.StatusNotFound, "task not found", nil)
				return
			}
			teamInfo, err := app.Team().FindTeamInfo(rawCtx, task.TeamID, userInfo.User.ID)
			if err != nil {
				appHttp.WriteErr(w, r, http.StatusInternalServerError, "error getting team info", err)
				return
			}
			if teamInfo == nil {
				appHttp.WriteErr(w, r, http.StatusUnauthorized, "you are not part of the task's team", nil)
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
func TeamInfoFromTaskProject(app core.App) HTTPMiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawCtx := r.Context()
			userInfo := contextstore.GetContextUserInfo(rawCtx)
			if userInfo == nil {
				appHttp.WriteErr(w, r, http.StatusUnauthorized, "unauthorized at middleware", nil)
				return
			}
			projectId := appHttp.GetParam(r, "task-project-id")
			if projectId == "" {
				next.ServeHTTP(w, r)
				return
			}
			parsedProjectID, err := uuid.Parse(projectId)
			if err != nil {
				appHttp.WriteErr(w, r, http.StatusBadRequest, "error parsing project id", err)
				return
			}
			project, err := app.Adapter().Task().FindTaskProjectByID(rawCtx, parsedProjectID)
			if err != nil {
				appHttp.WriteErr(w, r, http.StatusInternalServerError, "error getting project", err)
				return
			}
			if project == nil {
				appHttp.WriteErr(w, r, http.StatusNotFound, "project not found", nil)
				return
			}
			teamInfo, err := app.Team().FindTeamInfo(rawCtx, project.TeamID, userInfo.User.ID)
			if err != nil {
				appHttp.WriteErr(w, r, http.StatusInternalServerError, "error getting team info", err)
				return
			}
			if teamInfo == nil {
				appHttp.WriteErr(w, r, http.StatusNotFound, "team not found", nil)
				return
			}

			ctxx := contextstore.SetContextTeamInfo(rawCtx, teamInfo)
			r = r.WithContext(ctxx)
			next.ServeHTTP(w, r)
		})
	}
}

// TeamInfoFromTeamIDParam captures the {team-id} path param, and along with the user info, queries the teamInfo.
// If the user has membership in the team of the task project, that teamInfo is added to the context, otherwise it returns an error
func TeamInfoFromTeamIDParam(app core.App) HTTPMiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawCtx := r.Context()
			userInfo := contextstore.GetContextUserInfo(rawCtx)
			if userInfo == nil {
				appHttp.WriteErr(w, r, http.StatusUnauthorized, "unauthorized at middleware", nil)
				return
			}
			teamId := appHttp.GetParam(r, "team-id")
			if teamId == "" {
				appHttp.WriteErr(w, r, http.StatusBadRequest, "team id is required", nil)
				return
			}
			id, err := uuid.Parse(teamId)
			if err != nil {
				appHttp.WriteErr(w, r, http.StatusBadRequest, "error parsing team id", err)
				return
			}
			teamInfo, err := app.Team().FindTeamInfo(rawCtx, id, userInfo.User.ID)
			if err != nil {
				appHttp.WriteErr(w, r, http.StatusInternalServerError, "error getting team info", err)
				return
			}
			if teamInfo == nil {
				appHttp.WriteErr(w, r, http.StatusNotFound, "team not found", nil)
				return
			}
			ctxx := contextstore.SetContextTeamInfo(rawCtx, teamInfo)
			r = r.WithContext(ctxx)
			next.ServeHTTP(w, r)
		})
	}
}
