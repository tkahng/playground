package core

import (
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
