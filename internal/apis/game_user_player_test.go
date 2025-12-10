package apis_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stretchr/testify/assert"
	"github.com/tkahng/playground/internal/apis"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/test"
	"github.com/tkahng/playground/internal/tools/types"
)

func Test_PutUserPlayer_Success_SetDisplayName(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := SetupApi(t, ctx, db)
		scenario := &ApiScenario{
			Name:           "success put user player",
			Method:         http.MethodPut,
			URL:            "/games/players",
			ExpectedStatus: http.StatusOK,
			TestAppFactory: func(t testing.TB) *TestApi {
				return testApi
			},
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
				userInfo := core.CreateUserWithOptions(t, testApi.App)
				scenario.Store.Set("user_info", userInfo)
				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, userInfo.User.Email)
				scenario.Headers = []string{tokenHeader}
				dto := &apis.GamePutUserPlayerArgs{
					DisplayName: types.Pointer("test_display_name"),
				}
				data, err := json.Marshal(dto)
				if err != nil {
					t.Errorf("Error marshalling input: %v", err)
				}
				scenario.Body = strings.NewReader(string(data))
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
				userInfo, ok := scenario.Store.Get("user_info").(*models.UserInfo)
				if !ok {
					t.Fatal("user info not found")
				}
				result := test.MustUnMarshal[apis.ApiSingleResponse[*apis.ApiPlayer]](t, res.Body.Bytes())
				assert.NotNil(t, result)
				userId := *result.Data.UserID
				assert.Equal(t, userInfo.User.ID, userId)
				assert.Equal(t, userInfo.User.Email, result.Data.Email)
				assert.Equal(t, "test_display_name", *result.Data.DisplayName)
			},
		}
		scenario.Test(t)
	})
}
func Test_PutUserPlayer_Success_SetDisplayNameNil(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := SetupApi(t, ctx, db)
		scenario := &ApiScenario{
			Name:           "success put user player",
			Method:         http.MethodPut,
			URL:            "/games/players",
			ExpectedStatus: http.StatusOK,
			TestAppFactory: func(t testing.TB) *TestApi {
				return testApi
			},
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
				userInfo := core.CreateUserWithOptions(t, testApi.App)
				scenario.Store.Set("user_info", userInfo)
				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, userInfo.User.Email)
				scenario.Headers = []string{tokenHeader}
				dto := &apis.GamePutUserPlayerArgs{
					DisplayName: nil,
				}
				data, err := json.Marshal(dto)
				if err != nil {
					t.Errorf("Error marshalling input: %v", err)
				}
				scenario.Body = strings.NewReader(string(data))
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
				userInfo, ok := scenario.Store.Get("user_info").(*models.UserInfo)
				if !ok {
					t.Fatal("user info not found")
				}
				result := test.MustUnMarshal[apis.ApiSingleResponse[*apis.ApiPlayer]](t, res.Body.Bytes())
				assert.NotNil(t, result)
				userId := *result.Data.UserID
				assert.Equal(t, userInfo.User.ID, userId)
				assert.Equal(t, userInfo.User.Email, result.Data.Email)
				assert.Nil(t, result.Data.DisplayName)
			},
		}
		scenario.Test(t)
	})
}
func Test_PutUserPlayer_Fail_EmptyDisplayName(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := SetupApi(t, ctx, db)
		scenario := &ApiScenario{
			Name:           "success put user player",
			Method:         http.MethodPut,
			URL:            "/games/players",
			ExpectedStatus: http.StatusUnprocessableEntity,
			TestAppFactory: func(t testing.TB) *TestApi {
				return testApi
			},
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
				userInfo := core.CreateUserWithOptions(t, testApi.App)
				scenario.Store.Set("user_info", userInfo)
				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, userInfo.User.Email)
				scenario.Headers = []string{tokenHeader}
				dto := &apis.GamePutUserPlayerArgs{
					DisplayName: types.Pointer(""),
				}
				data, err := json.Marshal(dto)
				if err != nil {
					t.Errorf("Error marshalling input: %v", err)
				}
				scenario.Body = strings.NewReader(string(data))
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[huma.ErrorModel](t, res.Body.Bytes())
				assert.Equal(t, result.Status, 422)
				assert.Equal(t, result.Detail, "validation failed")
				assert.Equal(t, result.Errors[0].Message, "expected length >= 1")
			},
		}
		scenario.Test(t)
	})
}
func Test_GetUserPlayer_Success_HasPlayer(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := SetupApi(t, ctx, db)
		scenario := &ApiScenario{
			Name:           "success put user player",
			Method:         http.MethodGet,
			URL:            "/games/players",
			ExpectedStatus: http.StatusOK,
			TestAppFactory: func(t testing.TB) *TestApi {
				return testApi
			},
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
				userInfo := core.CreateUserWithOptions(t, testApi.App)
				scenario.Store.Set("user_info", userInfo)
				player, err := app.Adapter().Gaming().CreatePlayer(ctx, &models.Player{
					Email:  userInfo.User.Email,
					UserID: &userInfo.User.ID,
				})
				if err != nil {
					t.Fatal(err)
				}
				scenario.Store.Set("player", player)
				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, userInfo.User.Email)
				scenario.Headers = []string{tokenHeader}
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
				player, ok := scenario.Store.Get("player").(*models.Player)
				if !ok {
					t.Fatal("user info not found")
				}
				result := test.MustUnMarshal[apis.ApiSingleResponse[*apis.ApiPlayer]](t, res.Body.Bytes())
				assert.NotNil(t, result)
				userId := *result.Data.UserID
				assert.Equal(t, *player.UserID, userId)
				assert.Equal(t, player.Email, result.Data.Email)
			},
		}
		scenario.Test(t)
	})
}
func Test_GetUserPlayer_Success_HasNoPlayer(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := SetupApi(t, ctx, db)
		scenario := &ApiScenario{
			Name:           "success put user player",
			Method:         http.MethodGet,
			URL:            "/games/players",
			ExpectedStatus: http.StatusOK,
			TestAppFactory: func(t testing.TB) *TestApi {
				return testApi
			},
			ExpectedContent: []string{"null"},
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
				userInfo := core.CreateUserWithOptions(t, testApi.App)
				scenario.Store.Set("user_info", userInfo)
				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, userInfo.User.Email)
				scenario.Headers = []string{tokenHeader}
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[apis.ApiSingleResponse[*apis.ApiPlayer]](t, res.Body.Bytes())
				assert.Nil(t, result.Data)
			},
		}
		scenario.Test(t)
	})
}
