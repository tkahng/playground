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
		// Each game must be completed before creating the next: one active game per player.
		for range 5 {
			g := core.MustCreateGame(t, testApi.App, registeredPlayer.ID, registeredPlayer2.ID, models.RpsParticipantMovePaper)
			core.MustCompleteGame(t, testApi.App, g, models.RpsParticipantMoveRock)
		}
		for range 5 {
			g := core.MustCreateGame(t, testApi.App, registeredPlayer3.ID, registeredPlayer2.ID, models.RpsParticipantMovePaper)
			core.MustCompleteGame(t, testApi.App, g, models.RpsParticipantMoveRock)
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
					assert.NotNil(t, result.Data.RequestingParticipant.Move, "host move should be visible after completion")
					assert.Equal(t, apis.RpsParticipantMoveRock, *result.Data.RequestingParticipant.Move)
					assert.NotNil(t, result.Data.InvitedParticipant.Move, "guest move should be visible after completion")
					assert.Equal(t, apis.RpsParticipantMovePaper, *result.Data.InvitedParticipant.Move)
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

// --- Betting API tests ---

func Test_RequestGame_WithBetAmount_Success(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		host := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		guest := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		adapter := stores.NewDbAdapterDecorators(db)
		ledger := services.NewDbLedgerService(adapter)

		scenario := &apis.ApiScenario{
			Name:           "request game with bet_amount",
			Method:         http.MethodPost,
			URL:            "/games/rps/requests",
			ExpectedStatus: http.StatusOK,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				if err := services.FulfillPointsPurchase(ctx, adapter, ledger, services.PointsPurchaseFulfillInput{
					UserID:          *host.UserID,
					PointsAmount:    500,
					StripeSessionID: "cs_bet_request_host",
				}); err != nil {
					t.Fatalf("FulfillPointsPurchase() error = %v", err)
				}
				scenario.Headers = []string{core.CreateTokenHeader(t, app, host.Email)}
				betAmount := int64(100)
				body := &apis.RpsGameRequestInput{
					Move:             apis.RpsParticipantMoveRock,
					InvitingPlayerId: guest.ID,
					BetAmount:        &betAmount,
				}
				data, _ := json.Marshal(body)
				scenario.Body = strings.NewReader(string(data))
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[apis.ApiSingleResponse[*apis.RpsGameWithParticipants]](t, res.Body.Bytes())
				assert.NotNil(t, result.Data)
				assert.NotNil(t, result.Data.RpsGame.BetAmount, "BetAmount should be set in response")
				assert.Equal(t, int64(100), *result.Data.RpsGame.BetAmount)

				// Host available balance should be reduced by the escrow hold.
				avail, err := ledger.GetUserAvailableBalance(ctx, *host.UserID)
				assert.NoError(t, err)
				assert.Equal(t, int64(400), avail, "host available balance should be 400 after 100pt escrow hold")
			},
		}
		scenario.Test(t)
	})
}

func Test_RequestGame_WithBetAmount_InsufficientBalance(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		host := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		guest := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		scenario := &apis.ApiScenario{
			Name:            "bet_amount exceeds host balance",
			Method:          http.MethodPost,
			URL:             "/games/rps/requests",
			ExpectedStatus:  http.StatusInternalServerError,
			ExpectedContent: []string{"insufficient available balance"},
			TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				// No funds — wallet has 0 balance.
				scenario.Headers = []string{core.CreateTokenHeader(t, app, host.Email)}
				betAmount := int64(100)
				body := &apis.RpsGameRequestInput{
					Move:             apis.RpsParticipantMoveRock,
					InvitingPlayerId: guest.ID,
					BetAmount:        &betAmount,
				}
				data, _ := json.Marshal(body)
				scenario.Body = strings.NewReader(string(data))
			},
		}
		scenario.Test(t)
	})
}

func Test_RequestGame_WithBetAmountZero_ValidationError(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		host := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		guest := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		scenario := &apis.ApiScenario{
			Name:            "bet_amount=0 rejected by Huma minimum:1 validation",
			Method:          http.MethodPost,
			URL:             "/games/rps/requests",
			ExpectedStatus:  http.StatusUnprocessableEntity,
			ExpectedContent: []string{"expected number >= 1", "body.bet_amount"},
			TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				scenario.Headers = []string{core.CreateTokenHeader(t, app, host.Email)}
				betAmount := int64(0)
				body := &apis.RpsGameRequestInput{
					Move:             apis.RpsParticipantMoveRock,
					InvitingPlayerId: guest.ID,
					BetAmount:        &betAmount,
				}
				data, _ := json.Marshal(body)
				scenario.Body = strings.NewReader(string(data))
			},
		}
		scenario.Test(t)
	})
}

