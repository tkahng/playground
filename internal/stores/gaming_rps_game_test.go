package stores

import (
	"context"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/database/queries"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/test"
	"github.com/tkahng/playground/internal/tools/types"
)

func TestDBGamingStore_CreateRpsGame(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		gameStore := NewDBGamingStore(db)
		statusSelector := test.NewRandomeSelector(models.RpsGameStatusPending, models.RpsGameStatusCancelled, models.RpsGameStatusCompleted)
		for i := range 10 {
			game := &models.RpsGame{
				Status:    statusSelector.Select(),
				ExpiresAt: time.Now().UTC().Add(time.Hour * 1).UTC(),
				Metadata:  fmt.Appendf(nil, `{"idx": %d}`, i),
			}
			createdGame, err := gameStore.CreateRpsGame(ctx, game)
			if err != nil {
				t.Errorf("CreateRpsGame error: %v", err)
			}
			if createdGame.ID == uuid.Nil {
				t.Errorf("CreateRpsGame.ID() = %v, want id not nil", createdGame.ID)
			}
			if createdGame.Status != game.Status {
				t.Errorf("CreateRpsGame.Status() = %v, want status %v", createdGame.Status, game.Status)
			}
			if !createdGame.ExpiresAt.UTC().Equal(game.ExpiresAt.UTC()) {
				t.Errorf("CreateRpsGame.ExpiresAt() = %v, want expires at %v", createdGame.ExpiresAt, game.ExpiresAt)
			}
			if string(createdGame.Metadata) != string(game.Metadata) {
				t.Errorf("CreateRpsGame.Metadata() = %v, want metadata %v", createdGame.Metadata, game.Metadata)
			}
		}
	})
}

func TestDBGamingStore_UpdateRpsGame(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		gameStore := NewDBGamingStore(db)
		statusSelector := test.NewRandomeSelector(models.RpsGameStatusPending, models.RpsGameStatusCancelled, models.RpsGameStatusCompleted)
		for i := range 10 {
			game := &models.RpsGame{
				Status:    statusSelector.Select(),
				ExpiresAt: time.Now().UTC().Add(time.Hour * 1).UTC(),
				Metadata:  fmt.Appendf(nil, `{"idx": %d}`, i),
			}
			createdGame, err := gameStore.CreateRpsGame(ctx, game)
			if err != nil {
				t.Errorf("CreateRpsGame error: %v", err)
			}
			if createdGame.ID == uuid.Nil {
				t.Errorf("CreateRpsGame.ID() = %v, want id not nil", createdGame.ID)
			}
			if createdGame.Status != game.Status {
				t.Errorf("CreateRpsGame.Status() = %v, want status %v", createdGame.Status, game.Status)
			}
			if !createdGame.ExpiresAt.UTC().Truncate(time.Microsecond).Equal(game.ExpiresAt.UTC().Truncate(time.Microsecond)) {
				t.Errorf("CreateRpsGame.ExpiresAt() = %v, want expires at %v", createdGame.ExpiresAt, game.ExpiresAt)
			}
			if string(createdGame.Metadata) != string(game.Metadata) {
				t.Errorf("CreateRpsGame.Metadata() = %v, want metadata %v", createdGame.Metadata, game.Metadata)
			}
			createdGame.Status = statusSelector.Select()
			createdGame.Metadata = fmt.Appendf(nil, `{"idx": %d, "updated": true}`, i)
			createdGame.ExpiresAt = time.Now().Add(time.Minute * 3600).UTC()
			updatedGame, err := gameStore.UpdateRpsGame(ctx, createdGame)
			if err != nil {
				t.Errorf("UpdateRpsGame error: %v", err)
			}
			if updatedGame.ID != createdGame.ID {
				t.Errorf("UpdateRpsGame.ID() = %v, want id %v", updatedGame.ID, createdGame.ID)
			}
			if updatedGame.Status != createdGame.Status {
				t.Errorf("UpdateRpsGame.Status() = %v, want status %v", updatedGame.Status, createdGame.Status)
			}
			if !updatedGame.ExpiresAt.UTC().Truncate(time.Microsecond).Equal(createdGame.ExpiresAt.UTC().Truncate(time.Microsecond)) {
				t.Errorf("UpdateRpsGame.ExpiresAt() = %v, want expires at %v", updatedGame.ExpiresAt, createdGame.ExpiresAt)
			}
			if string(updatedGame.Metadata) != string(createdGame.Metadata) {
				t.Errorf("UpdateRpsGame.Metadata() = %v, want metadata %v", updatedGame.Metadata, createdGame.Metadata)
			}
		}
	})
}

