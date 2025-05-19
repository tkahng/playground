package apis

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/contextstore"
	"github.com/tkahng/authgo/internal/core"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/shared"
)

// func HumaRecoverer(next func(huma.Context)) func(ctx huma.Context, next func(huma.Context)) {
// 	return func(ctx huma.Context, next func(huma.Context)) {
// 		defer func() {
// 			if rvr := recover(); rvr != nil {
// 				if rvr == http.ErrAbortHandler {
// 					// we don't recover http.ErrAbortHandler so the response
// 					// to the client is aborted, this should not be logged
// 					panic(rvr)
// 				}

// 				logEntry := GetLogEntry(r)
// 				if logEntry != nil {
// 					logEntry.Panic(rvr, debug.Stack())
// 				} else {
// 					PrintPrettyStack(rvr)
// 				}

// 				if r.Header.Get("Connection") != "Upgrade" {
// 					w.WriteHeader(http.StatusInternalServerError)
// 				}
// 			}
// 		}()
// 		next(ctx)
// 	}
// }

// TokenFromCookie tries to retreive the token string from a cookie named
// "jwt".
func HumaTokenFromCookie(ctx huma.Context) string {
	cookie, err := huma.ReadCookie(ctx, "access_token")
	//  ctx.Header()
	if err != nil {
		return ""
	}
	return cookie.Value
}

// TokenFromHeader tries to retreive the token string from the
// "Authorization" reqeust header: "Authorization: BEARER T".
func HumaTokenFromHeader(ctx huma.Context) string {
	// Get token from authorization header.
	bearer := ctx.Header("Authorization")
	if len(bearer) > 7 && strings.ToUpper(bearer[0:6]) == "BEARER" {
		return bearer[7:]
	}
	return ""
}

var HumaTokenFuncs = []func(huma.Context) string{
	HumaTokenFromHeader,
	HumaTokenFromCookie,
}

func CheckTaskOwnerMiddleware(api huma.API, app core.App) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		db := app.Db()
		rawCtx := ctx.Context()
		taskId := ctx.Param("task-id")
		if taskId == "" {
			next(ctx)
			return
		}
		id, err := uuid.Parse(taskId)
		if err != nil {
			huma.WriteErr(api, ctx, http.StatusBadRequest, "invalid task id", err)
			return
		}
		task, err := queries.FindTaskByID(rawCtx, db, id)
		if err != nil {
			huma.WriteErr(api, ctx, http.StatusInternalServerError, "error getting task", err)
			return
		}
		if task == nil {
			huma.WriteErr(api, ctx, http.StatusNotFound, "task not found at middleware")
			return
		}
		userInfo := contextstore.GetContextUserInfo(rawCtx)
		if userInfo == nil {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "unauthorized at middleware")
			return
		}
		teamInfo := contextstore.GetContextTeamInfo(rawCtx)
		if teamInfo == nil {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "unauthorized at middleware")
			return
		}
		if task.CreatedBy != teamInfo.Member.ID {
			if slices.Contains(userInfo.Permissions, "superuser") {
				next(ctx)
				return
			}
			huma.WriteErr(api, ctx, http.StatusForbidden, "task user id does not match user id")
			return
		}
		next(ctx)
	}
}

func CheckPermissionsMiddleware(api huma.API, permissions ...string) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {

		if claims := contextstore.GetContextUserInfo(ctx.Context()); claims != nil {
			if len(permissions) == 0 {
				next(ctx)
				return
			}
			for _, permission := range claims.Permissions {
				if slices.Contains(permissions, permission) {
					next(ctx)
					return
				}
			}
		}
		huma.WriteErr(api, ctx, http.StatusForbidden, fmt.Sprintf("You do not have the required permissions: %v", permissions))
	}
}

