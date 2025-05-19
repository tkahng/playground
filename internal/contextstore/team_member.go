package contextstore

import (
	"context"

	"github.com/tkahng/authgo/internal/models"
)

const (
	ContextKeySelectedTeamMember ContextKey = "selected_team_member"
	ContextKeyLatestTeamMember   ContextKey = "latest_team_member"
)

func SetContextSelectedTeamMember(ctx context.Context, teamMember *models.TeamMember) context.Context {
	return context.WithValue(ctx, ContextKeySelectedTeamMember, teamMember)
}
func GetContextSelectedTeamMember(ctx context.Context) *models.TeamMember {
	if teamMember, ok := ctx.Value(ContextKeySelectedTeamMember).(*models.TeamMember); ok {
		return teamMember
	} else {
		return nil
	}
}

func SetContextLatestTeamMember(ctx context.Context, teamMember *models.TeamMember) context.Context {
	return context.WithValue(ctx, ContextKeyLatestTeamMember, teamMember)
}
func GetContextLatestTeamMember(ctx context.Context) *models.TeamMember {
	if teamMember, ok := ctx.Value(ContextKeyLatestTeamMember).(*models.TeamMember); ok {
		return teamMember
	} else {
		return nil
	}
}
