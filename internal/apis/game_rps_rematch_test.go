//go:build integration

package apis_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tkahng/playground/internal/apis"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/services"
	"github.com/tkahng/playground/internal/test"
)

// setupCompletedGame creates two registered players and a completed RPS game.
func setupCompletedGame(t testing.TB, testApi *apis.TestApi) (host, guest *models.Player, game *services.RpsGameWithParticipants) {
	t.Helper()
	host = core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
	guest = core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
	pending := core.MustCreateGame(t, testApi.App, host.ID, guest.ID, models.RpsParticipantMoveRock)
	game = core.MustCompleteGame(t, testApi.App, pending, models.RpsParticipantMoveScissors)
	return
}

// --- POST /games/rps/{game-id}/rematch ---

func Test_RequestRematch_Success(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		host, _, game := setupCompletedGame(t, testApi)

		scenario := &apis.ApiScenario{
			Name:           "host requests rematch",
			Method:         http.MethodPost,
			URL:            fmt.Sprintf("/games/rps/%s/rematch", game.RpsGame.ID),
			ExpectedStatus: http.StatusOK,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				header, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, host.Email)
				scenario.Headers = []string{header}
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[apis.ApiSingleResponse[*apis.RpsRematchRequest]](t, res.Body.Bytes())
				require.NotNil(t, result.Data)
				assert.Equal(t, game.RpsGame.ID.String(), result.Data.OriginalGameID.String())
				assert.Equal(t, string(apis.RpsRematchStatusPending), string(result.Data.Status))
				assert.Nil(t, result.Data.NewGameID)
			},
		}
		scenario.Test(t)
	})
}

func Test_RequestRematch_GuestCanAlsoRequest(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		_, guest, game := setupCompletedGame(t, testApi)

		scenario := &apis.ApiScenario{
			Name:            "guest requests rematch",
			Method:          http.MethodPost,
			URL:             fmt.Sprintf("/games/rps/%s/rematch", game.RpsGame.ID),
			ExpectedStatus:  http.StatusOK,
			ExpectedContent: []string{`"status":"pending"`},
			TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				header, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, guest.Email)
				scenario.Headers = []string{header}
			},
		}
		scenario.Test(t)
	})
}

func Test_RequestRematch_NonParticipantForbidden(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		_, _, game := setupCompletedGame(t, testApi)
		outsider := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		scenario := &apis.ApiScenario{
			Name:            "outsider forbidden",
			Method:          http.MethodPost,
			URL:             fmt.Sprintf("/games/rps/%s/rematch", game.RpsGame.ID),
			ExpectedStatus:  http.StatusForbidden,
			ExpectedContent: []string{"Forbidden"},
			TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				header, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, outsider.Email)
				scenario.Headers = []string{header}
			},
		}
		scenario.Test(t)
	})
}

func Test_RequestRematch_ConflictOnDuplicate(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		host, _, game := setupCompletedGame(t, testApi)

		header, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, host.Email)

		// First request succeeds.
		first := &apis.ApiScenario{
			Name:            "first request",
			Method:          http.MethodPost,
			URL:             fmt.Sprintf("/games/rps/%s/rematch", game.RpsGame.ID),
			ExpectedStatus:  http.StatusOK,
			ExpectedContent: []string{`"status":"pending"`},
			TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				scenario.Headers = []string{header}
			},
		}
		first.Test(t)

		// Second request conflicts.
		second := &apis.ApiScenario{
			Name:            "duplicate request",
			Method:          http.MethodPost,
			URL:             fmt.Sprintf("/games/rps/%s/rematch", game.RpsGame.ID),
			ExpectedStatus:  http.StatusConflict,
			ExpectedContent: []string{"Conflict"},
			TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				scenario.Headers = []string{header}
			},
		}
		second.Test(t)
	})
}

// --- POST /games/rps/rematches/{rematch-id}/accept ---

func Test_AcceptRematch_Success_CreatesNewGame(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		host, guest, game := setupCompletedGame(t, testApi)

		rematch, err := testApi.App.RpsGame().RequestRematch(ctx, &services.RematchRequestInput{
			OriginalGameID:     game.RpsGame.ID,
			RequestingPlayerID: host.ID,
			InvitedPlayerID:    guest.ID,
		})
		require.NoError(t, err)

		scenario := &apis.ApiScenario{
			Name:           "guest accepts",
			Method:         http.MethodPost,
			URL:            fmt.Sprintf("/games/rps/rematches/%s/accept", rematch.ID),
			ExpectedStatus: http.StatusOK,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				header, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, guest.Email)
				scenario.Headers = []string{header}
				body, _ := json.Marshal(apis.AcceptRematchInput{Move: apis.RpsParticipantMoveRock})
				scenario.Body = strings.NewReader(string(body))
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[apis.ApiSingleResponse[*apis.RpsRematchRequest]](t, res.Body.Bytes())
				require.NotNil(t, result.Data)
				assert.Equal(t, string(apis.RpsRematchStatusAccepted), string(result.Data.Status))
				assert.NotNil(t, result.Data.NewGameID)
			},
		}
		scenario.Test(t)
	})
}