func Test_SubmitMove_WithActiveBet_GuestWins(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		host := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		guest := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		adapter := stores.NewDbAdapterDecorators(db)
		ledger := services.NewDbLedgerService(adapter)

		betAmount := int64(100)
		var gameID string

		scenario := &apis.ApiScenario{
			Name:           "guest wins bet — balances settled correctly",
			Method:         http.MethodPost,
			URL:            "/games/rps/{game-id}/submit-move",
			ExpectedStatus: http.StatusOK,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				if err := services.FulfillPointsPurchase(ctx, adapter, ledger, services.PointsPurchaseFulfillInput{
					UserID:          *host.UserID,
					PointsAmount:    500,
					StripeSessionID: "cs_submit_host",
				}); err != nil {
					t.Fatalf("FulfillPointsPurchase(host) error = %v", err)
				}
				if err := services.FulfillPointsPurchase(ctx, adapter, ledger, services.PointsPurchaseFulfillInput{
					UserID:          *guest.UserID,
					PointsAmount:    500,
					StripeSessionID: "cs_submit_guest",
				}); err != nil {
					t.Fatalf("FulfillPointsPurchase(guest) error = %v", err)
				}

				// Create game with bet via service (host plays rock).
				game, err := app.RpsGame().RequestGame(ctx, &services.RpsGameRequestInput{
					RequestingPlayerID:   host.ID,
					InvitedPlayerID:      guest.ID,
					RequestingPlayerMove: models.RpsParticipantMoveRock,
					DurationSeconds:      3 * 24 * 60 * 60,
					BetAmount:            &betAmount,
					HostUserID:           host.UserID,
				})
				if err != nil {
					t.Fatalf("RequestGame() error = %v", err)
				}
				gameID = game.RpsGame.ID.String()

				scenario.Headers = []string{core.CreateTokenHeader(t, app, guest.Email)}
				scenario.URL = strings.ReplaceAll(scenario.URL, "{game-id}", gameID)
				// Guest plays paper — wins.
				body := &apis.SubmitMoveToGameInput{
					Move:   apis.RpsParticipantMovePaper,
					Status: apis.RpsGameStatusCompleted,
				}
				data, _ := json.Marshal(body)
				scenario.Body = strings.NewReader(string(data))
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[apis.ApiSingleResponse[*apis.RpsGameWithParticipants]](t, res.Body.Bytes())
				assert.NotNil(t, result.Data)
				assert.Equal(t, apis.RpsGameStatusCompleted, result.Data.RpsGame.Status)
				assert.Equal(t, apis.RpsParticipantResultWin, result.Data.InvitedParticipant.Result)

				// Guest wins: gains 100 (host's stake) → 600. Host loses 100 → 400.
				hostBal, err := ledger.GetUserBalance(ctx, *host.UserID)
				assert.NoError(t, err)
				assert.Equal(t, int64(400), hostBal, "host balance after losing bet")

				guestBal, err := ledger.GetUserBalance(ctx, *guest.UserID)
				assert.NoError(t, err)
				assert.Equal(t, int64(600), guestBal, "guest balance after winning bet")
			},
		}
		scenario.Test(t)
	})
}

