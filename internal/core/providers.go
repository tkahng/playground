package core

import (
	"context"

	"github.com/tkahng/playground/internal/conf"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/services"
	"github.com/tkahng/playground/internal/tools/mailer"
)

type Provider interface {
	Config() *conf.EnvConfig
	Db() database.Dbx
	Payment() services.PaymentClient
	Mailer() mailer.Mailer
}

type ResourceProvider struct {
	config  *conf.EnvConfig
	db      *database.Queries
	payment services.PaymentClient
	mailer  mailer.Mailer
}

func NewResourceProvider() *ResourceProvider {
	config := conf.AppConfigGetter()
	db := database.CreateQueriesContext(context.Background(), config.Db.GetDatabaseUrl())
	payment := services.NewPaymentClient(config.StripeConfig)
	mailer := mailer.NewSmtpMailer(config.SmtpConfig)
	return &ResourceProvider{
		config:  &config,
		db:      db,
		payment: payment,
		mailer:  mailer,
	}
}

func NewTestResourceProvider() *ResourceProvider {
	config := conf.ZeroEnvConfig()
	db := database.CreateQueriesContext(context.Background(), config.Db.GetDatabaseUrl())
	payment := services.NewTestPaymentClient()
	mailer := mailer.NewTestMailer()
	return &ResourceProvider{
		config:  &config,
		db:      db,
		payment: payment,
		mailer:  mailer,
	}
}
