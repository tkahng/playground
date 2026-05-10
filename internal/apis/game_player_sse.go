package apis

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	humasse "github.com/danielgtaylor/huma/v2/sse"
	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/contextstore"
	"github.com/tkahng/playground/internal/middleware"
	"github.com/tkahng/playground/internal/middleware/humamiddleware"
	"github.com/tkahng/playground/internal/notification"
	"github.com/tkahng/playground/internal/shared"
	appHttp "github.com/tkahng/playground/internal/tools/http"
	"github.com/tkahng/playground/internal/tools/sse"
)

func bindPlayerSseApi(api *Api) {
	api.IssuePlayerSSETicketBind(api.Api())
	api.PlayerSseEventsBind(api.Api())
}

type PlayerSseInput struct {
	PlayerID string `path:"player-id"`
	Ticket   string `query:"ticket"`
}

type IssuePlayerSSETicketResponse struct {
	Body struct {
		Ticket string `json:"ticket"`
	}
}

func (api *Api) IssuePlayerSSETicketBind(humapi huma.API) {
	huma.Register(
		humapi,
		huma.Operation{
			OperationID: "issue-player-sse-ticket",
			Method:      http.MethodPost,
			Path:        "/players/sse/ticket",
			Summary:     "Issue player SSE ticket",
			Description: "Issue a short-lived (60 s) ticket for player-level SSE notifications.",
			Tags:        []string{"Games", "Player"},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.RequireCurrentPlayerMiddelware(),
			),
			Errors: []int{http.StatusUnauthorized},
		},
		func(ctx context.Context, _ *struct{}) (*IssuePlayerSSETicketResponse, error) {
			userInfo := contextstore.GetContextUserInfo(ctx)
			if userInfo == nil {
				return nil, huma.Error401Unauthorized("unauthorized")
			}
			currentPlayer := contextstore.GetContextCurrentPlayer(ctx)
			if currentPlayer == nil {
				return nil, huma.Error401Unauthorized("no player found")
			}
			t := api.App().SseTickets().Issue(userInfo.User.ID, currentPlayer.ID)
			resp := &IssuePlayerSSETicketResponse{}
			resp.Body.Ticket = t
			return resp, nil
		},
	)
}

func playerSSETicketMiddleware(api *Api) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawCtx := r.Context()
			t := r.URL.Query().Get("ticket")
			if t == "" {
				appHttp.WriteErr(w, r, http.StatusUnauthorized, "missing SSE ticket")
				return
			}
			userID, playerID, ok := api.App().SseTickets().Validate(t)
			if !ok {
				appHttp.WriteErr(w, r, http.StatusUnauthorized, "invalid or expired SSE ticket")
				return
			}
			pathPlayerID := appHttp.GetParam(r, "player-id")
			if pathPlayerID != "" {
				parsed, err := uuid.Parse(pathPlayerID)
				if err != nil || parsed != playerID {
					appHttp.WriteErr(w, r, http.StatusUnauthorized, "SSE ticket does not match player")
					return
				}
			}
			user, err := api.App().Adapter().User().FindUserByID(rawCtx, userID)
			if err != nil || user == nil {
				appHttp.WriteErr(w, r, http.StatusUnauthorized, "user not found")
				return
			}
			userInfo, err := api.App().Adapter().User().GetUserInfo(rawCtx, user.Email)
			if err != nil || userInfo == nil {
				appHttp.WriteErr(w, r, http.StatusUnauthorized, "user info not found")
				return
			}
			r = r.WithContext(contextstore.SetContextUserInfo(rawCtx, userInfo))
			next.ServeHTTP(w, r)
		})
	}
}

func (api *Api) PlayerSseEventsBind(humapi huma.API) {
	handler := sse.ServeSSE(
		func(ctx context.Context, f func(any) error, input *PlayerSseInput) sse.Client {
			playerID := input.PlayerID
			return sse.NewClient(sse.PlayerChannel(playerID), f, slog.Default(), func() any {
				return &PingMessage{Message: "ping"}
			})
		},
		func(ctx context.Context, cf context.CancelFunc, c sse.Client) {
			api.app.SseManager().RegisterClient(ctx, cf, c)
		},
		func(c sse.Client) {
			api.app.SseManager().UnregisterClient(c)
		},
		30*time.Second,
	)
	humasse.Register(
		humapi,
		huma.Operation{
			OperationID: "player-sse-notifications",
			Method:      http.MethodGet,
			Path:        "/players/{player-id}/sse",
			Summary:     "Player SSE notifications",
			Description: "Subscribe to real-time player-level notifications (friend requests, etc.).",
			Tags:        []string{"Games", "Player"},
			Middlewares: humamiddleware.HumaChiMiddlewares(
				playerSSETicketMiddleware(api),
			),
			Errors: []int{http.StatusUnauthorized},
		},
		map[string]any{
			"friend_request": &notification.NotificationPayload[notification.FriendRequestNotificationData]{},
			"ping":           &PingMessage{},
		},
		handler,
	)
}
