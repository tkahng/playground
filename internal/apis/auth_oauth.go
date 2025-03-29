package apis

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

func (api *Api) OauthCallbackGetOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "oauth-callback-get",
		Method:      http.MethodGet,
		Path:        path,
		Summary:     "Oauth callback",
		Description: "Oauth callback",
		Tags:        []string{"Auth"},
		Errors:      []int{http.StatusNotFound},
	}
}

func (api *Api) OauthCallbackGet(context.Context, *struct{}) (*struct{}, error) {
	// panic("unimplemented")
	return nil, nil
}
