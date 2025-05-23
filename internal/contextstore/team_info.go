package contextstore

import (
	"context"

	"github.com/tkahng/authgo/internal/shared"
)

const (
	contextKeyTeamInfo contextKey = "team_info"
)

func SetContextTeamInfo(ctx context.Context, info *shared.TeamInfo) context.Context {
	return context.WithValue(ctx, contextKeyTeamInfo, info)
}
func GetContextTeamInfo(ctx context.Context) *shared.TeamInfo {
	if team, ok := ctx.Value(contextKeyTeamInfo).(*shared.TeamInfo); ok {
		return team
	} else {
		return nil
	}
}
