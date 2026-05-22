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

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tkahng/playground/internal/apis"
	"github.com/tkahng/playground/internal/core"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/test"
)

func Test_ListFriends_Empty(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		player := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		scenario := &apis.ApiScenario{
			Name:           "no friends",
			Method:         http.MethodGet,
			URL:            "/players/friends",
			ExpectedStatus: http.StatusOK,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, player.Email)
				scenario.Headers = []string{tokenHeader}
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[apis.ApiPaginatedResponse[*apis.Friendship]](t, res.Body.Bytes())
				assert.Equal(t, int64(0), result.Meta.Total)
			},
		}
		scenario.Test(t)
	})
}

func Test_ListFriends_WithAcceptedFriendships(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		player := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		friend1 := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		friend2 := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		// create accepted friendships
		_, err := testApi.App.Adapter().Gaming().CreateFriendship(ctx, &models.Friendship{
			RequestingPlayerID: player.ID,
			InvitedPlayerID:    friend1.ID,
			Status:             models.FriendshipStatusAccepted,
		})
		assert.NoError(t, err)
		_, err = testApi.App.Adapter().Gaming().CreateFriendship(ctx, &models.Friendship{
			RequestingPlayerID: friend2.ID,
			InvitedPlayerID:    player.ID,
			Status:             models.FriendshipStatusAccepted,
		})
		assert.NoError(t, err)
		// create a pending friendship — should NOT appear in friends list
		_, err = testApi.App.Adapter().Gaming().CreateFriendship(ctx, &models.Friendship{
			RequestingPlayerID: player.ID,
			InvitedPlayerID:    core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true)).ID,
			Status:             models.FriendshipStatusPending,
		})
		assert.NoError(t, err)

		scenario := &apis.ApiScenario{
			Name:           "two accepted friends",
			Method:         http.MethodGet,
			URL:            "/players/friends?page=0&per_page=10",
			ExpectedStatus: http.StatusOK,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, player.Email)
				scenario.Headers = []string{tokenHeader}
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[apis.ApiPaginatedResponse[*apis.Friendship]](t, res.Body.Bytes())
				assert.Equal(t, int64(2), result.Meta.Total)
				for _, f := range result.Data {
					assert.Equal(t, apis.FriendshipStatusAccepted, f.Status)
				}
			},
		}
		scenario.Test(t)
	})
}

func Test_ListFriendRequests_IncomingAndOutgoing(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		player := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		requester := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		invitee := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		// incoming request (requester → player)
		_, err := testApi.App.Adapter().Gaming().CreateFriendship(ctx, &models.Friendship{
			RequestingPlayerID: requester.ID,
			InvitedPlayerID:    player.ID,
			Status:             models.FriendshipStatusPending,
		})
		assert.NoError(t, err)
		// outgoing request (player → invitee)
		_, err = testApi.App.Adapter().Gaming().CreateFriendship(ctx, &models.Friendship{
			RequestingPlayerID: player.ID,
			InvitedPlayerID:    invitee.ID,
			Status:             models.FriendshipStatusPending,
		})
		assert.NoError(t, err)

		scenario := &apis.ApiScenario{
			Name:           "one incoming one outgoing",
			Method:         http.MethodGet,
			URL:            "/players/friends/requests?page=0&per_page=10",
			ExpectedStatus: http.StatusOK,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, player.Email)
				scenario.Headers = []string{tokenHeader}
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[apis.ApiPaginatedResponse[*apis.Friendship]](t, res.Body.Bytes())
				assert.Equal(t, int64(2), result.Meta.Total)
				for _, f := range result.Data {
					assert.Equal(t, apis.FriendshipStatusPending, f.Status)
				}
			},
		}
		scenario.Test(t)
	})
}

