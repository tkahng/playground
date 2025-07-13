package middleware

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/playground/internal/contextstore"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/models"
)

func SelectOrCreateOwnerCustomerFromTeam(api huma.API, app core.App) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		rawCtx := ctx.Context()

		teamInfo := contextstore.GetContextTeamInfo(rawCtx)
		if teamInfo == nil {
			huma.WriteErr(api, ctx, http.StatusForbidden, "no team info found")
			return
		}
		customer, err := app.Payment().FindCustomerByTeam(rawCtx, teamInfo.Team.ID)
		if err != nil {
			huma.WriteErr(api, ctx, http.StatusInternalServerError, "error getting customer", err)
			return
		}
		if customer == nil {
			userInfo := contextstore.GetContextUserInfo(rawCtx)
			if userInfo == nil {
				huma.WriteErr(api, ctx, http.StatusForbidden, "no user info found")
				return
			}
			customer, err = app.Payment().CreateTeamCustomer(rawCtx, &teamInfo.Team, &models.User{
				ID:    userInfo.User.ID,
				Name:  userInfo.User.Name,
				Email: userInfo.User.Email,
			})
			if err != nil {
				huma.WriteErr(api, ctx, http.StatusInternalServerError, "error creating customer", err)
				return
			}
			if customer == nil {
				huma.WriteErr(api, ctx, http.StatusInternalServerError, "error creating customer")
				return
			}
			newCtx := contextstore.SetContextCurrentCustomer(rawCtx, customer)
			ctx = huma.WithContext(ctx, newCtx)
			next(ctx)
			return
		}
		newCtx := contextstore.SetContextCurrentCustomer(rawCtx, customer)
		ctx = huma.WithContext(ctx, newCtx)
		next(ctx)
	}
}
func SelectCustomerFromTeam(api huma.API, app core.App) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		rawCtx := ctx.Context()

		teamInfo := contextstore.GetContextTeamInfo(rawCtx)
		if teamInfo == nil {
			huma.WriteErr(api, ctx, http.StatusForbidden, "no team info found")
			return
		}
		// if teamInfo.Member.Role != models.TeamMemberRoleOwner {
		// 	huma.WriteErr(api, ctx, http.StatusForbidden, "not a team owner")
		// 	return
		// }
		customer, err := app.Payment().FindCustomerByTeam(rawCtx, teamInfo.Team.ID)
		if err != nil {
			huma.WriteErr(api, ctx, http.StatusInternalServerError, "error getting customer", err)
			return
		}
		if customer == nil {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "customer not found")
			return
			// userInfo := contextstore.GetContextUserInfo(rawCtx)
			// if userInfo == nil {
			// 	huma.WriteErr(api, ctx, http.StatusForbidden, "no user info found")
			// 	return
			// }
			// customer, err = app.Payment().CreateTeamCustomer(rawCtx, &teamInfo.Team, &models.User{
			// 	ID:    userInfo.User.ID,
			// 	Name:  userInfo.User.Name,
			// 	Email: userInfo.User.Email,
			// })
			// if err != nil {
			// 	huma.WriteErr(api, ctx, http.StatusInternalServerError, "error creating customer", err)
			// 	return
			// }
			// if customer == nil {
			// 	huma.WriteErr(api, ctx, http.StatusInternalServerError, "error creating customer")
			// 	return
			// }
			// newCtx := contextstore.SetContextCurrentCustomer(rawCtx, customer)
			// ctx = huma.WithContext(ctx, newCtx)
			// next(ctx)
			// return
		}
		newCtx := contextstore.SetContextCurrentCustomer(rawCtx, customer)
		ctx = huma.WithContext(ctx, newCtx)
		next(ctx)
	}
}

func SelectCustomerFromUser(api huma.API, app core.App) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		rawCtx := ctx.Context()
		userInfo := contextstore.GetContextUserInfo(rawCtx)
		if userInfo == nil {
			huma.WriteErr(api, ctx, http.StatusForbidden, "no user info found")
			return
		}
		customer, err := app.Payment().FindCustomerByUser(rawCtx, userInfo.User.ID)
		if err != nil {
			huma.WriteErr(api, ctx, http.StatusInternalServerError, "error getting customer", err)
			return
		}
		if customer == nil {
			huma.WriteErr(api, ctx, http.StatusUnauthorized, "customer not found")
			return
			// customer, err = app.Payment().CreateUserCustomer(rawCtx, &models.User{
			// 	ID:    userInfo.User.ID,
			// 	Name:  userInfo.User.Name,
			// 	Email: userInfo.User.Email,
			// })
			// if err != nil {
			// 	huma.WriteErr(api, ctx, http.StatusInternalServerError, "error creating customer", err)
			// 	return
			// }
			// if customer == nil {
			// 	huma.WriteErr(api, ctx, http.StatusInternalServerError, "error creating customer")
			// 	return
			// }
			// newCtx := contextstore.SetContextCurrentCustomer(rawCtx, customer)
			// ctx = huma.WithContext(ctx, newCtx)
			// next(ctx)
			// return
		}
		newCtx := contextstore.SetContextCurrentCustomer(rawCtx, customer)
		ctx = huma.WithContext(ctx, newCtx)
		next(ctx)
	}
}
