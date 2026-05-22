//go:build integration

package apis_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tkahng/playground/internal/apis"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/test"
)

func Test_GetCurrentPlayerRpsStats_NewPlayer(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		player := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		scenario := &apis.ApiScenario{
			Name:           "new player has zero stats",
			Method:         http.MethodGet,
			URL:            "/players/current-player/rps-stats",
			ExpectedStatus: http.StatusOK,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				header, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, player.Email)
				scenario.Headers = []string{header}
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[apis.ApiSingleResponse[*apis.PlayerRpsStatsResponse]](t, res.Body.Bytes())
				assert.NotNil(t, result.Data)
				assert.Equal(t, int64(0), result.Data.TotalGames)
				assert.Equal(t, int64(0), result.Data.Wins)
				assert.Equal(t, int64(0), result.Data.Losses)
				assert.Equal(t, int64(0), result.Data.Ties)
			},
		}
		scenario.Test(t)
	})
}

func Test_GetCurrentPlayerRpsStats_AfterGames(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		host := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		guest := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		// Rock vs scissors → host wins.
		g1 := core.MustCreateGame(t, testApi.App, host.ID, guest.ID, models.RpsParticipantMoveRock)
		core.MustCompleteGame(t, testApi.App, g1, models.RpsParticipantMoveScissors)

		// Paper vs paper → tie.
		g2 := core.MustCreateGame(t, testApi.App, host.ID, guest.ID, models.RpsParticipantMovePaper)
		core.MustCompleteGame(t, testApi.App, g2, models.RpsParticipantMovePaper)

		// Scissors vs rock → host loses.
		g3 := core.MustCreateGame(t, testApi.App, host.ID, guest.ID, models.RpsParticipantMoveScissors)
		core.MustCompleteGame(t, testApi.App, g3, models.RpsParticipantMoveRock)

		scenario := &apis.ApiScenario{
			Name:           "stats reflect 3 completed games",
			Method:         http.MethodGet,
			URL:            "/players/current-player/rps-stats",
			ExpectedStatus: http.StatusOK,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				header, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, host.Email)
				scenario.Headers = []string{header}
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[apis.ApiSingleResponse[*apis.PlayerRpsStatsResponse]](t, res.Body.Bytes())
				assert.NotNil(t, result.Data)
				assert.Equal(t, int64(3), result.Data.TotalGames)
				assert.Equal(t, int64(1), result.Data.Wins)
				assert.Equal(t, int64(1), result.Data.Losses)
				assert.Equal(t, int64(1), result.Data.Ties)
			},
		}
		scenario.Test(t)
	})
}
