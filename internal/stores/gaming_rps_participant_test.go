package stores

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/test"
	"github.com/tkahng/playground/internal/tools/types"
)

func TestDBGamingStore_CreateRpsParticipant(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		gameStore := NewDBGamingStore(db)
		rpsGameStatusSelector := test.NewRandomeSelector(models.RpsGameStatusPending, models.RpsGameStatusCancelled, models.RpsGameStatusCompleted)
		moveSelector := test.NewRandomeSelector(models.RpsParticipantMoveRock, models.RpsParticipantMovePaper, models.RpsParticipantMoveScissors)
		typeSelector := test.NewRandomeSelector(models.RpsParticipantTypeGuest, models.RpsParticipantTypeHost)
		rpsParticipantStatusSelector := test.NewRandomeSelector(models.RpsParticipantStatusPending, models.RpsParticipantStatusDeclined, models.RpsParticipantStatusCompleted)
		rpsParticipantResultSelector := test.NewRandomeSelector(models.RpsParticipantResultWin, models.RpsParticipantResultTie, models.RpsParticipantResultLose)
		for i := range 10 {
			game := &models.RpsGame{
				Status:    rpsGameStatusSelector.Select(),
				ExpiresAt: time.Now().UTC().Add(time.Hour * 1).UTC(),
				Metadata:  fmt.Appendf(nil, `{"idx": %d}`, i),
			}
			createdGame, err := gameStore.CreateRpsGame(ctx, game)
			if err != nil {
				t.Errorf("CreateRpsGame error: %v", err)
			}
			player, err := gameStore.CreatePlayer(ctx, &models.Player{
				Email: fmt.Sprintf("%d@example.com", i),
			})
			if err != nil {
				t.Errorf("CreatePlayer error: %v", err)
			}
			participant := &models.RpsParticipant{
				GameID:   createdGame.ID,
				PlayerID: player.ID,
				Move:     moveSelector.Select(),
				Type:     typeSelector.Select(),
				Status:   rpsParticipantStatusSelector.Select(),
				Result:   rpsParticipantResultSelector.Select(),
				// RespondedAt: ,
			}
			createdParticipant, err := gameStore.CreateRpsParticipant(ctx, participant)
			if err != nil {
				t.Errorf("CreateRpsParticipant error: %v", err)
			}
			if createdParticipant.ID == uuid.Nil {
				t.Errorf("CreateRpsParticipant.ID() = %v, want id not nil", createdParticipant.ID)
			}
			if createdParticipant.GameID != participant.GameID {
				t.Errorf("CreateRpsParticipant.GameID() = %v, want game id %v", createdParticipant.GameID, participant.GameID)
			}
			if createdParticipant.PlayerID != participant.PlayerID {
				t.Errorf("CreateRpsParticipant.PlayerID() = %v, want player id %v", createdParticipant.PlayerID, participant.PlayerID)
			}
			if createdParticipant.Move != participant.Move {
				t.Errorf("CreateRpsParticipant.Move() = %v, want move %v", createdParticipant.Move, participant.Move)
			}
			if createdParticipant.Type != participant.Type {
				t.Errorf("CreateRpsParticipant.Type() = %v, want type %v", createdParticipant.Type, participant.Type)
			}
			if createdParticipant.Status != participant.Status {
				t.Errorf("CreateRpsParticipant.Status() = %v, want status %v", createdParticipant.Status, participant.Status)
			}
			if createdParticipant.Result != participant.Result {
				t.Errorf("CreateRpsParticipant.Result() = %v, want result %v", createdParticipant.Result, participant.Result)
			}
		}
	})
}
func TestDBGamingStore_CreateRpsParticipants(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		gameStore := NewDBGamingStore(db)
		rpsGameStatusSelector := test.NewRandomeSelector(models.RpsGameStatusPending, models.RpsGameStatusCancelled, models.RpsGameStatusCompleted)
		moveSelector := test.NewRandomeSelector(models.RpsParticipantMoveRock, models.RpsParticipantMovePaper, models.RpsParticipantMoveScissors)
		rpsParticipantStatusSelector := test.NewRandomeSelector(models.RpsParticipantStatusPending, models.RpsParticipantStatusDeclined, models.RpsParticipantStatusCompleted)
		rpsParticipantResultSelector := test.NewRandomeSelector(models.RpsParticipantResultWin, models.RpsParticipantResultTie, models.RpsParticipantResultLose)
		for i := range 10 {
			game := &models.RpsGame{
				Status:    rpsGameStatusSelector.Select(),
				ExpiresAt: time.Now().UTC().Add(time.Hour * 1).UTC(),
				Metadata:  fmt.Appendf(nil, `{"idx": %d}`, i),
			}
			createdGame, err := gameStore.CreateRpsGame(ctx, game)
			if err != nil {
				t.Errorf("CreateRpsGame error: %v", err)
			}
			player, err := gameStore.CreatePlayer(ctx, &models.Player{
				Email: fmt.Sprintf("%d@example.com", i),
			})
			if err != nil {
				t.Errorf("CreatePlayer error: %v", err)
			}
			player2, err := gameStore.CreatePlayer(ctx, &models.Player{
				Email: fmt.Sprintf("%d-2@example.com", i),
			})
			if err != nil {
				t.Errorf("CreatePlayer error: %v", err)
			}
			participant1 := &models.RpsParticipant{
				GameID:   createdGame.ID,
				PlayerID: player.ID,
				Move:     moveSelector.Select(),
				Type:     models.RpsParticipantTypeHost,
				Status:   rpsParticipantStatusSelector.Select(),
				Result:   rpsParticipantResultSelector.Select(),
			}
			participant2 := &models.RpsParticipant{
				GameID:   createdGame.ID,
				PlayerID: player2.ID,
				Move:     moveSelector.Select(),
				Type:     models.RpsParticipantTypeGuest,
				Status:   rpsParticipantStatusSelector.Select(),
				Result:   rpsParticipantResultSelector.Select(),
			}
			participants := []*models.RpsParticipant{participant1, participant2}
			createdParticipants, err := gameStore.CreateRpsParticipants(ctx, participants)
			if err != nil {
				t.Errorf("CreateRpsParticipant error: %v", err)
			}
			for idx, createdParticipant := range createdParticipants {
				participant := participants[idx]
				if createdParticipant.ID == uuid.Nil {
					t.Errorf("CreateRpsParticipant.ID() = %v, want id not nil", createdParticipant.ID)
				}
				if createdParticipant.GameID != participant.GameID {
					t.Errorf("CreateRpsParticipant.GameID() = %v, want game id %v", createdParticipant.GameID, participant.GameID)
				}
				if createdParticipant.PlayerID != participant.PlayerID {
					t.Errorf("CreateRpsParticipant.PlayerID() = %v, want player id %v", createdParticipant.PlayerID, participant.PlayerID)
				}
				if createdParticipant.Move != participant.Move {
					t.Errorf("CreateRpsParticipant.Move() = %v, want move %v", createdParticipant.Move, participant.Move)
				}
				if createdParticipant.Type != participant.Type {
					t.Errorf("CreateRpsParticipant.Type() = %v, want type %v", createdParticipant.Type, participant.Type)
				}
				if createdParticipant.Status != participant.Status {
					t.Errorf("CreateRpsParticipant.Status() = %v, want status %v", createdParticipant.Status, participant.Status)
				}
				if createdParticipant.Result != participant.Result {
					t.Errorf("CreateRpsParticipant.Result() = %v, want result %v", createdParticipant.Result, participant.Result)
				}
			}
		}
	})
}

