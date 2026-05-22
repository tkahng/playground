package apis_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
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
	"github.com/tkahng/playground/internal/tools/filesystem"
)

// skipContainer skips the test when -short is passed (container tests are slow).
func skipContainer(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping container test in short mode")
	}
}

// buildMultipart builds a multipart/form-data body containing one file field.
func buildMultipart(t testing.TB, fieldName, filename string, content []byte) (*bytes.Buffer, string) {
	t.Helper()
	buf := &bytes.Buffer{}
	w := multipart.NewWriter(buf)
	fw, err := w.CreateFormFile(fieldName, filename)
	require.NoError(t, err)
	_, err = io.Copy(fw, bytes.NewReader(content))
	require.NoError(t, err)
	require.NoError(t, w.Close())
	return buf, w.FormDataContentType()
}

// buildMedium constructs a Medium model from a filesystem FileDto and a user ID.
func buildMedium(userID uuid.UUID, dto *filesystem.FileDto) *models.Medium {
	pub := dto.PublicURL
	return &models.Medium{
		UserID:           &userID,
		StorageKey:       dto.StorageKey,
		PublicURL:        &pub,
		MimeType:         dto.MimeType,
		Size:             dto.Size,
		OriginalFilename: dto.OriginalName,
		Extension:        dto.Extension,
		Disk:             dto.Disk,
		Directory:        dto.Directory,
		Filename:         dto.Filename,
	}
}

// ── UploadMedia ───────────────────────────────────────────────────────────────

func TestApi_UploadMedia(t *testing.T) {
	skipContainer(t)

	filesystem.WithMinioContainer(t, func(fsCtx context.Context, fs filesystem.FileSystem, cfg conf.StorageConfig) {
		database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
			testApi := apis.SetupApi(t, ctx, db)
			testApi.App.SetFs(fs)

			user := core.CreateUserWithOptions(t, testApi.App, core.UserWithVerifiedNow())
			header, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, user.User.Email)

			content := []byte("hello from minio integration test")
			body, ct := buildMultipart(t, "files", "hello.txt", content)

			tests := []apis.ApiScenario{
				{
					Name:           "authenticated user uploads file — persisted in media store",
					Method:         http.MethodPost,
					URL:            "/media",
					Body:           body,
					ExpectedStatus: http.StatusNoContent,
					TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
					BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
						scenario.Headers = []string{header, "Content-Type: " + ct}
					},
					AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
						medias, err := app.Adapter().Media().FindMedia(ctx, nil)
						require.NoError(t, err)
						require.Len(t, medias, 1)
						assert.Equal(t, cfg.BucketName, medias[0].Disk)
						assert.Equal(t, "hello.txt", medias[0].OriginalFilename)
					},
				},
			}
			for _, tt := range tests {
				tt.Test(t)
			}
		})
	})
}

func TestApi_UploadMedia_Unauthenticated(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		body, ct := buildMultipart(t, "files", "file.txt", []byte("data"))

		tests := []apis.ApiScenario{
			{
				// Auth middleware rejects unauthenticated requests with 401 before the handler runs.
				Name:            "unauthenticated upload returns 401",
				Method:          http.MethodPost,
				URL:             "/media",
				Body:            body,
				ExpectedStatus:  http.StatusUnauthorized,
				ExpectedContent: []string{`"status":401`},
				TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					scenario.Headers = []string{"Content-Type: " + ct}
				},
			},
		}
		for _, tt := range tests {
			tt.Test(t)
		}
	})
}

// ── GetMedia ──────────────────────────────────────────────────────────────────