func Test_SendFriendRequest_Success(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		player := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		target := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		body, _ := json.Marshal(apis.SendFriendRequestBody{InvitedPlayerID: target.ID})
		scenario := &apis.ApiScenario{
			Name:           "success",
			Method:         http.MethodPost,
			URL:            "/players/friends/requests",
			ExpectedStatus: http.StatusOK,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, player.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.Body = strings.NewReader(string(body))
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[apis.ApiSingleResponse[*apis.Friendship]](t, res.Body.Bytes())
				assert.NotNil(t, result.Data)
				assert.Equal(t, apis.FriendshipStatusPending, result.Data.Status)
				assert.Equal(t, player.ID.String(), result.Data.RequestingPlayerID.String())
				assert.Equal(t, target.ID.String(), result.Data.InvitedPlayerID.String())
			},
		}
		scenario.Test(t)
	})
}

func Test_SendFriendRequest_ToSelf_Fails(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		player := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		body, _ := json.Marshal(apis.SendFriendRequestBody{InvitedPlayerID: player.ID})
		scenario := &apis.ApiScenario{
			Name:            "self request",
			Method:          http.MethodPost,
			URL:             "/players/friends/requests",
			ExpectedStatus:  http.StatusBadRequest,
			ExpectedContent: []string{"cannot send friend request to yourself"},
			TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, player.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.Body = strings.NewReader(string(body))
			},
		}
		scenario.Test(t)
	})
}

func Test_SendFriendRequest_AlreadyPending_Fails(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		player := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		target := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		_, err := testApi.App.Adapter().Gaming().CreateFriendship(ctx, &models.Friendship{
			RequestingPlayerID: player.ID,
			InvitedPlayerID:    target.ID,
			Status:             models.FriendshipStatusPending,
		})
		assert.NoError(t, err)

		body, _ := json.Marshal(apis.SendFriendRequestBody{InvitedPlayerID: target.ID})
		scenario := &apis.ApiScenario{
			Name:            "already pending",
			Method:          http.MethodPost,
			URL:             "/players/friends/requests",
			ExpectedStatus:  http.StatusConflict,
			ExpectedContent: []string{"friend request already pending"},
			TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, player.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.Body = strings.NewReader(string(body))
			},
		}
		scenario.Test(t)
	})
}

func Test_SendFriendRequest_AfterDecline_Succeeds(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		player := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		target := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		_, err := testApi.App.Adapter().Gaming().CreateFriendship(ctx, &models.Friendship{
			RequestingPlayerID: player.ID,
			InvitedPlayerID:    target.ID,
			Status:             models.FriendshipStatusDeclined,
		})
		assert.NoError(t, err)

		body, _ := json.Marshal(apis.SendFriendRequestBody{InvitedPlayerID: target.ID})
		scenario := &apis.ApiScenario{
			Name:           "re-request after decline",
			Method:         http.MethodPost,
			URL:            "/players/friends/requests",
			ExpectedStatus: http.StatusOK,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, player.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.Body = strings.NewReader(string(body))
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[apis.ApiSingleResponse[*apis.Friendship]](t, res.Body.Bytes())
				assert.NotNil(t, result.Data)
				assert.Equal(t, apis.FriendshipStatusPending, result.Data.Status)
			},
		}
		scenario.Test(t)
	})
}

func Test_AcceptFriendRequest_Success(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		requester := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		invitee := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		friendship, err := testApi.App.Adapter().Gaming().CreateFriendship(ctx, &models.Friendship{
			RequestingPlayerID: requester.ID,
			InvitedPlayerID:    invitee.ID,
			Status:             models.FriendshipStatusPending,
		})
		assert.NoError(t, err)

		scenario := &apis.ApiScenario{
			Name:           "accept",
			Method:         http.MethodPost,
			URL:            fmt.Sprintf("/players/friends/requests/%s/accept", friendship.ID),
			ExpectedStatus: http.StatusOK,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, invitee.Email)
				scenario.Headers = []string{tokenHeader}
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[apis.ApiSingleResponse[*apis.Friendship]](t, res.Body.Bytes())
				assert.Equal(t, apis.FriendshipStatusAccepted, result.Data.Status)
			},
		}
		scenario.Test(t)
	})
}

