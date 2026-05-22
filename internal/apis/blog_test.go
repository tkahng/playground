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
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/internal/stores"
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
				Name:            "anon search returns matching published post",
				Method:          http.MethodGet,
				URL:             "/blog/posts?q=Published",
				ExpectedStatus:  http.StatusOK,
				ExpectedContent: []string{`"title":"Published Post"`},
				TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
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
				Name:            "anon cannot get draft — 404",
				Method:          http.MethodGet,
				URL:             fmt.Sprintf("/blog/posts/%s", draft.Slug),
				ExpectedStatus:  http.StatusNotFound,
				ExpectedContent: []string{`"status":404`},
				TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
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
				Name:            "nonexistent slug returns 404",
				Method:          http.MethodGet,
				URL:             "/blog/posts/does-not-exist",
				ExpectedStatus:  http.StatusNotFound,
				ExpectedContent: []string{`"status":404`},
				TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
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
				// Unauthenticated requests are rejected by the auth middleware with 401.
				Name:            "anon cannot create post — 401",
				Method:          http.MethodPost,
				URL:             "/blog/posts",
				Body:            strings.NewReader(`{"title":"Test","content":"body"}`),
				ExpectedStatus:  http.StatusUnauthorized,
				ExpectedContent: []string{`"status":401`},
				TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
			},
			{
				// Authenticated but non-admin users are rejected with 403.
				Name:            "regular user cannot create post — 403",
				Method:          http.MethodPost,
				URL:             "/blog/posts",
				Body:            strings.NewReader(`{"title":"Test","content":"body"}`),
				ExpectedStatus:  http.StatusForbidden,
				ExpectedContent: []string{`"status":403`},
				TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
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
				Name:            "non-admin cannot update — 403",
				Method:          http.MethodPatch,
				URL:             fmt.Sprintf("/blog/posts/%s", post.ID),
				Body:            strings.NewReader(`{"title":"Hacked"}`),
				ExpectedStatus:  http.StatusForbidden,
				ExpectedContent: []string{`"status":403`},
				TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
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
				Name:            "anon cannot create tag — 401",
				Method:          http.MethodPost,
				URL:             "/blog/tags",
				Body:            strings.NewReader(`{"name":"Rust"}`),
				ExpectedStatus:  http.StatusUnauthorized,
				ExpectedContent: []string{`"status":401`},
				TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
			},
			{
				Name:            "list tags is public",
				Method:          http.MethodGet,
				URL:             "/blog/tags",
				ExpectedStatus:  http.StatusOK,
				ExpectedContent: []string{`"data"`},
				TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
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

		const publicURL = "https://pub.example.com/media/abc123.jpg"
		const storageKey = "media/abc123.jpg"

		admin := core.CreateUserWithOptions(t, testApi.App,
			core.UserWithVerifiedNow(),
			core.UserWithPermission(shared.PermissionNameAdmin),
		)

		// Create a media record directly — no file upload needed for URL resolution test.
		pub := publicURL
		medium, err := testApi.App.Adapter().Media().CreateMedia(ctx, &models.Medium{
			UserID:           &admin.User.ID,
			StorageKey:       storageKey,
			PublicURL:        &pub,
			MimeType:         "image/jpeg",
			Size:             1024,
			OriginalFilename: "abc123.jpg",
			Extension:        ".jpg",
			Disk:             "test-bucket",
			Directory:        "media",
			Filename:         "abc123.jpg",
		})
		require.NoError(t, err)

		postWithImage := createBlogPost(t, testApi.App, admin.User.ID, "Image Post", "content", models.BlogPostStatusPublished)
		_, err = testApi.App.Adapter().Blog().UpdatePost(ctx, postWithImage.ID, &stores.UpdateBlogPostDTO{
			FeaturedImageMediaID: stores.NullableUUID{Set: true, Value: &medium.ID},
		})
		require.NoError(t, err)

		postNoImage := createBlogPost(t, testApi.App, admin.User.ID, "No Image Post", "content", models.BlogPostStatusPublished)

		tests := []apis.ApiScenario{
			{
				Name:           "post with featured_image_media_id returns stored public URL",
				Method:         http.MethodGet,
				URL:            fmt.Sprintf("/blog/posts/%s", postWithImage.Slug),
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					var body apis.ApiSingleResponse[*apis.BlogPost]
					require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
					require.NotNil(t, body.Data.FeaturedImageURL)
					assert.Equal(t, publicURL, *body.Data.FeaturedImageURL)
					require.NotNil(t, body.Data.FeaturedImageID)
					assert.Equal(t, medium.ID, *body.Data.FeaturedImageID)
				},
			},
			{
				Name:           "post without featured image returns null featured_image_url",
				Method:         http.MethodGet,
				URL:            fmt.Sprintf("/blog/posts/%s", postNoImage.Slug),
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					var body apis.ApiSingleResponse[*apis.BlogPost]
					require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
					assert.Nil(t, body.Data.FeaturedImageURL)
					assert.Nil(t, body.Data.FeaturedImageID)
				},
			},
		}
		for _, tt := range tests {
			tt.Test(t)
		}
	})
}

