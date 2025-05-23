package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"slices"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/contextstore"
	"github.com/tkahng/authgo/internal/core"
	"github.com/tkahng/authgo/internal/models"
)

func TeamCanDeleteMiddleware(api huma.API, app core.App) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		rawCtx := ctx.Context()
		teamInfo := contextstore.GetContextTeamInfo(rawCtx)
		if teamInfo == nil {
			huma.WriteErr(api, ctx, http.StatusForbidden, "missing team membership", nil)
			return
		}
		can, err := app.Checker().TeamCannotHaveValidSubscription(rawCtx, teamInfo.Team.ID)
		if err != nil {
			huma.WriteErr(api, ctx, http.StatusInternalServerError, "error checking if team can be deleted", err)
			return
		}
		if !can {
			huma.WriteErr(api, ctx, http.StatusForbidden, "you are not allowed to delete this team", nil)
			return
		}
		next(ctx)
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
			huma.WriteErr(api, ctx, http.StatusBadRequest, "error parsing team id", err)
			return
		}
		teamInfo, err := app.Team().FindTeamInfo(rawCtx, id, userInfo.User.ID)
		if err != nil {
			huma.WriteErr(api, ctx, http.StatusInternalServerError, "error getting team info", err)
			return
		}
		if teamInfo == nil {
			// huma.WriteErr(api, ctx, http.StatusNotFound, "team not found. you might not be a member of this team or it might not exist", nil)
			next(ctx)
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
			if len(roles) == 0 {
				next(ctx)
				return
			}
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
