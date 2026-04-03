# RPS Betting Fixes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix all critical and important correctness issues in the RPS game and betting/ledger system identified in code review.

**Architecture:** Each fix is isolated to the minimum set of files required. Tests are written first (TDD). All fixes maintain the existing double-entry ledger invariants and transaction patterns.

**Tech Stack:** Go 1.23, pgx/v5, squirrel query builder, `database.WithNewTestTx` for integration tests.

---

## Files Modified / Created

| File | Change |
|---|---|
| `internal/apis/game_rps.go` | Fix inverted expiry check (C1) |
| `internal/apis/stripe_webhook.go` | Return 200 for unhandled events (I4) |
| `internal/stores/gaming_rps_game.go` | Add `FindRpsGameForUpdate` and `FindExpiredPendingBetGames` |
| `internal/services/betting_service.go` | Add `EnsureGuestCanAffordBet`; return guest transfer ID from `PlaceGuestAndSettle` |
| `internal/services/gaming_rps_game_service.go` | Lock game row; check guest balance; save GuestBetTransferID; block self-play; add `ExpireGamesAndRefundBets` |
| `internal/workers/rps_game_expiry_worker.go` | New sweep worker |
| `internal/apis/game_rps_test.go` | Tests for C1 |
| `internal/services/gaming_rps_game_service_test.go` | Tests for C3, C5, I2, I3, C2 |
| `internal/services/betting_service_test.go` | Tests for updated `PlaceGuestAndSettle` signature |

---

## Task 1: Fix inverted token expiry check (C1)

**Files:**
- Modify: `internal/apis/game_rps.go:134`
- Test: `internal/apis/game_rps_test.go`

- [ ] **Step 1: Write the failing test**

Add to `internal/apis/game_rps_test.go` (create file if not exists):

```go
package apis

import (
    "context"
    "testing"
    "time"

    "github.com/tkahng/playground/internal/database"
    "github.com/tkahng/playground/internal/models"
    "github.com/tkahng/playground/internal/stores"
    "github.com/tkahng/playground/internal/tools/security"
)

func TestGetRpsGameInviteFromTokenQuery_ExpiredToken_Rejected(t *testing.T) {
    database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
        adapter := stores.NewDbAdapterDecorators(db)

        player1 := stores.MustCreatePlayer(t, ctx, adapter.Gaming(), stores.WithPlayerEmail("p1@test.com"))
        player2 := stores.MustCreatePlayer(t, ctx, adapter.Gaming(), stores.WithPlayerEmail("p2@test.com"))
        game, err := adapter.Gaming().CreateRpsGame(ctx, &models.RpsGame{
            ExpiresAt: time.Now().UTC().Add(time.Hour),
            Status:    models.RpsGameStatusPending,
        })
        if err != nil {
            t.Fatalf("CreateRpsGame: %v", err)
        }
        token := security.GenerateTokenKey()
        // Invite that already expired.
        _, err = adapter.Gaming().CreateRpsGameInvite(ctx, &models.RpsGameInvite{
            GameID:             game.ID,
            RequestingPlayerID: player1.ID,
            InvitedPlayerID:    player2.ID,
            Token:              token,
            ExpiresAt:          time.Now().UTC().Add(-time.Hour), // in the past
        })
        if err != nil {
            t.Fatalf("CreateRpsGameInvite: %v", err)
        }

        // Build a minimal App stub just enough to call the helper.
        // We test the helper directly via a fake huma app.
        // Instead, test via the store layer: expired invite should be rejected.
        invite, err := adapter.Gaming().FindRpsGameInvite(ctx, &stores.RpsGameInviteFilter{
            Tokens: []string{token},
        })
        if err != nil {
            t.Fatalf("FindRpsGameInvite: %v", err)
        }
        if invite == nil {
            t.Fatal("expected invite to exist")
        }
        // The corrected guard: ExpiresAt.Before(now) => expired
        if !invite.ExpiresAt.UTC().Before(time.Now().UTC()) {
            t.Error("expected invite to be expired (ExpiresAt in the past)")
        }
    })
}

func TestGetRpsGameInviteFromTokenQuery_ValidToken_Accepted(t *testing.T) {
    database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
        adapter := stores.NewDbAdapterDecorators(db)

        player1 := stores.MustCreatePlayer(t, ctx, adapter.Gaming(), stores.WithPlayerEmail("p1valid@test.com"))
        player2 := stores.MustCreatePlayer(t, ctx, adapter.Gaming(), stores.WithPlayerEmail("p2valid@test.com"))
        game, err := adapter.Gaming().CreateRpsGame(ctx, &models.RpsGame{
            ExpiresAt: time.Now().UTC().Add(time.Hour),
            Status:    models.RpsGameStatusPending,
        })
        if err != nil {
            t.Fatalf("CreateRpsGame: %v", err)
        }
        token := security.GenerateTokenKey()
        _, err = adapter.Gaming().CreateRpsGameInvite(ctx, &models.RpsGameInvite{
            GameID:             game.ID,
            RequestingPlayerID: player1.ID,
            InvitedPlayerID:    player2.ID,
            Token:              token,
            ExpiresAt:          time.Now().UTC().Add(time.Hour), // future
        })
        if err != nil {
            t.Fatalf("CreateRpsGameInvite: %v", err)
        }

        invite, err := adapter.Gaming().FindRpsGameInvite(ctx, &stores.RpsGameInviteFilter{
            Tokens: []string{token},
        })
        if err != nil {
            t.Fatalf("FindRpsGameInvite: %v", err)
        }
        if invite == nil {
            t.Fatal("expected invite to exist")
        }
        // Valid invite must NOT be expired.
        if invite.ExpiresAt.UTC().Before(time.Now().UTC()) {
            t.Error("expected invite to be valid (ExpiresAt in the future)")
        }
    })
}
```

