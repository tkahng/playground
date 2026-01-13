package contextstore

import (
	"context"

	"github.com/tkahng/playground/internal/models"
)

const (
	contextKeyTaskProject contextKey = "task_project"
)

func SetContextTaskProject(ctx context.Context, info *models.TaskProject) context.Context {
	return context.WithValue(ctx, contextKeyTaskProject, info)
}
func GetContextTaskProject(ctx context.Context) *models.TaskProject {
	if team, ok := ctx.Value(contextKeyTaskProject).(*models.TaskProject); ok {
		return team
	} else {
		return nil
	}
}