func Test_AcceptRematch_WrongPlayerForbidden(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		host, guest, game := setupCompletedGame(t, testApi)

		rematch, err := testApi.App.RpsGame().RequestRematch(ctx, &services.RematchRequestInput{
			OriginalGameID:     game.RpsGame.ID,
			RequestingPlayerID: host.ID,
			InvitedPlayerID:    guest.ID,
		})
		require.NoError(t, err)

		// Host cannot accept their own rematch request.
		scenario := &apis.ApiScenario{
			Name:            "host self-accept forbidden",
			Method:          http.MethodPost,
			URL:             fmt.Sprintf("/games/rps/rematches/%s/accept", rematch.ID),
			ExpectedStatus:  http.StatusForbidden,
			ExpectedContent: []string{"Forbidden"},
			TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				header, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, host.Email)
				scenario.Headers = []string{header}
				body, _ := json.Marshal(apis.AcceptRematchInput{Move: apis.RpsParticipantMoveRock})
				scenario.Body = strings.NewReader(string(body))
			},
		}
		scenario.Test(t)
	})
}

// --- POST /games/rps/rematches/{rematch-id}/decline ---

func Test_DeclineRematch_Success(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		host, guest, game := setupCompletedGame(t, testApi)

		rematch, err := testApi.App.RpsGame().RequestRematch(ctx, &services.RematchRequestInput{
			OriginalGameID:     game.RpsGame.ID,
			RequestingPlayerID: host.ID,
			InvitedPlayerID:    guest.ID,
		})
		require.NoError(t, err)

		scenario := &apis.ApiScenario{
			Name:           "guest declines",
			Method:         http.MethodPost,
			URL:            fmt.Sprintf("/games/rps/rematches/%s/decline", rematch.ID),
			ExpectedStatus: http.StatusOK,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				header, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, guest.Email)
				scenario.Headers = []string{header}
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[apis.ApiSingleResponse[*apis.RpsRematchRequest]](t, res.Body.Bytes())
				require.NotNil(t, result.Data)
				assert.Equal(t, string(apis.RpsRematchStatusDeclined), string(result.Data.Status))
				assert.Nil(t, result.Data.NewGameID)
			},
		}
		scenario.Test(t)
	})
}

// --- GET /games/rps/{game-id}/rematch ---

func Test_GetRematch_ReturnsPendingRequest(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		host, guest, game := setupCompletedGame(t, testApi)

		rematch, err := testApi.App.RpsGame().RequestRematch(ctx, &services.RematchRequestInput{
			OriginalGameID:     game.RpsGame.ID,
			RequestingPlayerID: host.ID,
			InvitedPlayerID:    guest.ID,
		})
		require.NoError(t, err)

		scenario := &apis.ApiScenario{
			Name:           "get pending rematch",
			Method:         http.MethodGet,
			URL:            fmt.Sprintf("/games/rps/%s/rematch", game.RpsGame.ID),
			ExpectedStatus: http.StatusOK,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				header, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, host.Email)
				scenario.Headers = []string{header}
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[apis.ApiSingleResponse[*apis.RpsRematchRequest]](t, res.Body.Bytes())
				require.NotNil(t, result.Data)
				assert.Equal(t, rematch.ID.String(), result.Data.ID.String())
				assert.Equal(t, string(apis.RpsRematchStatusPending), string(result.Data.Status))
			},
		}
		scenario.Test(t)
	})
}

func Test_GetRematch_Returns404WhenNoneExists(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		host, _, game := setupCompletedGame(t, testApi)

		scenario := &apis.ApiScenario{
			Name:            "no rematch",
			Method:          http.MethodGet,
			URL:             fmt.Sprintf("/games/rps/%s/rematch", game.RpsGame.ID),
			ExpectedStatus:  http.StatusNotFound,
			ExpectedContent: []string{"Not Found"},
			TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				header, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, host.Email)
				scenario.Headers = []string{header}
			},
		}
		scenario.Test(t)
	})
}

// --- Serialization check ---

func Test_RpsRematchRequest_SerializesCorrectly(t *testing.T) {
	r := &apis.RpsRematchRequest{}
	b, err := json.Marshal(r)
	require.NoError(t, err)
	assert.Contains(t, string(b), `"original_game_id"`)
	assert.Contains(t, string(b), `"status"`)
	assert.Contains(t, string(b), `"expires_at"`)
}
