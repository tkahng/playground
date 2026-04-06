# Betting Comprehensive Integration Tests Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add 8 integration tests across two new files that verify no pending transfers survive any terminal game state and that total system funds are conserved across multiple games.

**Architecture:** Two new test files in `internal/services/` (same package as all other service tests). `betting_lifecycle_test.go` covers per-game explicit pending-transfer count assertions. `betting_invariants_test.go` covers multi-game aggregate conservation. No production code changes.

**Tech Stack:** Go test stdlib, `database.WithNewTestTx`, existing service helpers (`mustFundPlayerWallet`, `mustCreateUser`, `stores.MustCreatePlayer`), `ledger.CountTransfers` with `stores.LedgerTransferFilter`.

---

## File Map

| File | Action | Responsibility |
|------|--------|----------------|
| `internal/services/betting_lifecycle_test.go` | Create | 5 tests: pending-count = 0 after cancel / expiry / host-wins / tie / guest-retry |
| `internal/services/betting_invariants_test.go` | Create | 3 tests: multi-game total conservation, per-game escrow zero, no orphan pending |

**Helpers already available in package (do NOT redefine):**
- `mustFundWallet(t, ctx, adapter, ledger, userID uuid.UUID, amount int64)` — `betting_service_test.go`
- `mustFundPlayerWallet(t, ctx, adapter, ledger, userID *uuid.UUID, amount int64)` — `gaming_rps_game_service_test.go`
- `mustCreateUser(t, ctx, adapter, email string) *models.User` — `gaming_rps_game_service_test.go`
- `stores.MustCreatePlayer(t, ctx, adapter.Gaming(), stores.WithPlayerEmail(...), stores.WithUserID(...))` — stores package

**Standard test boilerplate:**
```go
database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
    adapter := stores.NewDbAdapterDecorators(db)
    ledger  := NewDbLedgerService(adapter)
    betting := NewDbBettingService(adapter, ledger)
    rps     := NewDbRpsGameService(adapter, betting)
    ...
})
```

**Standard pending-count assertion (used in every lifecycle test):**
```go
pendingCount, err := ledger.CountTransfers(ctx, &stores.LedgerTransferFilter{
    ReferenceIds: []uuid.UUID{game.RpsGame.ID},
    Statuses:     []models.LedgerTransferStatus{models.LedgerTransferStatusPending},
})
if err != nil {
    t.Fatalf("CountTransfers: %v", err)
}
if pendingCount != 0 {
    t.Errorf("pending transfers = %d, want 0", pendingCount)
}
```

---

## Task 1: Create `betting_lifecycle_test.go` — cancel and expiry tests

**Files:**
- Create: `internal/services/betting_lifecycle_test.go`

- [ ] **Step 1: Create the file with the first two tests**

