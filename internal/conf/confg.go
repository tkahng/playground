package conf

import (
	"github.com/caarlos0/env/v11"
	_ "github.com/joho/godotenv/autoload"
)

type JobsConfig struct {
	PollerInterval int64 `env:"POLLER_INTERVAL" envDefault:"1"` // Default
	JobTimeout     int64 `env:"JOB_TIMEOUT" envDefault:"30"`
}

// Duration: 3600, // 1hr
type StorageConfig struct {
	ClientId     string `env:"STORAGE_CLIENT_ID" required:"true" json:"client_id"`
	ClientSecret string `env:"STORAGE_CLIENT_SECRET" required:"true" json:"client_secret"`
	BucketName   string `env:"STORAGE_BUCKET_NAME" required:"true" json:"bucket_name"`
	EndpointUrl  string `env:"STORAGE_ENDPOINT_URL" required:"true" json:"endpoint_url"`
	Region       string `env:"STORAGE_REGION" required:"true" json:"region"`
}
type AppConfig struct {
	AppUrl        string `env:"APP_URL" envDefault:"http://127.0.0.1:8080"`
	AppName       string `env:"APP_NAME" envDefault:"Playground"`
	SenderName    string `env:"SENDER_NAME" envDefault:"info"`
	SenderAddress string `env:"SENDER_ADDRESS" envDefault:"Hb4k@notifications.k2dv.io"`
	EncryptionKey string `env:"ENCRYPTION_KEY" envDefault:"12345678901234567890123456789012"` //
}

type DBConfig struct {
	DatabaseUrl string `env:"DATABASE_URL" envDefault:"postgres://postgres:postgres@localhost:5432/authgo?sslmode=disable"`
}

type ResendConfig struct {
	ResendApiKey string `env:"RESEND_API_KEY" required:"false"`
}

type SmtpConfig struct {
	Host      string `env:"SMTP_HOST" required:"false"`
	Port      string `env:"SMTP_PORT" required:"false"`
	Username  string `env:"SMTP_USERNAME" required:"false"`
	EmailPass string `env:"SMTP_PASSWORD" required:"false"`
	TLS       bool   `env:"SMTP_TLS" required:"false"`
	Enabled   bool   `env:"SMTP_ENABLED" envDefault:"false"`
}
type GithubConfig struct {
	GithubClientId     string `env:"GITHUB_CLIENT_ID" required:"false"`
	GithubClientSecret string `env:"GITHUB_CLIENT_SECRET" required:"false"`
}

type GoogleConfig struct {
	GoogleClientId     string `env:"GOOGLE_CLIENT_ID" required:"false"`
	GoogleClientSecret string `env:"GOOGLE_CLIENT_SECRET" required:"false"`
}
type OAuth2Config struct {
	GithubConfig
	GoogleConfig
	AuthCallback string `env:"AUTH_CALLBACK" envDefault:"/api/auth/callback"`
}

type StripeConfig struct {
	PublicKey    string `env:"STRIPE_PUBLISHABLE_KEY"`
	ApiKey       string `env:"STRIPE_SECRET_KEY"`
	Webhook      string `env:"STRIPE_WEBHOOK_SECRET"`
	StripeAppUrl string `env:"APP_URL" envDefault:"http://localhost:5173"`
}

type AiConfig struct {
	GoogleGeminiApiKey string `env:"GOOGLE_GEMINI_API_KEY" required:"true"`
}

type Options struct {
	Debug bool `doc:"Enable debug logging" default:"true" short:"d"`

	Host string `doc:"Hostname to listen on." default:"localhost"`
	Port int    `doc:"Port to listen on." short:"p" default:"8080"`
}

func ZeroEnvConfig() EnvConfig {
	return EnvConfig{
		AuthOptions: NewTokenOptions(),
	}
}

type EnvConfig struct {
	Options
	Db DBConfig
	JobsConfig
	AppConfig
	ResendConfig
	OAuth2Config
	StripeConfig
	StorageConfig
	AiConfig
	SmtpConfig
	AuthOptions
}

func AppConfigGetter() EnvConfig {
	var config EnvConfig
	if err := env.ParseWithOptions(&config, env.Options{
		RequiredIfNoDef: true,
	}); err != nil {
		panic(err)
	}
	config.AuthOptions = NewTokenOptions()
	return config
}

func GetConfig[T any]() T {
	var config T
	if err := env.ParseWithOptions(&config, env.Options{
		RequiredIfNoDef: true,
	}); err != nil {
		panic(err)
	}
	return config
}

func NewEnvConfig() *EnvConfig {
	config := new(EnvConfig)
	config.AuthOptions = NewTokenOptions()
	return config
}
