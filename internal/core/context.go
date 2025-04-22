package core

import (
	"context"

	"github.com/tkahng/authgo/internal/shared"
)

type ContextKey string

const (
	ContextKeyUser           ContextKey = "user"
	ContextKeyStripeCustomer ContextKey = "stripe_customer"
	ContextKeyUserClaims     ContextKey = "user_claims"
	ContextKeyUserInfo       ContextKey = "user_info"
)

func SetContextUserClaims(ctx context.Context, claims *shared.UserInfoDto) context.Context {
	info := shared.ToUserInfo(claims)
	ctx = context.WithValue(ctx, ContextKeyUserInfo, info)
	return context.WithValue(ctx, ContextKeyUserClaims, claims)
}

func GetContextUserClaims(ctx context.Context) *shared.UserInfoDto {
	if user, ok := ctx.Value(ContextKeyUserClaims).(*shared.UserInfoDto); ok {
		return user
	} else {
		return nil
	}
}

func SetContextUserInfo(ctx context.Context, user *shared.UserInfo) context.Context {
	userDto := shared.ToUserInfoDto(user)
	ctx = context.WithValue(ctx, ContextKeyUserClaims, userDto)
	return context.WithValue(ctx, ContextKeyUserInfo, user)
}

func GetContextUserInfo(ctx context.Context) *shared.UserInfo {
	if user, ok := ctx.Value(ContextKeyUserInfo).(*shared.UserInfo); ok {
		return user
	} else {
		return nil
	}
}
