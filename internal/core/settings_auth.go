package core

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/stephenafamo/bob"
	"github.com/tkahng/authgo/internal/auth"
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
	OAuth2 OAuth2Config `form:"oauth2" json:"oauth2"`

	AccessToken        TokenConfig `form:"access_token" json:"access_token"`
	PasswordResetToken TokenConfig `form:"password_reset_token" json:"password_reset_token"`
	VerificationToken  TokenConfig `form:"verification_token" json:"verification_token"`
	RefreshToken       TokenConfig `form:"refresh_token" json:"refresh_token"`
	StateToken         TokenConfig `form:"state_token" json:"state_token"`

	// Default email templates
	// ---
	// VerificationTemplate       EmailTemplate `form:"verificationTemplate" json:"verificationTemplate"`
	// ResetPasswordTemplate      EmailTemplate `form:"resetPasswordTemplate" json:"resetPasswordTemplate"`
	// ConfirmEmailChangeTemplate EmailTemplate `form:"confirmEmailChangeTemplate" json:"confirmEmailChangeTemplate"`
}

func (o AuthOptions) Validate() error {
	err := validation.ValidateStruct(&o,
		// validation.Field(
		// 	&o.AuthRule,
		// 	validation.By(cv.checkRule),
		// 	validation.By(cv.ensureNoSystemRuleChange(cv.original.AuthRule)),
		// ),
		// validation.Field(
		// 	&o.ManageRule,
		// 	validation.NilOrNotEmpty,
		// 	validation.By(cv.checkRule),
		// 	validation.By(cv.ensureNoSystemRuleChange(cv.original.ManageRule)),
		// ),
		// validation.Field(&o.AuthAlert),
		// validation.Field(&o.PasswordAuth),
		validation.Field(&o.OAuth2),
		// validation.Field(&o.OTP),
		// validation.Field(&o.MFA),
		validation.Field(&o.AccessToken),
		validation.Field(&o.RefreshToken),
		validation.Field(&o.VerificationToken),
		validation.Field(&o.PasswordResetToken),
		validation.Field(&o.StateToken),
		// validation.Field(&o.FileToken),
		// validation.Field(&o.VerificationTemplate, validation.Required),
		// validation.Field(&o.ResetPasswordTemplate, validation.Required),
		// validation.Field(&o.ConfirmEmailChangeTemplate, validation.Required),
	)
	if err != nil {
		return err
	}

	// if o.MFA.Enabled {
	// 	// if MFA is enabled require at least 2 auth methods
	// 	//
	// 	// @todo maybe consider disabling the check because if custom auth methods
	// 	// are registered it may fail since we don't have mechanism to detect them at the moment
	// 	authsEnabled := 0
	// 	if o.PasswordAuth.Enabled {
	// 		authsEnabled++
	// 	}
	// 	if o.OAuth2.Enabled {
	// 		authsEnabled++
	// 	}
	// 	if o.OTP.Enabled {
	// 		authsEnabled++
	// 	}
	// 	if authsEnabled < 2 {
	// 		return validation.Errors{
	// 			"mfa": validation.Errors{
	// 				"enabled": validation.NewError("validation_mfa_not_enough_auths", "MFA requires at least 2 auth methods to be enabled."),
	// 			},
	// 		}
	// 	}

	// 	if o.MFA.Rule != "" {
	// 		mfaRuleValidators := []validation.RuleFunc{
	// 			cv.checkRule,
	// 			cv.ensureNoSystemRuleChange(&cv.original.MFA.Rule),
	// 		}

	// 		for _, validator := range mfaRuleValidators {
	// 			err := validator(&o.MFA.Rule)
	// 			if err != nil {
	// 				return validation.Errors{
	// 					"mfa": validation.Errors{
	// 						"rule": err,
	// 					},
	// 				}
	// 			}
	// 		}
	// 	}
	// }

	// // extra check to ensure that only unique identity fields are used
	// if o.PasswordAuth.Enabled {
	// 	err = validation.Validate(o.PasswordAuth.IdentityFields, validation.By(cv.checkFieldsForUniqueIndex))
	// 	if err != nil {
	// 		return validation.Errors{
	// 			"passwordAuth": validation.Errors{
	// 				"identityFields": err,
	// 			},
	// 		}
	// 	}
	// }

	return nil
}

func GetOrSetEncryptedAuthOptions(ctx context.Context, dbx bob.DB, encryptionKey string) (*AuthOptions, error) {
	var opts *AuthOptions
	var encryptedOpts *EncryptedAuthOptions
	// get the encrypted auth options from the db
	encryptedParams, err := repository.GetParams[EncryptedAuthOptions](ctx, dbx, EncryptedAuthOptionsKey)
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
		_, err = repository.SetParams(ctx, dbx, EncryptedAuthOptionsKey, encryptedOpts)
		if err != nil {
			return nil, err
		}

	}
	return opts, nil
}

func DefaultAuthSettings() *AuthOptions {
	return &AuthOptions{
		OAuth2: OAuth2Config{},

		// VerificationTemplate:       defaultVerificationTemplate,
		// ResetPasswordTemplate:      defaultResetPasswordTemplate,
		// ConfirmEmailChangeTemplate: defaultConfirmEmailChangeTemplate,

		VerificationToken: TokenConfig{
			// Type:     shared.VerificationTokenType,
			Secret:   string(shared.VerificationTokenType),
			Duration: 259200, // 3days
		},
		AccessToken: TokenConfig{
			// Type:     shared.AccessTokenType,
			Secret:   string(shared.AccessTokenType),
			Duration: 3600, // 1hr
		},
		PasswordResetToken: TokenConfig{
			// Type:     shared.PasswordResetTokenType,
			Secret:   string(shared.PasswordResetTokenType),
			Duration: 1800, // 30min
		},
		RefreshToken: TokenConfig{
			// Type:     shared.RefreshTokenType,
			Secret:   string(shared.RefreshTokenType),
			Duration: 604800, // 7days
		},
		StateToken: TokenConfig{
			// Type:     shared.StateTokenType,
			Secret:   string(shared.StateTokenType),
			Duration: 1800, // 30min
		},
	}
}

