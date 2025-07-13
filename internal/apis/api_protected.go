package apis

import (
	"context"
	"slices"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/authgo/internal/contextstore"
	"github.com/tkahng/authgo/internal/repository"
)

func (api *Api) ApiProtected(ctx context.Context, input *struct {
	PermissionName string `path:"permission-name"`
}) (*struct {
	Body string
}, error) {
	claims := contextstore.GetContextUserInfo(ctx)
	if claims == nil {
		return nil, huma.Error404NotFound("User not found")
	}
	dbx := api.App().Db()
	permission, err := repository.Permission.GetOne(
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
