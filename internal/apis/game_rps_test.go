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

func Test_FindCurrentPlayersRpsGames(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		registeredPlayer := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		registeredPlayer2 := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		registeredPlayer3 := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		for i := range 5 {
			switch i % 2 {
			case 0:
				_ = core.MustCreateGame(t, testApi.App, registeredPlayer.ID, registeredPlayer2.ID, models.RpsParticipantMovePaper)
			case 1:
				_ = core.MustCreateGame(t, testApi.App, registeredPlayer2.ID, registeredPlayer.ID, models.RpsParticipantMovePaper)
			}
		}
		for i := range 5 {
			switch i % 2 {
			case 0:
				_ = core.MustCreateGame(t, testApi.App, registeredPlayer3.ID, registeredPlayer2.ID, models.RpsParticipantMovePaper)
			case 1:
				_ = core.MustCreateGame(t, testApi.App, registeredPlayer2.ID, registeredPlayer3.ID, models.RpsParticipantMovePaper)
			}
		}
		count, err := testApi.App.Adapter().Gaming().CountRpsGames(ctx, nil)
		assert.NoError(t, err)
		assert.Equal(t, int64(10), count)
		scenarios := []*apis.ApiScenario{
			{
				Name:           "no games",
				Method:         http.MethodGet,
				URL:            "/players/current-player/games/rps",
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *apis.TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, registeredPlayer.Email)
					scenario.Headers = []string{tokenHeader}
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					result := test.MustUnMarshal[apis.ApiPaginatedResponse[*apis.RpsGameWithParticipants]](t, res.Body.Bytes())
					assert.NotEmpty(t, result.Data)
					for _, v := range result.Data {
						assert.NotNil(t, v)
						assert.NotNil(t, v.InvitedParticipant)
						assert.NotNil(t, v.RequestingParticipant)
						assert.NotNil(t, v.RpsGame)
						assert.NotNil(t, v.InvitedParticipant.Player)
						assert.NotNil(t, v.RequestingParticipant.Player)
					}
				},
			},
		}
		for _, scenario := range scenarios {
			scenario.Test(t)
		}
	})
}

func Test_SubmitMoveWithToken_Success(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		playerWithUser := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		unregisteredPlayer := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(false))
		scenarios := []*apis.ApiScenario{
			{
				Name:           "inviting player win",
				Method:         http.MethodPost,
				URL:            "/games/rps/token/submit-move",
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *apis.TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					_, err := apis.SendRpsGameRequestToUnregisteredPlayer(app, t.Context(), apis.UnregisteredPlayerInput{
						InvitingPlayerEmail: unregisteredPlayer.Email,
						Move:                apis.RpsParticipantMovePaper,
					}, playerWithUser)
					if err != nil {
						t.Fatalf("Error requesting game: %v", err)
					}
					ctx := t.Context()
					err = app.JobManager().PollOnce(ctx)
					assert.NoError(t, err)
					token := ExtractFistMessageTokenFromMailer(t, app)
					body := &apis.SubmitMoveWithTokenInput{
						Move:   apis.RpsParticipantMoveScissors,
						Token:  token,
						Status: apis.RpsGameStatusCompleted,
					}
					data, err := json.Marshal(body)
					if err != nil {
						t.Errorf("Error marshalling input: %v", err)
					}
					scenario.Body = strings.NewReader(string(data))
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					result := test.MustUnMarshal[apis.ApiSingleResponse[*apis.RpsGameWithParticipants]](t, res.Body.Bytes())
					assert.NotNil(t, result.Data)
					assert.Equal(t, playerWithUser.ID, result.Data.RequestingParticipant.PlayerID)
					assert.Equal(t, unregisteredPlayer.ID, result.Data.InvitedParticipant.PlayerID)
					assert.Equal(t, apis.RpsParticipantResultWin, result.Data.InvitedParticipant.Result)
					count, err := app.Adapter().Gaming().CountRpsGameInvites(t.Context(), nil)
					assert.NoError(t, err)
					assert.Equal(t, int64(0), count)
				},
			},
		}
		for _, scenario := range scenarios {
			scenario.Test(t)
		}
	})
}
func Test_VerifyRpsGameInvite_Success(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		playerWithUser := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		unregisteredPlayer := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(false))
		scenarios := []*apis.ApiScenario{
			{
				Name:           "inviting player win",
				Method:         http.MethodPost,
				URL:            "/games/rps/invites/verify",
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *apis.TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
					_, err := apis.SendRpsGameRequestToUnregisteredPlayer(app, t.Context(), apis.UnregisteredPlayerInput{
						InvitingPlayerEmail: unregisteredPlayer.Email,
						Move:                apis.RpsParticipantMovePaper,
					}, playerWithUser)
					if err != nil {
						t.Fatalf("Error requesting game: %v", err)
					}
					ctx := t.Context()
					err = app.JobManager().PollOnce(ctx)
					assert.NoError(t, err)
					token := ExtractFistMessageTokenFromMailer(t, app)
					body := &apis.VerifyRpsGameInviteInput{
						Token: token,
					}
					data, err := json.Marshal(body)
					if err != nil {
						t.Errorf("Error marshalling input: %v", err)
					}
					scenario.Body = strings.NewReader(string(data))
				},
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
					result := test.MustUnMarshal[apis.ApiSingleResponse[*apis.RpsGameWithParticipants]](t, res.Body.Bytes())
					assert.NotNil(t, result.Data)
					assert.Equal(t, playerWithUser.ID, result.Data.RequestingParticipant.PlayerID)
					assert.Equal(t, unregisteredPlayer.ID, result.Data.InvitedParticipant.PlayerID)
					assert.Equal(t, apis.RpsGameStatusPending, result.Data.RpsGame.Status)
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
		testApi := apis.SetupApi(t, ctx, db)
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
		scenarios := []*apis.ApiScenario{
			{
				Name:           "success",
				Method:         http.MethodPost,
				URL:            "/games/rps/{game-id}/submit-move",
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *apis.TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
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
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
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
func Test_SendGameRequestToUnRegisteredPlayer_Success(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		playerWithUser := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		unregisteredPlayer := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(false))
		scenarios := []*apis.ApiScenario{
			{
				Name:           "success",
				Method:         http.MethodPost,
				URL:            "/games/rps/requests/unregistered",
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *apis.TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
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
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
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

func Test_SendGameRequestToRegisteredPlayer_Success(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		playerWithUser := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		otherPlayerWithUser := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		scenarios := []*apis.ApiScenario{
			{
				Name:           "success",
				Method:         http.MethodPost,
				URL:            "/games/rps/requests",
				ExpectedStatus: http.StatusOK,
				TestAppFactory: func(t testing.TB) *apis.TestApi {
					return testApi
				},
				BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
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
				AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
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
