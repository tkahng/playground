package contextstore

import (
	"context"

	"github.com/tkahng/playground/internal/models"
)

const (
	contextKeyUserInfo contextKey = "user_info"
)

func SetContextUserInfo(ctx context.Context, user *models.UserInfo) context.Context {
	return context.WithValue(ctx, contextKeyUserInfo, user)
}
func GetContextUserInfo(ctx context.Context) *models.UserInfo {
	if user, ok := ctx.Value(contextKeyUserInfo).(*models.UserInfo); ok {
		return user
	} else {
		return nil
	}
}
