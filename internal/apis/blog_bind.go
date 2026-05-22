package apis

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/playground/internal/middleware"
	"github.com/tkahng/playground/internal/middleware/humamiddleware"
	"github.com/tkahng/playground/internal/shared"
)

func bindBlogApi(appApi *Api) {
	api := appApi.Api()

	// Public routes — no auth required
	publicGroup := huma.NewGroup(api, "/blog")

	huma.Register(publicGroup, huma.Operation{
		OperationID: "blog-post-list",
		Method:      http.MethodGet,
		Path:        "/posts",
		Summary:     "List blog posts",
		Tags:        []string{"Blog"},
	}, appApi.BlogPostList)

	huma.Register(publicGroup, huma.Operation{
		OperationID: "blog-post-get",
		Method:      http.MethodGet,
		Path:        "/posts/{slug}",
		Summary:     "Get blog post by slug",
		Tags:        []string{"Blog"},
		Errors:      []int{http.StatusNotFound},
	}, appApi.BlogPostGet)

	huma.Register(publicGroup, huma.Operation{
		OperationID: "blog-tag-list",
		Method:      http.MethodGet,
		Path:        "/tags",
		Summary:     "List blog tags",
		Tags:        []string{"Blog"},
	}, appApi.BlogTagList)

	// Admin routes — require superuser permission
	adminGroup := huma.NewGroup(api, "/blog")
	adminGroup.UseMiddleware(
		humamiddleware.HumaChiMiddlewares(
			middleware.CheckPermissionsMiddleware(shared.PermissionNameAdmin),
		)...,
	)

	huma.Register(adminGroup, huma.Operation{
		OperationID: "blog-post-create",
		Method:      http.MethodPost,
		Path:        "/posts",
		Summary:     "Create blog post",
		Tags:        []string{"Blog"},
		Errors:      []int{http.StatusBadRequest},
		Security:    []map[string][]string{{shared.BearerAuthSecurityKey: {}}},
	}, appApi.BlogPostCreate)

	huma.Register(adminGroup, huma.Operation{
		OperationID: "blog-post-update",
		Method:      http.MethodPatch,
		Path:        "/posts/{post-id}",
		Summary:     "Update blog post",
		Tags:        []string{"Blog"},
		Errors:      []int{http.StatusNotFound, http.StatusBadRequest},
		Security:    []map[string][]string{{shared.BearerAuthSecurityKey: {}}},
	}, appApi.BlogPostUpdate)

	huma.Register(adminGroup, huma.Operation{
		OperationID: "blog-post-publish",
		Method:      http.MethodPost,
		Path:        "/posts/{post-id}/publish",
		Summary:     "Publish blog post",
		Tags:        []string{"Blog"},
		Errors:      []int{http.StatusNotFound},
		Security:    []map[string][]string{{shared.BearerAuthSecurityKey: {}}},
	}, appApi.BlogPostPublish)

	huma.Register(adminGroup, huma.Operation{
		OperationID: "blog-post-unpublish",
		Method:      http.MethodPost,
		Path:        "/posts/{post-id}/unpublish",
		Summary:     "Unpublish blog post",
		Tags:        []string{"Blog"},
		Errors:      []int{http.StatusNotFound},
		Security:    []map[string][]string{{shared.BearerAuthSecurityKey: {}}},
	}, appApi.BlogPostUnpublish)

	huma.Register(adminGroup, huma.Operation{
		OperationID: "blog-post-archive",
		Method:      http.MethodPost,
		Path:        "/posts/{post-id}/archive",
		Summary:     "Archive blog post",
		Tags:        []string{"Blog"},
		Errors:      []int{http.StatusNotFound},
		Security:    []map[string][]string{{shared.BearerAuthSecurityKey: {}}},
	}, appApi.BlogPostArchive)

	huma.Register(adminGroup, huma.Operation{
		OperationID: "blog-post-delete",
		Method:      http.MethodDelete,
		Path:        "/posts/{post-id}",
		Summary:     "Delete blog post",
		Tags:        []string{"Blog"},
		Errors:      []int{http.StatusNotFound},
		Security:    []map[string][]string{{shared.BearerAuthSecurityKey: {}}},
	}, appApi.BlogPostDelete)

	huma.Register(adminGroup, huma.Operation{
		OperationID: "blog-tag-create",
		Method:      http.MethodPost,
		Path:        "/tags",
		Summary:     "Create blog tag",
		Tags:        []string{"Blog"},
		Errors:      []int{http.StatusBadRequest},
		Security:    []map[string][]string{{shared.BearerAuthSecurityKey: {}}},
	}, appApi.BlogTagCreate)
}