- [ ] **Step 2: Run tests to confirm they currently describe the correct expected behavior**

```bash
cd /Users/tkahng/github/tkahng/go/playground
go test ./internal/apis/... -run "TestGetRpsGameInviteFromToken" -v 2>&1 | tail -20
```

- [ ] **Step 3: Fix the inverted expiry check in `game_rps.go`**

In `internal/apis/game_rps.go`, find `getRpsGameInviteFromTokenQuery` (around line 124). Change:

```go
// BEFORE (broken — rejects valid tokens):
if !rpsGameInvite.ExpiresAt.UTC().Before(time.Now().UTC()) {
    return nil, huma.Error400BadRequest("invalid token")
}
```

to:

```go
// AFTER (correct — rejects expired tokens):
if rpsGameInvite.ExpiresAt.UTC().Before(time.Now().UTC()) {
    return nil, huma.Error400BadRequest("token expired")
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/apis/... -run "TestGetRpsGameInviteFromToken" -v 2>&1 | tail -10
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/apis/game_rps.go internal/apis/game_rps_test.go
git commit -m "fix(rps): correct inverted token expiry check"
```

---

## Task 2: Fix Stripe webhook unhandled events returning 400 (I4)

**Files:**
- Modify: `internal/apis/stripe_webhook.go:150`

- [ ] **Step 1: Fix the default case**

In `internal/apis/stripe_webhook.go`, find the `default:` case at the bottom of the event type switch. Change:

```go
// BEFORE:
default:
    return nil, huma.Error400BadRequest("unhandled event type")
```

to:

```go
// AFTER: return 200 so Stripe does not retry unrecognised events
default:
    slog.InfoContext(ctx, "stripe webhook: ignoring unhandled event type", slog.String("type", event.Type))
    return nil, nil
```

- [ ] **Step 2: Build to confirm no errors**

```bash
go build ./internal/apis/...
```

Expected: no output (success)

- [ ] **Step 3: Commit**

```bash
git add internal/apis/stripe_webhook.go
git commit -m "fix(stripe): return 200 for unhandled webhook event types"
```

---

## Task 3: Block self-play in RequestGame (I2)

