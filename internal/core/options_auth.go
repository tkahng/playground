package core

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/tkahng/authgo/internal/shared"
)

type AuthOptions struct {
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

func DefaultAuthSettings() *AuthOptions {
	return &AuthOptions{

		VerificationToken: TokenOption{
			Type:     shared.TokenTypesVerificationToken,
			Secret:   string(shared.TokenTypesVerificationToken),
			Duration: 259200, // 3days
		},
		AccessToken: TokenOption{
			Type:     shared.TokenTypesAccessToken,
			Secret:   string(shared.TokenTypesAccessToken),
			Duration: 3600, // 1hr
		},
		PasswordResetToken: TokenOption{
			Type:     shared.TokenTypesPasswordResetToken,
			Secret:   string(shared.TokenTypesPasswordResetToken),
			Duration: 1800, // 30min
		},
		RefreshToken: TokenOption{
			Type:     shared.TokenTypesRefreshToken,
			Secret:   string(shared.TokenTypesRefreshToken),
			Duration: 604800, // 7days
		},
		StateToken: TokenOption{
			Type:     shared.TokenTypesStateToken,
			Secret:   string(shared.TokenTypesStateToken),
			Duration: 1800, // 30min
		},
	}
}
