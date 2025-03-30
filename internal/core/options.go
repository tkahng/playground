package core

import (
	"context"
	"encoding/json"
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/stephenafamo/bob"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/tools/security"
)

const (
	// name of the param in db for encrypted auth options
	EncryptedAuthOptionsKey = "encrypted_auth"
	EncryptedAppOptionsKey  = "encrypted_app"
)

type EncryptedAppSettings struct {
	EncryptedAppSettings string `form:"encrypted_app_settings" json:"encrypted_app_settings"`
}

type MetaOption struct {
	AppName       string `form:"app_name" json:"app_name" envDefault:"AuthGo" default:"AuthGo"`
	AppURL        string `form:"app_url" json:"app_url" envDefault:"http://localhost:8080" default:"http://localhost:8080"`
	SenderName    string `form:"sender_name" json:"sender_name"`
	SenderAddress string `form:"sender_address" json:"sender_address"`
}

func (c MetaOption) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.AppName, validation.Required, validation.Length(1, 255)),
		validation.Field(&c.AppURL, validation.Required, is.URL),
		validation.Field(&c.SenderName, validation.Required, validation.Length(1, 255)),
		validation.Field(&c.SenderAddress, is.EmailFormat, validation.Required),
	)
}

// --

func NewDefaultSettings() *AppOptions {
	return &AppOptions{
		Meta: MetaOption{
			AppName:       "Acme",
			AppURL:        "http://localhost:8080",
			SenderName:    "Support",
			SenderAddress: "support@example.com",
		},
		Auth: *DefaultAuthSettings(),
		// Logs: LogsConfig{
		// 	MaxDays: 5,
		// 	LogIP:   true,
		// },
		SMTP: SMTPOption{
			Enabled:  false,
			Host:     "smtp.example.com",
			Port:     587,
			Username: "",
			Password: "",
			TLS:      false,
		},
		// Backups: BackupsConfig{
		// 	CronMaxKeep: 3,
		// },
		// Batch: BatchConfig{
		// 	Enabled:     false,
		// 	MaxRequests: 50,
		// 	Timeout:     3,
		// },
		// RateLimits: RateLimitsConfig{
		// 	Enabled: false, // @todo once tested enough enable by default for new installations
		// 	Rules: []RateLimitRule{
		// 		{Label: "*:auth", MaxRequests: 2, Duration: 3},
		// 		{Label: "*:create", MaxRequests: 20, Duration: 5},
		// 		{Label: "/api/batch", MaxRequests: 3, Duration: 1},
		// 		{Label: "/api/", MaxRequests: 300, Duration: 10},
		// 	},
		// },
	}
}

type AppOptions struct {
	Auth AuthOptions `form:"auth" json:"auth"`
	SMTP SMTPOption  `form:"smtp" json:"smtp"`
	// Backups      BackupsConfig      `form:"backups" json:"backups"`
	S3   S3option   `form:"s3" json:"s3"`
	Meta MetaOption `form:"meta" json:"meta"`
	// RateLimits   RateLimitsConfig   `form:"rateLimits" json:"rateLimits"`
	// TrustedProxy TrustedProxyConfig `form:"trustedProxy" json:"trustedProxy"`
	// Batch        BatchConfig        `form:"batch" json:"batch"`
	// Logs         LogsConfig         `form:"logs" json:"logs"`
}

func (s AppOptions) Validate() error {
	return validation.ValidateStruct(&s,
		validation.Field(&s.Meta),
		validation.Field(&s.Auth),
		// validation.Field(&s.Logs),
		validation.Field(&s.SMTP),
		validation.Field(&s.S3),
		// validation.Field(&s.Backups),
		// validation.Field(&s.Batch),
		// validation.Field(&s.RateLimits),
		// validation.Field(&s.TrustedProxy),
	)
}

func (s *AppOptions) PostValidate(ctx context.Context) error {
	// s.mu.RLock()
	// defer s.mu.RUnlock()

	return validation.ValidateStructWithContext(ctx, s,
		validation.Field(&s.Meta),
		validation.Field(&s.Auth),
		// validation.Field(&s.Logs),
		validation.Field(&s.SMTP),
		validation.Field(&s.S3),
		// validation.Field(&s.Backups),
		// validation.Field(&s.Batch),
		// validation.Field(&s.RateLimits),
		// validation.Field(&s.TrustedProxy),
	)
}

type PasswordAuthConfig struct {
	Enabled bool `form:"enabled" json:"enabled"`

	// IdentityFields is a list of field names that could be used as
	// identity during password authentication.
	//
	// Usually only fields that has single column UNIQUE index are accepted as values.

}

