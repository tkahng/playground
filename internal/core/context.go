package core

import (
	"context"

	"github.com/tkahng/authgo/internal/db/models"
	"github.com/tkahng/authgo/internal/shared"
)

type ContextKey string

const (
	// ContextKeyUser is the key used to store the user in the context.
	ContextKeyUser           ContextKey = "user"
	ContextKeyStripeCustomer ContextKey = "stripe_customer"
	ContextKeyUserClaims     ContextKey = "user_claims"
)

// type contextKey struct {
// 	name string
// }

// func (k *contextKey) String() string {
// 	return "jwtauth context value " + k.name
// }

// func SetDbUserAsBaseUser

func SetUser(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, ContextKeyUser, user)
}

func GetUser(ctx context.Context) *models.User {
	if user, ok := ctx.Value(ContextKeyUser).(*models.User); ok {
		return user
	} else {
		return nil
	}
}

func SetUserClaims(ctx context.Context, claims *shared.UserInfoDto) context.Context {
	return context.WithValue(ctx, ContextKeyUserClaims, claims)
}

func GetUserClaims(ctx context.Context) *shared.UserInfoDto {
	if user, ok := ctx.Value(ContextKeyUserClaims).(*shared.UserInfoDto); ok {
		return user
	} else {
		return nil
	}
}