func TestDBGamingStore_CreateRpsParticipant_ParticipantTypeConstraint(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		gameStore := NewDBGamingStore(db)
		rpsGameStatusSelector := test.NewRandomeSelector(models.RpsGameStatusPending, models.RpsGameStatusCancelled, models.RpsGameStatusCompleted)
		moveSelector := test.NewRandomeSelector(models.RpsParticipantMoveRock, models.RpsParticipantMovePaper, models.RpsParticipantMoveScissors)
		rpsParticipantStatusSelector := test.NewRandomeSelector(models.RpsParticipantStatusPending, models.RpsParticipantStatusDeclined, models.RpsParticipantStatusCompleted)
		rpsParticipantResultSelector := test.NewRandomeSelector(models.RpsParticipantResultWin, models.RpsParticipantResultTie, models.RpsParticipantResultLose)
		for i := range 1 {
			game := &models.RpsGame{
				Status:    rpsGameStatusSelector.Select(),
				ExpiresAt: time.Now().UTC().Add(time.Hour * 1).UTC(),
				Metadata:  fmt.Appendf(nil, `{"idx": %d}`, i),
			}
			createdGame, err := gameStore.CreateRpsGame(ctx, game)
			if err != nil {
				t.Fatalf("CreateRpsGame error: %v", err)
			}
			player, err := gameStore.CreatePlayer(ctx, &models.Player{
				Email: fmt.Sprintf("%d@example.com", i),
			})
			if err != nil {
				t.Fatalf("CreatePlayer error: %v", err)
			}
			player2, err := gameStore.CreatePlayer(ctx, &models.Player{
				Email: fmt.Sprintf("%d-2@example.com", i),
			})
			if err != nil {
				t.Fatalf("CreatePlayer error: %v", err)
			}
			player3, err := gameStore.CreatePlayer(ctx, &models.Player{
				Email: fmt.Sprintf("%d-3@example.com", i),
			})
			if err != nil {
				t.Fatalf("CreatePlayer error: %v", err)
			}
			participant := &models.RpsParticipant{
				GameID:   createdGame.ID,
				PlayerID: player.ID,
				Move:     moveSelector.Select(),
				Type:     models.RpsParticipantTypeGuest,
				Status:   rpsParticipantStatusSelector.Select(),
				Result:   rpsParticipantResultSelector.Select(),
			}
			_, err = gameStore.CreateRpsParticipant(ctx, participant)
			if err != nil {
				t.Errorf("CreateRpsParticipant error: %v", err)
			}
			participant2 := &models.RpsParticipant{
				GameID:   createdGame.ID,
				PlayerID: player2.ID,
				Move:     moveSelector.Select(),
				Type:     models.RpsParticipantTypeHost,
				Status:   rpsParticipantStatusSelector.Select(),
				Result:   rpsParticipantResultSelector.Select(),
			}
			_, err = gameStore.CreateRpsParticipant(ctx, participant2)
			if err != nil {
				t.Errorf("CreateRpsParticipant error: %v", err)
			}
			participant3 := &models.RpsParticipant{
				GameID:   createdGame.ID,
				PlayerID: player3.ID,
				Move:     moveSelector.Select(),
				Type:     models.RpsParticipantTypeHost,
				Status:   rpsParticipantStatusSelector.Select(),
				Result:   rpsParticipantResultSelector.Select(),
			}
			_, err = gameStore.CreateRpsParticipant(ctx, participant3)
			if err == nil {
				t.Errorf("CreateRpsParticipant error: %v", err)
			}
			if !strings.Contains(err.Error(), "violates unique constraint") {
				t.Errorf("CreateRpsParticipant error: %v", err)
			}
		}
	})
}
func TestDBGamingStore_CreateRpsParticipant_ParticipantPlayerIDConstraint(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		gameStore := NewDBGamingStore(db)
		rpsGameStatusSelector := test.NewRandomeSelector(models.RpsGameStatusPending, models.RpsGameStatusCancelled, models.RpsGameStatusCompleted)
		moveSelector := test.NewRandomeSelector(models.RpsParticipantMoveRock, models.RpsParticipantMovePaper, models.RpsParticipantMoveScissors)
		rpsParticipantStatusSelector := test.NewRandomeSelector(models.RpsParticipantStatusPending, models.RpsParticipantStatusDeclined, models.RpsParticipantStatusCompleted)
		rpsParticipantResultSelector := test.NewRandomeSelector(models.RpsParticipantResultWin, models.RpsParticipantResultTie, models.RpsParticipantResultLose)
		for i := range 1 {
			game := &models.RpsGame{
				Status:    rpsGameStatusSelector.Select(),
				ExpiresAt: time.Now().UTC().Add(time.Hour * 1).UTC(),
				Metadata:  fmt.Appendf(nil, `{"idx": %d}`, i),
			}
			createdGame, err := gameStore.CreateRpsGame(ctx, game)
			if err != nil {
				t.Fatalf("CreateRpsGame error: %v", err)
			}
			player, err := gameStore.CreatePlayer(ctx, &models.Player{
				Email: fmt.Sprintf("%d@example.com", i),
			})
			if err != nil {
				t.Fatalf("CreatePlayer error: %v", err)
			}

			participant := &models.RpsParticipant{
				GameID:   createdGame.ID,
				PlayerID: player.ID,
				Move:     moveSelector.Select(),
				Type:     models.RpsParticipantTypeGuest,
				Status:   rpsParticipantStatusSelector.Select(),
				Result:   rpsParticipantResultSelector.Select(),
			}
			_, err = gameStore.CreateRpsParticipant(ctx, participant)
			if err != nil {
				t.Errorf("CreateRpsParticipant error: %v", err)
			}
			participant2 := &models.RpsParticipant{
				GameID:   createdGame.ID,
				PlayerID: player.ID,
				Move:     moveSelector.Select(),
				Type:     models.RpsParticipantTypeHost,
				Status:   rpsParticipantStatusSelector.Select(),
				Result:   rpsParticipantResultSelector.Select(),
			}
			_, err = gameStore.CreateRpsParticipant(ctx, participant2)
			if err == nil {
				t.Errorf("CreateRpsParticipant error: %v", err)
			}
			if !strings.Contains(err.Error(), "violates unique constraint") {
				t.Errorf("CreateRpsParticipant error: %v", err)
			}
		}
	})
}

