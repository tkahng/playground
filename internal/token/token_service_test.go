package token_test

import (
	"context"
	"testing"
	"time"

	"github.com/tkahng/playground/internal/conf"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/test"
	"github.com/tkahng/playground/internal/token"
)

func TestTokenServiceImpl_GenerateToken(t *testing.T) {
	t.Run("Test success generate and validate token", func(t *testing.T) {
		test.WithSingletonTx(t, func(ctx context.Context, db database.Dbx) {
			cfg := conf.ZeroEnvConfig()
			store := stores.NewPostgresTokenStore(db)
			tokenService := token.NewTokenService(&cfg, store)
			email := "admin@k2dv.io"
			token, err := tokenService.GenerateToken(ctx, email, models.TokenTypesVerificationToken)
			if err != nil {
				t.Fatal(err)
			}
			if token == "" {
				t.Fatal("token is empty")
			}
			value, err := tokenService.ValidateToken(ctx, token, models.TokenTypesVerificationToken)
			if err != nil {
				t.Fatal(err)
			}
			if value != email {
				t.Fatalf("expected %s, got %s", email, value)
			}
		})
	})
	t.Run("Test success generate and validate token", func(t *testing.T) {
		test.WithSingletonTx(t, func(ctx context.Context, db database.Dbx) {
			cfg := conf.ZeroEnvConfig()
			cfg.AuthOptions.VerificationToken.Duration = 1
			store := stores.NewPostgresTokenStore(db)
			tokenService := token.NewTokenService(&cfg, store)
			email := "admin@k2dv.io"
			token, err := tokenService.GenerateToken(ctx, email, models.TokenTypesVerificationToken)
			if err != nil {
				t.Fatal(err)
			}
			if token == "" {
				t.Fatal("token is empty")
			}
			time.Sleep(time.Second * 1)
			_, err = tokenService.ValidateToken(ctx, token, models.TokenTypesVerificationToken)
			if err == nil {
				t.Fatal(err)
			}
		})
	})
}