func Test_SubmitMove_WithActiveBet_GuestInsufficientBalance(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		host := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		guest := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		adapter := stores.NewDbAdapterDecorators(db)
		ledger := services.NewDbLedgerService(adapter)

		betAmount := int64(100)

		scenario := &apis.ApiScenario{
			Name:            "guest cannot afford bet — submit-move rejected",
			Method:          http.MethodPost,
			URL:             "/games/rps/{game-id}/submit-move",
			ExpectedStatus:  http.StatusInternalServerError,
			ExpectedContent: []string{"insufficient balance"},
			TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				if err := services.FulfillPointsPurchase(ctx, adapter, ledger, services.PointsPurchaseFulfillInput{
					UserID:          *host.UserID,
					PointsAmount:    500,
					StripeSessionID: "cs_nobalance_host",
				}); err != nil {
					t.Fatalf("FulfillPointsPurchase(host) error = %v", err)
				}
				// Guest has no funds.

				game, err := app.RpsGame().RequestGame(ctx, &services.RpsGameRequestInput{
					RequestingPlayerID:   host.ID,
					InvitedPlayerID:      guest.ID,
					RequestingPlayerMove: models.RpsParticipantMoveRock,
					DurationSeconds:      3 * 24 * 60 * 60,
					BetAmount:            &betAmount,
					HostUserID:           host.UserID,
				})
				if err != nil {
					t.Fatalf("RequestGame() error = %v", err)
				}

				scenario.Headers = []string{core.CreateTokenHeader(t, app, guest.Email)}
				scenario.URL = strings.ReplaceAll(scenario.URL, "{game-id}", game.RpsGame.ID.String())
				body := &apis.SubmitMoveToGameInput{
					Move:   apis.RpsParticipantMovePaper,
					Status: apis.RpsGameStatusCompleted,
				}
				data, _ := json.Marshal(body)
				scenario.Body = strings.NewReader(string(data))
			},
		}
		scenario.Test(t)
	})
}

// Test_SendGameRequestToUnregisteredPlayer_NoBetSupport documents that the
// unregistered-player invite path does not support bet_amount. UnregisteredPlayerInput
// has no BetAmount field, and SendRpsGameRequestToUnregisteredPlayer never forwards
// one to RequestGame. A game created via this endpoint always has bet_amount nil.
func Test_SendGameRequestToUnregisteredPlayer_NoBetSupport(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		playerWithUser := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		unregisteredPlayer := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(false))

		scenario := &apis.ApiScenario{
			Name:           "unregistered-player invite never carries a bet",
			Method:         http.MethodPost,
			URL:            "/games/rps/requests/unregistered",
			ExpectedStatus: http.StatusOK,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				scenario.Headers = []string{core.CreateTokenHeader(t, app, playerWithUser.Email)}
				body := &apis.UnregisteredPlayerInput{
					Move:                apis.RpsParticipantMoveRock,
					InvitingPlayerEmail: unregisteredPlayer.Email,
				}
				data, _ := json.Marshal(body)
				scenario.Body = strings.NewReader(string(data))
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[apis.ApiSingleResponse[*apis.RpsGameWithParticipants]](t, res.Body.Bytes())
				if result.Data == nil {
					t.Fatal("expected game in response")
				}
				if result.Data.RpsGame.BetAmount != nil {
					t.Errorf("expected bet_amount nil on unregistered-player game, got %d", *result.Data.RpsGame.BetAmount)
				}
			},
		}
		scenario.Test(t)
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

func Test_SendGameRequest_BlockedPlayer_Fails(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		requester := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		target := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		// requester blocks target
		_, err := testApi.App.Adapter().Gaming().CreateFriendship(ctx, &models.Friendship{
			RequestingPlayerID: requester.ID,
			InvitedPlayerID:    target.ID,
			Status:             models.FriendshipStatusBlocked,
		})
		assert.NoError(t, err)

		body, _ := json.Marshal(map[string]any{
			"inviting_player_id": target.ID.String(),
			"move":               "rock",
		})
		scenario := &apis.ApiScenario{
			Name:            "blocked player cannot receive game request",
			Method:          http.MethodPost,
			URL:             "/games/rps/requests",
			ExpectedStatus:  http.StatusBadRequest,
			ExpectedContent: []string{"player can't play with invited player"},
			TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, requester.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.Body = strings.NewReader(string(body))
			},
		}
		scenario.Test(t)
	})
}

func Test_SendGameRequest_BlockedByTarget_Fails(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		requester := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		target := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		// target blocks requester (reverse direction)
		_, err := testApi.App.Adapter().Gaming().CreateFriendship(ctx, &models.Friendship{
			RequestingPlayerID: target.ID,
			InvitedPlayerID:    requester.ID,
			Status:             models.FriendshipStatusBlocked,
		})
		assert.NoError(t, err)

		body, _ := json.Marshal(map[string]any{
			"inviting_player_id": target.ID.String(),
			"move":               "rock",
		})
		scenario := &apis.ApiScenario{
			Name:            "target blocked requester — still cannot send game",
			Method:          http.MethodPost,
			URL:             "/games/rps/requests",
			ExpectedStatus:  http.StatusBadRequest,
			ExpectedContent: []string{"player can't play with invited player"},
			TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, requester.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.Body = strings.NewReader(string(body))
			},
		}
		scenario.Test(t)
	})
}

