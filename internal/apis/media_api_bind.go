package apis

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/playground/internal/shared"
)

func BindMediaApi(api huma.API, appApi *Api) {
	mediaGroup := huma.NewGroup(api)
	huma.Register(
		mediaGroup,
		huma.Operation{
			OperationID: "upload-media",
			Method:      http.MethodPost,
			Path:        "/media",
			Summary:     "Upload media",
			Description: "Upload a media file",
			Tags:        []string{"Media"},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Errors: []int{http.StatusUnauthorized, http.StatusBadRequest, http.StatusInternalServerError},
		},
		appApi.UploadMedia,
	)
	// ---- Get Media
	huma.Register(
		mediaGroup,
		huma.Operation{
			OperationID: "get-media",
			Method:      http.MethodGet,
			Path:        "/media/{id}",
			Summary:     "Get media",
			Description: "Get a media file by ID",
			Tags:        []string{"Media"},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Errors: []int{http.StatusUnauthorized, http.StatusNotFound, http.StatusInternalServerError},
		},
		appApi.GetMedia,
	)
	// ---- Get Media List
	huma.Register(
		mediaGroup,
		huma.Operation{
			OperationID: "list-media",
			Method:      http.MethodGet,
			Path:        "/media",
			Summary:     "List media",
			Description: "List all media files for the user",
			Tags:        []string{"Media"},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Errors: []int{http.StatusUnauthorized, http.StatusInternalServerError},
		},
		appApi.MediaList,
	)
}