func TestDBGamingStore_FindRpsGame_ByIds(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		gameStore := NewDBGamingStore(db)
		games := []*models.RpsGame{
			{
				Status:      models.RpsGameStatusCancelled,
				ExpiresAt:   time.Now().UTC().Add(time.Hour * 1).UTC(),
				CompletedAt: types.Pointer(time.Now().UTC().Add(time.Hour * 1).UTC()),
			},
			{
				Status:      models.RpsGameStatusCompleted,
				ExpiresAt:   time.Now().UTC().Add(time.Hour * 2).UTC(),
				CompletedAt: nil,
			},
			{
				Status:      models.RpsGameStatusPending,
				ExpiresAt:   time.Now().UTC().Add(time.Hour * 1).UTC(),
				CompletedAt: types.Pointer(time.Now().UTC().Add(time.Hour * 2).UTC()),
			},
		}
		createdGames := []*models.RpsGame{}
		for _, game := range games {
			createdGame, err := gameStore.CreateRpsGame(ctx, game)
			if err != nil {
				t.Fatalf("CreateRpsGame error: %v", err)
			}
			createdGames = append(createdGames, createdGame)
		}

		testCases := []struct {
			desc      string
			filter    *RpsGameFilter
			afterFunc func(t *testing.T, res []*models.RpsGame)
		}{
			{
				desc: "1,2",
				filter: &RpsGameFilter{
					Ids: []uuid.UUID{
						createdGames[0].ID,
						createdGames[1].ID,
					},
				},
				afterFunc: func(t *testing.T, result []*models.RpsGame) {
					if len(result) != 2 {
						t.Errorf("FindRpsGames() = %v, want %v", len(result), 2)
					}
					ids := []uuid.UUID{
						createdGames[0].ID,
						createdGames[1].ID,
					}
					for _, game := range result {
						if !slices.Contains(ids, game.ID) {
							t.Errorf("FindRpsGames() = %v, want %v", game.ID, ids)
						}
					}
				},
			},
			{
				desc: "1",
				filter: &RpsGameFilter{
					Ids: []uuid.UUID{
						createdGames[0].ID,
					},
				},
				afterFunc: func(t *testing.T, result []*models.RpsGame) {
					if len(result) != 1 {
						t.Errorf("FindRpsGames() = %v, want %v", len(result), 1)
					}
					ids := []uuid.UUID{
						createdGames[0].ID,
					}
					for _, game := range result {
						if !slices.Contains(ids, game.ID) {
							t.Errorf("FindRpsGames() = %v, want %v", game.ID, ids)
						}
					}
				},
			},
			{
				desc: "3",
				filter: &RpsGameFilter{
					Ids: []uuid.UUID{
						createdGames[2].ID,
					},
				},
				afterFunc: func(t *testing.T, result []*models.RpsGame) {
					if len(result) != 1 {
						t.Errorf("FindRpsGames() = %v, want %v", len(result), 1)
					}
					ids := []uuid.UUID{
						createdGames[2].ID,
					}
					for _, game := range result {
						if !slices.Contains(ids, game.ID) {
							t.Errorf("FindRpsGames() = %v, want %v", game.ID, ids)
						}
					}
				},
			},
		}
		for _, tC := range testCases {
			t.Run(tC.desc, func(t *testing.T) {
				result, err := gameStore.FindRpsGames(ctx, tC.filter)
				if err != nil {
					t.Fatalf("FindRpsGames error: %v", err)
				}
				tC.afterFunc(t, result)
			})
		}
	})
}