```go
package services

import (
	"context"
	"testing"
	"time"

	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
)

func TestBetting_NoPendingTransfers_AfterCancel(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)
		rpsService := NewDbRpsGameService(adapter, betting)

		host := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
			stores.WithPlayerEmail("cancel_host@example.com"),
			stores.WithUserID(mustCreateUser(t, ctx, adapter, "cancel_host@example.com").ID),
		)
		guest := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
			stores.WithPlayerEmail("cancel_guest@example.com"),
			stores.WithUserID(mustCreateUser(t, ctx, adapter, "cancel_guest@example.com").ID),
		)
		mustFundPlayerWallet(t, ctx, adapter, ledger, host.UserID, 500)

		betAmount := int64(100)
		game, err := rpsService.RequestGame(ctx, &RpsGameRequestInput{
			RequestingPlayerID:   host.ID,
			InvitedPlayerID:      guest.ID,
			RequestingPlayerMove: models.RpsParticipantMoveRock,
			DurationSeconds:      60 * 60 * 24,
			BetAmount:            &betAmount,
			HostUserID:           host.UserID,
		})
		if err != nil {
			t.Fatalf("RequestGame() error = %v", err)
		}

		_, err = rpsService.RespondToGameRequest(ctx, &GameRequestResponse{
			InvitedPlayerID: guest.ID,
			GameID:          game.RpsGame.ID,
			Status:          models.RpsGameStatusCancelled,
		})
		if err != nil {
			t.Fatalf("RespondToGameRequest(cancel) error = %v", err)
		}

		pendingCount, err := ledger.CountTransfers(ctx, &stores.LedgerTransferFilter{
			ReferenceIds: []uuid.UUID{game.RpsGame.ID},
			Statuses:     []models.LedgerTransferStatus{models.LedgerTransferStatusPending},
		})
		if err != nil {
			t.Fatalf("CountTransfers: %v", err)
		}
		if pendingCount != 0 {
			t.Errorf("pending transfers after cancel = %d, want 0", pendingCount)
		}

		avail, err := ledger.GetUserAvailableBalance(ctx, *host.UserID)
		if err != nil {
			t.Fatalf("GetUserAvailableBalance: %v", err)
		}
		if avail != 500 {
			t.Errorf("host available balance = %d, want 500", avail)
		}
	})
}

func TestBetting_NoPendingTransfers_AfterExpiry(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)
		rpsService := NewDbRpsGameService(adapter, betting)

		host := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
			stores.WithPlayerEmail("expiry_host@example.com"),
			stores.WithUserID(mustCreateUser(t, ctx, adapter, "expiry_host@example.com").ID),
		)
		guest := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
			stores.WithPlayerEmail("expiry_guest@example.com"),
			stores.WithUserID(mustCreateUser(t, ctx, adapter, "expiry_guest@example.com").ID),
		)
		mustFundPlayerWallet(t, ctx, adapter, ledger, host.UserID, 500)

		betAmount := int64(100)
		game, err := rpsService.RequestGame(ctx, &RpsGameRequestInput{
			RequestingPlayerID:   host.ID,
			InvitedPlayerID:      guest.ID,
			RequestingPlayerMove: models.RpsParticipantMoveRock,
			DurationSeconds:      1,
			BetAmount:            &betAmount,
			HostUserID:           host.UserID,
		})
		if err != nil {
			t.Fatalf("RequestGame() error = %v", err)
		}

		time.Sleep(2 * time.Second)

		processed, err := rpsService.ExpireGamesAndRefundBets(ctx)
		if err != nil {
			t.Fatalf("ExpireGamesAndRefundBets() error = %v", err)
		}
		if processed != 1 {
			t.Errorf("processed = %d, want 1", processed)
		}

		pendingCount, err := ledger.CountTransfers(ctx, &stores.LedgerTransferFilter{
			ReferenceIds: []uuid.UUID{game.RpsGame.ID},
			Statuses:     []models.LedgerTransferStatus{models.LedgerTransferStatusPending},
		})
		if err != nil {
			t.Fatalf("CountTransfers: %v", err)
		}
		if pendingCount != 0 {
			t.Errorf("pending transfers after expiry = %d, want 0", pendingCount)
		}

		avail, err := ledger.GetUserAvailableBalance(ctx, *host.UserID)
		if err != nil {
			t.Fatalf("GetUserAvailableBalance: %v", err)
		}
		if avail != 500 {
			t.Errorf("host available balance after sweep = %d, want 500", avail)
		}

		updated, err := adapter.Gaming().FindRpsGame(ctx, &stores.RpsGameFilter{Ids: []uuid.UUID{game.RpsGame.ID}})
		if err != nil {
			t.Fatalf("FindRpsGame: %v", err)
		}
		if updated.Status != models.RpsGameStatusCancelled {
			t.Errorf("game status = %v, want cancelled", updated.Status)
		}
	})
}
```

Note: the `uuid` import is missing — you will add it in Step 2.

- [ ] **Step 2: Fix the import block** — `uuid` is needed for `[]uuid.UUID{...}` in the filter. Update the import:

```go
import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
)
```

- [ ] **Step 3: Run these two tests to verify they pass**

```bash
cd /Users/tkahng/github/tkahng/go/playground
go test ./internal/services/ -run "TestBetting_NoPendingTransfers_AfterCancel|TestBetting_NoPendingTransfers_AfterExpiry" -v -count=1
```

Expected output: both tests `PASS`. The expiry test will take ~2 seconds due to `time.Sleep`.

- [ ] **Step 4: Commit**

```bash
git add internal/services/betting_lifecycle_test.go
git commit -m "test: pending-transfer count asserts for cancel and expiry paths"
```

---

## Task 2: Add complete and guest-retry lifecycle tests

**Files:**
- Modify: `internal/services/betting_lifecycle_test.go`

- [ ] **Step 1: Append the three remaining lifecycle tests to the file**

