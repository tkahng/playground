package core

import (
	"context"
	"encoding/json"
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/stephenafamo/bob"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/security"
)

const (
	SignupVerification             = "signup"
	RecoveryVerification           = "recovery"
	InviteVerification             = "invite"
	MagicLinkVerification          = "magiclink"
	EmailChangeVerification        = "email_change"
	EmailOTPVerification           = "email"
	EmailChangeCurrentVerification = "email_change_current"
	EmailChangeNewVerification     = "email_change_new"
	ReauthenticationVerification   = "reauthentication"
)

// struct to save in the value of the param in the db
type EncryptedAuthOptions struct {
	EncryptedAuthOptions string `form:"encrypted_auth_options" json:"encrypted_auth_options"`
}

type AuthOptions struct {
	// OAuth2             OAuth2Option `form:"oauth2" json:"oauth2"`
	OAuth2Config       OAuth2Config `form:"oauth2_config" json:"oauth2_config"`
	AccessToken        TokenOption  `form:"access_token" json:"access_token"`
	PasswordResetToken TokenOption  `form:"password_reset_token" json:"password_reset_token"`
	VerificationToken  TokenOption  `form:"verification_token" json:"verification_token"`
	RefreshToken       TokenOption  `form:"refresh_token" json:"refresh_token"`
	StateToken         TokenOption  `form:"state_token" json:"state_token"`
}

func (o AuthOptions) Validate() error {
	err := validation.ValidateStruct(&o,

		validation.Field(&o.OAuth2Config),

		validation.Field(&o.AccessToken),
		validation.Field(&o.RefreshToken),
		validation.Field(&o.VerificationToken),
		validation.Field(&o.PasswordResetToken),
		validation.Field(&o.StateToken),
	)
	if err != nil {
		return err
	}

	return nil
}

func GetOrSetEncryptedAuthOptions(ctx context.Context, dbx bob.Executor, encryptionKey string) (*AuthOptions, error) {
	var opts *AuthOptions
	var encryptedOpts *EncryptedAuthOptions
	// get the encrypted auth options from the db
	encryptedParams, err := repository.FindParams[EncryptedAuthOptions](ctx, dbx, EncryptedAuthOptionsKey)
	if err != nil {
		return nil, fmt.Errorf("error getting encrypted auth options from db: %w", err)
	}
	if encryptedParams != nil {
		encryptedOpts = &encryptedParams.Value.Val
	}
	// if the encrypted auth options are not nil, decrypt them
	if encryptedOpts != nil {
		decryptedOptString, err := security.Decrypt(encryptedOpts.EncryptedAuthOptions, encryptionKey)
		if err != nil {
			return nil, fmt.Errorf("error decrypting auth options: %w", err)
		}
		var authOpts *AuthOptions
		err = json.Unmarshal(decryptedOptString, &authOpts)
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling auth options: %w", err)
		}
		opts = authOpts
	}
	if opts == nil {
		opts = DefaultAuthSettings()
		optsStr, err := json.Marshal(opts)
		if err != nil {
			return nil, fmt.Errorf("error marshalling auth options: %w", err)
		}
		encryptedOptsStr, err := security.Encrypt(optsStr, encryptionKey)
		if err != nil {
			return nil, fmt.Errorf("error encrypting auth options: %w", err)
		}

		encryptedOpts = &EncryptedAuthOptions{
			EncryptedAuthOptions: encryptedOptsStr,
		}
		err = repository.SetParams(ctx, dbx, EncryptedAuthOptionsKey, encryptedOpts)
		if err != nil {
			return nil, err
		}

	}
	return opts, nil
}

func DefaultAuthSettings() *AuthOptions {
	return &AuthOptions{

		VerificationToken: TokenOption{
			Type:     shared.VerificationTokenType,
			Secret:   string(shared.VerificationTokenType),
			Duration: 259200, // 3days
		},
		AccessToken: TokenOption{
			Type:     shared.AccessTokenType,
			Secret:   string(shared.AccessTokenType),
			Duration: 60, // 1hr
		},
		PasswordResetToken: TokenOption{
			Type:     shared.PasswordResetTokenType,
			Secret:   string(shared.PasswordResetTokenType),
			Duration: 1800, // 30min
		},
		RefreshToken: TokenOption{
			Type:     shared.RefreshTokenType,
			Secret:   string(shared.RefreshTokenType),
			Duration: 604800, // 7days
		},
		StateToken: TokenOption{
			Type:     shared.StateTokenType,
			Secret:   string(shared.StateTokenType),
			Duration: 1800, // 30min
		},
	}
}
