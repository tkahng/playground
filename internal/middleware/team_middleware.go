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
)

func TeamInfoFromTeamMemberID(app core.App) HttpMiddelwareFunc {
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

func TeamCanDelete(app core.App) HttpMiddelwareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			slog.Info("starting TeamCanDelete middleware")
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
				_ = appHttp.WriteErr(w, r, http.StatusNotFound, "team not found", nil)
				return
			}

			ctxx := contextstore.SetContextTeamInfo(rawCtx, teamInfo)
			r = r.WithContext(ctxx)
			next.ServeHTTP(w, r)
		})
	}
}

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

func TeamInfoFromParam(app core.App) HttpMiddelwareFunc {
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
			slog.Info("found team info")
			ctxx := contextstore.SetContextTeamInfo(rawCtx, teamInfo)
			r = r.WithContext(ctxx)
			next.ServeHTTP(w, r)
		})
	}
}

func RequireTeamMemberRolesMiddleware(roles ...models.TeamMemberRole) HttpMiddelwareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			slog.Info("starting RequireTeamMemberRolesMiddleware")
			rawCtx := r.Context()
			if info := contextstore.GetContextTeamInfo(rawCtx); info != nil {
				slog.Info("found team info in context")
				if len(roles) == 0 {
					next.ServeHTTP(w, r)
					return
				}
				if slices.Contains(roles, info.Member.Role) {
					slog.Info("user has required team member role")
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