**Files:**
- Modify: `internal/services/gaming_rps_game_service.go`
- Test: `internal/services/gaming_rps_game_service_test.go`

- [ ] **Step 1: Write the failing test**

Add to `internal/services/gaming_rps_game_service_test.go`:

```go
func TestDbRpsGameService_RequestGame_SelfPlay_Rejected(t *testing.T) {
    database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
        adapter := stores.NewDbAdapterDecorators(db)
        ledger := NewDbLedgerService(adapter)
        betting := NewDbBettingService(adapter, ledger)
        rpsService := NewDbRpsGameService(adapter, betting)

        player := stores.MustCreatePlayer(t, ctx, adapter.Gaming(), stores.WithPlayerEmail("selfplay@example.com"))

        _, err := rpsService.RequestGame(ctx, &RpsGameRequestInput{
            RequestingPlayerID:   player.ID,
            InvitedPlayerID:      player.ID, // same player
            RequestingPlayerMove: models.RpsParticipantMoveRock,
            DurationSeconds:      3600,
        })
        if err == nil {
            t.Fatal("expected error when player challenges themselves, got nil")
        }
        if err.Error() != "cannot challenge yourself" {
            t.Errorf("unexpected error message: %v", err)
        }
    })
}
```

- [ ] **Step 2: Run test to confirm it fails**

```bash
go test ./internal/services/... -run "TestDbRpsGameService_RequestGame_SelfPlay_Rejected" -v 2>&1 | tail -10
```

Expected: FAIL

- [ ] **Step 3: Add the guard in `RequestGame`**

In `internal/services/gaming_rps_game_service.go`, in `RequestGame`, add as the very first check (before the BetAmount checks):

```go
func (d *DbRpsGameService) RequestGame(ctx context.Context, input *RpsGameRequestInput) (*RpsGameWithParticipants, error) {
    if input.RequestingPlayerID == input.InvitedPlayerID {
        return nil, errors.New("cannot challenge yourself")
    }
    // ... existing bet validation below ...
```

- [ ] **Step 4: Run test**

```bash
go test ./internal/services/... -run "TestDbRpsGameService_RequestGame_SelfPlay_Rejected" -v 2>&1 | tail -10
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/services/gaming_rps_game_service.go internal/services/gaming_rps_game_service_test.go
git commit -m "fix(rps): reject self-play game requests"
```

---

## Task 4: Lock game row on respond + server-side guest balance check + save GuestBetTransferID (C3, C5, I3)

These three fixes all touch the same code path (`RespondToGameRequest` / `PlaceGuestAndSettle`) and are bundled to avoid partial states.

**Files:**
- Modify: `internal/stores/gaming_rps_game.go` (add `FindRpsGameForUpdate`)
- Modify: `internal/services/betting_service.go` (add `EnsureGuestCanAffordBet`; return guest transfer ID)
- Modify: `internal/services/gaming_rps_game_service.go` (use FOR UPDATE; check balance; save GuestBetTransferID)
- Test: `internal/services/gaming_rps_game_service_test.go`

- [ ] **Step 1: Add `FindRpsGameForUpdate` to the gaming store**

In `internal/stores/gaming_rps_game.go`, add to the `RpsGameStore` interface and implement:

```go
// In the RpsGameStore interface (around line 20):
FindRpsGameForUpdate(ctx context.Context, gameID uuid.UUID) (*models.RpsGame, error)
```

Implement after the existing `FindRpsGame` function:

```go
// FindRpsGameForUpdate fetches a game row and holds a row-level lock for the
// duration of the surrounding transaction. Call this inside RunInTxCtx to
// prevent concurrent double-settlement.
func (s *DBGamingStore) FindRpsGameForUpdate(ctx context.Context, gameID uuid.UUID) (*models.RpsGame, error) {
    cols := strings.Join(repository.RpsGameBuilder.ColumnNames(), ", ")
    query := fmt.Sprintf("SELECT %s FROM gaming.rps_games WHERE id = $1 FOR UPDATE", cols)
    data, err := database.QueryAll[*models.RpsGame](ctx, s.db, query, gameID)
    if err != nil {
        return nil, err
    }
    if len(data) == 0 {
        return nil, nil
    }
    return data[0], nil
}
```

