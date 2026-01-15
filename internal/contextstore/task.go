package contextstore

import (
	"context"

	"github.com/tkahng/playground/internal/models"
)

const (
	contextKeyTask contextKey = "task"
)

func SetContextTask(ctx context.Context, info *models.Task) context.Context {
	return context.WithValue(ctx, contextKeyTask, info)
}
func GetContextTask(ctx context.Context) *models.Task {
	if team, ok := ctx.Value(contextKeyTask).(*models.Task); ok {
		return team
	} else {
		return nil
	}
}
