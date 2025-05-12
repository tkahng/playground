package apis

import (
	"context"

	"github.com/tkahng/authgo/internal/core"
)

type AppSettingsout struct {
	Body *core.AppOptions
}

func (api *Api) GetAppSettings(context context.Context, input *struct{}) (*AppSettingsout, error) {
	return &AppSettingsout{
		Body: api.app.Settings(),
	}, nil
}

type AppSettingsInput struct {
	Body core.AppOptions
}