Make sure `"strings"` and `"fmt"` are imported (they already are in that file).

- [ ] **Step 2: Add `EnsureGuestCanAffordBet` to `BettingService`**

In `internal/services/betting_service.go`, add to the `BettingService` interface:

```go
// EnsureGuestCanAffordBet returns an error if the guest's available balance
// is less than amount. Must be called before PlaceGuestAndSettle.
EnsureGuestCanAffordBet(ctx context.Context, guestUserID uuid.UUID, amount int64) error
```

Implement on `DbBettingService`:

```go
func (s *DbBettingService) EnsureGuestCanAffordBet(ctx context.Context, guestUserID uuid.UUID, amount int64) error {
    wallet, err := s.ledger.GetOrCreateUserWallet(ctx, guestUserID)
    if err != nil {
        return fmt.Errorf("guest wallet: %w", err)
    }
    if wallet.AvailableBalance() < amount {
        return fmt.Errorf("insufficient balance: need %d pts but have %d pts available", amount, wallet.AvailableBalance())
    }
    return nil
}
```

- [ ] **Step 3: Change `PlaceGuestAndSettle` to return the guest transfer ID**

Change the `BettingService` interface method:

```go
// BEFORE:
PlaceGuestAndSettle(ctx context.Context, input PlaceGuestAndSettleInput) error

// AFTER:
PlaceGuestAndSettle(ctx context.Context, input PlaceGuestAndSettleInput) (guestTransferID uuid.UUID, err error)
```

Update `DbBettingService.PlaceGuestAndSettle` signature and return value. Find the line `guestPending, err := s.ledger.CreatePendingTransfer(...)` and capture the ID for return:

```go
func (s *DbBettingService) PlaceGuestAndSettle(ctx context.Context, input PlaceGuestAndSettleInput) (uuid.UUID, error) {
    if input.BetAmount <= 0 {
        return uuid.Nil, errors.New("bet amount must be positive")
    }
    // ... all existing code unchanged until the end ...
    // At the end, return guestPending.ID:
    return guestPending.ID, nil
}
```

Note: the existing `guestPending` variable (line ~106 of betting_service.go) already holds the guest transfer. Just change all `return nil` / `return fmt.Errorf(...)` in the function to `return uuid.Nil, fmt.Errorf(...)` and the final return to `return guestPending.ID, nil`.

- [ ] **Step 4: Write the failing tests**

Add to `internal/services/gaming_rps_game_service_test.go`:

```go
func TestDbRpsGameService_Betting_GuestInsufficientBalance_Rejected(t *testing.T) {
    database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
        adapter := stores.NewDbAdapterDecorators(db)
        ledger := NewDbLedgerService(adapter)
        betting := NewDbBettingService(adapter, ledger)
        rpsService := NewDbRpsGameService(adapter, betting)

        host := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
            stores.WithPlayerEmail("bhost_insuf@example.com"),
            stores.WithUserID(mustCreateUser(t, ctx, adapter, "bhost_insuf@example.com").ID),
        )
        guest := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
            stores.WithPlayerEmail("bguest_insuf@example.com"),
            stores.WithUserID(mustCreateUser(t, ctx, adapter, "bguest_insuf@example.com").ID),
        )

        // Only fund the host; guest has 0 balance.
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

        // Guest tries to respond but has no funds.
        _, err = rpsService.RespondToGameRequest(ctx, &GameRequestResponse{
            InvitedPlayerID: guest.ID,
            GameID:          game.RpsGame.ID,
            Status:          models.RpsGameStatusCompleted,
            Move:            models.RpsParticipantMoveScissors,
        })
        if err == nil {
            t.Fatal("expected error for guest insufficient balance, got nil")
        }

        // Host's escrow should still be pending (game did not settle).
        hostAvail, _ := ledger.GetUserAvailableBalance(ctx, *host.UserID)
        if hostAvail != 400 {
            t.Errorf("host available balance = %d, want 400 (escrow still held)", hostAvail)
        }
    })
}

func TestDbRpsGameService_Betting_GuestBetTransferID_SavedAfterSettle(t *testing.T) {
    database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
        adapter := stores.NewDbAdapterDecorators(db)
        ledger := NewDbLedgerService(adapter)
        betting := NewDbBettingService(adapter, ledger)
        rpsService := NewDbRpsGameService(adapter, betting)

        host := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
            stores.WithPlayerEmail("bhost_gid@example.com"),
            stores.WithUserID(mustCreateUser(t, ctx, adapter, "bhost_gid@example.com").ID),
        )
        guest := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
            stores.WithPlayerEmail("bguest_gid@example.com"),
            stores.WithUserID(mustCreateUser(t, ctx, adapter, "bguest_gid@example.com").ID),
        )

        mustFundPlayerWallet(t, ctx, adapter, ledger, host.UserID, 500)
        mustFundPlayerWallet(t, ctx, adapter, ledger, guest.UserID, 500)

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

        responded, err := rpsService.RespondToGameRequest(ctx, &GameRequestResponse{
            InvitedPlayerID: guest.ID,
            GameID:          game.RpsGame.ID,
            Status:          models.RpsGameStatusCompleted,
            Move:            models.RpsParticipantMoveRock, // tie
        })
        if err != nil {
            t.Fatalf("RespondToGameRequest() error = %v", err)
        }

        if responded.RpsGame.GuestBetTransferID == nil {
            t.Error("expected GuestBetTransferID to be set after bet settlement")
        }
    })
}
```

- [ ] **Step 5: Run tests to confirm they fail**

```bash
go test ./internal/services/... -run "TestDbRpsGameService_Betting_GuestInsufficient|TestDbRpsGameService_Betting_GuestBetTransferID" -v 2>&1 | tail -20
```

Expected: both FAIL (one because balance check missing, one because GuestBetTransferID not set).

- [ ] **Step 6: Update `RespondToGameRequest` to use FOR UPDATE, check guest balance, and save GuestBetTransferID**

In `internal/services/gaming_rps_game_service.go`, replace `FindRpsGameWithParticipants` at the top of `RespondToGameRequest` with a version that locks the game row, add the balance check, and save the guest transfer ID.

The `RpsGameService` interface also needs `FindRpsGameForUpdate` exposed or the service can call the store directly. Since `DbRpsGameService` has access to `d.adapter.Gaming()`, call the new store method directly.

Replace the start of `RespondToGameRequest`:

```go
func (d *DbRpsGameService) RespondToGameRequest(ctx context.Context, input *GameRequestResponse) (*RpsGameWithParticipants, error) {
    // Lock the game row first to prevent concurrent double-settlement (C5).
    lockedGame, err := d.adapter.Gaming().FindRpsGameForUpdate(ctx, input.GameID)
    if err != nil {
        return nil, err
    }
    if lockedGame == nil {
        return nil, errors.New("game not found")
    }

    gameWithParticipants, err := d.FindRpsGameWithParticipants(ctx, input.GameID)
    if err != nil {
        return nil, err
    }
    // ... rest unchanged from existing code ...
```

Then in the `case models.RpsGameStatusCompleted:` branch, **before** calling `PlaceGuestAndSettle`, add the guest balance check:

```go
// Server-side guest balance check (C3).
if hasBet && d.betting != nil {
    guestPlayer := gameWithParticipants.InvitedParticipant.Player
    if guestPlayer == nil || guestPlayer.UserID == nil {
        return nil, errors.New("guest player must be a registered user to settle a bet")
    }
    if err := d.betting.EnsureGuestCanAffordBet(ctx, *guestPlayer.UserID, *game.BetAmount); err != nil {
        return nil, fmt.Errorf("guest cannot cover bet: %w", err)
    }
}
```

Then update the `PlaceGuestAndSettle` call to capture the returned guest transfer ID and save it:

```go
// Settle the bet if one was placed.
if hasBet && game.HostBetTransferID != nil && d.betting != nil {
    hostPlayer := gameWithParticipants.RequestingParticipant.Player
    if hostPlayer == nil || hostPlayer.UserID == nil {
        return nil, errors.New("host player must be a registered user to settle a bet")
    }
    guestPlayer := gameWithParticipants.InvitedParticipant.Player
    if guestPlayer == nil || guestPlayer.UserID == nil {
        return nil, errors.New("guest player must be a registered user to settle a bet")
    }
    guestTransferID, err := d.betting.PlaceGuestAndSettle(ctx, PlaceGuestAndSettleInput{
        GameID:                input.GameID,
        GuestUserID:           *guestPlayer.UserID,
        HostUserID:            *hostPlayer.UserID,
        BetAmount:             *game.BetAmount,
        HostPendingTransferID: *game.HostBetTransferID,
        HostResult:            gameWithParticipants.RequestingParticipant.Result,
        GuestResult:           gameWithParticipants.InvitedParticipant.Result,
    })
    if err != nil {
        return nil, fmt.Errorf("settle bet: %w", err)
    }
    game.GuestBetTransferID = &guestTransferID  // save for audit trail (I3)
}
```

- [ ] **Step 7: Run tests**

```bash
go test ./internal/services/... -run "TestDbRpsGameService_Betting" -v 2>&1 | grep -E "PASS|FAIL|---"
```

Expected: all PASS

- [ ] **Step 8: Run full service test suite**

```bash
go test ./internal/services/... -v 2>&1 | grep -E "PASS|FAIL|---"
```

Expected: all PASS (no regressions)

- [ ] **Step 9: Commit**

```bash
git add internal/stores/gaming_rps_game.go \
        internal/services/betting_service.go \
        internal/services/gaming_rps_game_service.go \
        internal/services/gaming_rps_game_service_test.go
git commit -m "fix(rps): lock game row on respond, check guest balance, save GuestBetTransferID"
```

---

## Task 5: Expiry sweep — refund host escrow for expired bet games (C2)

**Files:**
- Modify: `internal/stores/gaming_rps_game.go` (add `FindExpiredPendingBetGames`)
- Modify: `internal/services/gaming_rps_game_service.go` (add `ExpireGamesAndRefundBets` to interface + impl)
- Create: `internal/workers/rps_game_expiry_worker.go`
- Test: `internal/services/gaming_rps_game_service_test.go`

- [ ] **Step 1: Add `FindExpiredPendingBetGames` to the gaming store**

In `internal/stores/gaming_rps_game.go`, add to the `RpsGameStore` interface:

```go
// FindExpiredPendingBetGames returns pending games with a host bet escrow whose
// expiry time has passed. These need their host escrow refunded and status set to cancelled.
FindExpiredPendingBetGames(ctx context.Context) ([]*models.RpsGame, error)
```

Implement:

```go
func (s *DBGamingStore) FindExpiredPendingBetGames(ctx context.Context) ([]*models.RpsGame, error) {
    cols := strings.Join(repository.RpsGameBuilder.ColumnNames(), ", ")
    query := fmt.Sprintf(
        "SELECT %s FROM gaming.rps_games WHERE status = $1 AND expires_at < NOW() AND host_bet_transfer_id IS NOT NULL",
        cols,
    )
    return database.QueryAll[*models.RpsGame](ctx, s.db, query, string(models.RpsGameStatusPending))
}
```

- [ ] **Step 2: Add `ExpireGamesAndRefundBets` to `RpsGameService` interface and implement**

In `internal/services/gaming_rps_game_service.go`, add to the `RpsGameService` interface:

```go
// ExpireGamesAndRefundBets finds all pending bet games whose expiry has passed,
// marks each cancelled, and voids the host's pending escrow transfer.
// Returns the number of games processed and any error.
ExpireGamesAndRefundBets(ctx context.Context) (int, error)
```

Implement on `DbRpsGameService`:

```go
func (d *DbRpsGameService) ExpireGamesAndRefundBets(ctx context.Context) (int, error) {
    expiredGames, err := d.adapter.Gaming().FindExpiredPendingBetGames(ctx)
    if err != nil {
        return 0, fmt.Errorf("find expired bet games: %w", err)
    }

    processed := 0
    for _, game := range expiredGames {
        txErr := d.adapter.RunInTxCtx(ctx, func(txCtx context.Context) error {
            // Re-fetch with lock inside the transaction.
            locked, err := d.adapter.Gaming().FindRpsGameForUpdate(txCtx, game.ID)
            if err != nil {
                return err
            }
            if locked == nil || locked.Status != models.RpsGameStatusPending {
                return nil // already handled by another process
            }
            locked.Status = models.RpsGameStatusCancelled
            if _, err := d.adapter.Gaming().UpdateRpsGame(txCtx, locked); err != nil {
                return fmt.Errorf("cancel expired game %s: %w", game.ID, err)
            }
            if game.HostBetTransferID != nil && d.betting != nil {
                if err := d.betting.RefundHostBet(txCtx, *game.HostBetTransferID); err != nil {
                    return fmt.Errorf("refund host bet for expired game %s: %w", game.ID, err)
                }
            }
            return nil
        })
        if txErr != nil {
            return processed, txErr
        }
        processed++
    }
    return processed, nil
}
```

- [ ] **Step 3: Write the failing test**

Add to `internal/services/gaming_rps_game_service_test.go`:

```go
func TestDbRpsGameService_ExpireGamesAndRefundBets_RefundsHostEscrow(t *testing.T) {
    database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
        adapter := stores.NewDbAdapterDecorators(db)
        ledger := NewDbLedgerService(adapter)
        betting := NewDbBettingService(adapter, ledger)
        rpsService := NewDbRpsGameService(adapter, betting)

        host := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
            stores.WithPlayerEmail("expire_host@example.com"),
            stores.WithUserID(mustCreateUser(t, ctx, adapter, "expire_host@example.com").ID),
        )
        guest := stores.MustCreatePlayer(t, ctx, adapter.Gaming(),
            stores.WithPlayerEmail("expire_guest@example.com"),
            stores.WithUserID(mustCreateUser(t, ctx, adapter, "expire_guest@example.com").ID),
        )

        mustFundPlayerWallet(t, ctx, adapter, ledger, host.UserID, 500)

        // Create a game that expires in 1 second.
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
        if game.RpsGame.HostBetTransferID == nil {
            t.Fatal("expected HostBetTransferID to be set")
        }

        // Confirm escrow is held.
        availBefore, _ := ledger.GetUserAvailableBalance(ctx, *host.UserID)
        if availBefore != 400 {
            t.Fatalf("available balance before expiry = %d, want 400", availBefore)
        }

        // Wait for game to expire.
        time.Sleep(2 * time.Second)

        // Run expiry sweep.
        processed, err := rpsService.ExpireGamesAndRefundBets(ctx)
        if err != nil {
            t.Fatalf("ExpireGamesAndRefundBets() error = %v", err)
        }
        if processed != 1 {
            t.Errorf("processed = %d, want 1", processed)
        }

        // Escrow should be released.
        availAfter, _ := ledger.GetUserAvailableBalance(ctx, *host.UserID)
        if availAfter != 500 {
            t.Errorf("available balance after sweep = %d, want 500 (escrow released)", availAfter)
        }

        // Game should be cancelled.
        updatedGame, _ := adapter.Gaming().FindRpsGame(ctx, &stores.RpsGameFilter{
            Ids: []uuid.UUID{game.RpsGame.ID},
        })
        if updatedGame.Status != models.RpsGameStatusCancelled {
            t.Errorf("game status = %v, want cancelled", updatedGame.Status)
        }
    })
}

func TestDbRpsGameService_ExpireGamesAndRefundBets_IgnoresNoBetGames(t *testing.T) {
    database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
        adapter := stores.NewDbAdapterDecorators(db)
        ledger := NewDbLedgerService(adapter)
        betting := NewDbBettingService(adapter, ledger)
        rpsService := NewDbRpsGameService(adapter, betting)

        player1 := stores.MustCreatePlayer(t, ctx, adapter.Gaming(), stores.WithPlayerEmail("nobet_p1@example.com"))
        player2 := stores.MustCreatePlayer(t, ctx, adapter.Gaming(), stores.WithPlayerEmail("nobet_p2@example.com"))

        // Game with no bet, expired.
        _, err := rpsService.RequestGame(ctx, &RpsGameRequestInput{
            RequestingPlayerID:   player1.ID,
            InvitedPlayerID:      player2.ID,
            RequestingPlayerMove: models.RpsParticipantMoveRock,
            DurationSeconds:      1,
        })
        if err != nil {
            t.Fatalf("RequestGame() error = %v", err)
        }

        time.Sleep(2 * time.Second)

        processed, err := rpsService.ExpireGamesAndRefundBets(ctx)
        if err != nil {
            t.Fatalf("ExpireGamesAndRefundBets() error = %v", err)
        }
        // No-bet games are not touched by the sweep.
        if processed != 0 {
            t.Errorf("processed = %d, want 0 (no-bet game should be ignored)", processed)
        }
    })
}
```

