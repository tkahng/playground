package apis

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/contextstore"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/database/repository"
	"github.com/tkahng/playground/internal/middleware"
	"github.com/tkahng/playground/internal/middleware/humamiddleware"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/notification"
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/mapper"
)

func bindFriendApi(humaApi huma.API, app core.App) {
	bindListFriendsApi(humaApi, app)
	bindListFriendRequestsApi(humaApi, app)
	bindSendFriendRequestApi(humaApi, app)
	bindAcceptFriendRequestApi(humaApi, app)
	bindDeclineFriendRequestApi(humaApi, app)
	bindRemoveFriendApi(humaApi, app)
	bindBlockPlayerApi(humaApi, app)
	bindUnblockPlayerApi(humaApi, app)
	bindGetFriendshipApi(humaApi, app)
}

// populateFriendshipPlayers batch-fetches all unique players referenced by the
// given friendships in a single DB query and assigns them in-place.
func populateFriendshipPlayers(ctx context.Context, adapter stores.StorageAdapterInterface, friendships []*models.Friendship) error {
	if len(friendships) == 0 {
		return nil
	}
	seen := make(map[uuid.UUID]struct{}, len(friendships)*2)
	ids := make([]uuid.UUID, 0, len(friendships)*2)
	for _, f := range friendships {
		if _, ok := seen[f.RequestingPlayerID]; !ok {
			seen[f.RequestingPlayerID] = struct{}{}
			ids = append(ids, f.RequestingPlayerID)
		}
		if _, ok := seen[f.InvitedPlayerID]; !ok {
			seen[f.InvitedPlayerID] = struct{}{}
			ids = append(ids, f.InvitedPlayerID)
		}
	}
	players, err := adapter.Gaming().FindPlayers(ctx, &stores.PlayersFilter{
		Ids:            ids,
		PaginatedInput: repository.PaginatedInput{Page: 0, PerPage: int64(len(ids))},
	})
	if err != nil {
		return err
	}
	byID := make(map[uuid.UUID]*models.Player, len(players))
	for _, p := range players {
		byID[p.ID] = p
	}
	for _, f := range friendships {
		f.RequestingPlayer = byID[f.RequestingPlayerID]
		f.InvitedPlayer = byID[f.InvitedPlayerID]
	}
	return nil
}

// findFriendshipBetween finds the friendship record between two specific players (in either direction)
// using a single DB query.
func findFriendshipBetween(ctx context.Context, adapter stores.StorageAdapterInterface, playerA, playerB uuid.UUID) (*models.Friendship, error) {
	pair := [2]uuid.UUID{playerA, playerB}
	return adapter.Gaming().FindFriendship(ctx, &stores.FriendshipFilter{
		PlayerPair: &pair,
	})
}

func bindListFriendsApi(api huma.API, app core.App) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "list-friends",
			Method:      http.MethodGet,
			Path:        "/players/friends",
			Summary:     "List friends",
			Description: "List accepted friends for the current player.",
			Tags:        []string{"Games", "Friends"},
			Errors:      []int{http.StatusUnauthorized},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.RequireCurrentPlayerMiddelware(),
			),
		},
		func(ctx context.Context, input *PaginatedInput) (*ApiPaginatedOutput[*Friendship], error) {
			currentPlayer := contextstore.GetContextCurrentPlayer(ctx)
			if currentPlayer == nil {
				return nil, huma.Error401Unauthorized("no player found")
			}
			filter := &stores.FriendshipFilter{
				PaginatedInput:               repository.PaginatedInput{Page: input.Page, PerPage: input.PerPage},
				RequestingOrInvitedPlayerIds: []uuid.UUID{currentPlayer.ID},
				Statuses:                     []models.FriendshipStatus{models.FriendshipStatusAccepted},
			}
			friendships, err := app.Adapter().Gaming().FindFriendships(ctx, filter)
			if err != nil {
				return nil, err
			}
			count, err := app.Adapter().Gaming().CountFriendships(ctx, filter)
			if err != nil {
				return nil, err
			}
			
			if err := populateFriendshipPlayers(ctx, app.Adapter(), friendships); err != nil {
				return nil, err
			}
			return &ApiPaginatedOutput[*Friendship]{
				Body: ApiPaginatedResponse[*Friendship]{
					Data: mapper.Map(friendships, ToApiFriendship),
					Meta: ApiGenerateMeta(input, count),
				},
			}, nil
		},
	)
}