// Test_HostMoveHiddenWhenPending verifies that the host's move is not returned
// to the guest via the invite-verify endpoint while the game is still pending.
// The invite token is created directly so this test does not depend on the job
// queue being committed within the test transaction.
func Test_HostMoveHiddenWhenPending(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		host := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		guest := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(false))

		// Create game and invite directly without going through the mailer.
		game, err := testApi.App.RpsGame().RequestGame(ctx, &services.RpsGameRequestInput{
			RequestingPlayerID:   host.ID,
			InvitedPlayerID:      guest.ID,
			RequestingPlayerMove: models.RpsParticipantMoveScissors,
			DurationSeconds:      services.GameDurationSeconds,
		})
		if err != nil {
			t.Fatalf("RequestGame error: %v", err)
		}
		invite, err := testApi.App.Adapter().Gaming().CreateRpsGameInvite(ctx, &models.RpsGameInvite{
			GameID:             game.RpsGame.ID,
			RequestingPlayerID: host.ID,
			InvitedPlayerID:    guest.ID,
			Token:              "test-token-hidden-move-check",
			ExpiresAt:          game.RpsGame.ExpiresAt,
		})
		if err != nil {
			t.Fatalf("CreateRpsGameInvite error: %v", err)
		}

		scenario := &apis.ApiScenario{
			Name:           "host move is hidden from guest on pending game via invite-verify",
			Method:         http.MethodPost,
			URL:            "/games/rps/invites/verify",
			ExpectedStatus: http.StatusOK,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				body, _ := json.Marshal(apis.VerifyRpsGameInviteInput{Token: invite.Token})
				scenario.Body = strings.NewReader(string(body))
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[apis.ApiSingleResponse[*apis.RpsGameWithParticipants]](t, res.Body.Bytes())
				assert.NotNil(t, result.Data)
				assert.Equal(t, apis.RpsGameStatusPending, result.Data.RpsGame.Status)
				// Host's move must be hidden from the guest.
				assert.Nil(t, result.Data.RequestingParticipant.Move,
					"host move must not be exposed to guest while game is pending")
				assert.Equal(t, guest.ID, result.Data.InvitedParticipant.PlayerID)
			},
		}
		scenario.Test(t)
	})
}

// Test_HostMoveVisibleAfterCompletion verifies that after the game is complete,
// both moves are returned to both players.
func Test_HostMoveVisibleAfterCompletion(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		host := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		guest := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		game, err := testApi.App.RpsGame().RequestGame(ctx, &services.RpsGameRequestInput{
			RequestingPlayerID:   host.ID,
			InvitedPlayerID:      guest.ID,
			RequestingPlayerMove: models.RpsParticipantMoveScissors,
			DurationSeconds:      services.GameDurationSeconds,
		})
		if err != nil {
			t.Fatalf("RequestGame error: %v", err)
		}

		// Guest submits rock — host (scissors) loses.
		scenario := &apis.ApiScenario{
			Name:           "both moves visible after completion",
			Method:         http.MethodPost,
			URL:            fmt.Sprintf("/games/rps/%s/submit-move", game.RpsGame.ID),
			ExpectedStatus: http.StatusOK,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, guest.Email)
				scenario.Headers = []string{tokenHeader}
				body, _ := json.Marshal(apis.SubmitMoveToGameInput{
					Move:   apis.RpsParticipantMoveRock,
					Status: apis.RpsGameStatusCompleted,
				})
				scenario.Body = strings.NewReader(string(body))
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[apis.ApiSingleResponse[*apis.RpsGameWithParticipants]](t, res.Body.Bytes())
				assert.NotNil(t, result.Data)
				assert.Equal(t, apis.RpsGameStatusCompleted, result.Data.RpsGame.Status)
				assert.NotNil(t, result.Data.RequestingParticipant.Move, "host move should be visible after completion")
				assert.Equal(t, apis.RpsParticipantMoveScissors, *result.Data.RequestingParticipant.Move)
				assert.NotNil(t, result.Data.InvitedParticipant.Move, "guest move should be visible after completion")
				assert.Equal(t, apis.RpsParticipantMoveRock, *result.Data.InvitedParticipant.Move)
				assert.Equal(t, apis.RpsParticipantResultLose, result.Data.RequestingParticipant.Result)
				assert.Equal(t, apis.RpsParticipantResultWin, result.Data.InvitedParticipant.Result)
			},
		}
		scenario.Test(t)
	})
}