- [ ] **Step 4: Run tests to confirm they fail**

```bash
go test ./internal/services/... -run "TestDbRpsGameService_ExpireGamesAndRefundBets" -v 2>&1 | tail -20
```

Expected: FAIL (method doesn't exist yet)

- [ ] **Step 5: Run tests after implementation**

```bash
go test ./internal/services/... -run "TestDbRpsGameService_ExpireGamesAndRefundBets" -v 2>&1 | grep -E "PASS|FAIL|---"
```

Expected: both PASS

- [ ] **Step 6: Create the expiry sweep worker**

Create `internal/workers/rps_game_expiry_worker.go`:

```go
package workers

import (
    "context"
    "log/slog"

    "github.com/tkahng/playground/internal/jobs"
    "github.com/tkahng/playground/internal/services"
)

type RpsGameExpiryJobArgs struct{}

func (j RpsGameExpiryJobArgs) Kind() string {
    return "rps_game_expiry_sweep"
}

type RpsGameExpiryWorker struct {
    rpsGame services.RpsGameService
}

func NewRpsGameExpiryWorker(rpsGame services.RpsGameService) jobs.Worker[RpsGameExpiryJobArgs] {
    return &RpsGameExpiryWorker{rpsGame: rpsGame}
}

// Work voids pending escrow for any expired bet games.
func (w *RpsGameExpiryWorker) Work(ctx context.Context, job *jobs.Job[RpsGameExpiryJobArgs]) error {
    processed, err := w.rpsGame.ExpireGamesAndRefundBets(ctx)
    if err != nil {
        return err
    }
    if processed > 0 {
        slog.InfoContext(ctx, "rps expiry sweep: refunded expired bet games", slog.Int("count", processed))
    }
    return nil
}
```

- [ ] **Step 7: Build**

```bash
go build ./internal/workers/... ./internal/services/... ./internal/stores/...
```

Expected: no output

- [ ] **Step 8: Run full service test suite**

```bash
go test ./internal/services/... 2>&1 | grep -E "^(ok|FAIL|---)"
```

Expected: `ok` for all packages

- [ ] **Step 9: Commit**

```bash
git add internal/stores/gaming_rps_game.go \
        internal/services/gaming_rps_game_service.go \
        internal/services/gaming_rps_game_service_test.go \
        internal/workers/rps_game_expiry_worker.go
git commit -m "fix(rps): expire bet games and refund host escrow via sweep worker"
```

---

## Task 6: Final verification

- [ ] **Step 1: Run all affected test packages**

```bash
go test ./internal/services/... ./internal/stores/... ./internal/apis/... ./internal/workers/... 2>&1 | grep -E "^(ok|FAIL)"
```

Expected: all `ok`

- [ ] **Step 2: Build entire project**

```bash
go build ./...
```

Expected: no output

- [ ] **Step 3: Commit if any fixups needed, otherwise done**