func bindListFriendRequestsApi(api huma.API, app core.App) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "list-friend-requests",
			Method:      http.MethodGet,
			Path:        "/players/friends/requests",
			Summary:     "List friend requests",
			Description: "List pending friend requests (incoming and outgoing) for the current player.",
			Tags:        []string{"Games", "Friends"},
			Errors:      []int{http.StatusUnauthorized},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.RequireCurrentPlayerMiddelware(),
			),
		},
		func(ctx context.Context, input *PaginatedInput) (*ApiPaginatedOutput[*Friendship], error) {
			currentPlayer := contextstore.GetContextCurrentPlayer(ctx)
			if currentPlayer == nil {
				return nil, huma.Error401Unauthorized("no player found")
			}
			cutoff := time.Now().UTC().AddDate(0, 0, -30)
			filter := &stores.FriendshipFilter{
				PaginatedInput:               repository.PaginatedInput{Page: input.Page, PerPage: input.PerPage},
				RequestingOrInvitedPlayerIds: []uuid.UUID{currentPlayer.ID},
				Statuses:                     []models.FriendshipStatus{models.FriendshipStatusPending},
				CreatedAfter:                 &cutoff,
			}
			friendships, err := app.Adapter().Gaming().FindFriendships(ctx, filter)
			if err != nil {
				return nil, err
			}
			count, err := app.Adapter().Gaming().CountFriendships(ctx, filter)
			if err != nil {
				return nil, err
			}
			
			if err := populateFriendshipPlayers(ctx, app.Adapter(), friendships); err != nil {
				return nil, err
			}
			return &ApiPaginatedOutput[*Friendship]{
				Body: ApiPaginatedResponse[*Friendship]{
					Data: mapper.Map(friendships, ToApiFriendship),
					Meta: ApiGenerateMeta(input, count),
				},
			}, nil
		},
	)
}

type SendFriendRequestBody struct {
	InvitedPlayerID uuid.UUID `json:"invited_player_id" required:"true" format:"uuid"`
}

func bindSendFriendRequestApi(api huma.API, app core.App) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "send-friend-request",
			Method:      http.MethodPost,
			Path:        "/players/friends/requests",
			Summary:     "Send friend request",
			Description: "Send a friend request to another player.",
			Tags:        []string{"Games", "Friends"},
			Errors:      []int{http.StatusUnauthorized, http.StatusBadRequest, http.StatusConflict},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.RequireCurrentPlayerMiddelware(),
			),
		},
		func(ctx context.Context, input *struct {
			Body SendFriendRequestBody
		}) (*ApiSingleOutput[*Friendship], error) {
			currentPlayer := contextstore.GetContextCurrentPlayer(ctx)
			if currentPlayer == nil {
				return nil, huma.Error401Unauthorized("no player found")
			}
			if currentPlayer.ID == input.Body.InvitedPlayerID {
				return nil, huma.Error400BadRequest("cannot send friend request to yourself")
			}
			target, err := app.Adapter().Gaming().FindPlayer(ctx, &stores.PlayersFilter{
				Ids: []uuid.UUID{input.Body.InvitedPlayerID},
			})
			if err != nil {
				return nil, err
			}
			if target == nil {
				return nil, huma.Error400BadRequest("player not found")
			}
			existing, err := findFriendshipBetween(ctx, app.Adapter(), currentPlayer.ID, input.Body.InvitedPlayerID)
			if err != nil {
				return nil, err
			}
			if existing != nil {
				switch existing.Status {
				case models.FriendshipStatusAccepted:
					return nil, huma.Error409Conflict("already friends")
				case models.FriendshipStatusPending:
					return nil, huma.Error409Conflict("friend request already pending")
				case models.FriendshipStatusBlocked:
					return nil, huma.Error409Conflict("cannot send friend request to blocked player")
				case models.FriendshipStatusDeclined:
					_, err = app.Adapter().Gaming().DeleteFriendships(ctx, &stores.FriendshipFilter{
						Ids: []uuid.UUID{existing.ID},
					})
					if err != nil {
						return nil, err
					}
				}
			}
			friendship, err := app.Adapter().Gaming().CreateFriendship(ctx, &models.Friendship{
				RequestingPlayerID: currentPlayer.ID,
				InvitedPlayerID:    input.Body.InvitedPlayerID,
				Status:             models.FriendshipStatusPending,
			})
			if err != nil {
				return nil, err
			}
			friendship.RequestingPlayer = currentPlayer
			friendship.InvitedPlayer = target
			// Notify the invited player — persist + SSE (best-effort; failure is non-fatal).
			payload := notification.NewNotificationPayload(
				"New friend request",
				currentPlayer.Email+" wants to be your friend",
				notification.FriendRequestNotificationData{
					RequestingPlayerID: currentPlayer.ID,
					RequestingEmail:    currentPlayer.Email,
					FriendshipID:       friendship.ID,
				},
			)
			if err := app.PlayerNotificationPublisher().Notify(ctx, target.ID, notification.FriendRequestNotificationData{}.Kind(), payload); err != nil {
				slog.WarnContext(ctx, "friend request notify failed", "error", err)
			}
			return &ApiSingleOutput[*Friendship]{
				Body: ApiSingleResponse[*Friendship]{
					Data: ToApiFriendship(friendship),
				},
			}, nil
		},
	)
}