// Test_HostMoveHiddenInGameList verifies that when the guest views the game list
// on a pending game, the host's move is nil.
func Test_HostMoveHiddenInGameList(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		host := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		guest := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		_, err := testApi.App.RpsGame().RequestGame(ctx, &services.RpsGameRequestInput{
			RequestingPlayerID:   host.ID,
			InvitedPlayerID:      guest.ID,
			RequestingPlayerMove: models.RpsParticipantMovePaper,
			DurationSeconds:      services.GameDurationSeconds,
		})
		if err != nil {
			t.Fatalf("RequestGame error: %v", err)
		}

		scenario := &apis.ApiScenario{
			Name:           "host move hidden from guest in game list",
			Method:         http.MethodGet,
			URL:            "/players/current-player/games/rps",
			ExpectedStatus: http.StatusOK,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				// Authenticated as guest.
				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, guest.Email)
				scenario.Headers = []string{tokenHeader}
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[apis.ApiPaginatedResponse[*apis.RpsGameWithParticipants]](t, res.Body.Bytes())
				assert.NotEmpty(t, result.Data)
				g := result.Data[0]
				assert.Equal(t, apis.RpsGameStatusPending, g.RpsGame.Status)
				assert.Nil(t, g.RequestingParticipant.Move,
					"host move must be nil when guest views a pending game")
			},
		}
		scenario.Test(t)
	})
}

// Test_SubmitMove_WithActiveBet_HostWins verifies that the host receives the
// full bet payout when they win.
func Test_SubmitMove_WithActiveBet_HostWins(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		host := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		guest := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		adapter := stores.NewDbAdapterDecorators(db)
		ledger := services.NewDbLedgerService(adapter)

		betAmount := int64(100)
		var gameID string

		scenario := &apis.ApiScenario{
			Name:           "host wins bet — balances settled correctly",
			Method:         http.MethodPost,
			URL:            "/games/rps/{game-id}/submit-move",
			ExpectedStatus: http.StatusOK,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				if err := services.FulfillPointsPurchase(ctx, adapter, ledger, services.PointsPurchaseFulfillInput{
					UserID:          *host.UserID,
					PointsAmount:    500,
					StripeSessionID: "cs_hostwins_host",
				}); err != nil {
					t.Fatalf("FulfillPointsPurchase(host): %v", err)
				}
				if err := services.FulfillPointsPurchase(ctx, adapter, ledger, services.PointsPurchaseFulfillInput{
					UserID:          *guest.UserID,
					PointsAmount:    500,
					StripeSessionID: "cs_hostwins_guest",
				}); err != nil {
					t.Fatalf("FulfillPointsPurchase(guest): %v", err)
				}

				// Host plays paper.
				game, err := app.RpsGame().RequestGame(ctx, &services.RpsGameRequestInput{
					RequestingPlayerID:   host.ID,
					InvitedPlayerID:      guest.ID,
					RequestingPlayerMove: models.RpsParticipantMovePaper,
					DurationSeconds:      services.GameDurationSeconds,
					BetAmount:            &betAmount,
					HostUserID:           host.UserID,
				})
				if err != nil {
					t.Fatalf("RequestGame: %v", err)
				}
				gameID = game.RpsGame.ID.String()

				scenario.Headers = []string{core.CreateTokenHeader(t, app, guest.Email)}
				scenario.URL = strings.ReplaceAll(scenario.URL, "{game-id}", gameID)
				// Guest plays rock — loses.
				body, _ := json.Marshal(apis.SubmitMoveToGameInput{
					Move:   apis.RpsParticipantMoveRock,
					Status: apis.RpsGameStatusCompleted,
				})
				scenario.Body = strings.NewReader(string(body))
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[apis.ApiSingleResponse[*apis.RpsGameWithParticipants]](t, res.Body.Bytes())
				assert.NotNil(t, result.Data)
				assert.Equal(t, apis.RpsGameStatusCompleted, result.Data.RpsGame.Status)
				assert.Equal(t, apis.RpsParticipantResultWin, result.Data.RequestingParticipant.Result)

				// Host wins: gains 100 → 600. Guest loses 100 → 400.
				hostBal, err := ledger.GetUserBalance(ctx, *host.UserID)
				assert.NoError(t, err)
				assert.Equal(t, int64(600), hostBal, "host balance after winning bet")

				guestBal, err := ledger.GetUserBalance(ctx, *guest.UserID)
				assert.NoError(t, err)
				assert.Equal(t, int64(400), guestBal, "guest balance after losing bet")
			},
		}
		scenario.Test(t)
	})
}

