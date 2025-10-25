package middleware

import (
	"net/http"

	"github.com/tkahng/playground/internal/contextstore"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/models"
	appHttp "github.com/tkahng/playground/internal/tools/http"
)

func SelectOrCreateOwnerCustomerFromTeam(app core.App) HttpMiddelwareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawCtx := r.Context()

			teamInfo := contextstore.GetContextTeamInfo(rawCtx)
			if teamInfo == nil {
				_ = appHttp.WriteErr(w, r, http.StatusForbidden, "no team info found")
				return
			}
			customer, err := app.Payment().FindCustomerByTeamId(rawCtx, teamInfo.Team.ID)
			if err != nil {
				_ = appHttp.WriteErr(w, r, http.StatusInternalServerError, "error getting customer", err)
				return
			}
			if customer == nil {
				userInfo := contextstore.GetContextUserInfo(rawCtx)
				if userInfo == nil {
					_ = appHttp.WriteErr(w, r, http.StatusForbidden, "no user info found")
					return
				}
				customer, err = app.Payment().CreateTeamCustomer(rawCtx, &teamInfo.Team, &models.User{
					ID:    userInfo.User.ID,
					Name:  userInfo.User.Name,
					Email: userInfo.User.Email,
				})
				if err != nil {
					_ = appHttp.WriteErr(w, r, http.StatusInternalServerError, "error creating customer", err)
					return
				}
				if customer == nil {
					_ = appHttp.WriteErr(w, r, http.StatusInternalServerError, "error creating customer")
					return
				}
				newCtx := contextstore.SetContextCurrentCustomer(rawCtx, customer)
				r = r.WithContext(newCtx)
				next.ServeHTTP(w, r)
				return
			}
			newCtx := contextstore.SetContextCurrentCustomer(rawCtx, customer)
			r = r.WithContext(newCtx)
			next.ServeHTTP(w, r)
		})
	}
}
func SelectCustomerFromTeam(app core.App) HttpMiddelwareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawCtx := r.Context()

			teamInfo := contextstore.GetContextTeamInfo(rawCtx)
			if teamInfo == nil {
				_ = appHttp.WriteErr(w, r, http.StatusForbidden, "no team info found")
				return
			}
			customer, err := app.Payment().FindCustomerByTeamId(rawCtx, teamInfo.Team.ID)
			if err != nil {
				_ = appHttp.WriteErr(w, r, http.StatusInternalServerError, "error getting customer", err)
				return
			}
			if customer == nil {
				_ = appHttp.WriteErr(w, r, http.StatusUnauthorized, "customer not found")
				return
			}
			newCtx := contextstore.SetContextCurrentCustomer(rawCtx, customer)
			r = r.WithContext(newCtx)
			next.ServeHTTP(w, r)
		})
	}
}

func SelectCustomerFromUser(app core.App) HttpMiddelwareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawCtx := r.Context()
			userInfo := contextstore.GetContextUserInfo(rawCtx)
			if userInfo == nil {
				_ = appHttp.WriteErr(w, r, http.StatusForbidden, "no user info found")
				return
			}
			customer, err := app.Payment().FindCustomerByUserId(rawCtx, userInfo.User.ID)
			if err != nil {
				_ = appHttp.WriteErr(w, r, http.StatusInternalServerError, "error getting customer", err)
				return
			}
			if customer == nil {
				_ = appHttp.WriteErr(w, r, http.StatusUnauthorized, "customer not found")
				return
			}
			newCtx := contextstore.SetContextCurrentCustomer(rawCtx, customer)
			r = r.WithContext(newCtx)
			next.ServeHTTP(w, r)
		})
	}

}
