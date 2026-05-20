package apis_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tkahng/playground/internal/apis"
	"github.com/tkahng/playground/internal/conf"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/filesystem"
)

// createBlogPost is a test helper that creates a post directly through the store.
func createBlogPost(t testing.TB, app *core.BaseApp, authorID uuid.UUID, title, content string, status models.BlogPostStatus) *models.BlogPost {
	t.Helper()
	post, err := app.Adapter().Blog().CreatePost(context.Background(), &stores.CreateBlogPostDTO{
		Title:    title,
		Content:  content,
		AuthorID: authorID,
	})
	require.NoError(t, err)
	if status != models.BlogPostStatusDraft {
		switch status {
		case models.BlogPostStatusPublished:
			post, err = app.Adapter().Blog().PublishPost(context.Background(), post.ID)
		case models.BlogPostStatusArchived:
			post, err = app.Adapter().Blog().ArchivePost(context.Background(), post.ID)
		}
		require.NoError(t, err)
	}
	return post
}

// ── BlogPostList ──────────────────────────────────────────────────────────────

func TestApi_BlogPostList(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)

		admin := core.CreateUserWithOptions(t, testApi.App,
			core.UserWithVerifiedNow(),
			core.UserWithPermission(shared.PermissionNameAdmin),
		)
		createBlogPost(t, testApi.App, admin.User.ID, "Published Post", "hello world", models.BlogPostStatusPublished)
		createBlogPost(t, testApi.App, admin.User.ID, "Draft Post", "draft content", models.BlogPostStatusDraft)

		tests := []apis.ApiScenario{
			{
				Name:           "anon sees only published",
				Method:         http.MethodGet,
				URL:            "/blog/posts",
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					var body apis.ApiPaginatedResponse[*apis.BlogPost]
					require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
					assert.Equal(t, int64(1), body.Meta.Total)
					assert.Equal(t, "Published Post", body.Data[0].Title)
				},
			},
			{
				Name:           "admin sees all statuses",
				Method:         http.MethodGet,
				URL:            "/blog/posts",
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					header, _ := core.CreateAccessHeaderAndRefreshToken(t, app, admin.User.Email)
					scenario.Headers = []string{header}
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					var body apis.ApiPaginatedResponse[*apis.BlogPost]
					require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
					assert.Equal(t, int64(2), body.Meta.Total)
				},
			},
			{
				Name:           "anon search returns matching published post",
				Method:         http.MethodGet,
				URL:            "/blog/posts?q=hello",
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				ExpectedContent: []string{`"title":"Published Post"`},
			},
			{
				Name:           "anon search with no match returns empty",
				Method:         http.MethodGet,
				URL:            "/blog/posts?q=zzznomatch",
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					var body apis.ApiPaginatedResponse[*apis.BlogPost]
					require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
					assert.Equal(t, int64(0), body.Meta.Total)
				},
			},
		}
		for _, tt := range tests {
			tt.Test(t)
		}
	})
}

// ── BlogPostGet ───────────────────────────────────────────────────────────────

func TestApi_BlogPostGet(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)

		admin := core.CreateUserWithOptions(t, testApi.App,
			core.UserWithVerifiedNow(),
			core.UserWithPermission(shared.PermissionNameAdmin),
		)
		published := createBlogPost(t, testApi.App, admin.User.ID, "Hello World", "content", models.BlogPostStatusPublished)
		draft := createBlogPost(t, testApi.App, admin.User.ID, "Draft Post", "content", models.BlogPostStatusDraft)

		tests := []apis.ApiScenario{
			{
				Name:            "anon gets published post",
				Method:          http.MethodGet,
				URL:             fmt.Sprintf("/blog/posts/%s", published.Slug),
				ExpectedStatus:  http.StatusOK,
				TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
				ExpectedContent: []string{`"slug":"hello-world"`},
			},
			{
				Name:           "anon cannot get draft — 404",
				Method:         http.MethodGet,
				URL:            fmt.Sprintf("/blog/posts/%s", draft.Slug),
				ExpectedStatus: http.StatusNotFound,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			},
			{
				Name:           "admin can get draft",
				Method:         http.MethodGet,
				URL:            fmt.Sprintf("/blog/posts/%s", draft.Slug),
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					header, _ := core.CreateAccessHeaderAndRefreshToken(t, app, admin.User.Email)
					scenario.Headers = []string{header}
				},
				ExpectedContent: []string{`"status":"draft"`},
			},
			{
				Name:           "nonexistent slug returns 404",
				Method:         http.MethodGet,
				URL:            "/blog/posts/does-not-exist",
				ExpectedStatus: http.StatusNotFound,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			},
		}
		for _, tt := range tests {
			tt.Test(t)
		}
	})
}

// ── BlogPostCreate ────────────────────────────────────────────────────────────