// Test_SubmitMove_WithActiveBet_Tie verifies that both bets are refunded on a tie.
func Test_SubmitMove_WithActiveBet_Tie(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		host := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		guest := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		adapter := stores.NewDbAdapterDecorators(db)
		ledger := services.NewDbLedgerService(adapter)

		betAmount := int64(100)
		var gameID string

		scenario := &apis.ApiScenario{
			Name:           "tie — both bets refunded",
			Method:         http.MethodPost,
			URL:            "/games/rps/{game-id}/submit-move",
			ExpectedStatus: http.StatusOK,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				if err := services.FulfillPointsPurchase(ctx, adapter, ledger, services.PointsPurchaseFulfillInput{
					UserID:          *host.UserID,
					PointsAmount:    500,
					StripeSessionID: "cs_tie_host",
				}); err != nil {
					t.Fatalf("FulfillPointsPurchase(host): %v", err)
				}
				if err := services.FulfillPointsPurchase(ctx, adapter, ledger, services.PointsPurchaseFulfillInput{
					UserID:          *guest.UserID,
					PointsAmount:    500,
					StripeSessionID: "cs_tie_guest",
				}); err != nil {
					t.Fatalf("FulfillPointsPurchase(guest): %v", err)
				}

				// Both play rock.
				game, err := app.RpsGame().RequestGame(ctx, &services.RpsGameRequestInput{
					RequestingPlayerID:   host.ID,
					InvitedPlayerID:      guest.ID,
					RequestingPlayerMove: models.RpsParticipantMoveRock,
					DurationSeconds:      services.GameDurationSeconds,
					BetAmount:            &betAmount,
					HostUserID:           host.UserID,
				})
				if err != nil {
					t.Fatalf("RequestGame: %v", err)
				}
				gameID = game.RpsGame.ID.String()

				scenario.Headers = []string{core.CreateTokenHeader(t, app, guest.Email)}
				scenario.URL = strings.ReplaceAll(scenario.URL, "{game-id}", gameID)
				body, _ := json.Marshal(apis.SubmitMoveToGameInput{
					Move:   apis.RpsParticipantMoveRock,
					Status: apis.RpsGameStatusCompleted,
				})
				scenario.Body = strings.NewReader(string(body))
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[apis.ApiSingleResponse[*apis.RpsGameWithParticipants]](t, res.Body.Bytes())
				assert.NotNil(t, result.Data)
				assert.Equal(t, apis.RpsGameStatusCompleted, result.Data.RpsGame.Status)
				assert.Equal(t, apis.RpsParticipantResultTie, result.Data.RequestingParticipant.Result)
				assert.Equal(t, apis.RpsParticipantResultTie, result.Data.InvitedParticipant.Result)

				// Both should have their original 500 restored after tie refund.
				hostBal, err := ledger.GetUserBalance(ctx, *host.UserID)
				assert.NoError(t, err)
				assert.Equal(t, int64(500), hostBal, "host balance unchanged on tie")

				guestBal, err := ledger.GetUserBalance(ctx, *guest.UserID)
				assert.NoError(t, err)
				assert.Equal(t, int64(500), guestBal, "guest balance unchanged on tie")
			},
		}
		scenario.Test(t)
	})
}