func Test_AcceptFriendRequest_WrongPlayer_Fails(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		requester := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		invitee := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		bystander := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		friendship, err := testApi.App.Adapter().Gaming().CreateFriendship(ctx, &models.Friendship{
			RequestingPlayerID: requester.ID,
			InvitedPlayerID:    invitee.ID,
			Status:             models.FriendshipStatusPending,
		})
		assert.NoError(t, err)

		scenario := &apis.ApiScenario{
			Name:            "wrong player",
			Method:          http.MethodPost,
			URL:             fmt.Sprintf("/players/friends/requests/%s/accept", friendship.ID),
			ExpectedStatus:  http.StatusForbidden,
			ExpectedContent: []string{"not authorized to accept this request"},
			TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, bystander.Email)
				scenario.Headers = []string{tokenHeader}
			},
		}
		scenario.Test(t)
	})
}

func Test_DeclineFriendRequest_Success(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		requester := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		invitee := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		friendship, err := testApi.App.Adapter().Gaming().CreateFriendship(ctx, &models.Friendship{
			RequestingPlayerID: requester.ID,
			InvitedPlayerID:    invitee.ID,
			Status:             models.FriendshipStatusPending,
		})
		assert.NoError(t, err)

		scenario := &apis.ApiScenario{
			Name:           "decline",
			Method:         http.MethodPost,
			URL:            fmt.Sprintf("/players/friends/requests/%s/decline", friendship.ID),
			ExpectedStatus: http.StatusOK,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, invitee.Email)
				scenario.Headers = []string{tokenHeader}
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[apis.ApiSingleResponse[*apis.Friendship]](t, res.Body.Bytes())
				assert.Equal(t, apis.FriendshipStatusDeclined, result.Data.Status)
			},
		}
		scenario.Test(t)
	})
}

func Test_RemoveFriend_Success(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		player := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		friend := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		friendship, err := testApi.App.Adapter().Gaming().CreateFriendship(ctx, &models.Friendship{
			RequestingPlayerID: player.ID,
			InvitedPlayerID:    friend.ID,
			Status:             models.FriendshipStatusAccepted,
		})
		assert.NoError(t, err)

		scenario := &apis.ApiScenario{
			Name:           "remove friend",
			Method:         http.MethodDelete,
			URL:            fmt.Sprintf("/players/friends/%s", friendship.ID),
			ExpectedStatus: http.StatusNoContent,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, player.Email)
				scenario.Headers = []string{tokenHeader}
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				count, err := testApi.App.Adapter().Gaming().CountFriendships(ctx, &stores.FriendshipFilter{
					Ids: []uuid.UUID{friendship.ID},
				})
				assert.NoError(t, err)
				assert.Equal(t, int64(0), count)
			},
		}
		scenario.Test(t)
	})
}

func Test_RemoveFriend_NotInFriendship_Fails(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		player1 := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		player2 := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		bystander := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		friendship, err := testApi.App.Adapter().Gaming().CreateFriendship(ctx, &models.Friendship{
			RequestingPlayerID: player1.ID,
			InvitedPlayerID:    player2.ID,
			Status:             models.FriendshipStatusAccepted,
		})
		assert.NoError(t, err)

		scenario := &apis.ApiScenario{
			Name:            "bystander cannot remove",
			Method:          http.MethodDelete,
			URL:             fmt.Sprintf("/players/friends/%s", friendship.ID),
			ExpectedStatus:  http.StatusForbidden,
			ExpectedContent: []string{"not authorized"},
			TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, bystander.Email)
				scenario.Headers = []string{tokenHeader}
			},
		}
		scenario.Test(t)
	})
}

func Test_BlockPlayer_Success(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		player := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		target := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		body, _ := json.Marshal(apis.BlockPlayerBody{PlayerID: target.ID})
		scenario := &apis.ApiScenario{
			Name:           "block player",
			Method:         http.MethodPost,
			URL:            "/players/block",
			ExpectedStatus: http.StatusOK,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, player.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.Body = strings.NewReader(string(body))
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[apis.ApiSingleResponse[*apis.Friendship]](t, res.Body.Bytes())
				assert.Equal(t, apis.FriendshipStatusBlocked, result.Data.Status)
				assert.Equal(t, player.ID.String(), result.Data.RequestingPlayerID.String())
				assert.Equal(t, target.ID.String(), result.Data.InvitedPlayerID.String())
			},
		}
		scenario.Test(t)
	})
}

