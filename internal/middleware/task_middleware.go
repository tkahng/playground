package middleware

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/contextstore"
	"github.com/tkahng/playground/internal/core"

	apphttp "github.com/tkahng/playground/internal/tools/http"
)

func TaskFromParam(app core.App) HttpMiddelwareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// context
			// get task-id
			key := "task-id"
			taskId := apphttp.GetParam(r, key)
			// if task-id is empty then move on
			if taskId == "" {
				next.ServeHTTP(w, r)
				return
			}
			// parse task-id to uuid
			parsedTaskID, err := uuid.Parse(taskId)
			if err != nil {
				slog.ErrorContext(
					r.Context(),
					"error while parsing taskID in TaskFromParam middleware",
					slog.Any("error", err),
					slog.String(key, taskId),
				)
				_ = apphttp.WriteErr(w, r, http.StatusBadRequest, "error parsing task id. invalid UUID format")
				return
			}
			rawCtx := r.Context()
			// query task from task-id.
			// do not filter by active, we will filter it later
			task, err := app.Adapter().Task().FindTaskByID(rawCtx, parsedTaskID)
			if err != nil {
				slog.ErrorContext(
					rawCtx,
					"error while querying task in TaskFromParam middleware",
					slog.Any("error", err),
					slog.String(key, taskId),
				)
				_ = apphttp.WriteErr(w, r, http.StatusInternalServerError, "error while querying task")
				return
			}
			// if task id was not nil and valid uuid,
			// it is reasonable to assume that the task exists,
			// assuming the uuid was provided within our system.
			// however this is not a security check, so we will log it and move on.
			if task == nil {
				slog.WarnContext(
					rawCtx,
					"no task with given task-id found",
					slog.String(key, taskId),
				)
				next.ServeHTTP(w, r)
				return
			}
			// add team to context
			newCtx := contextstore.SetContextTask(rawCtx, task)
			r = r.WithContext(newCtx)
			next.ServeHTTP(w, r)
		})
	}
}

func TaskProjectFromParam(app core.App) HttpMiddelwareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// context
			// get task-project-id
			key := "task-project-id"
			projectId := apphttp.GetParam(r, key)
			// if task-project-id is empty then move on
			if projectId == "" {
				next.ServeHTTP(w, r)
				return
			}
			// parse task-project-id to uuid
			parsedProjectID, err := uuid.Parse(projectId)
			if err != nil {
				slog.ErrorContext(
					r.Context(),
					"error while parsing taskID in TaskProjectFromParam middleware",
					slog.Any("error", err),
					slog.String(key, projectId),
				)
				_ = apphttp.WriteErr(w, r, http.StatusBadRequest, "error parsing task project id. invalid UUID format")
				return
			}
			rawCtx := r.Context()
			// query task project from task-project-id.
			// do not filter by active, we will filter it later
			task, err := app.Adapter().Task().FindTaskProjectByID(rawCtx, parsedProjectID)
			if err != nil {
				slog.ErrorContext(
					rawCtx,
					"error while querying task in TaskFromParam middleware",
					slog.Any("error", err),
					slog.String(key, projectId),
				)
				_ = apphttp.WriteErr(w, r, http.StatusInternalServerError, "error while querying task")
				return
			}
			// if task project id was not nil and valid uuid,
			// it is reasonable to assume that the task exists,
			// assuming the uuid was provided within our system.
			// however this is not a security check, so we will log it and move on.
			if task == nil {
				slog.WarnContext(
					rawCtx,
					"no task with given task-project-id found",
					slog.String(key, projectId),
				)
				next.ServeHTTP(w, r)
				return
			}
			// add project to context
			newCtx := contextstore.SetContextTaskProject(rawCtx, task)
			r = r.WithContext(newCtx)
			next.ServeHTTP(w, r)
		})
	}
}

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
