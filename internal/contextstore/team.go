package contextstore

import (
	"context"

	"github.com/tkahng/authgo/internal/models"
)

const (
	ContextKeySelectedTeam ContextKey = "selected_team"
)

func SetContextSelectedTeam(ctx context.Context, team *models.Team) context.Context {
	return context.WithValue(ctx, ContextKeySelectedTeam, team)
}
func GetContextSelectedTeam(ctx context.Context) *models.Team {
	if team, ok := ctx.Value(ContextKeySelectedTeam).(*models.Team); ok {
		return team
	} else {
		return nil
	}
}
