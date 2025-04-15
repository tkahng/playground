package apis

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/shared"
)

func (a *Api) ApiProtectedBasicPermissionOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "api-protected-basic-permission",
		Method:      http.MethodGet,
		Path:        path,
		Summary:     "Api protected basic permission",
		Description: "Api protected basic permission",
		Tags:        []string{"Protected"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {shared.PermissionNameBasic}},
		},
	}
}

func (a *Api) ApiProtectedBasicPermission(ctx context.Context, input *struct{}) (*struct {
	Body string
}, error) {
	return &struct{ Body string }{Body: "Api protected basic permission"}, nil
}

func (a *Api) ApiProtectedProPermissionOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "api-protected-pro-permission",
		Method:      http.MethodGet,
		Path:        path,
		Summary:     "Api protected pro permission",
		Description: "Api protected pro permission",
		Tags:        []string{"Protected"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {shared.PermissionNamePro}},
		},
	}
}
func (a *Api) ApiProtectedProPermission(ctx context.Context, input *struct{}) (*struct {
	Body string
}, error) {
	return &struct{ Body string }{Body: "Api protected pro permission"}, nil
}
func (a *Api) ApiProtectedAdvancedPermissionOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "api-protected-advanced-permission",
		Method:      http.MethodGet,
		Path:        path,
		Summary:     "Api protected advanced permission",
		Description: "Api protected advanced permission",
		Tags:        []string{"Protected"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {shared.PermissionNameAdvanced}},
		},
	}
}
func (a *Api) ApiProtectedAdvancedPermission(ctx context.Context, input *struct{}) (*struct {
	Body string
}, error) {
	return &struct{ Body string }{Body: "Api protected advanced permission"}, nil
}