func bindAcceptFriendRequestApi(api huma.API, app core.App) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "accept-friend-request",
			Method:      http.MethodPost,
			Path:        "/players/friends/requests/{request-id}/accept",
			Summary:     "Accept friend request",
			Description: "Accept an incoming friend request.",
			Tags:        []string{"Games", "Friends"},
			Errors:      []int{http.StatusUnauthorized, http.StatusNotFound, http.StatusForbidden},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.RequireCurrentPlayerMiddelware(),
			),
		},
		func(ctx context.Context, input *struct {
			RequestID uuid.UUID `path:"request-id" required:"true" format:"uuid"`
		}) (*ApiSingleOutput[*Friendship], error) {
			currentPlayer := contextstore.GetContextCurrentPlayer(ctx)
			if currentPlayer == nil {
				return nil, huma.Error401Unauthorized("no player found")
			}
			friendship, err := app.Adapter().Gaming().FindFriendship(ctx, &stores.FriendshipFilter{
				Ids: []uuid.UUID{input.RequestID},
			})
			if err != nil {
				return nil, err
			}
			if friendship == nil {
				return nil, huma.Error404NotFound("friend request not found")
			}
			if friendship.InvitedPlayerID != currentPlayer.ID {
				return nil, huma.Error403Forbidden("not authorized to accept this request")
			}
			if friendship.Status != models.FriendshipStatusPending {
				return nil, huma.Error400BadRequest("request is not pending")
			}
			now := time.Now().UTC()
			friendship.Status = models.FriendshipStatusAccepted
			friendship.RespondedAt = &now
			updated, err := app.Adapter().Gaming().UpdateFriendship(ctx, friendship)
			if err != nil {
				return nil, err
			}
			
			if err := populateFriendshipPlayers(ctx, app.Adapter(), []*models.Friendship{updated}); err != nil {
				return nil, err
			}
			return &ApiSingleOutput[*Friendship]{
				Body: ApiSingleResponse[*Friendship]{
					Data: ToApiFriendship(updated),
				},
			}, nil
		},
	)
}

