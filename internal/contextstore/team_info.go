package contextstore

import (
	"context"

	"github.com/tkahng/authgo/internal/shared"
)

const (
	contextKeyTeamInfo contextKey = "team_info"
)

func SetContextTeamInfo(ctx context.Context, info *shared.TeamInfoModel) context.Context {
	return context.WithValue(ctx, contextKeyTeamInfo, info)
}
func GetContextTeamInfo(ctx context.Context) *shared.TeamInfoModel {
	if team, ok := ctx.Value(contextKeyTeamInfo).(*shared.TeamInfoModel); ok {
		return team
	} else {
		return nil
	}
}