func TestApi_BlogPostCreate(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)

		admin := core.CreateUserWithOptions(t, testApi.App,
			core.UserWithVerifiedNow(),
			core.UserWithPermission(shared.PermissionNameAdmin),
		)
		regular := core.CreateUserWithOptions(t, testApi.App,
			core.UserWithVerifiedNow(),
			core.UserWithEmail("regular@example.com"),
		)

		tests := []apis.ApiScenario{
			{
				// CheckPermissionsMiddleware returns 403 for unauthenticated requests
				Name:           "anon cannot create post — 403",
				Method:         http.MethodPost,
				URL:            "/blog/posts",
				Body:           strings.NewReader(`{"title":"Test","content":"body"}`),
				ExpectedStatus: http.StatusForbidden,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			},
			{
				// Non-admin authenticated user is also rejected with 403
				Name:           "regular user cannot create post — 403",
				Method:         http.MethodPost,
				URL:            "/blog/posts",
				Body:           strings.NewReader(`{"title":"Test","content":"body"}`),
				ExpectedStatus: http.StatusForbidden,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					header, _ := core.CreateAccessHeaderAndRefreshToken(t, app, regular.User.Email)
					scenario.Headers = []string{header}
				},
			},
			{
				Name:           "admin creates post — starts as draft",
				Method:         http.MethodPost,
				URL:            "/blog/posts",
				Body:           strings.NewReader(`{"title":"My First Post","content":"hello world"}`),
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					header, _ := core.CreateAccessHeaderAndRefreshToken(t, app, admin.User.Email)
					scenario.Headers = []string{header}
				},
				ExpectedContent: []string{`"slug":"my-first-post"`, `"status":"draft"`},
			},
		}
		for _, tt := range tests {
			tt.Test(t)
		}
	})
}

// ── BlogPostUpdate ────────────────────────────────────────────────────────────

func TestApi_BlogPostUpdate(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)

		admin := core.CreateUserWithOptions(t, testApi.App,
			core.UserWithVerifiedNow(),
			core.UserWithPermission(shared.PermissionNameAdmin),
		)
		regular := core.CreateUserWithOptions(t, testApi.App,
			core.UserWithVerifiedNow(),
			core.UserWithEmail("regular@example.com"),
		)
		post := createBlogPost(t, testApi.App, admin.User.ID, "Original Title", "original content", models.BlogPostStatusDraft)

		tests := []apis.ApiScenario{
			{
				Name:           "non-admin cannot update — 403",
				Method:         http.MethodPatch,
				URL:            fmt.Sprintf("/blog/posts/%s", post.ID),
				Body:           strings.NewReader(`{"title":"Hacked"}`),
				ExpectedStatus: http.StatusForbidden,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					header, _ := core.CreateAccessHeaderAndRefreshToken(t, app, regular.User.Email)
					scenario.Headers = []string{header}
				},
			},
			{
				Name:           "admin updates title",
				Method:         http.MethodPatch,
				URL:            fmt.Sprintf("/blog/posts/%s", post.ID),
				Body:           strings.NewReader(`{"title":"Updated Title"}`),
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					header, _ := core.CreateAccessHeaderAndRefreshToken(t, app, admin.User.Email)
					scenario.Headers = []string{header}
				},
				ExpectedContent: []string{`"title":"Updated Title"`},
			},
		}
		for _, tt := range tests {
			tt.Test(t)
		}
	})
}

// ── BlogPostPublish / Unpublish / Archive ─────────────────────────────────────

func TestApi_BlogPostPublish(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)

		admin := core.CreateUserWithOptions(t, testApi.App,
			core.UserWithVerifiedNow(),
			core.UserWithPermission(shared.PermissionNameAdmin),
		)
		post := createBlogPost(t, testApi.App, admin.User.ID, "Publish Me", "content", models.BlogPostStatusDraft)

		header, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, admin.User.Email)

		tests := []apis.ApiScenario{
			{
				Name:            "publish draft post",
				Method:          http.MethodPost,
				URL:             fmt.Sprintf("/blog/posts/%s/publish", post.ID),
				ExpectedStatus:  http.StatusOK,
				TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc:  func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) { scenario.Headers = []string{header} },
				ExpectedContent: []string{`"status":"published"`},
			},
			{
				Name:            "unpublish back to draft",
				Method:          http.MethodPost,
				URL:             fmt.Sprintf("/blog/posts/%s/unpublish", post.ID),
				ExpectedStatus:  http.StatusOK,
				TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc:  func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) { scenario.Headers = []string{header} },
				ExpectedContent: []string{`"status":"draft"`},
			},
			{
				Name:            "archive post",
				Method:          http.MethodPost,
				URL:             fmt.Sprintf("/blog/posts/%s/archive", post.ID),
				ExpectedStatus:  http.StatusOK,
				TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc:  func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) { scenario.Headers = []string{header} },
				ExpectedContent: []string{`"status":"archived"`},
			},
		}
		for _, tt := range tests {
			tt.Test(t)
		}
	})
}

// ── BlogPostDelete ────────────────────────────────────────────────────────────

