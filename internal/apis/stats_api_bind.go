package apis

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/playground/internal/shared"
)

func BindStatsApi(api huma.API, appApi *Api) {
	statsGroup := huma.NewGroup(api)
	huma.Register(
		statsGroup,
		huma.Operation{
			OperationID: "stats-get",
			Method:      http.MethodGet,
			Path:        "/stats",
			Summary:     "Get stats",
			Description: "Get stats",
			Tags:        []string{"Stats"},
			Errors:      []int{http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
		},
		appApi.Stats,
	)
}
