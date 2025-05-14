package apis

import (
	"context"

	"github.com/tkahng/authgo/internal/conf"
)

type AppSettingsout struct {
	Body *conf.AppOptions
}

func (api *Api) GetAppSettings(context context.Context, input *struct{}) (*AppSettingsout, error) {
	return &AppSettingsout{
		Body: api.app.Settings(),
	}, nil
}

type AppSettingsInput struct {
	Body *conf.AppOptions
}