type EmailContentOption struct {
	Invite           string `json:"invite"`
	Confirmation     string `json:"confirmation"`
	Recovery         string `json:"recovery"`
	EmailChange      string `json:"email_change" split_words:"true"`
	MagicLink        string `json:"magic_link" split_words:"true"`
	Reauthentication string `json:"reauthentication"`
}

type SMTPOption struct {
	Enabled  bool   `form:"enabled" json:"enabled" envDefault:"false" default:"false"`
	Port     int    `form:"port" json:"port"`
	Host     string `form:"host" json:"host"`
	Username string `form:"username" json:"username"`
	Password string `form:"password" json:"password,omitempty"`
	TLS      bool   `form:"tls" json:"tls"`
}

func (c SMTPOption) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(
			&c.Host,
			validation.When(c.Enabled, validation.Required),
			is.Host,
		),
		validation.Field(
			&c.Port,
			validation.When(c.Enabled, validation.Required),
			validation.Min(0),
		),
		validation.Field(
			&c.Username,
			validation.When(c.Enabled, validation.Required),
		),
		validation.Field(
			&c.Password,
			validation.When(c.Enabled, validation.Required),
		),
		validation.Field(
			&c.TLS,
			validation.When(c.Enabled, validation.Required),
		),
	)
}

type S3option struct {
	Enabled        bool   `form:"enabled" json:"enabled"`
	Bucket         string `form:"bucket" json:"bucket"`
	Region         string `form:"region" json:"region"`
	Endpoint       string `form:"endpoint" json:"endpoint"`
	AccessKey      string `form:"access_key" json:"access_key"`
	Secret         string `form:"secret" json:"secret,omitempty"`
	ForcePathStyle bool   `form:"forcePathStyle" json:"forcePathStyle"`
	//
}

func (c S3option) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Endpoint, is.URL, validation.When(c.Enabled, validation.Required)),
		validation.Field(&c.Bucket, validation.When(c.Enabled, validation.Required)),
		validation.Field(&c.Region, validation.When(c.Enabled, validation.Required)),
		validation.Field(&c.AccessKey, validation.When(c.Enabled, validation.Required)),
		validation.Field(&c.Secret, validation.When(c.Enabled, validation.Required)),
	)
}

type EncryptedAppOptions struct {
	EncryptedAppOptions string `form:"encrypted_app_options" json:"encrypted_app_options"`
}

func GetOrSetEncryptedAppOptions(ctx context.Context, dbx bob.DB, encryptionKey string) (*AppOptions, error) {
	var opts *AppOptions
	var encryptedOpts *EncryptedAppOptions
	// get the encrypted auth options from the db
	encryptedParams, err := repository.GetParams[EncryptedAppOptions](ctx, dbx, EncryptedAppOptionsKey)
	if err != nil {
		return nil, fmt.Errorf("error getting encrypted auth options from db: %w", err)
	}
	if encryptedParams != nil {
		encryptedOpts = &encryptedParams.Value.Val
	}
	// if the encrypted auth options are not nil, decrypt them
	if encryptedOpts != nil {
		decryptedOptString, err := security.Decrypt(encryptedOpts.EncryptedAppOptions, encryptionKey)
		if err != nil {
			return nil, fmt.Errorf("error decrypting auth options: %w", err)
		}
		var appOpts *AppOptions
		err = json.Unmarshal(decryptedOptString, &appOpts)
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling auth options: %w", err)
		}
		opts = appOpts
	}
	if opts == nil {
		opts = NewDefaultSettings()
		err1 := EncryptAndSetSettings(ctx, dbx, opts, encryptionKey)
		if err1 != nil {
			return nil, err1
		}
		if opts != nil {
			return opts, nil
		}
	}
	return opts, nil
}

func EncryptAndSetSettings(ctx context.Context, dbx bob.DB, opts *AppOptions, encryptionKey string) error {
	optsStr, err := json.Marshal(opts)
	if err != nil {
		return fmt.Errorf("error marshalling auth options: %w", err)
	}
	encryptedOptsStr, err := security.Encrypt(optsStr, encryptionKey)
	if err != nil {
		return fmt.Errorf("error encrypting auth options: %w", err)
	}

	encryptedOpts := &EncryptedAppOptions{
		EncryptedAppOptions: encryptedOptsStr,
	}
	err = repository.SetParams(ctx, dbx, EncryptedAuthOptionsKey, encryptedOpts)
	if err != nil {
		return err
	}
	return nil
}

// regex to