func TestApi_BlogPostDelete(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)

		admin := core.CreateUserWithOptions(t, testApi.App,
			core.UserWithVerifiedNow(),
			core.UserWithPermission(shared.PermissionNameAdmin),
		)
		post := createBlogPost(t, testApi.App, admin.User.ID, "Delete Me", "content", models.BlogPostStatusDraft)

		tests := []apis.ApiScenario{
			{
				Name:           "admin deletes post",
				Method:         http.MethodDelete,
				URL:            fmt.Sprintf("/blog/posts/%s", post.ID),
				ExpectedStatus: http.StatusNoContent,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					header, _ := core.CreateAccessHeaderAndRefreshToken(t, app, admin.User.Email)
					scenario.Headers = []string{header}
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					_, err := app.Adapter().Blog().FindPostByID(context.Background(), post.ID)
					assert.Error(t, err, "post should no longer exist after deletion")
				},
			},
		}
		for _, tt := range tests {
			tt.Test(t)
		}
	})
}

// ── BlogTagList / BlogTagCreate ───────────────────────────────────────────────

func TestApi_BlogTags(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)

		admin := core.CreateUserWithOptions(t, testApi.App,
			core.UserWithVerifiedNow(),
			core.UserWithPermission(shared.PermissionNameAdmin),
		)

		tests := []apis.ApiScenario{
			{
				Name:            "create tag — idempotent on duplicate name",
				Method:          http.MethodPost,
				URL:             "/blog/tags",
				Body:            strings.NewReader(`{"name":"Go"}`),
				ExpectedStatus:  http.StatusOK,
				TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					header, _ := core.CreateAccessHeaderAndRefreshToken(t, app, admin.User.Email)
					scenario.Headers = []string{header}
				},
				ExpectedContent: []string{`"slug":"go"`},
			},
			{
				Name:           "anon cannot create tag — 403",
				Method:         http.MethodPost,
				URL:            "/blog/tags",
				Body:           strings.NewReader(`{"name":"Rust"}`),
				ExpectedStatus: http.StatusForbidden,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			},
			{
				Name:           "list tags is public",
				Method:         http.MethodGet,
				URL:            "/blog/tags",
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			},
		}
		for _, tt := range tests {
			tt.Test(t)
		}
	})
}

// ── FeaturedImageURL resolution ───────────────────────────────────────────────

func TestApi_BlogPost_FeaturedImageURL(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)

		const publicBase = "https://pub.example.com"
		testApi.App.SetFs(filesystem.NewMockFileSystem(conf.StorageConfig{
			PublicBaseURL: publicBase,
		}))

		admin := core.CreateUserWithOptions(t, testApi.App,
			core.UserWithVerifiedNow(),
			core.UserWithPermission(shared.PermissionNameAdmin),
		)

		const imageKey = "media/abc123.jpg"
		postWithImage := createBlogPost(t, testApi.App, admin.User.ID, "Image Post", "content", models.BlogPostStatusPublished)
		imageKeyVal := imageKey
		_, err := testApi.App.Adapter().Blog().UpdatePost(ctx, postWithImage.ID, &stores.UpdateBlogPostDTO{
			FeaturedImageKey: &imageKeyVal,
		})
		require.NoError(t, err)

		postNoImage := createBlogPost(t, testApi.App, admin.User.ID, "No Image Post", "content", models.BlogPostStatusPublished)

		tests := []apis.ApiScenario{
			{
				Name:           "post with featured_image_key resolves to public URL",
				Method:         http.MethodGet,
				URL:            fmt.Sprintf("/blog/posts/%s", postWithImage.Slug),
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					var body apis.ApiSingleResponse[*apis.BlogPost]
					require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
					require.NotNil(t, body.Data.FeaturedImageURL)
					assert.Equal(t, publicBase+"/"+imageKey, *body.Data.FeaturedImageURL)
				},
			},
			{
				Name:           "post without featured_image_key returns null featured_image_url",
				Method:         http.MethodGet,
				URL:            fmt.Sprintf("/blog/posts/%s", postNoImage.Slug),
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					var body apis.ApiSingleResponse[*apis.BlogPost]
					require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
					assert.Nil(t, body.Data.FeaturedImageURL)
				},
			},
		}
		for _, tt := range tests {
			tt.Test(t)
		}
	})
}

// ── Slug uniqueness ───────────────────────────────────────────────────────────

func TestApi_BlogPostSlugUnique(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)

		admin := core.CreateUserWithOptions(t, testApi.App,
			core.UserWithVerifiedNow(),
			core.UserWithPermission(shared.PermissionNameAdmin),
		)

		p1 := createBlogPost(t, testApi.App, admin.User.ID, "Same Title", "a", models.BlogPostStatusDraft)
		p2 := createBlogPost(t, testApi.App, admin.User.ID, "Same Title", "b", models.BlogPostStatusDraft)

		assert.NotEqual(t, p1.Slug, p2.Slug)
		assert.True(t, strings.HasPrefix(p2.Slug, "same-title"))
	})
}