type OAuth2KnownFields struct {
	Id        string `form:"id" json:"id"`
	Name      string `form:"name" json:"name"`
	Username  string `form:"username" json:"username"`
	AvatarURL string `form:"avatar_url" json:"avatarURL"`
}

type ProviderInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	State       string `json:"state"`
	AuthURL     string `json:"auth_url"`

	// @todo
	// deprecated: use AuthURL instead
	// AuthUrl will be removed after dropping v0.22 support
	AuthUrl string `json:"authUrl"`

	// technically could be omitted if the provider doesn't support PKCE,
	// but to avoid breaking existing typed clients we'll return them as empty string
	CodeVerifier        string `json:"code_verifier"`
	CodeChallenge       string `json:"code_challenge"`
	CodeChallengeMethod string `json:"code_challenge_method"`
}

type OAuth2ProviderConfig struct {
	// PKCE overwrites the default provider PKCE config option.
	//
	// This usually shouldn't be needed but some OAuth2 vendors, like the LinkedIn OIDC,
	// may require manual adjustment due to returning error if extra parameters are added to the request
	// (https://github.com/tkahng/authgo/internal/discussions/3799#discussioncomment-7640312)
	PKCE *bool `form:"pkce" json:"pkce"`

	Name         string         `form:"name" json:"name"`
	ClientId     string         `form:"client_id" json:"client_id"`
	ClientSecret string         `form:"client_secret" json:"client_secret"`
	AuthURL      string         `form:"auth_url" json:"auth_url"`
	TokenURL     string         `form:"token_url" json:"token_url"`
	UserInfoURL  string         `form:"user_info_url" json:"user_info_url"`
	DisplayName  string         `form:"display_name" json:"display_name"`
	Extra        map[string]any `form:"extra" json:"extra"`
}

func (c OAuth2ProviderConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Name, validation.Required, validation.By(checkProviderName)),
		validation.Field(&c.ClientId, validation.Required),
		validation.Field(&c.ClientSecret, validation.Required),
		validation.Field(&c.AuthURL, is.URL),
		validation.Field(&c.TokenURL, is.URL),
		validation.Field(&c.UserInfoURL, is.URL),
	)
}
func checkProviderName(value any) error {
	name, _ := value.(string)
	if name == "" {
		return nil // nothing to check
	}

	if _, err := auth.NewProviderByName(name); err != nil {
		return validation.NewError("validation_missing_provider", "Invalid or missing provider with name {{.name}}.").
			SetParams(map[string]any{"name": name})
	}

	return nil
}

type OAuth2Config struct {
	Providers []OAuth2ProviderConfig `form:"providers" json:"providers"`

	MappedFields OAuth2KnownFields `form:"mapped_fields" json:"mapped_fields"`

	Enabled bool `form:"enabled" json:"enabled"`
}

func (c OAuth2Config) Validate() error {
	if !c.Enabled {
		return nil // no need to validate
	}

	return validation.ValidateStruct(&c,
		// note: don't require providers for now as they could be externally registered/removed
		validation.Field(&c.Providers, validation.By(checkForDuplicatedProviders)),
	)
}

func checkForDuplicatedProviders(value any) error {
	configs, _ := value.([]OAuth2ProviderConfig)

	existing := map[string]struct{}{}

	for i, c := range configs {
		if c.Name == "" {
			continue // the name nonempty state is validated separately
		}
		if _, ok := existing[c.Name]; ok {
			return validation.Errors{
				strconv.Itoa(i): validation.Errors{
					"name": validation.NewError("validation_duplicated_provider", "The provider {{.name}} is already registered.").
						SetParams(map[string]any{"name": c.Name}),
				},
			}
		}
		existing[c.Name] = struct{}{}
	}

	return nil
}

func (c OAuth2Config) GetProviderConfig(name string) (config OAuth2ProviderConfig, exists bool) {
	for _, p := range c.Providers {
		if p.Name == name {
			return p, true
		}
	}
	return
}

func (c OAuth2ProviderConfig) InitProvider() (auth.Provider, error) {
	provider, err := auth.NewProviderByName(c.Name)
	if err != nil {
		return nil, err
	}

	if c.ClientId != "" {
		provider.SetClientId(c.ClientId)
	}

	if c.ClientSecret != "" {
		provider.SetClientSecret(c.ClientSecret)
	}

	if c.AuthURL != "" {
		provider.SetAuthURL(c.AuthURL)
	}

	if c.UserInfoURL != "" {
		provider.SetUserInfoURL(c.UserInfoURL)
	}

	if c.TokenURL != "" {
		provider.SetTokenURL(c.TokenURL)
	}

	if c.DisplayName != "" {
		provider.SetDisplayName(c.DisplayName)
	}

	if c.PKCE != nil {
		provider.SetPKCE(*c.PKCE)
	}

	if c.Extra != nil {
		provider.SetExtra(c.Extra)
	}

	return provider, nil
}