func RequireAuthMiddleware(api huma.API) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		if ctx.Operation().Security == nil {
			next(ctx)
			return
		}
		var anyOfNeededScopes []string
		isAuthorizationRequired := false

		for _, opScheme := range ctx.Operation().Security {
			var ok bool
			if anyOfNeededScopes, ok = opScheme[shared.BearerAuthSecurityKey]; ok {
				isAuthorizationRequired = true
				break
			}
		}

		if !isAuthorizationRequired {
			next(ctx)
			return
		}

		c := contextstore.GetContextUserInfo(ctx.Context())
		if c != nil {
			if len(anyOfNeededScopes) == 0 {
				next(ctx)
				return
			}
			for _, scope := range c.Permissions {
				if slices.Contains(anyOfNeededScopes, string(scope)) {
					next(ctx)
					return
				}
			}
		}
		huma.WriteErr(api, ctx, http.StatusForbidden, "Forbidden")
	}
}

func TeamInfoFromParamMiddleware(api huma.API, app core.App) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		rawCtx := ctx.Context()
		userInfo := contextstore.GetContextUserInfo(rawCtx)
		if userInfo == nil {
			next(ctx)
			return
		}
		teamId := ctx.Param("team-id")
		if teamId == "" {
			next(ctx)
			return
		}
		id, err := uuid.Parse(teamId)
		if err != nil {
			huma.WriteErr(api, ctx, http.StatusBadRequest, "invalid team id", err)
			return
		}
		teamInfo, err := app.Team().FindTeamInfo(rawCtx, id, userInfo.User.ID)
		if err != nil {
			huma.WriteErr(api, ctx, http.StatusInternalServerError, "error getting team info", err)
			return
		}
		if teamInfo == nil {
			huma.WriteErr(api, ctx, http.StatusNotFound, "team not found at middleware")
			return
		}
		ctxx := contextstore.SetContextTeamInfo(rawCtx, teamInfo)
		ctx = huma.WithContext(ctx, ctxx)
		next(ctx)
	}
}

func RequireTeamMemberRolesMiddleware(api huma.API, roles ...models.TeamMemberRole) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		rawctx := ctx.Context()
		if info := contextstore.GetContextTeamInfo(rawctx); info != nil {
			if slices.Contains(roles, info.Member.Role) {
				next(ctx)
				return
			}
		} else {
			huma.WriteErr(
				api,
				ctx,
				http.StatusForbidden,
				fmt.Sprintf("You do not have the required team member roles: %v", roles),
			)
		}
	}
}

func LatestTeamMiddleware(api huma.API, app core.App) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		rawCtx := ctx.Context()
		user := contextstore.GetContextUserInfo(rawCtx)
		if user == nil {
			next(ctx)
			return
		}
		info, err := app.Team().FindLatestTeamInfo(rawCtx, user.User.ID)
		if err != nil {
			slog.ErrorContext(
				rawCtx,
				"error getting team info",
				slog.String("user_id", user.User.ID.String()),
				slog.Any("error", err),
			)
			next(ctx)
			return
		}
		if info == nil {
			next(ctx)
			return
		}
		ctxx := contextstore.SetContextTeamInfo(rawCtx, info)
		ctx = huma.WithContext(ctx, ctxx)
		next(ctx)
	}
}

// Auth creates a middleware that will authorize requests based on the required scopes for the operation.
func AuthMiddleware(api huma.API, app core.App) func(ctx huma.Context, next func(huma.Context)) {

	return func(ctx huma.Context, next func(huma.Context)) {
		ctxx := ctx.Context()
		action := app.Auth()
		// check if already has user claims
		if claims := contextstore.GetContextUserInfo(ctxx); claims != nil {
			log.Println("already has user claims")
			next(ctx)
			return
		}
		var token string
		for _, f := range HumaTokenFuncs {
			token = f(ctx)
			if len(token) > 0 {
				break
			}
		}
		if len(token) == 0 {
			next(ctx)
			return
		}
		user, err := action.HandleAccessToken(ctxx, token)
		if err != nil {
			log.Println(err)
			next(ctx)
			return
		}
		ctxx = contextstore.SetContextUserInfo(ctxx, user)
		ctx = huma.WithContext(ctx, ctxx)

		next(ctx)
	}
}