func TestDBGamingStore_UpdateRpsParticipant(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		gameStore := NewDBGamingStore(db)
		rpsGameStatusSelector := test.NewRandomeSelector(models.RpsGameStatusPending, models.RpsGameStatusCancelled, models.RpsGameStatusCompleted)
		moveSelector := test.NewRandomeSelector(models.RpsParticipantMoveRock, models.RpsParticipantMovePaper, models.RpsParticipantMoveScissors)
		typeSelector := test.NewRandomeSelector(models.RpsParticipantTypeGuest, models.RpsParticipantTypeHost)
		rpsParticipantStatusSelector := test.NewRandomeSelector(models.RpsParticipantStatusPending, models.RpsParticipantStatusDeclined, models.RpsParticipantStatusCompleted)
		rpsParticipantResultSelector := test.NewRandomeSelector(models.RpsParticipantResultWin, models.RpsParticipantResultTie, models.RpsParticipantResultLose)
		for i := range 10 {
			game := &models.RpsGame{
				Status:    rpsGameStatusSelector.Select(),
				ExpiresAt: time.Now().UTC().Add(time.Hour * 1).UTC(),
				Metadata:  fmt.Appendf(nil, `{"idx": %d}`, i),
			}
			createdGame, err := gameStore.CreateRpsGame(ctx, game)
			if err != nil {
				t.Errorf("CreateRpsGame error: %v", err)
			}
			player, err := gameStore.CreatePlayer(ctx, &models.Player{
				Email: fmt.Sprintf("%d@example.com", i),
			})
			if err != nil {
				t.Errorf("CreatePlayer error: %v", err)
			}
			participant := &models.RpsParticipant{
				GameID:   createdGame.ID,
				PlayerID: player.ID,
				Move:     moveSelector.Select(),
				Type:     typeSelector.Select(),
				Status:   rpsParticipantStatusSelector.Select(),
				Result:   rpsParticipantResultSelector.Select(),
				// RespondedAt: ,
			}
			createdParticipant, err := gameStore.CreateRpsParticipant(ctx, participant)
			if err != nil {
				t.Errorf("CreateRpsParticipant error: %v", err)
			}
			if createdParticipant.ID == uuid.Nil {
				t.Errorf("CreateRpsParticipant.ID() = %v, want id not nil", createdParticipant.ID)
			}
			if createdParticipant.GameID != participant.GameID {
				t.Errorf("CreateRpsParticipant.GameID() = %v, want game id %v", createdParticipant.GameID, participant.GameID)
			}
			if createdParticipant.PlayerID != participant.PlayerID {
				t.Errorf("CreateRpsParticipant.PlayerID() = %v, want player id %v", createdParticipant.PlayerID, participant.PlayerID)
			}
			if createdParticipant.Move != participant.Move {
				t.Errorf("CreateRpsParticipant.Move() = %v, want move %v", createdParticipant.Move, participant.Move)
			}
			if createdParticipant.Type != participant.Type {
				t.Errorf("CreateRpsParticipant.Type() = %v, want type %v", createdParticipant.Type, participant.Type)
			}
			if createdParticipant.Status != participant.Status {
				t.Errorf("CreateRpsParticipant.Status() = %v, want status %v", createdParticipant.Status, participant.Status)
			}
			if createdParticipant.Result != participant.Result {
				t.Errorf("CreateRpsParticipant.Result() = %v, want result %v", createdParticipant.Result, participant.Result)
			}

			// update
			createdParticipant.Move = moveSelector.Select()
			createdParticipant.Type = typeSelector.Select()
			createdParticipant.Status = rpsParticipantStatusSelector.Select()
			createdParticipant.Result = rpsParticipantResultSelector.Select()
			updatedParticipant, err := gameStore.UpdateRpsParticipant(ctx, createdParticipant)
			if err != nil {
				t.Errorf("UpdateRpsParticipant error: %v", err)
			}
			if updatedParticipant.ID != createdParticipant.ID {
				t.Errorf("UpdateRpsParticipant.ID() = %v, want id %v", updatedParticipant.ID, createdParticipant.ID)
			}
			if updatedParticipant.GameID != createdParticipant.GameID {
				t.Errorf("UpdateRpsParticipant.GameID() = %v, want game id %v", updatedParticipant.GameID, createdParticipant.GameID)
			}
			if updatedParticipant.PlayerID != createdParticipant.PlayerID {
				t.Errorf("UpdateRpsParticipant.PlayerID() = %v, want player id %v", updatedParticipant.PlayerID, createdParticipant.PlayerID)
			}
			if updatedParticipant.Move != createdParticipant.Move {
				t.Errorf("UpdateRpsParticipant.Move() = %v, want move %v", updatedParticipant.Move, createdParticipant.Move)
			}
			if updatedParticipant.Type != createdParticipant.Type {
				t.Errorf("UpdateRpsParticipant.Type() = %v, want type %v", updatedParticipant.Type, createdParticipant.Type)
			}
			if updatedParticipant.Status != createdParticipant.Status {
				t.Errorf("UpdateRpsParticipant.Status() = %v, want status %v", updatedParticipant.Status, createdParticipant.Status)
			}
			if updatedParticipant.Result != createdParticipant.Result {
				t.Errorf("UpdateRpsParticipant.Result() = %v, want result %v", updatedParticipant.Result, createdParticipant.Result)
			}
		}
	})
}

