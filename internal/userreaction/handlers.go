package userreaction

import (
	"context"
	"log/slog"

	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/mapper"
	"github.com/tkahng/playground/internal/tools/sse"
)

type UserReactionHandler interface {
	OnUserReactionCreated(ctx context.Context, event *UserReactionCreated) error
}
type UserReactionEventHandler struct {
	logger     *slog.Logger
	store      stores.UserReactionStore
	sseManager sse.Manager
	// rateLimiter limiter.Limiter
}

func (u *UserReactionEventHandler) OnUserReactionCreated(ctx context.Context, event *UserReactionCreated) error {
	// if !u.rateLimiter.Allow() {
	// 	return nil
	// }
	stats := new(UserReactionStats)
	stats.LastCreated = FromModelUserReaction(event.UserReaction)
	recent, err := u.store.CountByCountry(ctx, &stores.UserReactionFilter{
		PaginatedInput: stores.PaginatedInput{
			PerPage: 5,
		},
	})
	if err != nil {
		u.logger.Error("failed to get recent user reactions", slog.Any("error", err))
	}
	stats.TopFiveCountries = mapper.Map(recent, func(r *stores.ReactionByCountry) ReactionByCountry {
		return ReactionByCountry{
			Country:        r.Country,
			TotalReactions: r.TotalReactions,
		}
	})
	count, err := u.store.CountUserReactions(ctx, nil)
	if err != nil {
		u.logger.Error("failed to get recent user reactions", slog.Any("error", err))
	}
	stats.TotalReactions = count
	err = u.sseManager.Send(sse.UserReactionsChannel, LatestUserReactionStatsSseEvent{
		UserReactionStats: stats,
	})
	if err != nil {
		u.logger.Error("failed to send sse", slog.Any("error", err))
	}
	return nil
}

var _ UserReactionHandler = (*UserReactionEventHandler)(nil)

func NewUserReactionEventHandler(logger *slog.Logger, store stores.UserReactionStore, sseManager sse.Manager) UserReactionHandler {
	return &UserReactionEventHandler{
		logger:     logger,
		store:      store,
		sseManager: sseManager,
		// rateLimiter: limiter.NewRateLimiter(
		// 	10,
		// 	10*time.Second,
		// ),
	}
}
