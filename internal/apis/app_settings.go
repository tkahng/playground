package apis

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/shared"
)

func (api *Api) AppSettings(context context.Context, input *struct{}) (*struct{}, error) {
	return nil, nil
}

func (api *Api) AppSettingsOperation(s string) huma.Operation {
	return huma.Operation{
		OperationID: "app-settings",
		Method:      http.MethodGet,
		Path:        s,
		Summary:     "App settings",
		Description: "App settings",
		Tags:        []string{"App", "Settings"},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}