func TestDBGamingStore_FindRpsGame_ByStatus(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		gameStore := NewDBGamingStore(db)
		games := []*models.RpsGame{
			{
				Status:      models.RpsGameStatusCancelled,
				ExpiresAt:   time.Now().UTC().Add(time.Hour * 1).UTC(),
				CompletedAt: types.Pointer(time.Now().UTC().Add(time.Hour * 1).UTC()),
			},
			{
				Status:      models.RpsGameStatusCompleted,
				ExpiresAt:   time.Now().UTC().Add(time.Hour * 2).UTC(),
				CompletedAt: nil,
			},
			{
				Status:      models.RpsGameStatusPending,
				ExpiresAt:   time.Now().UTC().Add(time.Hour * 1).UTC(),
				CompletedAt: types.Pointer(time.Now().UTC().Add(time.Hour * 2).UTC()),
			},
		}
		createdGames := []*models.RpsGame{}
		for _, game := range games {
			createdGame, err := gameStore.CreateRpsGame(ctx, game)
			if err != nil {
				t.Fatalf("CreateRpsGame error: %v", err)
			}
			createdGames = append(createdGames, createdGame)
		}

		testCases := []struct {
			desc      string
			filter    *RpsGameFilter
			afterFunc func(t *testing.T, res []*models.RpsGame)
		}{
			{
				desc: "cancelled, completed",
				filter: &RpsGameFilter{
					Statuses: []models.RpsGameStatus{
						models.RpsGameStatusCancelled,
						models.RpsGameStatusCompleted,
					},
				},
				afterFunc: func(t *testing.T, result []*models.RpsGame) {
					if len(result) != 2 {
						t.Errorf("FindRpsGames() = %v, want %v", len(result), 2)
					}
					ids := []uuid.UUID{
						createdGames[0].ID,
						createdGames[1].ID,
					}
					for _, game := range result {
						if !slices.Contains(ids, game.ID) {
							t.Errorf("FindRpsGames() = %v, want %v", game.ID, ids)
						}
					}
				},
			},
			{
				desc: "cancelled",
				filter: &RpsGameFilter{
					Statuses: []models.RpsGameStatus{
						models.RpsGameStatusCancelled,
					},
				},
				afterFunc: func(t *testing.T, result []*models.RpsGame) {
					if len(result) != 1 {
						t.Errorf("FindRpsGames() = %v, want %v", len(result), 1)
					}
					ids := []uuid.UUID{
						createdGames[0].ID,
					}
					for _, game := range result {
						if !slices.Contains(ids, game.ID) {
							t.Errorf("FindRpsGames() = %v, want %v", game.ID, ids)
						}
					}
				},
			},
			{
				desc: "pending",
				filter: &RpsGameFilter{
					Statuses: []models.RpsGameStatus{
						models.RpsGameStatusPending,
					},
				},
				afterFunc: func(t *testing.T, result []*models.RpsGame) {
					if len(result) != 1 {
						t.Errorf("FindRpsGames() = %v, want %v", len(result), 1)
					}
					ids := []uuid.UUID{
						createdGames[2].ID,
					}
					for _, game := range result {
						if !slices.Contains(ids, game.ID) {
							t.Errorf("FindRpsGames() = %v, want %v", game.ID, ids)
						}
					}
				},
			},
		}
		for _, tC := range testCases {
			t.Run(tC.desc, func(t *testing.T) {
				result, err := gameStore.FindRpsGames(ctx, tC.filter)
				if err != nil {
					t.Fatalf("FindRpsGames error: %v", err)
				}
				tC.afterFunc(t, result)
			})
		}
	})
}