func TestDBGamingStore_FindRpsParticipant(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		now := time.Now().UTC()
		gameStore := NewDBGamingStore(db)
		// player1 is the guest (status=completed) across all games — no constraint issue.
		// Each game gets its own distinct host player (status=pending) so no player
		// appears in more than one pending participant row, satisfying the partial
		// unique index idx_rps_participants_one_pending_per_player.
		player1, err := gameStore.CreatePlayer(ctx, &models.Player{Email: "findpart_p1@gmail.com"})
		require.NoError(t, err)
		hostPlayers := make([]*models.Player, 3)
		for i := range hostPlayers {
			hostPlayers[i], err = gameStore.CreatePlayer(ctx, &models.Player{
				Email: fmt.Sprintf("findpart_host%d@gmail.com", i),
			})
			require.NoError(t, err)
		}

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
		participants := []*models.RpsParticipant{}
		createdGames := []*models.RpsGame{}
		for _, game := range games {
			createdGame, err := gameStore.CreateRpsGame(ctx, game)
			if err != nil {
				t.Fatalf("CreateRpsGame error: %v", err)
			}
			createdGames = append(createdGames, createdGame)
		}
		for i, createdGame := range createdGames {
			ps := []*models.RpsParticipant{
				{
					GameID:      createdGame.ID,
					PlayerID:    player1.ID,
					Type:        models.RpsParticipantTypeGuest,
					Status:      models.RpsParticipantStatusCompleted,
					Move:        models.RpsParticipantMovePaper,
					Result:      models.RpsParticipantResultTie,
					RespondedAt: nil,
				},
				{
					GameID:      createdGame.ID,
					PlayerID:    hostPlayers[i].ID, // distinct host per game — satisfies unique index
					Type:        models.RpsParticipantTypeHost,
					Status:      models.RpsParticipantStatusPending,
					Move:        models.RpsParticipantMovePaper,
					Result:      models.RpsParticipantResultTie,
					RespondedAt: nil,
				},
			}
			for _, p := range ps {
				createdParticipant, err := gameStore.CreateRpsParticipant(ctx, p)
				if err != nil {
					t.Fatalf("CreateRpsParticipant error: %v", err)
				}
				participants = append(participants, createdParticipant)
			}
		}
		testCases := []struct {
			desc      string
			filter    *RpsParticipantFilter
			afterFunc func(t *testing.T, res []*models.RpsParticipant)
		}{
			{
				desc: "by status pending",
				filter: &RpsParticipantFilter{
					Statuses: []models.RpsParticipantStatus{
						models.RpsParticipantStatusPending,
					},
				},
				afterFunc: func(t *testing.T, result []*models.RpsParticipant) {
					if len(result) != 3 {
						t.Errorf("FindRpsGames() = %v, want %v", len(result), 3)
					}
				},
			},
		}
		for _, tC := range testCases {
			t.Run(tC.desc, func(t *testing.T) {
				result, err := gameStore.FindRpsParticipants(ctx, tC.filter)
				if err != nil {
					t.Fatalf("FindRpsGames error: %v", err)
				}
				tC.afterFunc(t, result)
			})
		}
	})
}

