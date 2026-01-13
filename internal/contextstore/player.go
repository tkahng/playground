package contextstore

import (
	"context"

	"github.com/tkahng/playground/internal/models"
)

const (
	contextKeyCurrentPlayer contextKey = "current_player"
)

func SetContextCurrentPlayer(ctx context.Context, info *models.Player) context.Context {
	return context.WithValue(ctx, contextKeyCurrentPlayer, info)
}
func GetContextCurrentPlayer(ctx context.Context) *models.Player {
	if val, ok := ctx.Value(contextKeyCurrentPlayer).(*models.Player); ok {
		return val
	} else {
		return nil
	}
}