```go
func TestBetting_NoPendingTransfers_AfterComplete_HostWins(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)
		rpsService := NewDbRpsGameService(adapter, betting)

		host := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
			stores.WithPlayerEmail("hwins_host@example.com"),
			stores.WithUserID(mustCreateUser(t, ctx, adapter, "hwins_host@example.com").ID),
		)
		guest := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
			stores.WithPlayerEmail("hwins_guest@example.com"),
			stores.WithUserID(mustCreateUser(t, ctx, adapter, "hwins_guest@example.com").ID),
		)
		mustFundPlayerWallet(t, ctx, adapter, ledger, host.UserID, 500)
		mustFundPlayerWallet(t, ctx, adapter, ledger, guest.UserID, 500)

		betAmount := int64(100)
		// Host plays Rock; guest plays Scissors → host wins.
		game, err := rpsService.RequestGame(ctx, &RpsGameRequestInput{
			RequestingPlayerID:   host.ID,
			InvitedPlayerID:      guest.ID,
			RequestingPlayerMove: models.RpsParticipantMoveRock,
			DurationSeconds:      60 * 60 * 24,
			BetAmount:            &betAmount,
			HostUserID:           host.UserID,
		})
		if err != nil {
			t.Fatalf("RequestGame() error = %v", err)
		}

		_, err = rpsService.RespondToGameRequest(ctx, &GameRequestResponse{
			InvitedPlayerID: guest.ID,
			GameID:          game.RpsGame.ID,
			Status:          models.RpsGameStatusCompleted,
			Move:            models.RpsParticipantMoveScissors,
		})
		if err != nil {
			t.Fatalf("RespondToGameRequest() error = %v", err)
		}

		pendingCount, err := ledger.CountTransfers(ctx, &stores.LedgerTransferFilter{
			ReferenceIds: []uuid.UUID{game.RpsGame.ID},
			Statuses:     []models.LedgerTransferStatus{models.LedgerTransferStatusPending},
		})
		if err != nil {
			t.Fatalf("CountTransfers: %v", err)
		}
		if pendingCount != 0 {
			t.Errorf("pending transfers after host win = %d, want 0", pendingCount)
		}

		hostBal, err := ledger.GetUserBalance(ctx, *host.UserID)
		if err != nil {
			t.Fatalf("GetUserBalance host: %v", err)
		}
		guestBal, err := ledger.GetUserBalance(ctx, *guest.UserID)
		if err != nil {
			t.Fatalf("GetUserBalance guest: %v", err)
		}
		if hostBal != 600 {
			t.Errorf("host balance = %d, want 600", hostBal)
		}
		if guestBal != 400 {
			t.Errorf("guest balance = %d, want 400", guestBal)
		}
	})
}

func TestBetting_NoPendingTransfers_AfterComplete_Tie(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)
		rpsService := NewDbRpsGameService(adapter, betting)

		host := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
			stores.WithPlayerEmail("tie_host@example.com"),
			stores.WithUserID(mustCreateUser(t, ctx, adapter, "tie_host@example.com").ID),
		)
		guest := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
			stores.WithPlayerEmail("tie_guest@example.com"),
			stores.WithUserID(mustCreateUser(t, ctx, adapter, "tie_guest@example.com").ID),
		)
		mustFundPlayerWallet(t, ctx, adapter, ledger, host.UserID, 500)
		mustFundPlayerWallet(t, ctx, adapter, ledger, guest.UserID, 500)

		betAmount := int64(100)
		// Rock vs Rock → tie.
		game, err := rpsService.RequestGame(ctx, &RpsGameRequestInput{
			RequestingPlayerID:   host.ID,
			InvitedPlayerID:      guest.ID,
			RequestingPlayerMove: models.RpsParticipantMoveRock,
			DurationSeconds:      60 * 60 * 24,
			BetAmount:            &betAmount,
			HostUserID:           host.UserID,
		})
		if err != nil {
			t.Fatalf("RequestGame() error = %v", err)
		}

		_, err = rpsService.RespondToGameRequest(ctx, &GameRequestResponse{
			InvitedPlayerID: guest.ID,
			GameID:          game.RpsGame.ID,
			Status:          models.RpsGameStatusCompleted,
			Move:            models.RpsParticipantMoveRock,
		})
		if err != nil {
			t.Fatalf("RespondToGameRequest() error = %v", err)
		}

		pendingCount, err := ledger.CountTransfers(ctx, &stores.LedgerTransferFilter{
			ReferenceIds: []uuid.UUID{game.RpsGame.ID},
			Statuses:     []models.LedgerTransferStatus{models.LedgerTransferStatusPending},
		})
		if err != nil {
			t.Fatalf("CountTransfers: %v", err)
		}
		if pendingCount != 0 {
			t.Errorf("pending transfers after tie = %d, want 0", pendingCount)
		}

		hostBal, err := ledger.GetUserBalance(ctx, *host.UserID)
		if err != nil {
			t.Fatalf("GetUserBalance host: %v", err)
		}
		guestBal, err := ledger.GetUserBalance(ctx, *guest.UserID)
		if err != nil {
			t.Fatalf("GetUserBalance guest: %v", err)
		}
		if hostBal != 500 {
			t.Errorf("host balance after tie = %d, want 500", hostBal)
		}
		if guestBal != 500 {
			t.Errorf("guest balance after tie = %d, want 500", guestBal)
		}
	})
}

func TestBetting_GuestCanRetry_AfterInsufficientFunds(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)
		rpsService := NewDbRpsGameService(adapter, betting)

		host := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
			stores.WithPlayerEmail("retry_host@example.com"),
			stores.WithUserID(mustCreateUser(t, ctx, adapter, "retry_host@example.com").ID),
		)
		guest := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
			stores.WithPlayerEmail("retry_guest@example.com"),
			stores.WithUserID(mustCreateUser(t, ctx, adapter, "retry_guest@example.com").ID),
		)
		mustFundPlayerWallet(t, ctx, adapter, ledger, host.UserID, 500)
		// Guest starts with 0 balance — no funding here.

		betAmount := int64(100)
		game, err := rpsService.RequestGame(ctx, &RpsGameRequestInput{
			RequestingPlayerID:   host.ID,
			InvitedPlayerID:      guest.ID,
			RequestingPlayerMove: models.RpsParticipantMoveRock,
			DurationSeconds:      60 * 60 * 24,
			BetAmount:            &betAmount,
			HostUserID:           host.UserID,
		})
		if err != nil {
			t.Fatalf("RequestGame() error = %v", err)
		}

		// First attempt: must fail — guest has no funds.
		_, err = rpsService.RespondToGameRequest(ctx, &GameRequestResponse{
			InvitedPlayerID: guest.ID,
			GameID:          game.RpsGame.ID,
			Status:          models.RpsGameStatusCompleted,
			Move:            models.RpsParticipantMoveScissors,
		})
		if err == nil {
			t.Fatal("expected error when guest has no funds, got nil")
		}

		// Game must still be pending after failed attempt.
		currentGame, err := adapter.Gaming().FindRpsGame(ctx, &stores.RpsGameFilter{Ids: []uuid.UUID{game.RpsGame.ID}})
		if err != nil {
			t.Fatalf("FindRpsGame: %v", err)
		}
		if currentGame.Status != models.RpsGameStatusPending {
			t.Errorf("game status after failed response = %v, want pending", currentGame.Status)
		}

		// Host escrow is still held.
		hostAvail, err := ledger.GetUserAvailableBalance(ctx, *host.UserID)
		if err != nil {
			t.Fatalf("GetUserAvailableBalance: %v", err)
		}
		if hostAvail != 400 {
			t.Errorf("host available after failed attempt = %d, want 400 (escrow still held)", hostAvail)
		}

		// Fund the guest, then retry — must succeed.
		mustFundPlayerWallet(t, ctx, adapter, ledger, guest.UserID, 500)

		// Host plays Rock, guest plays Scissors → host wins.
		_, err = rpsService.RespondToGameRequest(ctx, &GameRequestResponse{
			InvitedPlayerID: guest.ID,
			GameID:          game.RpsGame.ID,
			Status:          models.RpsGameStatusCompleted,
			Move:            models.RpsParticipantMoveScissors,
		})
		if err != nil {
			t.Fatalf("RespondToGameRequest after funding guest: %v", err)
		}

		pendingCount, err := ledger.CountTransfers(ctx, &stores.LedgerTransferFilter{
			ReferenceIds: []uuid.UUID{game.RpsGame.ID},
			Statuses:     []models.LedgerTransferStatus{models.LedgerTransferStatusPending},
		})
		if err != nil {
			t.Fatalf("CountTransfers: %v", err)
		}
		if pendingCount != 0 {
			t.Errorf("pending transfers after retry success = %d, want 0", pendingCount)
		}

		// Host wins: 500 - 100 (escrow posted) + 200 (pot) = 600
		hostBal, err := ledger.GetUserBalance(ctx, *host.UserID)
		if err != nil {
			t.Fatalf("GetUserBalance host: %v", err)
		}
		guestBal, err := ledger.GetUserBalance(ctx, *guest.UserID)
		if err != nil {
			t.Fatalf("GetUserBalance guest: %v", err)
		}
		if hostBal != 600 {
			t.Errorf("host balance = %d, want 600", hostBal)
		}
		if guestBal != 400 {
			t.Errorf("guest balance = %d, want 400 (500 funded - 100 bet)", guestBal)
		}
	})
}
```