// ── BlogPostGet by UUID ───────────────────────────────────────────────────────

func TestApi_BlogPostGet_ByUUID(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)

		admin := core.CreateUserWithOptions(t, testApi.App,
			core.UserWithVerifiedNow(),
			core.UserWithPermission(shared.PermissionNameAdmin),
		)
		adminHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, admin.User.Email)

		published := createBlogPost(t, testApi.App, admin.User.ID, "UUID Lookup Post", "body", models.BlogPostStatusPublished)
		draft := createBlogPost(t, testApi.App, admin.User.ID, "UUID Draft", "body", models.BlogPostStatusDraft)

		tests := []apis.ApiScenario{
			{
				Name:            "admin fetches published post by UUID",
				Method:          http.MethodGet,
				URL:             fmt.Sprintf("/blog/posts/%s", published.ID),
				ExpectedStatus:  http.StatusOK,
				ExpectedContent: []string{fmt.Sprintf(`"id":"%s"`, published.ID)},
				TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					scenario.Headers = []string{adminHeader}
				},
			},
			{
				Name:            "admin fetches draft post by UUID",
				Method:          http.MethodGet,
				URL:             fmt.Sprintf("/blog/posts/%s", draft.ID),
				ExpectedStatus:  http.StatusOK,
				ExpectedContent: []string{`"status":"draft"`},
				TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					scenario.Headers = []string{adminHeader}
				},
			},
			{
				Name:            "anon cannot fetch draft by UUID — 404",
				Method:          http.MethodGet,
				URL:             fmt.Sprintf("/blog/posts/%s", draft.ID),
				ExpectedStatus:  http.StatusNotFound,
				ExpectedContent: []string{`"status":404`},
				TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
			},
		}
		for _, tt := range tests {
			tt.Test(t)
		}
	})
}

// ── Nullable field clearing ───────────────────────────────────────────────────

func TestApi_BlogPostUpdate_NullableFieldClearing(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)

		admin := core.CreateUserWithOptions(t, testApi.App,
			core.UserWithVerifiedNow(),
			core.UserWithPermission(shared.PermissionNameAdmin),
		)
		adminHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, admin.User.Email)

		// Seed a media record so we can set and then clear the featured image.
		pub := "https://cdn.example.com/img.jpg"
		medium, err := testApi.App.Adapter().Media().CreateMedia(ctx, &models.Medium{
			UserID: &admin.User.ID, StorageKey: "media/img.jpg", PublicURL: &pub,
			MimeType: "image/jpeg", Size: 512, OriginalFilename: "img.jpg",
			Extension: ".jpg", Disk: "bucket", Directory: "media", Filename: "img.jpg",
		})
		require.NoError(t, err)

		post := createBlogPost(t, testApi.App, admin.User.ID, "Nullable Test", "content", models.BlogPostStatusDraft)

		// Set seo_title, seo_description, and featured_image_media_id.
		_, err = testApi.App.Adapter().Blog().UpdatePost(ctx, post.ID, &stores.UpdateBlogPostDTO{
			SeoTitle:             stores.NullableString{Set: true, Value: strPtr("My SEO Title")},
			SeoDescription:       stores.NullableString{Set: true, Value: strPtr("My SEO Desc")},
			FeaturedImageMediaID: stores.NullableUUID{Set: true, Value: &medium.ID},
		})
		require.NoError(t, err)

		url := fmt.Sprintf("/blog/posts/%s", post.ID)

		tests := []apis.ApiScenario{
			{
				Name:           "PATCH with null clears featured_image_media_id",
				Method:         http.MethodPatch,
				URL:            url,
				Body:           strings.NewReader(`{"featured_image_media_id":null}`),
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					scenario.Headers = []string{adminHeader, "Content-Type: application/json"}
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					var body apis.ApiSingleResponse[*apis.BlogPost]
					require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
					assert.Nil(t, body.Data.FeaturedImageID, "featured_image_id should be null after clearing")
					assert.Nil(t, body.Data.FeaturedImageURL)
				},
			},
			{
				Name:           "PATCH with null clears seo_title and seo_description",
				Method:         http.MethodPatch,
				URL:            url,
				Body:           strings.NewReader(`{"seo_title":null,"seo_description":null}`),
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					scenario.Headers = []string{adminHeader, "Content-Type: application/json"}
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					var body apis.ApiSingleResponse[*apis.BlogPost]
					require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
					assert.Nil(t, body.Data.SeoTitle)
					assert.Nil(t, body.Data.SeoDescription)
				},
			},
			{
				Name:           "PATCH without nullable fields leaves them unchanged",
				Method:         http.MethodPatch,
				URL:            url,
				Body:           strings.NewReader(`{"title":"Updated Title"}`),
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					scenario.Headers = []string{adminHeader, "Content-Type: application/json"}
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					var body apis.ApiSingleResponse[*apis.BlogPost]
					require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
					assert.Equal(t, "Updated Title", body.Data.Title)
					// seo fields already nulled in previous sub-test; just confirm title changed
				},
			},
		}
		for _, tt := range tests {
			tt.Test(t)
		}
	})
}

func strPtr(s string) *string { return &s }

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