func Test_BlockPlayer_RemovesExistingFriendship(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		player := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		target := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		// Pre-existing accepted friendship
		_, err := testApi.App.Adapter().Gaming().CreateFriendship(ctx, &models.Friendship{
			RequestingPlayerID: player.ID,
			InvitedPlayerID:    target.ID,
			Status:             models.FriendshipStatusAccepted,
		})
		assert.NoError(t, err)

		body, _ := json.Marshal(apis.BlockPlayerBody{PlayerID: target.ID})
		scenario := &apis.ApiScenario{
			Name:           "block replaces friendship",
			Method:         http.MethodPost,
			URL:            "/players/block",
			ExpectedStatus: http.StatusOK,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, player.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.Body = strings.NewReader(string(body))
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[apis.ApiSingleResponse[*apis.Friendship]](t, res.Body.Bytes())
				assert.Equal(t, apis.FriendshipStatusBlocked, result.Data.Status)
				// Only one record should exist between these two players
				count, err := testApi.App.Adapter().Gaming().CountFriendships(ctx, &stores.FriendshipFilter{
					RequestingPlayerIds: []uuid.UUID{player.ID},
					InvitedPlayerIds:    []uuid.UUID{target.ID},
				})
				assert.NoError(t, err)
				assert.Equal(t, int64(1), count)
			},
		}
		scenario.Test(t)
	})
}

func Test_BlockPlayer_CannotSendFriendRequest_After(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		player := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		target := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		// block first
		_, err := testApi.App.Adapter().Gaming().CreateFriendship(ctx, &models.Friendship{
			RequestingPlayerID: player.ID,
			InvitedPlayerID:    target.ID,
			Status:             models.FriendshipStatusBlocked,
		})
		assert.NoError(t, err)

		body, _ := json.Marshal(apis.SendFriendRequestBody{InvitedPlayerID: target.ID})
		scenario := &apis.ApiScenario{
			Name:            "cannot request blocked player",
			Method:          http.MethodPost,
			URL:             "/players/friends/requests",
			ExpectedStatus:  http.StatusConflict,
			ExpectedContent: []string{"cannot send friend request to blocked player"},
			TestAppFactory:  func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, player.Email)
				scenario.Headers = []string{tokenHeader}
				scenario.Body = strings.NewReader(string(body))
			},
		}
		scenario.Test(t)
	})
}

func Test_UnblockPlayer_Success(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		player := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		target := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		_, err := testApi.App.Adapter().Gaming().CreateFriendship(ctx, &models.Friendship{
			RequestingPlayerID: player.ID,
			InvitedPlayerID:    target.ID,
			Status:             models.FriendshipStatusBlocked,
		})
		assert.NoError(t, err)

		scenario := &apis.ApiScenario{
			Name:           "unblock",
			Method:         http.MethodDelete,
			URL:            fmt.Sprintf("/players/block/%s", target.ID),
			ExpectedStatus: http.StatusNoContent,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, player.Email)
				scenario.Headers = []string{tokenHeader}
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				count, _ := testApi.App.Adapter().Gaming().CountFriendships(ctx, nil)
				assert.Equal(t, int64(0), count)
			},
		}
		scenario.Test(t)
	})
}

func Test_GetFriendship_None(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		player := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		other := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		scenario := &apis.ApiScenario{
			Name:           "no relationship",
			Method:         http.MethodGet,
			URL:            fmt.Sprintf("/players/%s/friendship", other.ID),
			ExpectedStatus: http.StatusOK,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, player.Email)
				scenario.Headers = []string{tokenHeader}
			},
			ExpectedContent: []string{`"data":null`},
		}
		scenario.Test(t)
	})
}

