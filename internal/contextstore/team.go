package contextstore

import (
	"context"

	"github.com/tkahng/playground/internal/models"
)

const (
	contextKeyTeam contextKey = "team"
)

func SetContextTeam(ctx context.Context, info *models.Team) context.Context {
	return context.WithValue(ctx, contextKeyTeam, info)
}
func GetContextTeam(ctx context.Context) *models.Team {
	if team, ok := ctx.Value(contextKeyTeam).(*models.Team); ok {
		return team
	} else {
		return nil
	}
}