func bindDeclineFriendRequestApi(api huma.API, app core.App) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "decline-friend-request",
			Method:      http.MethodPost,
			Path:        "/players/friends/requests/{request-id}/decline",
			Summary:     "Decline friend request",
			Description: "Decline an incoming friend request.",
			Tags:        []string{"Games", "Friends"},
			Errors:      []int{http.StatusUnauthorized, http.StatusNotFound, http.StatusForbidden},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.RequireCurrentPlayerMiddelware(),
			),
		},
		func(ctx context.Context, input *struct {
			RequestID uuid.UUID `path:"request-id" required:"true" format:"uuid"`
		}) (*ApiSingleOutput[*Friendship], error) {
			currentPlayer := contextstore.GetContextCurrentPlayer(ctx)
			if currentPlayer == nil {
				return nil, huma.Error401Unauthorized("no player found")
			}
			friendship, err := app.Adapter().Gaming().FindFriendship(ctx, &stores.FriendshipFilter{
				Ids: []uuid.UUID{input.RequestID},
			})
			if err != nil {
				return nil, err
			}
			if friendship == nil {
				return nil, huma.Error404NotFound("friend request not found")
			}
			if friendship.InvitedPlayerID != currentPlayer.ID {
				return nil, huma.Error403Forbidden("not authorized to decline this request")
			}
			if friendship.Status != models.FriendshipStatusPending {
				return nil, huma.Error400BadRequest("request is not pending")
			}
			now := time.Now().UTC()
			friendship.Status = models.FriendshipStatusDeclined
			friendship.RespondedAt = &now
			updated, err := app.Adapter().Gaming().UpdateFriendship(ctx, friendship)
			if err != nil {
				return nil, err
			}
			
			if err := populateFriendshipPlayers(ctx, app.Adapter(), []*models.Friendship{updated}); err != nil {
				return nil, err
			}
			return &ApiSingleOutput[*Friendship]{
				Body: ApiSingleResponse[*Friendship]{
					Data: ToApiFriendship(updated),
				},
			}, nil
		},
	)
}

func bindRemoveFriendApi(api huma.API, app core.App) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "remove-friend",
			Method:      http.MethodDelete,
			Path:        "/players/friends/{friendship-id}",
			Summary:     "Remove friend",
			Description: "Remove a friend from the friend list.",
			Tags:        []string{"Games", "Friends"},
			Errors:      []int{http.StatusUnauthorized, http.StatusNotFound, http.StatusForbidden},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.RequireCurrentPlayerMiddelware(),
			),
		},
		func(ctx context.Context, input *struct {
			FriendshipID uuid.UUID `path:"friendship-id" required:"true" format:"uuid"`
		}) (*struct{}, error) {
			currentPlayer := contextstore.GetContextCurrentPlayer(ctx)
			if currentPlayer == nil {
				return nil, huma.Error401Unauthorized("no player found")
			}
			friendship, err := app.Adapter().Gaming().FindFriendship(ctx, &stores.FriendshipFilter{
				Ids: []uuid.UUID{input.FriendshipID},
			})
			if err != nil {
				return nil, err
			}
			if friendship == nil {
				return nil, huma.Error404NotFound("friendship not found")
			}
			if friendship.RequestingPlayerID != currentPlayer.ID && friendship.InvitedPlayerID != currentPlayer.ID {
				return nil, huma.Error403Forbidden("not authorized")
			}
			_, err = app.Adapter().Gaming().DeleteFriendships(ctx, &stores.FriendshipFilter{
				Ids: []uuid.UUID{input.FriendshipID},
			})
			if err != nil {
				return nil, err
			}
			return nil, nil
		},
	)
}

type BlockPlayerBody struct {
	PlayerID uuid.UUID `json:"player_id" required:"true" format:"uuid"`
}

