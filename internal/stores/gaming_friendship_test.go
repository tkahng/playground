package stores

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/database/repository"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/test"
)

func TestDBGamingStore_CreateUpdateFindCount(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		// init
		gamingStore := NewDBGamingStore(db)

		// declaration
		players := []*models.Player{}
		friendships := []*models.Frindship{}
		friendshipsMap := map[uuid.UUID]*models.Frindship{}
		playerIdFriendshipMap := map[uuid.UUID][]*models.Frindship{}

		playerCount := 10

		// create players
		for i := range playerCount {
			player := MustCreatePlayer(t, ctx, gamingStore, WithPlayerEmail(fmt.Sprintf("user%da@example.com", i)))
			players = append(players, player)
		}

		// create -------------------------------------------------------------------
		// for each player in players
		for i, player1 := range players {
			// for each progressive player, increment i as to avoid duplicate friendships
			for j := i + 1; j < len(players); j++ {
				player2 := players[j]
				friendship, err := gamingStore.CreateFriendship(ctx, &models.Frindship{
					RequestingPlayerID: player1.ID,
					InvitedPlayerID:    player2.ID,
				})
				friendships = append(friendships, friendship)
				friendshipsMap[friendship.ID] = friendship
				playerIdFriendshipMap[friendship.InvitedPlayerID] = append(playerIdFriendshipMap[friendship.InvitedPlayerID], friendship)
				playerIdFriendshipMap[friendship.RequestingPlayerID] = append(playerIdFriendshipMap[friendship.RequestingPlayerID], friendship)
				if err != nil {
					t.Fatalf("CreateFriendship() error = %v", err)
				}
				if friendship == nil {
					t.Errorf("CreateFriendship() = %v, want not nil", friendship)
				}
				if friendship.ID == uuid.Nil {
					t.Errorf("CreateFriendship() = %v, want id not nil", friendship)
				}
				if friendship.InvitedPlayerID != player2.ID {
					t.Errorf("CreateFriendship() = %v, want invited player id %v", friendship, player1.ID)
				}
				if friendship.RequestingPlayerID != player1.ID {
					t.Errorf("CreateFriendship() = %v, want requesting player id %v", friendship, player2.ID)
				}
				if friendship.Status != models.FriendshipStatusPending {
					t.Errorf("CreateFriendship() = %v, want status %v", friendship, models.FriendshipStatusPending)
				}
			}
		}
		// find -------------------------------------------------------------------
		// for each player in playerIdFriendshipMap
		for k, m := range playerIdFriendshipMap {
			// find player
			player, err := gamingStore.FindPlayer(ctx, &PlayersFilter{
				Ids: []uuid.UUID{k},
			})
			if err != nil {
				t.Fatalf("FindFriendship() error = %v", err)
			}
			if player == nil {
				t.Errorf("FindFriendship() = %v, want not nil", player)
			}
			// friendship count should be less thant total player count
			if len(m) != (playerCount - 1) {
				t.Errorf("Friendship count for player id %v should be one less than total player count. got %v, want %v", k, len(m), playerCount-1)
			}
		}
		// update -------------------------------------------------------------------
		// random selector
		providerSelector := test.NewRandomeSelector(
			models.FriendshipStatusAccepted,
			models.FriendshipStatusDeclined,
		)
		// declare slices
		accepted := []*models.Frindship{}
		declined := []*models.Frindship{}

		// for each friendship
		for _, f := range friendships {
			// randome select status
			status := providerSelector.Select()

			// update status
			f.Status = status
			newf, err := gamingStore.UpdateFriendship(ctx, f)
			if err != nil {
				t.Fatalf("UpdateFriendshipStatus() error = %v", err)
			}
			if newf.Status != status {
				t.Errorf("UpdateFriendshipStatus() = %v, want status %v", f.Status, status)
			}
			// add to slice
			switch status {
			case models.FriendshipStatusAccepted:
				accepted = append(accepted, newf)
			case models.FriendshipStatusDeclined:
				declined = append(declined, newf)
			}
		}
		// The sum of the lengths of accepted and declined should be equal to the length of friendships.
		if len(accepted)+len(declined) != len(friendships) {
			t.Errorf("The sum of the lengths of accepted and declined should be equal to the length of friendships. got %v, want %v", len(accepted)+len(declined), len(friendships))
		}
		// find -------------------------------------------------------------------
		acceptedMap := map[uuid.UUID]*models.Frindship{}
		declinedMap := map[uuid.UUID]*models.Frindship{}
		for _, player := range players {
			// find friendships accepted
			playerFriendshipsAccepted, err := gamingStore.FindFriendships(ctx, &FriendshipFilter{
				PaginatedInput: repository.PaginatedInput{
					Page:    0,
					PerPage: 100,
				},
				RequestingOrInvitedPlayerIds: []uuid.UUID{player.ID},
				Statuses:                     []models.FriendshipStatus{models.FriendshipStatusAccepted},
			})
			if err != nil {
				t.Fatalf("FindFriendship() error = %v", err)
			}
			for _, v := range playerFriendshipsAccepted {
				acceptedMap[v.ID] = v
			}
			// find friendships declined
			playerFriendshipsDeclined, err := gamingStore.FindFriendships(ctx, &FriendshipFilter{
				PaginatedInput: repository.PaginatedInput{
					Page:    0,
					PerPage: 100,
				},
				RequestingOrInvitedPlayerIds: []uuid.UUID{player.ID},
				Statuses:                     []models.FriendshipStatus{models.FriendshipStatusDeclined},
			})
			if err != nil {
				t.Fatalf("FindFriendship() error = %v", err)
			}
			for _, v := range playerFriendshipsDeclined {
				declinedMap[v.ID] = v
			}
		}
		if len(acceptedMap)+len(declinedMap) != len(friendships) {
			t.Errorf("The sum of the lengths of accepted and declined should be equal to the length of friendships. got %v, want %v", len(acceptedMap)+len(declinedMap), len(friendships))
		}
		if len(acceptedMap) != len(accepted) {
			t.Errorf("The sum of the lengths of accepted map should be equal to the length of accepted. got %v, want %v", len(acceptedMap), len(accepted))
		}
		if len(declinedMap) != len(declined) {
			t.Errorf("The sum of the lengths of accepted and declined should be equal to the length of friendships. got %v, want %v", len(declinedMap), len(declined))
		}
		// count -------------------------------------------------------------------
		totalFriendShipCount, err := gamingStore.CountFriendships(ctx, nil)
		if err != nil {
			t.Fatalf("CountFriendships() error = %v", err)
		}
		if totalFriendShipCount != int64(len(friendships)) {
			t.Errorf("CountFriendships() = %v, want %v", totalFriendShipCount, len(friendships))
		}
		totalAcceptedFriendShipCount, err := gamingStore.CountFriendships(ctx, &FriendshipFilter{
			PaginatedInput: repository.PaginatedInput{
				Page:    0,
				PerPage: 100,
			},
			Statuses: []models.FriendshipStatus{models.FriendshipStatusAccepted},
		})
		if err != nil {
			t.Fatalf("CountFriendships() error = %v", err)
		}
		if totalAcceptedFriendShipCount != int64(len(accepted)) {
			t.Errorf("CountFriendships() = %v, want %v", totalAcceptedFriendShipCount, len(accepted))
		}
		totalDeclinedFriendShipCount, err := gamingStore.CountFriendships(ctx, &FriendshipFilter{
			PaginatedInput: repository.PaginatedInput{
				Page:    0,
				PerPage: 100,
			},
			Statuses: []models.FriendshipStatus{models.FriendshipStatusDeclined},
		})
		if err != nil {
			t.Fatalf("CountFriendships() error = %v", err)
		}
		if totalDeclinedFriendShipCount != int64(len(declined)) {
			t.Errorf("CountFriendships() = %v, want %v", totalDeclinedFriendShipCount, len(declined))
		}
		if totalAcceptedFriendShipCount+totalDeclinedFriendShipCount != totalFriendShipCount {
			t.Errorf("CountFriendships() = %v, want %v", totalAcceptedFriendShipCount+totalDeclinedFriendShipCount, totalFriendShipCount)
		}
		// count again
		totalFriendShipCount, err = gamingStore.CountFriendships(ctx, nil)
		if err != nil {
			t.Fatalf("CountFriendships() error = %v", err)
		}
		if totalFriendShipCount != 45 {
			t.Errorf("CountFriendships() = %v, want %v", totalFriendShipCount, 0)
		}
		// delete
		toDeleteIds := []uuid.UUID{}
		for _, f := range accepted {
			toDeleteIds = append(toDeleteIds, f.ID)
		}
		for _, f := range declined {
			toDeleteIds = append(toDeleteIds, f.ID)
		}
		deleted, err := gamingStore.DeleteFriendships(ctx, &FriendshipFilter{
			Ids: toDeleteIds,
		})
		if err != nil {
			t.Fatalf("DeleteFriendships() error = %v", err)
		}
		if deleted != int64(len(toDeleteIds)) {
			t.Errorf("DeleteFriendships() = %v, want %v", deleted, len(toDeleteIds))
		}
	})
}
