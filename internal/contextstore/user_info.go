package contextstore

import (
	"context"

	"github.com/tkahng/authgo/internal/shared"
)

const (
	contextKeyUserInfo contextKey = "user_info"
)

func SetContextUserInfo(ctx context.Context, user *shared.UserInfo) context.Context {
	return context.WithValue(ctx, contextKeyUserInfo, user)
}
func GetContextUserInfo(ctx context.Context) *shared.UserInfo {
	if user, ok := ctx.Value(contextKeyUserInfo).(*shared.UserInfo); ok {
		return user
	} else {
		return nil
	}
}
