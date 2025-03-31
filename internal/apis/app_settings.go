package apis

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/core"
	"github.com/tkahng/authgo/internal/shared"
)

func (api *Api) GetAppSettingsOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "app-settings-get",
		Method:      http.MethodGet,
		Path:        path,
		Summary:     "App settings",
		Description: "App settings",
		Tags:        []string{"Admin", "Settings"},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

type AppSettingsout struct {
	Body *core.AppOptions
}

func (api *Api) GetAppSettings(context context.Context, input *struct{}) (*AppSettingsout, error) {
	return &AppSettingsout{
		Body: api.app.Settings(),
	}, nil
}

func (api *Api) PostAppSettingsOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "app-settings-post",
		Method:      http.MethodPost,
		Path:        path,
		Summary:     "Update App settings",
		Description: "Update App settings",
		Tags:        []string{"App", "Settings"},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

type AppSettingsInput struct {
	Body core.AppOptions
}