func TestApi_GetMedia(t *testing.T) {
	skipContainer(t)

	filesystem.WithMinioContainer(t, func(fsCtx context.Context, fs filesystem.FileSystem, cfg conf.StorageConfig) {
		database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
			testApi := apis.SetupApi(t, ctx, db)
			testApi.App.SetFs(fs)

			user := core.CreateUserWithOptions(t, testApi.App, core.UserWithVerifiedNow())
			header, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, user.User.Email)

			// Upload a real file directly to MinIO and record it in the DB.
			content := []byte("get-media integration test content")
			dto, err := fs.PutFileFromBytes(fsCtx, content, "test-get.txt")
			require.NoError(t, err)

			medium, err := testApi.App.Adapter().Media().CreateMedia(ctx, buildMedium(user.User.ID, dto))
			require.NoError(t, err)

			tests := []apis.ApiScenario{
				{
					Name:           "get media returns record with stable public URL",
					Method:         http.MethodGet,
					URL:            fmt.Sprintf("/media/%s", medium.ID),
					ExpectedStatus: http.StatusOK,
					TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
					BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
						scenario.Headers = []string{header}
					},
					AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
						var result struct {
							ID       uuid.UUID `json:"id"`
							Filename string    `json:"filename"`
							URL      string    `json:"url"`
						}
						require.NoError(t, json.Unmarshal(res.Body.Bytes(), &result))
						assert.Equal(t, medium.ID, result.ID)
						assert.NotEmpty(t, result.URL)
						// Public URL must be credential-free and serve the uploaded bytes.
						resp, err := http.Get(result.URL) //nolint:noctx
						require.NoError(t, err)
						defer resp.Body.Close()
						require.Equal(t, http.StatusOK, resp.StatusCode)
						got, err := io.ReadAll(resp.Body)
						require.NoError(t, err)
						assert.Equal(t, content, got)
					},
				},
				{
					Name:            "get nonexistent media ID returns 404",
					Method:          http.MethodGet,
					URL:             fmt.Sprintf("/media/%s", uuid.New()),
					ExpectedStatus:  http.StatusNotFound,
					ExpectedContent: []string{`"status":404`},
					TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
					BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
						scenario.Headers = []string{header}
					},
				},
			}
			for _, tt := range tests {
				tt.Test(t)
			}
		})
	})
}

// ── UpdateMedia ───────────────────────────────────────────────────────────────

func TestApi_UpdateMedia_Ownership(t *testing.T) {
	skipContainer(t)

	filesystem.WithMinioContainer(t, func(fsCtx context.Context, fs filesystem.FileSystem, cfg conf.StorageConfig) {
		database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
			testApi := apis.SetupApi(t, ctx, db)
			testApi.App.SetFs(fs)

			owner := core.CreateUserWithOptions(t, testApi.App,
				core.UserWithEmail("media-owner@test.local"), core.UserWithVerifiedNow())
			other := core.CreateUserWithOptions(t, testApi.App,
				core.UserWithEmail("media-other@test.local"), core.UserWithVerifiedNow())
			admin := core.CreateUserWithOptions(t, testApi.App,
				core.UserWithEmail("media-admin@test.local"),
				core.UserWithVerifiedNow(),
				core.UserWithPermission(shared.PermissionNameAdmin),
			)

			ownerHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, owner.User.Email)
			otherHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, other.User.Email)
			adminHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, admin.User.Email)

			dto, err := fs.PutFileFromBytes(fsCtx, []byte("ownership test"), "ownership.txt")
			require.NoError(t, err)
			medium, err := testApi.App.Adapter().Media().CreateMedia(ctx, buildMedium(owner.User.ID, dto))
			require.NoError(t, err)

			altText := "updated alt text"
			body, _ := json.Marshal(map[string]any{"alt_text": altText})
			url := fmt.Sprintf("/media/%s", medium.ID)

			tests := []apis.ApiScenario{
				{
					Name:            "non-owner cannot update another user's media",
					Method:          http.MethodPatch,
					URL:             url,
					Body:            bytes.NewReader(body),
					ExpectedStatus:  http.StatusForbidden,
					ExpectedContent: []string{`"status":403`},
					TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
					BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
						scenario.Headers = []string{otherHeader, "Content-Type: application/json"}
					},
				},
				{
					Name:           "owner can update their own media",
					Method:         http.MethodPatch,
					URL:            url,
					Body:           bytes.NewReader(body),
					ExpectedStatus: http.StatusOK,
					TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
					BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
						scenario.Headers = []string{ownerHeader, "Content-Type: application/json"}
					},
					AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
						var result apis.Media
						require.NoError(t, json.Unmarshal(res.Body.Bytes(), &result))
						require.NotNil(t, result.AltText)
						assert.Equal(t, altText, *result.AltText)
					},
				},
				{
					Name:           "admin can update any user's media",
					Method:         http.MethodPatch,
					URL:            url,
					Body:           bytes.NewReader(body),
					ExpectedStatus: http.StatusOK,
					TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
					BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
						scenario.Headers = []string{adminHeader, "Content-Type: application/json"}
					},
					AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
						var result apis.Media
						require.NoError(t, json.Unmarshal(res.Body.Bytes(), &result))
						assert.Equal(t, medium.ID, result.ID)
					},
				},
			}
			for _, tt := range tests {
				tt.Test(t)
			}
		})
	})
}

// ── DeleteMedia ───────────────────────────────────────────────────────────────