- [ ] **Step 2: Run all five lifecycle tests**

```bash
cd /Users/tkahng/github/tkahng/go/playground
go test ./internal/services/ -run "TestBetting_NoPendingTransfers|TestBetting_GuestCanRetry" -v -count=1
```

Expected: all 5 `PASS`.

- [ ] **Step 3: Commit**

```bash
git add internal/services/betting_lifecycle_test.go
git commit -m "test: pending-transfer count asserts for complete and guest-retry paths"
```

---

## Task 3: Create `betting_invariants_test.go` — multi-game conservation test

**Files:**
- Create: `internal/services/betting_invariants_test.go`

- [ ] **Step 1: Create the file with `TestBettingInvariant_MultiGame_TotalSystemConservation`**

```go
package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
)

// TestBettingInvariant_MultiGame_TotalSystemConservation runs six games covering
// every terminal path (host-win × 2, guest-win × 2, tie × 1, expire × 1) and
// asserts that the total of all user balances plus the escrow account equals the
// total points ever issued. No money is created or destroyed.
func TestBettingInvariant_MultiGame_TotalSystemConservation(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)
		rpsService := NewDbRpsGameService(adapter, betting)

		const startBalance int64 = 1000
		const numPairs = 6
		const totalIssued int64 = startBalance * numPairs * 2 // 12 000

		type playerPair struct {
			host  *models.Player
			guest *models.Player
		}

		makePair := func(i int) playerPair {
			hEmail := fmt.Sprintf("inv_host%d@example.com", i)
			gEmail := fmt.Sprintf("inv_guest%d@example.com", i)
			hUser := mustCreateUser(t, ctx, adapter, hEmail)
			gUser := mustCreateUser(t, ctx, adapter, gEmail)
			h := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
				stores.WithPlayerEmail(hEmail),
				stores.WithUserID(hUser.ID),
			)
			g := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
				stores.WithPlayerEmail(gEmail),
				stores.WithUserID(gUser.ID),
			)
			mustFundPlayerWallet(t, ctx, adapter, ledger, h.UserID, startBalance)
			mustFundPlayerWallet(t, ctx, adapter, ledger, g.UserID, startBalance)
			return playerPair{h, g}
		}

		pairs := make([]playerPair, numPairs)
		for i := range pairs {
			pairs[i] = makePair(i)
		}

		escrow, err := ledger.GetSystemAccount(ctx, models.SystemAccountGameEscrow)
		if err != nil {
			t.Fatalf("GetSystemAccount escrow: %v", err)
		}
		escrowStart := escrow.Balance()

		betPtr := func(v int64) *int64 { return &v }

		// Game 0: host wins (100 pt bet) — Rock beats Scissors
		{
			p := pairs[0]
			g, err := rpsService.RequestGame(ctx, &RpsGameRequestInput{
				RequestingPlayerID:   p.host.ID,
				InvitedPlayerID:      p.guest.ID,
				RequestingPlayerMove: models.RpsParticipantMoveRock,
				DurationSeconds:      3600,
				BetAmount:            betPtr(100),
				HostUserID:           p.host.UserID,
			})
			if err != nil {
				t.Fatalf("game0 RequestGame: %v", err)
			}
			if _, err = rpsService.RespondToGameRequest(ctx, &GameRequestResponse{
				InvitedPlayerID: p.guest.ID,
				GameID:          g.RpsGame.ID,
				Status:          models.RpsGameStatusCompleted,
				Move:            models.RpsParticipantMoveScissors,
			}); err != nil {
				t.Fatalf("game0 Respond: %v", err)
			}
		}

		// Game 1: guest wins (100 pt bet) — Paper beats Rock
		{
			p := pairs[1]
			g, err := rpsService.RequestGame(ctx, &RpsGameRequestInput{
				RequestingPlayerID:   p.host.ID,
				InvitedPlayerID:      p.guest.ID,
				RequestingPlayerMove: models.RpsParticipantMoveRock,
				DurationSeconds:      3600,
				BetAmount:            betPtr(100),
				HostUserID:           p.host.UserID,
			})
			if err != nil {
				t.Fatalf("game1 RequestGame: %v", err)
			}
			if _, err = rpsService.RespondToGameRequest(ctx, &GameRequestResponse{
				InvitedPlayerID: p.guest.ID,
				GameID:          g.RpsGame.ID,
				Status:          models.RpsGameStatusCompleted,
				Move:            models.RpsParticipantMovePaper,
			}); err != nil {
				t.Fatalf("game1 Respond: %v", err)
			}
		}

		// Game 2: host wins (200 pt bet) — Rock beats Scissors
		{
			p := pairs[2]
			g, err := rpsService.RequestGame(ctx, &RpsGameRequestInput{
				RequestingPlayerID:   p.host.ID,
				InvitedPlayerID:      p.guest.ID,
				RequestingPlayerMove: models.RpsParticipantMoveRock,
				DurationSeconds:      3600,
				BetAmount:            betPtr(200),
				HostUserID:           p.host.UserID,
			})
			if err != nil {
				t.Fatalf("game2 RequestGame: %v", err)
			}
			if _, err = rpsService.RespondToGameRequest(ctx, &GameRequestResponse{
				InvitedPlayerID: p.guest.ID,
				GameID:          g.RpsGame.ID,
				Status:          models.RpsGameStatusCompleted,
				Move:            models.RpsParticipantMoveScissors,
			}); err != nil {
				t.Fatalf("game2 Respond: %v", err)
			}
		}

		// Game 3: guest wins (200 pt bet) — Paper beats Rock
		{
			p := pairs[3]
			g, err := rpsService.RequestGame(ctx, &RpsGameRequestInput{
				RequestingPlayerID:   p.host.ID,
				InvitedPlayerID:      p.guest.ID,
				RequestingPlayerMove: models.RpsParticipantMoveRock,
				DurationSeconds:      3600,
				BetAmount:            betPtr(200),
				HostUserID:           p.host.UserID,
			})
			if err != nil {
				t.Fatalf("game3 RequestGame: %v", err)
			}
			if _, err = rpsService.RespondToGameRequest(ctx, &GameRequestResponse{
				InvitedPlayerID: p.guest.ID,
				GameID:          g.RpsGame.ID,
				Status:          models.RpsGameStatusCompleted,
				Move:            models.RpsParticipantMovePaper,
			}); err != nil {
				t.Fatalf("game3 Respond: %v", err)
			}
		}

		// Game 4: tie (150 pt bet) — Rock vs Rock
		{
			p := pairs[4]
			g, err := rpsService.RequestGame(ctx, &RpsGameRequestInput{
				RequestingPlayerID:   p.host.ID,
				InvitedPlayerID:      p.guest.ID,
				RequestingPlayerMove: models.RpsParticipantMoveRock,
				DurationSeconds:      3600,
				BetAmount:            betPtr(150),
				HostUserID:           p.host.UserID,
			})
			if err != nil {
				t.Fatalf("game4 RequestGame: %v", err)
			}
			if _, err = rpsService.RespondToGameRequest(ctx, &GameRequestResponse{
				InvitedPlayerID: p.guest.ID,
				GameID:          g.RpsGame.ID,
				Status:          models.RpsGameStatusCompleted,
				Move:            models.RpsParticipantMoveRock,
			}); err != nil {
				t.Fatalf("game4 Respond: %v", err)
			}
		}

		// Game 5: expires (100 pt bet)
		{
			p := pairs[5]
			_, err := rpsService.RequestGame(ctx, &RpsGameRequestInput{
				RequestingPlayerID:   p.host.ID,
				InvitedPlayerID:      p.guest.ID,
				RequestingPlayerMove: models.RpsParticipantMoveRock,
				DurationSeconds:      1,
				BetAmount:            betPtr(100),
				HostUserID:           p.host.UserID,
			})
			if err != nil {
				t.Fatalf("game5 RequestGame: %v", err)
			}
		}
		time.Sleep(2 * time.Second)
		if _, err := rpsService.ExpireGamesAndRefundBets(ctx); err != nil {
			t.Fatalf("ExpireGamesAndRefundBets: %v", err)
		}

		// --- Invariant assertions ---

		// 1. Total system conservation.
		var totalUserBalance int64
		for i, p := range pairs {
			hBal, err := ledger.GetUserBalance(ctx, *p.host.UserID)
			if err != nil {
				t.Fatalf("GetUserBalance host[%d]: %v", i, err)
			}
			gBal, err := ledger.GetUserBalance(ctx, *p.guest.UserID)
			if err != nil {
				t.Fatalf("GetUserBalance guest[%d]: %v", i, err)
			}
			totalUserBalance += hBal + gBal
		}

		escrow, err = ledger.GetSystemAccount(ctx, models.SystemAccountGameEscrow)
		if err != nil {
			t.Fatalf("re-fetch escrow: %v", err)
		}
		totalAfter := totalUserBalance + escrow.Balance()
		if totalAfter != totalIssued {
			t.Errorf("total system balance = %d, want %d (money conservation violated)", totalAfter, totalIssued)
		}

		// 2. Escrow nets to zero.
		if escrow.Balance() != escrowStart {
			t.Errorf("escrow balance = %d, want %d (must return to start)", escrow.Balance(), escrowStart)
		}

		// 3. No pending transfers remain anywhere in the ledger.
		pendingCount, err := ledger.CountTransfers(ctx, &stores.LedgerTransferFilter{
			Statuses: []models.LedgerTransferStatus{models.LedgerTransferStatusPending},
		})
		if err != nil {
			t.Fatalf("CountTransfers pending: %v", err)
		}
		if pendingCount != 0 {
			t.Errorf("pending transfers remaining = %d, want 0", pendingCount)
		}
	})
}
```