// Test_SubmitMove_PendingStatusRejected verifies that submitting status="pending" is
// rejected by the API schema (422 Unprocessable Entity) rather than reaching the service.
func Test_SubmitMove_PendingStatusRejected(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		host := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		guest := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		game, err := testApi.App.RpsGame().RequestGame(ctx, &services.RpsGameRequestInput{
			RequestingPlayerID:   host.ID,
			InvitedPlayerID:      guest.ID,
			RequestingPlayerMove: models.RpsParticipantMoveRock,
			DurationSeconds:      services.GameDurationSeconds,
		})
		if err != nil {
			t.Fatalf("RequestGame: %v", err)
		}

		scenario := &apis.ApiScenario{
			Name:            "status=pending rejected by schema validation",
			Method:          http.MethodPost,
			URL:             fmt.Sprintf("/games/rps/%s/submit-move", game.RpsGame.ID),
			ExpectedStatus:  http.StatusUnprocessableEntity,
			ExpectedContent: []string{"body.status"},
			TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, guest.Email)
				scenario.Headers = []string{tokenHeader}
				body, _ := json.Marshal(map[string]string{
					"move":   "rock",
					"status": "pending",
				})
				scenario.Body = strings.NewReader(string(body))
			},
		}
		scenario.Test(t)
	})
}

// Test_CancelGame_HostCancel verifies that the host can cancel their own pending game.
func Test_CancelGame_HostCancel(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		host := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		guest := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		game, err := testApi.App.RpsGame().RequestGame(ctx, &services.RpsGameRequestInput{
			RequestingPlayerID:   host.ID,
			InvitedPlayerID:      guest.ID,
			RequestingPlayerMove: models.RpsParticipantMoveRock,
			DurationSeconds:      services.GameDurationSeconds,
		})
		if err != nil {
			t.Fatalf("RequestGame: %v", err)
		}

		scenario := &apis.ApiScenario{
			Name:           "host cancels pending game",
			Method:         http.MethodPost,
			URL:            fmt.Sprintf("/games/rps/%s/cancel", game.RpsGame.ID),
			ExpectedStatus: http.StatusOK,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, host.Email)
				scenario.Headers = []string{tokenHeader}
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[apis.ApiSingleResponse[*apis.RpsGameWithParticipants]](t, res.Body.Bytes())
				assert.NotNil(t, result.Data)
				assert.Equal(t, apis.RpsGameStatusCancelled, result.Data.RpsGame.Status)
				assert.Equal(t, apis.RpsParticipantStatusDeclined, result.Data.InvitedParticipant.Status)
			},
		}
		scenario.Test(t)
	})
}

// Test_CancelGame_GuestForbidden verifies that the guest cannot cancel the host's game.
func Test_CancelGame_GuestForbidden(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		host := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		guest := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		game, err := testApi.App.RpsGame().RequestGame(ctx, &services.RpsGameRequestInput{
			RequestingPlayerID:   host.ID,
			InvitedPlayerID:      guest.ID,
			RequestingPlayerMove: models.RpsParticipantMoveRock,
			DurationSeconds:      services.GameDurationSeconds,
		})
		if err != nil {
			t.Fatalf("RequestGame: %v", err)
		}

		scenario := &apis.ApiScenario{
			Name:            "guest cannot cancel host's game",
			Method:          http.MethodPost,
			URL:             fmt.Sprintf("/games/rps/%s/cancel", game.RpsGame.ID),
			ExpectedStatus:  http.StatusForbidden,
			ExpectedContent: []string{"only the host may cancel"},
			TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, guest.Email)
				scenario.Headers = []string{tokenHeader}
			},
		}
		scenario.Test(t)
	})
}

// Test_CancelGame_AlreadyCompleted verifies that a completed game cannot be cancelled.
func Test_CancelGame_AlreadyCompleted(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		host := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		guest := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		game := core.MustCreateGame(t, testApi.App, host.ID, guest.ID, models.RpsParticipantMoveRock)
		core.MustCompleteGame(t, testApi.App, game, models.RpsParticipantMoveScissors)

		scenario := &apis.ApiScenario{
			Name:            "cannot cancel completed game",
			Method:          http.MethodPost,
			URL:             fmt.Sprintf("/games/rps/%s/cancel", game.RpsGame.ID),
			ExpectedStatus:  http.StatusBadRequest,
			ExpectedContent: []string{"only pending games"},
			TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, host.Email)
				scenario.Headers = []string{tokenHeader}
			},
		}
		scenario.Test(t)
	})
}

