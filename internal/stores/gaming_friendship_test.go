//go:build integration

package stores

import (
	"context"
	"fmt"
	"testing"
	"time"

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
		friendships := []*models.Friendship{}
		friendshipsMap := map[uuid.UUID]*models.Friendship{}
		playerIdFriendshipMap := map[uuid.UUID][]*models.Friendship{}

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
				friendship, err := gamingStore.CreateFriendship(ctx, &models.Friendship{
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
		accepted := []*models.Friendship{}
		declined := []*models.Friendship{}

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
		acceptedMap := map[uuid.UUID]*models.Friendship{}
		declinedMap := map[uuid.UUID]*models.Friendship{}
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
		// verify blocked status can be persisted (after clearing all records)
		blockedFriendship, err := gamingStore.CreateFriendship(ctx, &models.Friendship{
			RequestingPlayerID: players[0].ID,
			InvitedPlayerID:    players[1].ID,
			Status:             models.FriendshipStatusBlocked,
		})
		if err != nil {
			t.Fatalf("CreateFriendship(blocked) error = %v", err)
		}
		if blockedFriendship.Status != models.FriendshipStatusBlocked {
			t.Errorf("CreateFriendship(blocked) status = %v, want %v", blockedFriendship.Status, models.FriendshipStatusBlocked)
		}
		blockedCount, err := gamingStore.CountFriendships(ctx, &FriendshipFilter{
			Statuses: []models.FriendshipStatus{models.FriendshipStatusBlocked},
		})
		if err != nil {
			t.Fatalf("CountFriendships(blocked) error = %v", err)
		}
		if blockedCount != 1 {
			t.Errorf("CountFriendships(blocked) = %v, want 1", blockedCount)
		}
	})
}

func TestFriendshipFilter_PlayerPair(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		gs := NewDBGamingStore(db)
		a := MustCreatePlayer(t, ctx, gs, WithPlayerEmail("pair_a@example.com"))
		b := MustCreatePlayer(t, ctx, gs, WithPlayerEmail("pair_b@example.com"))
		c := MustCreatePlayer(t, ctx, gs, WithPlayerEmail("pair_c@example.com"))

		// a→b friendship
		MustCreateFriendship(t, ctx, gs, a, b, WithStatus(models.FriendshipStatusAccepted))
		// b→c unrelated friendship
		MustCreateFriendship(t, ctx, gs, b, c, WithStatus(models.FriendshipStatusPending))

		pair := [2]uuid.UUID{a.ID, b.ID}

		// PlayerPair finds a↔b regardless of direction
		f, err := gs.FindFriendship(ctx, &FriendshipFilter{PlayerPair: &pair})
		if err != nil {
			t.Fatalf("FindFriendship(PlayerPair a→b) error = %v", err)
		}
		if f == nil {
			t.Fatal("FindFriendship(PlayerPair a→b) = nil, want record")
		}

		// Reverse pair also finds it
		pairRev := [2]uuid.UUID{b.ID, a.ID}
		fRev, err := gs.FindFriendship(ctx, &FriendshipFilter{PlayerPair: &pairRev})
		if err != nil {
			t.Fatalf("FindFriendship(PlayerPair b→a) error = %v", err)
		}
		if fRev == nil || fRev.ID != f.ID {
			t.Error("PlayerPair reverse should return the same record")
		}

		// a↔c returns nothing
		pairAC := [2]uuid.UUID{a.ID, c.ID}
		fNone, err := gs.FindFriendship(ctx, &FriendshipFilter{PlayerPair: &pairAC})
		if err != nil {
			t.Fatalf("FindFriendship(PlayerPair a→c) error = %v", err)
		}
		if fNone != nil {
			t.Errorf("FindFriendship(PlayerPair a→c) = %v, want nil", fNone)
		}

		// PlayerPair + status filter only matches when status aligns
		fBlocked, err := gs.FindFriendship(ctx, &FriendshipFilter{
			PlayerPair: &pair,
			Statuses:   []models.FriendshipStatus{models.FriendshipStatusBlocked},
		})
		if err != nil {
			t.Fatalf("FindFriendship(PlayerPair+blocked) error = %v", err)
		}
		if fBlocked != nil {
			t.Error("FindFriendship(PlayerPair+blocked) should return nil for accepted friendship")
		}
	})
}

func TestFriendshipFilter_CreatedAfter(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		gs := NewDBGamingStore(db)
		a := MustCreatePlayer(t, ctx, gs, WithPlayerEmail("ca_a@example.com"))
		b := MustCreatePlayer(t, ctx, gs, WithPlayerEmail("ca_b@example.com"))

		MustCreateFriendship(t, ctx, gs, a, b, WithStatus(models.FriendshipStatusPending))

		// CreatedAfter=now excludes the just-created record
		future := time.Now().UTC().Add(time.Minute)
		count, err := gs.CountFriendships(ctx, &FriendshipFilter{CreatedAfter: &future})
		if err != nil {
			t.Fatalf("CountFriendships(CreatedAfter=future) error = %v", err)
		}
		if count != 0 {
			t.Errorf("CountFriendships(CreatedAfter=future) = %d, want 0", count)
		}

		// CreatedAfter=one hour ago includes the record
		past := time.Now().UTC().Add(-time.Hour)
		count, err = gs.CountFriendships(ctx, &FriendshipFilter{CreatedAfter: &past})
		if err != nil {
			t.Fatalf("CountFriendships(CreatedAfter=past) error = %v", err)
		}
		if count != 1 {
			t.Errorf("CountFriendships(CreatedAfter=past) = %d, want 1", count)
		}
	})
}
