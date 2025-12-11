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
	"github.com/tkahng/playground/internal/apis"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/services"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/test"
)

func Test_SendGameRequestToUnRegisteredPlayer_Success(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := SetupApi(t, ctx, db)
		playerWithUser := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		unregisteredPlayer := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(false))
		scenarios := []*ApiScenario{
			{
				Name:           "success",
				Method:         http.MethodPost,
				URL:            "/games/rps/requests/unregistered",
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
					tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, playerWithUser.Email)
					scenario.Headers = []string{tokenHeader}
					body := &apis.UnregisteredPlayerInput{
						Move:                apis.RpsParticipantMoveRock,
						InvitingPlayerEmail: unregisteredPlayer.Email,
					}
					data, err := json.Marshal(body)
					if err != nil {
						t.Errorf("Error marshalling input: %v", err)
					}
					scenario.Body = strings.NewReader(string(data))
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
					result := test.MustUnMarshal[apis.ApiSingleResponse[*apis.RpsGameWithParticipants]](t, res.Body.Bytes())
					assert.NotNil(t, result.Data)
					assert.Equal(t, playerWithUser.ID, result.Data.RequestingParticipant.PlayerID)
					assert.Equal(t, unregisteredPlayer.ID, result.Data.InvitedParticipant.PlayerID)
					ctx := t.Context()
					err := app.JobManager().PollOnce(ctx)
					assert.NoError(t, err)
					token := ExtractFistMessageTokenFromMailer(t, app)
					invite, err := app.Adapter().Gaming().FindRpsGameInvite(ctx, &stores.RpsGameInviteFilter{
						Tokens: []string{token},
					})
					assert.NoError(t, err)
					assert.Equal(t, token, invite.Token)
				},
			},
		}
		for _, scenario := range scenarios {
			scenario.Test(t)
		}
	})
}

func Test_SubmitMove_Success(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := SetupApi(t, ctx, db)
		playerWithUser := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		otherPlayerWithUser := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		gameWithParticipants, err := testApi.App.RpsGame().RequestGame(ctx, &services.RpsGameRequestInput{
			RequestingPlayerID:   playerWithUser.ID,
			InvitedPlayerID:      otherPlayerWithUser.ID,
			RequestingPlayerMove: models.RpsParticipantMoveRock,
			DurationSeconds:      3 * 24 * 60 * 60,
		})
		if err != nil {
			t.Errorf("Error requesting game: %v", err)
		}
		scenarios := []*ApiScenario{
			{
				Name:           "success",
				Method:         http.MethodPost,
				URL:            "/games/rps/{game-id}/submit-move",
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
					tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, otherPlayerWithUser.Email)
					scenario.Headers = []string{tokenHeader}
					scenario.URL = strings.ReplaceAll(scenario.URL, "{game-id}", fmt.Sprintf("%s", gameWithParticipants.RpsGame.ID))
					body := &apis.SubmitMoveToGameInput{
						Move:   apis.RpsParticipantMovePaper,
						Status: apis.RpsGameStatusCompleted,
					}
					data, err := json.Marshal(body)
					if err != nil {
						t.Errorf("Error marshalling input: %v", err)
					}
					scenario.Body = strings.NewReader(string(data))
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
					result := test.MustUnMarshal[apis.ApiSingleResponse[*apis.RpsGameWithParticipants]](t, res.Body.Bytes())
					assert.NotNil(t, result.Data)
					assert.Equal(t, playerWithUser.ID, result.Data.RequestingParticipant.PlayerID)
					assert.Equal(t, otherPlayerWithUser.ID, result.Data.InvitedParticipant.PlayerID)
					assert.Equal(t, apis.RpsParticipantMoveRock, result.Data.RequestingParticipant.Move)
					assert.Equal(t, apis.RpsParticipantMovePaper, result.Data.InvitedParticipant.Move)
					assert.Equal(t, apis.RpsGameStatusCompleted, result.Data.RpsGame.Status)
					assert.Equal(t, result.Data.InvitedParticipant.Result, apis.RpsParticipantResultWin)
				},
			},
		}
		for _, scenario := range scenarios {
			scenario.Test(t)
		}
	})
}

func Test_SendGameRequestToRegisteredPlayer_Success(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := SetupApi(t, ctx, db)
		playerWithUser := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		otherPlayerWithUser := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		scenarios := []*ApiScenario{
			{
				Name:           "success",
				Method:         http.MethodPost,
				URL:            "/games/rps/requests",
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario) {
					tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, playerWithUser.Email)
					scenario.Headers = []string{tokenHeader}
					body := &apis.RpsGameRequestInput{
						Move:             apis.RpsParticipantMoveRock,
						InvitingPlayerId: otherPlayerWithUser.ID,
					}
					data, err := json.Marshal(body)
					if err != nil {
						t.Errorf("Error marshalling input: %v", err)
					}
					scenario.Body = strings.NewReader(string(data))
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *ApiScenario, res *httptest.ResponseRecorder) {
					result := test.MustUnMarshal[apis.ApiSingleResponse[*apis.RpsGameWithParticipants]](t, res.Body.Bytes())
					assert.NotNil(t, result.Data)
					assert.Equal(t, playerWithUser.ID, result.Data.RequestingParticipant.PlayerID)
					assert.Equal(t, otherPlayerWithUser.ID, result.Data.InvitedParticipant.PlayerID)
				},
			},
		}
		for _, scenario := range scenarios {
			scenario.Test(t)
		}
	})
}