func TestApi_DeleteMedia_Ownership(t *testing.T) {
	skipContainer(t)

	filesystem.WithMinioContainer(t, func(fsCtx context.Context, fs filesystem.FileSystem, cfg conf.StorageConfig) {
		database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
			testApi := apis.SetupApi(t, ctx, db)
			testApi.App.SetFs(fs)

			owner := core.CreateUserWithOptions(t, testApi.App,
				core.UserWithEmail("del-owner@test.local"), core.UserWithVerifiedNow())
			other := core.CreateUserWithOptions(t, testApi.App,
				core.UserWithEmail("del-other@test.local"), core.UserWithVerifiedNow())
			admin := core.CreateUserWithOptions(t, testApi.App,
				core.UserWithEmail("del-admin@test.local"),
				core.UserWithVerifiedNow(),
				core.UserWithPermission(shared.PermissionNameAdmin),
			)

			ownerHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, owner.User.Email)
			otherHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, other.User.Email)
			adminHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, admin.User.Email)

			// Helper to create a fresh media record for each sub-test.
			newMedium := func() *models.Medium {
				dto, err := fs.PutFileFromBytes(fsCtx, []byte("delete ownership test"), "delete-test.txt")
				require.NoError(t, err)
				m, err := testApi.App.Adapter().Media().CreateMedia(ctx, buildMedium(owner.User.ID, dto))
				require.NoError(t, err)
				return m
			}

			medium := newMedium()
			adminTarget := newMedium()

			tests := []apis.ApiScenario{
				{
					Name:            "non-owner cannot delete another user's media",
					Method:          http.MethodDelete,
					URL:             fmt.Sprintf("/media/%s", medium.ID),
					ExpectedStatus:  http.StatusForbidden,
					ExpectedContent: []string{`"status":403`},
					TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
					BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
						scenario.Headers = []string{otherHeader}
					},
					AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
						// Record must still exist.
						got, err := app.Adapter().Media().FindMediaByID(ctx, medium.ID)
						require.NoError(t, err)
						assert.NotNil(t, got)
					},
				},
				{
					Name:           "owner can delete their own media",
					Method:         http.MethodDelete,
					URL:            fmt.Sprintf("/media/%s", medium.ID),
					ExpectedStatus: http.StatusNoContent,
					TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
					BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
						scenario.Headers = []string{ownerHeader}
					},
					AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
						got, err := app.Adapter().Media().FindMediaByID(ctx, medium.ID)
						require.NoError(t, err)
						assert.Nil(t, got)
					},
				},
				{
					Name:           "admin can delete any user's media",
					Method:         http.MethodDelete,
					URL:            fmt.Sprintf("/media/%s", adminTarget.ID),
					ExpectedStatus: http.StatusNoContent,
					TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
					BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
						scenario.Headers = []string{adminHeader}
					},
				},
			}
			for _, tt := range tests {
				tt.Test(t)
			}
		})
	})
}

// ── MediaList ─────────────────────────────────────────────────────────────────

func TestApi_MediaList(t *testing.T) {
	skipContainer(t)

	filesystem.WithMinioContainer(t, func(fsCtx context.Context, fs filesystem.FileSystem, cfg conf.StorageConfig) {
		database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
			testApi := apis.SetupApi(t, ctx, db)
			testApi.App.SetFs(fs)

			user := core.CreateUserWithOptions(t, testApi.App, core.UserWithVerifiedNow())
			header, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, user.User.Email)

			// Upload two files and persist them.
			for _, name := range []string{"alpha.txt", "beta.txt"} {
				dto, err := fs.PutFileFromBytes(fsCtx, []byte("content-"+name), name)
				require.NoError(t, err)
				_, err = testApi.App.Adapter().Media().CreateMedia(ctx, buildMedium(user.User.ID, dto))
				require.NoError(t, err)
			}

			tests := []apis.ApiScenario{
				{
					Name:           "list media returns all files with public URLs",
					Method:         http.MethodGet,
					URL:            "/media",
					ExpectedStatus: http.StatusOK,
					TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
					BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
						scenario.Headers = []string{header}
					},
					AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
						var result apis.ApiPaginatedResponse[*apis.Media]
						require.NoError(t, json.Unmarshal(res.Body.Bytes(), &result))
						assert.Equal(t, int64(2), result.Meta.Total)
						for _, m := range result.Data {
							assert.NotEmpty(t, m.URL, "every item must have a public URL")
						}
					},
				},
			}
			for _, tt := range tests {
				tt.Test(t)
			}
		})
	})
}