func bindBlockPlayerApi(api huma.API, app core.App) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "block-player",
			Method:      http.MethodPost,
			Path:        "/players/block",
			Summary:     "Block player",
			Description: "Block a player. Removes any existing friendship and prevents future matching.",
			Tags:        []string{"Games", "Friends"},
			Errors:      []int{http.StatusUnauthorized, http.StatusBadRequest},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.RequireCurrentPlayerMiddelware(),
			),
		},
		func(ctx context.Context, input *struct {
			Body BlockPlayerBody
		}) (*ApiSingleOutput[*Friendship], error) {
			currentPlayer := contextstore.GetContextCurrentPlayer(ctx)
			if currentPlayer == nil {
				return nil, huma.Error401Unauthorized("no player found")
			}
			if currentPlayer.ID == input.Body.PlayerID {
				return nil, huma.Error400BadRequest("cannot block yourself")
			}
			target, err := app.Adapter().Gaming().FindPlayer(ctx, &stores.PlayersFilter{
				Ids: []uuid.UUID{input.Body.PlayerID},
			})
			if err != nil {
				return nil, err
			}
			if target == nil {
				return nil, huma.Error400BadRequest("player not found")
			}
			// Delete any existing friendship between the two players in either direction
			existing, err := findFriendshipBetween(ctx, app.Adapter(), currentPlayer.ID, input.Body.PlayerID)
			if err != nil {
				return nil, err
			}
			if existing != nil {
				_, err = app.Adapter().Gaming().DeleteFriendships(ctx, &stores.FriendshipFilter{
					Ids: []uuid.UUID{existing.ID},
				})
				if err != nil {
					return nil, err
				}
			}
			now := time.Now().UTC()
			blocked, err := app.Adapter().Gaming().CreateFriendship(ctx, &models.Friendship{
				RequestingPlayerID: currentPlayer.ID,
				InvitedPlayerID:    input.Body.PlayerID,
				Status:             models.FriendshipStatusBlocked,
				RespondedAt:        &now,
			})
			if err != nil {
				return nil, err
			}
			blocked.RequestingPlayer = currentPlayer
			blocked.InvitedPlayer = target
			return &ApiSingleOutput[*Friendship]{
				Body: ApiSingleResponse[*Friendship]{
					Data: ToApiFriendship(blocked),
				},
			}, nil
		},
	)
}

func bindUnblockPlayerApi(api huma.API, app core.App) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "unblock-player",
			Method:      http.MethodDelete,
			Path:        "/players/block/{player-id}",
			Summary:     "Unblock player",
			Description: "Unblock a previously blocked player.",
			Tags:        []string{"Games", "Friends"},
			Errors:      []int{http.StatusUnauthorized},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.RequireCurrentPlayerMiddelware(),
			),
		},
		func(ctx context.Context, input *struct {
			PlayerID uuid.UUID `path:"player-id" required:"true" format:"uuid"`
		}) (*struct{}, error) {
			currentPlayer := contextstore.GetContextCurrentPlayer(ctx)
			if currentPlayer == nil {
				return nil, huma.Error401Unauthorized("no player found")
			}
			_, err := app.Adapter().Gaming().DeleteFriendships(ctx, &stores.FriendshipFilter{
				RequestingPlayerIds: []uuid.UUID{currentPlayer.ID},
				InvitedPlayerIds:    []uuid.UUID{input.PlayerID},
				Statuses:            []models.FriendshipStatus{models.FriendshipStatusBlocked},
			})
			if err != nil {
				return nil, err
			}
			return nil, nil
		},
	)
}

func bindGetFriendshipApi(api huma.API, app core.App) {
	huma.Register(
		api,
		huma.Operation{
			OperationID: "get-friendship",
			Method:      http.MethodGet,
			Path:        "/players/{player-id}/friendship",
			Summary:     "Get friendship status",
			Description: "Get the friendship status between the current player and another player.",
			Tags:        []string{"Games", "Friends"},
			Errors:      []int{http.StatusUnauthorized},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.RequireCurrentPlayerMiddelware(),
			),
		},
		func(ctx context.Context, input *struct {
			PlayerID uuid.UUID `path:"player-id" required:"true" format:"uuid"`
		}) (*ApiSingleOutput[*Friendship], error) {
			currentPlayer := contextstore.GetContextCurrentPlayer(ctx)
			if currentPlayer == nil {
				return nil, huma.Error401Unauthorized("no player found")
			}
			friendship, err := findFriendshipBetween(ctx, app.Adapter(), currentPlayer.ID, input.PlayerID)
			if err != nil {
				return nil, err
			}
			if friendship != nil {
				
				if err := populateFriendshipPlayers(ctx, app.Adapter(), []*models.Friendship{friendship}); err != nil {
					return nil, err
				}
			}
			return &ApiSingleOutput[*Friendship]{
				Body: ApiSingleResponse[*Friendship]{
					Data: ToApiFriendship(friendship),
				},
			}, nil
		},
	)
}
