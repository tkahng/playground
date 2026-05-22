package apis

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/playground/internal/shared"
)

func bindMediaApi(appApi *Api) {
	api := appApi.Api()
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
	// ---- Update Media (alt text, etc.)
	huma.Register(
		mediaGroup,
		huma.Operation{
			OperationID: "update-media",
			Method:      http.MethodPatch,
			Path:        "/media/{id}",
			Summary:     "Update media metadata",
			Tags:        []string{"Media"},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Errors: []int{http.StatusUnauthorized, http.StatusNotFound, http.StatusInternalServerError},
		},
		appApi.UpdateMedia,
	)
	// ---- Delete Media
	huma.Register(
		mediaGroup,
		huma.Operation{
			OperationID: "delete-media",
			Method:      http.MethodDelete,
			Path:        "/media/{id}",
			Summary:     "Delete media",
			Tags:        []string{"Media"},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Errors: []int{http.StatusUnauthorized, http.StatusNotFound, http.StatusInternalServerError},
		},
		appApi.DeleteMedia,
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
