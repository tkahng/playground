package apis

import (
	"context"

	"github.com/tkahng/playground/internal/conf"
)

type AppOptions struct {
	Auth conf.AuthOptions `form:"auth" json:"auth"`
	Smtp conf.SmtpConfig  `form:"smtp" json:"smtp"`
	Meta conf.AppConfig   `form:"meta" json:"meta"`
}
type AppSettingsout struct {
	Body *AppOptions
}

func (api *Api) GetAppSettings(context context.Context, input *struct{}) (*AppSettingsout, error) {
	cfg := api.App().Config()
	return &AppSettingsout{
		Body: &AppOptions{
			Auth: cfg.AuthOptions,
			Smtp: cfg.SmtpConfig,
			Meta: cfg.AppConfig,
		},
	}, nil
}