- [ ] **Step 2: Run this test**

```bash
cd /Users/tkahng/github/tkahng/go/playground
go test ./internal/services/ -run "TestBettingInvariant_MultiGame_TotalSystemConservation" -v -count=1 -timeout 60s
```

Expected: `PASS`. The test runs ~2 seconds due to the expiry sleep.

- [ ] **Step 3: Commit**

```bash
git add internal/services/betting_invariants_test.go
git commit -m "test: multi-game total system conservation invariant"
```

---

## Task 4: Add escrow-per-game and no-orphan invariant tests

**Files:**
- Modify: `internal/services/betting_invariants_test.go`

- [ ] **Step 1: Append the two remaining invariant tests to the file**

```go
// TestBettingInvariant_EscrowNetsZeroAfterEachGame verifies that the escrow
// account returns to its starting balance immediately after each individual game
// settles, not just in aggregate. This catches a leak in one game being masked
// by another game's accounting.
func TestBettingInvariant_EscrowNetsZeroAfterEachGame(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)
		rpsService := NewDbRpsGameService(adapter, betting)

		escrow, err := ledger.GetSystemAccount(ctx, models.SystemAccountGameEscrow)
		if err != nil {
			t.Fatalf("GetSystemAccount: %v", err)
		}
		escrowStart := escrow.Balance()

		assertEscrowAtStart := func(label string) {
			t.Helper()
			e, err := ledger.GetSystemAccount(ctx, models.SystemAccountGameEscrow)
			if err != nil {
				t.Fatalf("%s: re-fetch escrow: %v", label, err)
			}
			if e.Balance() != escrowStart {
				t.Errorf("%s: escrow = %d, want %d", label, e.Balance(), escrowStart)
			}
		}

		betPtr := func(v int64) *int64 { return &v }

		makePlayerPair := func(tag string) (host, guest *models.Player) {
			hUser := mustCreateUser(t, ctx, adapter, tag+"_h@example.com")
			gUser := mustCreateUser(t, ctx, adapter, tag+"_g@example.com")
			h := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
				stores.WithPlayerEmail(tag+"_h@example.com"),
				stores.WithUserID(hUser.ID),
			)
			g := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
				stores.WithPlayerEmail(tag+"_g@example.com"),
				stores.WithUserID(gUser.ID),
			)
			mustFundPlayerWallet(t, ctx, adapter, ledger, h.UserID, 500)
			mustFundPlayerWallet(t, ctx, adapter, ledger, g.UserID, 500)
			return h, g
		}

		// Game 1: complete (host wins) — escrow must return to start immediately.
		{
			h, g := makePlayerPair("ezg1")
			game, err := rpsService.RequestGame(ctx, &RpsGameRequestInput{
				RequestingPlayerID:   h.ID,
				InvitedPlayerID:      g.ID,
				RequestingPlayerMove: models.RpsParticipantMoveRock,
				DurationSeconds:      3600,
				BetAmount:            betPtr(100),
				HostUserID:           h.UserID,
			})
			if err != nil {
				t.Fatalf("game1 RequestGame: %v", err)
			}
			if _, err = rpsService.RespondToGameRequest(ctx, &GameRequestResponse{
				InvitedPlayerID: g.ID,
				GameID:          game.RpsGame.ID,
				Status:          models.RpsGameStatusCompleted,
				Move:            models.RpsParticipantMoveScissors,
			}); err != nil {
				t.Fatalf("game1 Respond: %v", err)
			}
			assertEscrowAtStart("after game1 (host wins)")
		}

		// Game 2: cancelled — escrow must return to start immediately.
		{
			h, g := makePlayerPair("ezg2")
			game, err := rpsService.RequestGame(ctx, &RpsGameRequestInput{
				RequestingPlayerID:   h.ID,
				InvitedPlayerID:      g.ID,
				RequestingPlayerMove: models.RpsParticipantMoveRock,
				DurationSeconds:      3600,
				BetAmount:            betPtr(100),
				HostUserID:           h.UserID,
			})
			if err != nil {
				t.Fatalf("game2 RequestGame: %v", err)
			}
			if _, err = rpsService.RespondToGameRequest(ctx, &GameRequestResponse{
				InvitedPlayerID: g.ID,
				GameID:          game.RpsGame.ID,
				Status:          models.RpsGameStatusCancelled,
			}); err != nil {
				t.Fatalf("game2 Respond: %v", err)
			}
			assertEscrowAtStart("after game2 (cancelled)")
		}

		// Game 3: expired — escrow must return to start after sweep.
		{
			h, g := makePlayerPair("ezg3")
			_, err := rpsService.RequestGame(ctx, &RpsGameRequestInput{
				RequestingPlayerID:   h.ID,
				InvitedPlayerID:      g.ID,
				RequestingPlayerMove: models.RpsParticipantMoveRock,
				DurationSeconds:      1,
				BetAmount:            betPtr(100),
				HostUserID:           h.UserID,
			})
			if err != nil {
				t.Fatalf("game3 RequestGame: %v", err)
			}
			time.Sleep(2 * time.Second)
			if _, err = rpsService.ExpireGamesAndRefundBets(ctx); err != nil {
				t.Fatalf("game3 ExpireGamesAndRefundBets: %v", err)
			}
			assertEscrowAtStart("after game3 (expired)")
		}
	})
}

// TestBettingInvariant_NoOrphanPendingAfterAllPaths verifies that after running
// all three terminal game paths (complete / cancel / expire), the ledger transfer
// table contains zero pending bet_escrow rows. This is a direct audit of the
// transfer table — stronger than checking balances alone.
func TestBettingInvariant_NoOrphanPendingAfterAllPaths(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)
		rpsService := NewDbRpsGameService(adapter, betting)

		betPtr := func(v int64) *int64 { return &v }

		makePlayerPair := func(tag string) (host, guest *models.Player) {
			hUser := mustCreateUser(t, ctx, adapter, tag+"_h@example.com")
			gUser := mustCreateUser(t, ctx, adapter, tag+"_g@example.com")
			h := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
				stores.WithPlayerEmail(tag+"_h@example.com"),
				stores.WithUserID(hUser.ID),
			)
			g := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
				stores.WithPlayerEmail(tag+"_g@example.com"),
				stores.WithUserID(gUser.ID),
			)
			mustFundPlayerWallet(t, ctx, adapter, ledger, h.UserID, 500)
			mustFundPlayerWallet(t, ctx, adapter, ledger, g.UserID, 500)
			return h, g
		}

		// Path 1: complete (host wins).
		{
			h, g := makePlayerPair("orp1")
			game, err := rpsService.RequestGame(ctx, &RpsGameRequestInput{
				RequestingPlayerID:   h.ID,
				InvitedPlayerID:      g.ID,
				RequestingPlayerMove: models.RpsParticipantMoveRock,
				DurationSeconds:      3600,
				BetAmount:            betPtr(100),
				HostUserID:           h.UserID,
			})
			if err != nil {
				t.Fatalf("path1 RequestGame: %v", err)
			}
			if _, err = rpsService.RespondToGameRequest(ctx, &GameRequestResponse{
				InvitedPlayerID: g.ID,
				GameID:          game.RpsGame.ID,
				Status:          models.RpsGameStatusCompleted,
				Move:            models.RpsParticipantMoveScissors,
			}); err != nil {
				t.Fatalf("path1 Respond: %v", err)
			}
		}

		// Path 2: cancelled.
		{
			h, g := makePlayerPair("orp2")
			game, err := rpsService.RequestGame(ctx, &RpsGameRequestInput{
				RequestingPlayerID:   h.ID,
				InvitedPlayerID:      g.ID,
				RequestingPlayerMove: models.RpsParticipantMoveRock,
				DurationSeconds:      3600,
				BetAmount:            betPtr(100),
				HostUserID:           h.UserID,
			})
			if err != nil {
				t.Fatalf("path2 RequestGame: %v", err)
			}
			if _, err = rpsService.RespondToGameRequest(ctx, &GameRequestResponse{
				InvitedPlayerID: g.ID,
				GameID:          game.RpsGame.ID,
				Status:          models.RpsGameStatusCancelled,
			}); err != nil {
				t.Fatalf("path2 Respond: %v", err)
			}
		}

		// Path 3: expired.
		{
			h, g := makePlayerPair("orp3")
			_, err := rpsService.RequestGame(ctx, &RpsGameRequestInput{
				RequestingPlayerID:   h.ID,
				InvitedPlayerID:      g.ID,
				RequestingPlayerMove: models.RpsParticipantMoveRock,
				DurationSeconds:      1,
				BetAmount:            betPtr(100),
				HostUserID:           h.UserID,
			})
			if err != nil {
				t.Fatalf("path3 RequestGame: %v", err)
			}
			time.Sleep(2 * time.Second)
			if _, err = rpsService.ExpireGamesAndRefundBets(ctx); err != nil {
				t.Fatalf("path3 ExpireGamesAndRefundBets: %v", err)
			}
		}

		// No pending bet_escrow transfers anywhere in the ledger.
		count, err := ledger.CountTransfers(ctx, &stores.LedgerTransferFilter{
			TransferCodes: []string{models.TransferCodeBetEscrow},
			Statuses:      []models.LedgerTransferStatus{models.LedgerTransferStatusPending},
		})
		if err != nil {
			t.Fatalf("CountTransfers: %v", err)
		}
		if count != 0 {
			t.Errorf("orphaned pending bet_escrow transfers = %d, want 0", count)
		}
	})
}
```

