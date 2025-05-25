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

func EmailVerifiedMiddleware(api huma.API) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		rawCtx := ctx.Context()
		userInfo := contextstore.GetContextUserInfo(rawCtx)
		if userInfo == nil {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "unauthorized", nil)
			return
		}
		if userInfo.User.EmailVerifiedAt == nil {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "email not verified", nil)
			return
		}
		next(ctx)
	}
}
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

func TeamInfoFromTaskMiddleware(api huma.API, app core.App) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		rawCtx := ctx.Context()
		userInfo := contextstore.GetContextUserInfo(rawCtx)
		if userInfo == nil {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "unauthorized at middleware", nil)
			return
		}
		taskId := ctx.Param("task-id")
		if taskId == "" {
			next(ctx)
			return
		}
		parsedTaskId, err := uuid.Parse(taskId)
		if err != nil {
			huma.WriteErr(api, ctx, http.StatusBadRequest, "error parsing task id", err)
			return
		}
		task, err := app.Task().Store().FindTaskByID(rawCtx, parsedTaskId)
		if err != nil {
			huma.WriteErr(api, ctx, http.StatusInternalServerError, "error getting task", err)
			return
		}
		if task == nil {
			huma.WriteErr(api, ctx, http.StatusNotFound, "task not found", nil)
			return
		}
		teamInfo, err := app.Team().FindTeamInfo(rawCtx, task.TeamID, userInfo.User.ID)
		if err != nil {
			huma.WriteErr(api, ctx, http.StatusInternalServerError, "error getting team info", err)
			return
		}
		if teamInfo == nil {
			huma.WriteErr(api, ctx, http.StatusNotFound, "team not found", nil)
			return
		}
		ctxx := contextstore.SetContextTeamInfo(rawCtx, teamInfo)
		ctx = huma.WithContext(ctx, ctxx)
		next(ctx)
	}
}

func TeamInfoFromTaskProjectMiddleware(api huma.API, app core.App) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		rawCtx := ctx.Context()
		userInfo := contextstore.GetContextUserInfo(rawCtx)
		if userInfo == nil {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "unauthorized at middleware", nil)
			return
		}
		projectId := ctx.Param("task-project-id")
		if projectId == "" {
			next(ctx)
			return
		}
		parsedProjectID, err := uuid.Parse(projectId)
		if err != nil {
			huma.WriteErr(api, ctx, http.StatusBadRequest, "error parsing project id", err)
			return
		}
		project, err := app.Task().Store().FindTaskProjectByID(rawCtx, parsedProjectID)
		if err != nil {
			huma.WriteErr(api, ctx, http.StatusInternalServerError, "error getting project", err)
			return
		}
		if project == nil {
			huma.WriteErr(api, ctx, http.StatusNotFound, "project not found", nil)
			return
		}
		teamInfo, err := app.Team().FindTeamInfo(rawCtx, project.TeamID, userInfo.User.ID)
		if err != nil {
			huma.WriteErr(api, ctx, http.StatusInternalServerError, "error getting team info", err)
			return
		}
		if teamInfo == nil {
			huma.WriteErr(api, ctx, http.StatusNotFound, "team not found", nil)
			return
		}
		ctxx := contextstore.SetContextTeamInfo(rawCtx, teamInfo)
		ctx = huma.WithContext(ctx, ctxx)
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