func TestDBGamingStore_FindRpsParticipant_Single(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		gameStore := NewDBGamingStore(db)

		game, err := gameStore.CreateRpsGame(ctx, &models.RpsGame{
			Status:   models.RpsGameStatusPending,
			ExpiresAt: time.Now().UTC().Add(time.Hour),
			Metadata:  []byte("{}"),
		})
		if err != nil {
			t.Fatalf("CreateRpsGame error: %v", err)
		}
		player, err := gameStore.CreatePlayer(ctx, &models.Player{Email: "single_p@example.com"})
		if err != nil {
			t.Fatalf("CreatePlayer error: %v", err)
		}
		created, err := gameStore.CreateRpsParticipant(ctx, &models.RpsParticipant{
			GameID:   game.ID,
			PlayerID: player.ID,
			Move:     models.RpsParticipantMoveRock,
			Type:     models.RpsParticipantTypeHost,
			Status:   models.RpsParticipantStatusPending,
			Result:   models.RpsParticipantResultTie,
		})
		if err != nil {
			t.Fatalf("CreateRpsParticipant error: %v", err)
		}

		// FindRpsParticipant (single) by game ID.
		found, err := gameStore.FindRpsParticipant(ctx, &RpsParticipantFilter{
			RpsGameIds: []uuid.UUID{game.ID},
		})
		if err != nil {
			t.Fatalf("FindRpsParticipant() error = %v", err)
		}
		if found == nil {
			t.Fatal("FindRpsParticipant() returned nil, want participant")
		}
		if found.ID != created.ID {
			t.Errorf("FindRpsParticipant() ID = %v, want %v", found.ID, created.ID)
		}

		// FindRpsParticipant for non-existent game returns nil.
		missing, err := gameStore.FindRpsParticipant(ctx, &RpsParticipantFilter{
			RpsGameIds: []uuid.UUID{uuid.New()},
		})
		if err != nil {
			t.Fatalf("FindRpsParticipant(missing) error = %v", err)
		}
		if missing != nil {
			t.Errorf("FindRpsParticipant(missing) = %v, want nil", missing)
		}
	})
}

