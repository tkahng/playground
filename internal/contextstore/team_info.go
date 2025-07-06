package contextstore

import (
	"context"

	"github.com/tkahng/authgo/internal/models"
)

const (
	contextKeyTeamInfo contextKey = "team_info"
)

func SetContextTeamInfo(ctx context.Context, info *models.TeamInfoModel) context.Context {
	return context.WithValue(ctx, contextKeyTeamInfo, info)
}
func GetContextTeamInfo(ctx context.Context) *models.TeamInfoModel {
	if team, ok := ctx.Value(contextKeyTeamInfo).(*models.TeamInfoModel); ok {
		return team
	} else {
		return nil
	}
}
