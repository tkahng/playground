package middleware

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/contextstore"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/stores"
	appHttp "github.com/tkahng/playground/internal/tools/http"
)

func RequireCurrentPlayerMiddelware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			player := contextstore.GetContextCurrentPlayer(ctx)
			if player == nil {
				appHttp.WriteErr(w, r, http.StatusUnauthorized, "unauthorized. player not found", nil)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func SetCurrentPlayerMiddleware(app core.App) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			userInfo := contextstore.GetContextUserInfo(ctx)
			if userInfo == nil {
				appHttp.WriteErr(w, r, http.StatusUnauthorized, "unauthorized. user info not found", nil)
				return
			}
			player, err := app.Adapter().Gaming().FindPlayer(ctx, &stores.PlayersFilter{
				UserIds: []uuid.UUID{userInfo.User.ID},
				Emails:  []string{userInfo.User.Email},
			})
			if err != nil {
				appHttp.WriteErr(w, r, http.StatusInternalServerError, "error getting player", err)
				return
			}
			if player == nil {
				next.ServeHTTP(w, r)
				return
			}
			newCtx := contextstore.SetContextCurrentPlayer(ctx, player)
			r = r.WithContext(newCtx)
			next.ServeHTTP(w, r)
		})
	}
}
