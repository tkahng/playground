package contextstore

import (
	"context"

	"github.com/tkahng/playground/internal/models"
)

const (
	contextKeyTeamMember contextKey = "team_member"
)

func SetContextTeamMember(ctx context.Context, info *models.TeamMember) context.Context {
	return context.WithValue(ctx, contextKeyTeamMember, info)
}
func GetContextTeamMember(ctx context.Context) *models.TeamMember {
	if team, ok := ctx.Value(contextKeyTeamMember).(*models.TeamMember); ok {
		return team
	} else {
		return nil
	}
}
