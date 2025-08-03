package middleware

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/contextstore"
	"github.com/tkahng/playground/internal/core"

	apphttp "github.com/tkahng/playground/internal/tools/http"
)

func CheckTaskOwnerMiddleware(app core.App) HttpMiddelwareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawCtx := r.Context()
			taskId := apphttp.GetParam(r, "task-id")
			if taskId == "" {
				next.ServeHTTP(w, r)
				return
			}
			id, err := uuid.Parse(taskId)
			if err != nil {
				_ = apphttp.WriteErr(w, r, http.StatusBadRequest, "invalid task id", err)
				return
			}
			task, err := app.Adapter().Task().FindTaskByID(rawCtx, id)
			if err != nil {
				_ = apphttp.WriteErr(w, r, http.StatusInternalServerError, "error getting task", err)
				return
			}
			if task == nil {
				_ = apphttp.WriteErr(w, r, http.StatusNotFound, "task not found at middleware")
				return
			}
			userInfo := contextstore.GetContextUserInfo(rawCtx)
			if userInfo == nil {
				_ = apphttp.WriteErr(w, r, http.StatusUnauthorized, "unauthorized at middleware")
				return
			}
			teamInfo := contextstore.GetContextTeamInfo(rawCtx)
			if teamInfo == nil {
				_ = apphttp.WriteErr(w, r, http.StatusUnauthorized, "unauthorized at middleware")
				return
			}
			// if task.CreatedByMemberID != teamInfo.Member.ID {
			// 	if slices.Contains(userInfo.Permissions, "superuser") {
			// 		next(ctx)
			// 		return
			// 	}
			// 	_ = apphttp.WriteErr(w, r, http.StatusForbidden, "task user id does not match user id")
			// 	return
			// }
			next.ServeHTTP(w, r)
		})
	}

}
