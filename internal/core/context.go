package core

import (
	"context"

	"github.com/tkahng/authgo/internal/shared"
)

type ContextKey string

const (
	ContextKeyUser           ContextKey = "user"
	ContextKeyStripeCustomer ContextKey = "stripe_customer"
	ContextKeyUserInfo       ContextKey = "user_info"
)

func SetContextUserInfo(ctx context.Context, user *shared.UserInfo) context.Context {
	return context.WithValue(ctx, ContextKeyUserInfo, user)
}

func GetContextUserInfo(ctx context.Context) *shared.UserInfo {
	if user, ok := ctx.Value(ContextKeyUserInfo).(*shared.UserInfo); ok {
		return user
	} else {
		return nil
	}
}