func TestDBGamingStore_CountRpsParticipants(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		gameStore := NewDBGamingStore(db)

		game, err := gameStore.CreateRpsGame(ctx, &models.RpsGame{
			Status:   models.RpsGameStatusPending,
			ExpiresAt: time.Now().UTC().Add(time.Hour),
			Metadata:  []byte("{}"),
		})
		if err != nil {
			t.Fatalf("CreateRpsGame error: %v", err)
		}
		for i, pType := range []models.RpsParticipantType{models.RpsParticipantTypeHost, models.RpsParticipantTypeGuest} {
			player, err := gameStore.CreatePlayer(ctx, &models.Player{
				Email: fmt.Sprintf("count_p%d@example.com", i),
			})
			if err != nil {
				t.Fatalf("CreatePlayer error: %v", err)
			}
			_, err = gameStore.CreateRpsParticipant(ctx, &models.RpsParticipant{
				GameID:   game.ID,
				PlayerID: player.ID,
				Move:     models.RpsParticipantMoveRock,
				Type:     pType,
				Status:   models.RpsParticipantStatusPending,
				Result:   models.RpsParticipantResultTie,
			})
			if err != nil {
				t.Fatalf("CreateRpsParticipant error: %v", err)
			}
		}

		count, err := gameStore.CountRpsParticipants(ctx, &RpsParticipantFilter{
			RpsGameIds: []uuid.UUID{game.ID},
		})
		if err != nil {
			t.Fatalf("CountRpsParticipants() error = %v", err)
		}
		if count != 2 {
			t.Errorf("CountRpsParticipants() = %d, want 2", count)
		}

		// Filter by status with no matches.
		count, err = gameStore.CountRpsParticipants(ctx, &RpsParticipantFilter{
			RpsGameIds: []uuid.UUID{game.ID},
			Statuses:   []models.RpsParticipantStatus{models.RpsParticipantStatusCompleted},
		})
		if err != nil {
			t.Fatalf("CountRpsParticipants(completed) error = %v", err)
		}
		if count != 0 {
			t.Errorf("CountRpsParticipants(completed) = %d, want 0", count)
		}
	})
}