func TestDBGamingStore_FindRpsGame_ByCompletedAt(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		now := time.Now().UTC()
		gameStore := NewDBGamingStore(db)
		games := []*models.RpsGame{
			{
				Status:      models.RpsGameStatusCancelled,
				ExpiresAt:   now.Add(time.Hour * 1).UTC(),
				CompletedAt: types.Pointer(now.Add(time.Hour * 1).UTC()),
			},
			{
				Status:      models.RpsGameStatusCompleted,
				ExpiresAt:   now.Add(time.Hour * 2).UTC(),
				CompletedAt: nil,
			},
			{
				Status:      models.RpsGameStatusPending,
				ExpiresAt:   now.Add(time.Hour * 1).UTC(),
				CompletedAt: types.Pointer(now.Add(time.Hour * 2).UTC()),
			},
		}
		createdGames := []*models.RpsGame{}
		for _, game := range games {
			createdGame, err := gameStore.CreateRpsGame(ctx, game)
			if err != nil {
				t.Fatalf("CreateRpsGame error: %v", err)
			}
			createdGames = append(createdGames, createdGame)
		}

		testCases := []struct {
			desc      string
			filter    *RpsGameFilter
			afterFunc func(t *testing.T, res []*models.RpsGame)
		}{
			{
				desc: "later than now",
				filter: &RpsGameFilter{
					CompletedAt: types.OptionalParam[time.Time]{
						Value: now,
						IsSet: true,
					},
					CompletedAtOp: queries.FilterOperatorGte,
				},
				afterFunc: func(t *testing.T, result []*models.RpsGame) {
					if len(result) != 2 {
						t.Errorf("FindRpsGames() = %v, want %v", len(result), 2)
					}
					ids := []uuid.UUID{
						createdGames[0].ID,
						createdGames[2].ID,
					}
					for _, game := range result {
						if !slices.Contains(ids, game.ID) {
							t.Errorf("FindRpsGames() = %v, want %v", game.ID, ids)
						}
					}
				},
			},
			{
				desc: "later than hour from now",
				filter: &RpsGameFilter{
					CompletedAt: types.OptionalParam[time.Time]{
						Value: now.Add(time.Hour * 1).UTC(),
						IsSet: true,
					},
					CompletedAtOp: queries.FilterOperatorGte,
				},
				afterFunc: func(t *testing.T, result []*models.RpsGame) {
					if len(result) != 2 {
						t.Errorf("FindRpsGames() = %v, want %v", len(result), 2)
					}
					ids := []uuid.UUID{
						createdGames[0].ID,
						createdGames[2].ID,
					}
					for _, game := range result {
						if !slices.Contains(ids, game.ID) {
							t.Errorf("FindRpsGames() = %v, want %v", game.ID, ids)
						}
					}
				},
			},
			{
				desc: "later than 2 hour from now",
				filter: &RpsGameFilter{
					CompletedAt: types.OptionalParam[time.Time]{
						Value: now.Add(time.Hour * 2).UTC(),
						IsSet: true,
					},
					CompletedAtOp: queries.FilterOperatorGte,
				},
				afterFunc: func(t *testing.T, result []*models.RpsGame) {
					if len(result) != 1 {
						t.Errorf("FindRpsGames() = %v, want %v", len(result), 2)
					}
					ids := []uuid.UUID{
						createdGames[2].ID,
					}
					for _, game := range result {
						if !slices.Contains(ids, game.ID) {
							t.Errorf("FindRpsGames() = %v, want %v", game.ID, ids)
						}
					}
				},
			},
			{
				desc: "completed at null",
				filter: &RpsGameFilter{
					CompletedAtOp: queries.FilterOperatorNull,
				},
				afterFunc: func(t *testing.T, result []*models.RpsGame) {
					if len(result) != 1 {
						t.Errorf("FindRpsGames() = %v, want %v", len(result), 2)
					}
					ids := []uuid.UUID{
						createdGames[1].ID,
					}
					for _, game := range result {
						if !slices.Contains(ids, game.ID) {
							t.Errorf("FindRpsGames() = %v, want %v", game.ID, ids)
						}
					}
				},
			},
			{
				desc: "completed at not null",
				filter: &RpsGameFilter{
					CompletedAtOp: queries.FilterOperatorNotNull,
				},
				afterFunc: func(t *testing.T, result []*models.RpsGame) {
					if len(result) != 2 {
						t.Errorf("FindRpsGames() = %v, want %v", len(result), 2)
					}
					ids := []uuid.UUID{
						createdGames[0].ID,
						createdGames[2].ID,
					}
					for _, game := range result {
						if !slices.Contains(ids, game.ID) {
							t.Errorf("FindRpsGames() = %v, want %v", game.ID, ids)
						}
					}
				},
			},
		}
		for _, tC := range testCases {
			t.Run(tC.desc, func(t *testing.T) {
				result, err := gameStore.FindRpsGames(ctx, tC.filter)
				if err != nil {
					t.Fatalf("FindRpsGames error: %v", err)
				}
				tC.afterFunc(t, result)
			})
		}
	})
}

