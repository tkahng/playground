package core

import (
	"context"
	"fmt"

	"github.com/stephenafamo/bob"
	"github.com/tkahng/authgo/internal/db/models"
)

// SendVerificationEmail implements App.
func (app *BaseApp) SendVerificationEmail(ctx context.Context, db bob.Executor, user *models.User, redirectTo string) error {
	opts := app.Settings().Auth
	config := app.Settings()
	client := app.NewMailClient()
	payload := app.TokenVerifier().CreateVerificationPayload(user, redirectTo)

	tokenHash, err := app.TokenVerifier().CreateVerificationToken(payload, opts.VerificationToken)
	if err != nil {
		return fmt.Errorf("error at creating verification token: %w", err)
	}

	err = app.TokenStorage().PersistVerificationToken(ctx, db, payload, opts.VerificationToken)
	if err != nil {
		return fmt.Errorf("error at storing verification token: %w", err)
	}
	mailParams, err := createVerificationMailParams(tokenHash, payload, config)
	if err != nil {
		return fmt.Errorf("error creating verification token: %w", err)
	}
	err = client.Send(mailParams)
	if err != nil {
		return fmt.Errorf("error sending verification token: %w", err)
	}
	return nil
}