func Test_GetFriendship_Pending_RequestingDirection(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		player := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		other := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		_, err := testApi.App.Adapter().Gaming().CreateFriendship(ctx, &models.Friendship{
			RequestingPlayerID: player.ID,
			InvitedPlayerID:    other.ID,
			Status:             models.FriendshipStatusPending,
		})
		assert.NoError(t, err)

		scenario := &apis.ApiScenario{
			Name:           "pending outgoing",
			Method:         http.MethodGet,
			URL:            fmt.Sprintf("/players/%s/friendship", other.ID),
			ExpectedStatus: http.StatusOK,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, player.Email)
				scenario.Headers = []string{tokenHeader}
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[apis.ApiSingleResponse[*apis.Friendship]](t, res.Body.Bytes())
				assert.NotNil(t, result.Data)
				assert.Equal(t, apis.FriendshipStatusPending, result.Data.Status)
			},
		}
		scenario.Test(t)
	})
}

func Test_GetFriendship_Pending_InvitedDirection(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		player := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		other := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		_, err := testApi.App.Adapter().Gaming().CreateFriendship(ctx, &models.Friendship{
			RequestingPlayerID: other.ID,
			InvitedPlayerID:    player.ID,
			Status:             models.FriendshipStatusPending,
		})
		assert.NoError(t, err)

		scenario := &apis.ApiScenario{
			Name:           "pending incoming",
			Method:         http.MethodGet,
			URL:            fmt.Sprintf("/players/%s/friendship", other.ID),
			ExpectedStatus: http.StatusOK,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, player.Email)
				scenario.Headers = []string{tokenHeader}
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[apis.ApiSingleResponse[*apis.Friendship]](t, res.Body.Bytes())
				assert.NotNil(t, result.Data)
				assert.Equal(t, apis.FriendshipStatusPending, result.Data.Status)
			},
		}
		scenario.Test(t)
	})
}

func Test_GetFriendship_Accepted(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		player := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		other := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		_, err := testApi.App.Adapter().Gaming().CreateFriendship(ctx, &models.Friendship{
			RequestingPlayerID: player.ID,
			InvitedPlayerID:    other.ID,
			Status:             models.FriendshipStatusAccepted,
		})
		assert.NoError(t, err)

		scenario := &apis.ApiScenario{
			Name:           "accepted",
			Method:         http.MethodGet,
			URL:            fmt.Sprintf("/players/%s/friendship", other.ID),
			ExpectedStatus: http.StatusOK,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, player.Email)
				scenario.Headers = []string{tokenHeader}
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[apis.ApiSingleResponse[*apis.Friendship]](t, res.Body.Bytes())
				assert.NotNil(t, result.Data)
				assert.Equal(t, apis.FriendshipStatusAccepted, result.Data.Status)
			},
		}
		scenario.Test(t)
	})
}

func Test_GetFriendship_Blocked(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		testApi := apis.SetupApi(t, ctx, db)
		player := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))
		other := core.MustCreatePlayerWithOptions(t, testApi.App, core.WithPlayerRegistered(true))

		_, err := testApi.App.Adapter().Gaming().CreateFriendship(ctx, &models.Friendship{
			RequestingPlayerID: player.ID,
			InvitedPlayerID:    other.ID,
			Status:             models.FriendshipStatusBlocked,
		})
		assert.NoError(t, err)

		scenario := &apis.ApiScenario{
			Name:           "blocked",
			Method:         http.MethodGet,
			URL:            fmt.Sprintf("/players/%s/friendship", other.ID),
			ExpectedStatus: http.StatusOK,
			TestAppFactory: func(t testing.TB) *apis.TestApi { return testApi },
			BeforeTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario) {
				tokenHeader, _ := core.CreateAccessHeaderAndRefreshToken(t, testApi.App, player.Email)
				scenario.Headers = []string{tokenHeader}
			},
			AfterTestFunc: func(t testing.TB, app *core.BaseApp, scenario *apis.ApiScenario, res *httptest.ResponseRecorder) {
				result := test.MustUnMarshal[apis.ApiSingleResponse[*apis.Friendship]](t, res.Body.Bytes())
				assert.NotNil(t, result.Data)
				assert.Equal(t, apis.FriendshipStatusBlocked, result.Data.Status)
			},
		}
		scenario.Test(t)
	})
}
