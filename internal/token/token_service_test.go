package token_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tkahng/playground/internal/conf"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/token"
)

func TestTokenServiceImpl_GenerateToken(t *testing.T) {
	t.Run("Test success generate and validate token", func(t *testing.T) {
		database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
			cfg := conf.ZeroEnvConfig()
			store := stores.NewPostgresTokenStore(db)
			tokenService := token.NewTokenService(cfg, store)
			email := "admin@k2dv.io"
			token1, err := tokenService.GenerateToken(ctx, &token.GenerateTokenOptions{
				Email: email,
				Type:  models.TokenTypesVerificationToken,
			})
			if err != nil {
				t.Fatal(err)
			}
			if token1 == "" {
				t.Fatal("token is empty")
			}
			value, err := tokenService.ValidateToken(ctx, &token.ValidateTokenOptions{
				Token: token1,
				Type:  models.TokenTypesVerificationToken,
			})
			if err != nil {
				t.Fatal(err)
			}
			if value != email {
				t.Fatalf("expected %s, got %s", email, value)
			}
		})
	})
	t.Run("Test success generate and validate token", func(t *testing.T) {
		database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
			cfg := conf.ZeroEnvConfig()
			cfg.AuthOptions.VerificationToken.Duration = 1
			store := stores.NewPostgresTokenStore(db)
			tokenService := token.NewTokenService(cfg, store)
			email := "admin@k2dv.io"
			token1, err := tokenService.GenerateToken(ctx, &token.GenerateTokenOptions{
				Email: email,
				Type:  models.TokenTypesVerificationToken,
			})
			if err != nil {
				t.Fatal(err)
			}
			if token1 == "" {
				t.Fatal("token is empty")
			}
			require.Eventually(t, func() bool {
				_, err := tokenService.ValidateToken(ctx, &token.ValidateTokenOptions{
					Token: token1,
					Type:  models.TokenTypesVerificationToken,
				})
				return err != nil
			}, 3*time.Second, 50*time.Millisecond)
		})
	})
}
