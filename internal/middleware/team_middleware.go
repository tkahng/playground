package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"slices"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/contextstore"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/models"
)

func TeamInfoFromTeamMemberID(api huma.API, app core.App) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		rawCtx := ctx.Context()
		userInfo := contextstore.GetContextUserInfo(rawCtx)
		if userInfo == nil {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "unauthorized at middleware", nil)
			return
		}
		teamMemberID := ctx.Param("team-member-id")
		if teamMemberID == "" {
			huma.WriteErr(api, ctx, http.StatusBadRequest, "team slug is required", nil)
			return
		}
		parsedTeamMemberID, err := uuid.Parse(teamMemberID)
		if err != nil {
			huma.WriteErr(api, ctx, http.StatusBadRequest, "error parsing team member id", err)
			return
		}
		teamInfo, err := app.Team().FindTeamInfoByMemberID(rawCtx, parsedTeamMemberID)
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
func TeamCanDelete(api huma.API, app core.App) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		slog.Info("starting TeamCanDelete middleware")
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

func TeamInfoFromTask(api huma.API, app core.App) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		rawCtx := ctx.Context()
		userInfo := contextstore.GetContextUserInfo(rawCtx)
		if userInfo == nil {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "unauthorized at middleware", nil)
			return
		}
		taskId := ctx.Param("task-id")
		if taskId == "" {
			huma.WriteErr(api, ctx, http.StatusBadRequest, "task id is required", nil)
			return
		}
		parsedTaskId, err := uuid.Parse(taskId)
		if err != nil {
			huma.WriteErr(api, ctx, http.StatusBadRequest, "error parsing task id", err)
			return
		}
		task, err := app.Adapter().Task().FindTaskByID(rawCtx, parsedTaskId)
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

func TeamInfoFromTaskProject(api huma.API, app core.App) func(ctx huma.Context, next func(huma.Context)) {
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
		project, err := app.Adapter().Task().FindTaskProjectByID(rawCtx, parsedProjectID)
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

func TeamInfoFromTeamSlug(api huma.API, app core.App) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		rawCtx := ctx.Context()
		userInfo := contextstore.GetContextUserInfo(rawCtx)
		if userInfo == nil {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "unauthorized at middleware", nil)
			return
		}
		teamSlug := ctx.Param("team-slug")
		if teamSlug == "" {
			huma.WriteErr(api, ctx, http.StatusBadRequest, "team slug is required", nil)
			return
		}
		teamInfo, err := app.Team().FindTeamInfoBySlug(rawCtx, teamSlug, userInfo.User.ID)
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

func TeamInfoFromParam(api huma.API, app core.App) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		rawCtx := ctx.Context()
		userInfo := contextstore.GetContextUserInfo(rawCtx)
		if userInfo == nil {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "unauthorized at middleware", nil)
			return
		}
		teamId := ctx.Param("team-id")
		if teamId == "" {
			huma.WriteErr(api, ctx, http.StatusBadRequest, "team id is required", nil)
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
			huma.WriteErr(api, ctx, http.StatusNotFound, "team not found", nil)
			return
		}
		slog.Info("found team info")
		ctxx := contextstore.SetContextTeamInfo(rawCtx, teamInfo)
		ctx = huma.WithContext(ctx, ctxx)
		next(ctx)
	}
}

func RequireTeamMemberRolesMiddleware(api huma.API, roles ...models.TeamMemberRole) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		slog.Info("starting RequireTeamMemberRolesMiddleware")
		rawctx := ctx.Context()
		if info := contextstore.GetContextTeamInfo(rawctx); info != nil {
			slog.Info("found team info in context")
			if len(roles) == 0 {
				next(ctx)
				return
			}
			if slices.Contains(roles, info.Member.Role) {
				slog.Info("user has required team member role")
				next(ctx)
				return
			}
			huma.WriteErr(
				api,
				ctx,
				http.StatusForbidden,
				fmt.Sprintf("You do not have the required team member role: %v", info.Member.Role),
			)
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
		userInfo := contextstore.GetContextUserInfo(rawCtx)
		if userInfo == nil {
			next(ctx)
			return
		}
		info, err := app.Team().FindLatestTeamInfo(rawCtx, userInfo.User.ID)
		if err != nil {
			slog.ErrorContext(
				rawCtx,
				"error getting team info",
				slog.String("user_id", userInfo.User.ID.String()),
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
