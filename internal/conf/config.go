package conf

import (
	"fmt"

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
	AppUrl        string `env:"APP_URL" envDefault:"http://localhost:8080"`
	AppName       string `env:"APP_NAME" envDefault:"Playground"`
	SenderName    string `env:"SENDER_NAME" envDefault:"info"`
	SenderAddress string `env:"SENDER_ADDRESS" envDefault:"Hb4k@notifications.k2dv.io"`
	EncryptionKey string `env:"ENCRYPTION_KEY" envDefault:"12345678901234567890123456789012"`
	AppEnv        string `env:"APP_ENV" envDefault:"development"` // can be development, staging, production, test, or debug
}

type DBConfig struct {
	User     string `env:"DATABASE_USER,expand" envDefault:"postgres"`
	Password string `env:"DATABASE_PASSWORD,expand" envDefault:"postgres"`
	Host     string `env:"DATABASE_HOST,expand" envDefault:"localhost"`
	Port     string `env:"DATABASE_PORT,expand" envDefault:"5432"`
	Db       string `env:"DATABASE_DB,expand" envDefault:"playground"`
	SSL      string `env:"DATABASE_SSL,expand" envDefault:"disable"`
	// DatabaseUrl string `env:"DATABASE_URL,expand" envDefault:"postgres://${DATABASE_USER}:${DATABASE_PASSWORD}@${DATABASE_HOST}:${DATABASE_PORT}/${DATABASE_DB}?sslmode=${DATABASE_SSL}"`
}

func (c *DBConfig) GetDatabaseUrl() string {
	url := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.Db, c.SSL,
	)
	return url
}

type ResendConfig struct {
	ResendApiKey string `env:"RESEND_API_KEY" required:"false"`
}

type SmtpConfig struct {
	Host      string `env:"SMTP_HOST" envDefault:""`
	Port      string `env:"SMTP_PORT" envDefault:""`
	Username  string `env:"SMTP_USERNAME" envDefault:""`
	EmailPass string `env:"SMTP_PASSWORD" envDefault:""`
	TLS       bool   `env:"SMTP_TLS" envDefault:"false"`
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
	Debug bool   `doc:"Enable debug logging" default:"true" short:"d"`
	Host  string `doc:"Hostname to listen on." default:"localhost"`
	Port  int    `doc:"Port to listen on." short:"p" default:"8080"`
}

func ZeroEnvConfig() *EnvConfig {
	// nolint:exhaustruct
	return &EnvConfig{
		Db: DBConfig{
			User:     "postgres",
			Password: "postgres",
			Host:     "localhost",
			Port:     "5432",
			Db:       "playground_test",
			SSL:      "disable",
		},
		AppConfig: AppConfig{
			AppUrl:        "http://localhost:8080",
			AppName:       "Playground",
			SenderName:    "info",
			SenderAddress: "Hb4k@notifications.k2dv.io",
			EncryptionKey: "12345678901234567890123456789012",
			AppEnv:        "dev",
		},
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

func AppConfigGetter() *EnvConfig {
	var config EnvConfig
	if err := env.ParseWithOptions(&config, env.Options{
		RequiredIfNoDef: true,
	}); err != nil {
		panic(err)
	}
	config.AuthOptions = NewTokenOptions()
	return &config
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
