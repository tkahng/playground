package contextstore

import (
	"context"

	"github.com/tkahng/authgo/internal/shared"
)

const (
	ContextKeyTeamInfo ContextKey = "team_info"
)

func SetContextTeamInfo(ctx context.Context, info *shared.TeamInfo) context.Context {
	return context.WithValue(ctx, ContextKeySelectedTeam, info)
}
func GetContextTeamInfo(ctx context.Context) *shared.TeamInfo {
	if team, ok := ctx.Value(ContextKeySelectedTeam).(*shared.TeamInfo); ok {
		return team
	} else {
		return nil
	}
}
