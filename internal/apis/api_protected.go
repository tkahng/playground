package apis

import (
	"context"
	"slices"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/core"
	"github.com/tkahng/authgo/internal/crudrepo"
)

func (a *Api) ApiProtected(ctx context.Context, input *struct {
	PermissionName string `path:"permission-name"`
}) (*struct {
	Body string
}, error) {
	claims := core.GetContextUserInfo(ctx)
	if claims == nil {
		return nil, huma.Error404NotFound("User not found")
	}
	dbx := a.app.Db()
	permission, err := crudrepo.Permission.GetOne(
		ctx,
		dbx,
		&map[string]any{
			"name": map[string]any{
				"_eq": input.PermissionName,
			},
		},
	)
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