- [ ] **Step 2: Run both new invariant tests**

```bash
cd /Users/tkahng/github/tkahng/go/playground
go test ./internal/services/ -run "TestBettingInvariant_EscrowNetsZeroAfterEachGame|TestBettingInvariant_NoOrphanPendingAfterAllPaths" -v -count=1 -timeout 60s
```

Expected: both `PASS`. Each test sleeps ~2 seconds for the expiry game.

- [ ] **Step 3: Run the full service test suite to confirm no regressions**

```bash
cd /Users/tkahng/github/tkahng/go/playground
go test ./internal/services/ -v -count=1 -timeout 120s 2>&1 | tail -30
```

Expected: all tests pass. Check that the final line contains `ok` not `FAIL`.

- [ ] **Step 4: Commit**

```bash
git add internal/services/betting_invariants_test.go
git commit -m "test: escrow-per-game and no-orphan-pending invariant tests"
```

---

## Self-Review Against Spec

**Spec requirement → Task coverage:**

| Spec requirement | Task |
|-----------------|------|
| `TestBetting_NoPendingTransfers_AfterCancel` | Task 1 |
| `TestBetting_NoPendingTransfers_AfterExpiry` | Task 1 |
| `TestBetting_NoPendingTransfers_AfterComplete_HostWins` | Task 2 |
| `TestBetting_NoPendingTransfers_AfterComplete_Tie` | Task 2 |
| `TestBetting_GuestCanRetry_AfterInsufficientFunds` | Task 2 |
| `TestBettingInvariant_MultiGame_TotalSystemConservation` | Task 3 |
| `TestBettingInvariant_EscrowNetsZeroAfterEachGame` | Task 4 |
| `TestBettingInvariant_NoOrphanPendingAfterAllPaths` | Task 4 |

All 8 tests from the spec are covered. No placeholders. All method names (`CountTransfers`, `GetUserBalance`, `GetUserAvailableBalance`, `ExpireGamesAndRefundBets`, `FindRpsGame`) match what exists in the codebase. Filter fields (`ReferenceIds`, `Statuses`, `TransferCodes`) match `stores.LedgerTransferFilter`. All helpers (`mustCreateUser`, `mustFundPlayerWallet`, `MustCreatePlayer`) are used exactly as defined — no redefinitions.
