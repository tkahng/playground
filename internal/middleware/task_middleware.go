package middleware

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/contextstore"
	"github.com/tkahng/authgo/internal/core"
)

func CheckTaskOwnerMiddleware(api huma.API, app core.App) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
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
		task, err := app.Task().Store().FindTaskByID(rawCtx, id)
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
		// if task.CreatedByMemberID != teamInfo.Member.ID {
		// 	if slices.Contains(userInfo.Permissions, "superuser") {
		// 		next(ctx)
		// 		return
		// 	}
		// 	huma.WriteErr(api, ctx, http.StatusForbidden, "task user id does not match user id")
		// 	return
		// }
		next(ctx)
	}
}
