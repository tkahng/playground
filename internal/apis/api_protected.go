package apis

import (
	"context"
	"net/http"
	"slices"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/core"
	"github.com/tkahng/authgo/internal/queries"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/utils"
)

func (a *Api) ApiProtectedOperation(path string) huma.Operation {
	return huma.Operation{
		OperationID: "api-protected",
		Method:      http.MethodGet,
		Path:        path,
		Summary:     "Api protected",
		Description: "Api protected",
		Tags:        []string{"Protected"},
		Errors:      []int{http.StatusNotFound},
		Security: []map[string][]string{
			{shared.BearerAuthSecurityKey: {}},
		},
	}
}

func (a *Api) ApiProtected(ctx context.Context, input *struct {
	PermissionName string `path:"permission-name"`
}) (*struct {
	Body string
}, error) {
	claims := core.GetContextUserInfo(ctx)
	if claims == nil {
		return nil, huma.Error404NotFound("User not found")
	}
	utils.PrettyPrintJSON(claims)
	permission, err := queries.FindPermissionByName(ctx, a.app.Db(), input.PermissionName)
	if err != nil {
		return nil, err
	}
	if permission == nil {
		return nil, huma.Error404NotFound("Permission not found")
	}
	if slices.Contains(claims.Permissions, permission.Name) {
		return &struct{ Body string }{Body: "Api protected"}, nil
	}
	return nil, huma.Error401Unauthorized("Not authorized")
}