// Test_CancelGame_WithBetRefunded verifies that a bet escrow is refunded when the host cancels.
func Test_CancelGame_WithBetRefunded(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		host := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		guest := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		adapter := stores.NewDbAdapterDecorators(db)
		ledger := services.NewDbLedgerService(adapter)

		if err := services.FulfillPointsPurchase(ctx, adapter, ledger, services.PointsPurchaseFulfillInput{
			UserID:          *host.UserID,
			PointsAmount:    500,
			StripeSessionID: "cs_cancel_refund",
		}); err != nil {
			t.Fatalf("FulfillPointsPurchase: %v", err)
		}

		betAmount := int64(100)
		game, err := testApi.App.RpsGame().RequestGame(ctx, &services.RpsGameRequestInput{
			RequestingPlayerID:   host.ID,
			InvitedPlayerID:      guest.ID,
			RequestingPlayerMove: models.RpsParticipantMoveRock,
			DurationSeconds:      services.GameDurationSeconds,
			BetAmount:            &betAmount,
			HostUserID:           host.UserID,
		})
		if err != nil {
			t.Fatalf("RequestGame: %v", err)
		}

		// Verify escrow hold is active.
		avail, _ := ledger.GetUserAvailableBalance(ctx, *host.UserID)
		assert.Equal(t, int64(400), avail, "escrow should reduce available balance")

		scenario := &apis.ApiScenario{
			Name:           "bet escrow refunded on host cancel",
			Method:         http.MethodPost,
			URL:            fmt.Sprintf("/games/rps/%s/cancel", game.RpsGame.ID),
			ExpectedStatus: http.StatusOK,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, host.Email)
				scenario.Headers = []string{tokenHeader}
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[apis.ApiSingleResponse[*apis.RpsGameWithParticipants]](t, res.Body.Bytes())
				assert.Equal(t, apis.RpsGameStatusCancelled, result.Data.RpsGame.Status)

				// Escrow must be released.
				availAfter, _ := ledger.GetUserAvailableBalance(ctx, *host.UserID)
				assert.Equal(t, int64(500), availAfter, "escrow should be refunded after cancel")
			},
		}
		scenario.Test(t)
	})
}

// Test_AcceptRematch_WithMove verifies that the accept rematch endpoint requires
// a move, creates the new host participant as completed, and leaves the guest pending.
func Test_AcceptRematch_WithMove(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		host := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		guest := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		// Play and complete a game.
		game := core.MustCreateGame(t, testApi.App, host.ID, guest.ID, models.RpsParticipantMoveRock)
		core.MustCompleteGame(t, testApi.App, game, models.RpsParticipantMoveScissors)

		// Host requests a rematch.
		rematch, err := testApi.App.RpsGame().RequestRematch(ctx, &services.RematchRequestInput{
			OriginalGameID:     game.RpsGame.ID,
			RequestingPlayerID: host.ID,
			InvitedPlayerID:    guest.ID,
		})
		if err != nil {
			t.Fatalf("RequestRematch: %v", err)
		}

		scenario := &apis.ApiScenario{
			Name:           "guest accepts rematch with paper move",
			Method:         http.MethodPost,
			URL:            fmt.Sprintf("/games/rps/rematches/%s/accept", rematch.ID),
			ExpectedStatus: http.StatusOK,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, guest.Email)
				scenario.Headers = []string{tokenHeader}
				body, _ := json.Marshal(apis.AcceptRematchInput{Move: apis.RpsParticipantMovePaper})
				scenario.Body = strings.NewReader(string(body))
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[apis.ApiSingleResponse[*apis.RpsRematchRequest]](t, res.Body.Bytes())
				assert.NotNil(t, result.Data)
				assert.Equal(t, apis.RpsRematchStatusAccepted, result.Data.Status)
				assert.NotNil(t, result.Data.NewGameID)

				// Inspect the new game: guest (accepting player) is now host with move set.
				newGame, err := testApi.App.RpsGame().FindRpsGameWithParticipants(ctx, *result.Data.NewGameID)
				if err != nil {
					t.Fatalf("FindRpsGameWithParticipants: %v", err)
				}
				assert.Equal(t, guest.ID, newGame.RequestingParticipant.PlayerID)
				assert.Equal(t, models.RpsParticipantStatusCompleted, newGame.RequestingParticipant.Status)
				assert.Equal(t, models.RpsParticipantMovePaper, newGame.RequestingParticipant.Move)
				// Previous host is now guest — still pending.
				assert.Equal(t, host.ID, newGame.InvitedParticipant.PlayerID)
				assert.Equal(t, models.RpsParticipantStatusPending, newGame.InvitedParticipant.Status)
			},
		}
		scenario.Test(t)
	})
}
