package conf

import (
	_ "github.com/joho/godotenv/autoload"
)

type AppConfig struct {
	AppUrl        string `env:"APP_URL" envDefault:"http://localhost:8080"`
	EncryptionKey string `env:"ENCRYPTION_KEY" envDefault:"12345678901234567890123456789012"` //
}

type DBConfig struct {
	DatabaseUrl string `env:"DATABASE_URL" envDefault:"postgres://postgres:postgres@localhost:5432/authgo?sslmode=disable"`
	// DatabaseUrl string `env:"DATABASE_URL" envDefault:"host=host.docker.internal user=postgres password=postgres dbname=db port=5432 sslmode=disable"`
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
	AuthBaseUrl  string `env:"AUTH_BASE_URL" envDefault:"http://127.0.0.1:8080/api/auth"`
	AuthCallback string `env:"AUTH_CALLBACK" envDefault:"http://127.0.0.1:8080/api/auth/callback"`
}

// type StripeConfig struct {
// 	PublicKey string `env:"STRIPE_PUBLISHABLE_KEY"`
// 	ApiKey    string `env:"STRIPE_SECRET_KEY"`
// 	Webhook   string `env:"STRIPE_WEBHOOK_SECRET"`
// }
// type AiConfig struct {
// 	GoogleGeminiApiKey string `env:"GOOGLE_GEMINI_API_KEY"`
// }

type Options struct {
	Debug bool `doc:"Enable debug logging" default:"true" short:"d"`

	Host string `doc:"Hostname to listen on." default:"localhost"`
	Port int    `doc:"Port to listen on." short:"p" default:"8080"`
}

type EnvConfig struct {
	Db DBConfig
	AppConfig
	ResendConfig
	OAuth2Config
	// DBConfig
}

// const (
// 	AuthBaseUrl  string = "http://127.0.0.1:8080/api/auth"
// 	AuthCallback string = "http://127.0.0.1:8080/api/auth/callback"
// )