func TestDBGamingStore_FindRpsGame_ByExpiresAt(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		now := time.Now().UTC()
		gameStore := NewDBGamingStore(db)
		games := []*models.RpsGame{
			{
				Status:    models.RpsGameStatusCancelled,
				ExpiresAt: now.Add(time.Hour * 1).UTC(),
			},
			{
				Status:    models.RpsGameStatusCompleted,
				ExpiresAt: now.Add(time.Hour * 2).UTC(),
			},
			{
				Status:    models.RpsGameStatusPending,
				ExpiresAt: now.Add(time.Hour * 3).UTC(),
			},
		}
		createdGames := []*models.RpsGame{}
		for _, game := range games {
			createdGame, err := gameStore.CreateRpsGame(ctx, game)
			if err != nil {
				t.Fatalf("CreateRpsGame error: %v", err)
			}
			createdGames = append(createdGames, createdGame)
		}

		testCases := []struct {
			desc      string
			filter    *RpsGameFilter
			afterFunc func(t *testing.T, res []*models.RpsGame)
		}{
			{
				desc: "later than hour from now",
				filter: &RpsGameFilter{
					ExpiresAt: types.OptionalParam[time.Time]{
						Value: now.Add(time.Hour * 1).UTC(),
						IsSet: true,
					},
					ExpiresAtOp: queries.FilterOperatorGte,
				},
				afterFunc: func(t *testing.T, result []*models.RpsGame) {
					if len(result) != 3 {
						t.Errorf("FindRpsGames() = %v, want %v", len(result), 2)
					}
					ids := []uuid.UUID{
						createdGames[0].ID,
						createdGames[1].ID,
						createdGames[2].ID,
					}
					for _, game := range result {
						if !slices.Contains(ids, game.ID) {
							t.Errorf("FindRpsGames() = %v, want %v", game.ID, ids)
						}
					}
				},
			},
			{
				desc: "later than 2 hour from now",
				filter: &RpsGameFilter{
					ExpiresAt: types.OptionalParam[time.Time]{
						Value: now.Add(time.Hour * 2).UTC(),
						IsSet: true,
					},
					ExpiresAtOp: queries.FilterOperatorGte,
				},
				afterFunc: func(t *testing.T, result []*models.RpsGame) {
					if len(result) != 2 {
						t.Errorf("FindRpsGames() = %v, want %v", len(result), 2)
					}
					ids := []uuid.UUID{
						createdGames[1].ID,
						createdGames[2].ID,
					}
					for _, game := range result {
						if !slices.Contains(ids, game.ID) {
							t.Errorf("FindRpsGames() = %v, want %v", game.ID, ids)
						}
					}
				},
			},
		}
		for _, tC := range testCases {
			t.Run(tC.desc, func(t *testing.T) {
				result, err := gameStore.FindRpsGames(ctx, tC.filter)
				if err != nil {
					t.Fatalf("FindRpsGames error: %v", err)
				}
				tC.afterFunc(t, result)
			})
		}
	})
}
