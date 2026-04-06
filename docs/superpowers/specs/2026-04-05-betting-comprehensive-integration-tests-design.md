# Betting Comprehensive Integration Tests — Design Spec

**Date:** 2026-04-05  
**Branch:** feat/betting  
**Scope:** Service-layer correctness — data integrity, no leaked funds, no orphaned pending transfers

---

## Context

The existing betting/ledger test suite covers ~65 tests across service, store, worker, and concurrency layers. Per-game lifecycle paths (host wins, guest wins, tie, cancel, expiry) are tested, ledger constraints and state machine transitions are covered, and concurrent serialization is verified.

Two categories of coverage are missing:

1. **Explicit pending-transfer count assertions** — existing tests check `available_balance` after terminal states, which implies pending debits are cleared. But balance-based checks do not directly verify the ledger transfer table. If a void and a new pending accidentally cancel out numerically, balance looks correct while the ledger has orphaned rows.

2. **Multi-game aggregate conservation** — no test exercises multiple games with mixed outcomes and verifies the total system: `sum(all user wallets) + escrow_balance == total_issued_points`. This is the single assertion that proves funds are never created or destroyed across any combination of paths.

---

## What Is NOT Changing

- No new production code.
- No changes to `BettingService`, `LedgerService`, `RpsGameService`, or any worker.
- No new test infrastructure helpers (reuse `WithNewTestTx`, `mustFundWallet`, `mustFundPlayerWallet`, `mustCreateUser`, `MustCreatePlayer`).

---

## New Files

Both files live in `internal/services/` (same package as all other service tests).

### `internal/services/betting_lifecycle_test.go`

Five tests that complete a full game lifecycle through `RpsGameService` and assert the explicit pending-transfer count for that game is zero at the terminal state.

**Shared assertion pattern used in all five:**

```go
pendingCount, err := ledger.CountTransfers(ctx, &stores.LedgerTransferFilter{
    ReferenceIds: []uuid.UUID{gameID},
    Statuses:     []models.LedgerTransferStatus{models.LedgerTransferStatusPending},
})
// assert pendingCount == 0
```

#### Tests

| Name | Path | Key assertions |
|------|------|----------------|
| `TestBetting_NoPendingTransfers_AfterCancel` | Host bets → guest cancels | pending count for game = 0; host available balance = starting balance |
| `TestBetting_NoPendingTransfers_AfterExpiry` | Host bets → game expires → sweep runs | pending count for game = 0; host available balance = starting balance; game status = cancelled |
| `TestBetting_NoPendingTransfers_AfterComplete_HostWins` | Full game, host wins | pending count for game = 0; host balance = start + bet; guest balance = start - bet |
| `TestBetting_NoPendingTransfers_AfterComplete_Tie` | Full game, tie | pending count for game = 0; both balances = starting balance |
| `TestBetting_GuestCanRetry_AfterInsufficientFunds` | Guest fails response → gets funded → retries | After failed attempt: game still `pending`, host escrow still held (host available = start - bet). After funded retry: game `completed`, balances correct, pending count = 0 |

---

### `internal/services/betting_invariants_test.go`

Three tests that verify system-wide invariants across multiple games.

#### `TestBettingInvariant_MultiGame_TotalSystemConservation`

The comprehensive conservation test.

**Setup:**
- 12 users (6 host/guest pairs), each funded with 1000 points
- Total issued: 12,000 points
- Record `escrowStart := escrow.Balance()`

**Games run sequentially in same transaction:**
1. Game A: host wins (100 pt bet)
2. Game B: guest wins (100 pt bet)
3. Game C: host wins (200 pt bet)
4. Game D: guest wins (200 pt bet)
5. Game E: tie (150 pt bet)
6. Game F: bet placed, expires, sweep runs (100 pt bet)

**Assertions after all games complete:**
- `sum(all 12 user balances) + escrow.Balance() == 12,000` — total system conservation
- `escrow.Balance() == escrowStart` — escrow nets to zero
- `CountTransfers(status=pending) == 0` globally across all transfers — no orphaned holds anywhere in the ledger

#### `TestBettingInvariant_EscrowNetsZeroAfterEachGame`

Verifies that escrow returns to its starting balance after each individual game, not just the aggregate.

**Setup:** 3 pairs of users, each funded with 500 points.

**For each of 3 games (run, assert, repeat):**
- Record `escrowBefore`
- Play game to terminal state (cancel / complete / expire)
- Assert `escrow.Balance() == escrowBefore` immediately after that game

This catches the case where one game's escrow leak is masked by another game's behavior in the aggregate test.

#### `TestBettingInvariant_NoOrphanPendingAfterAllPaths`

Explicitly verifies the ledger transfer table has zero pending `bet_escrow` rows after running all terminal paths.

**Setup:** 3 pairs of users.

**Paths run:**
1. Complete game (settle with winner)
2. Cancelled game (guest declines)
3. Expired game (expiry sweep)

**Final assertion:**
```go
count, _ := ledger.CountTransfers(ctx, &stores.LedgerTransferFilter{
    TransferCodes: []string{models.TransferCodeBetEscrow},
    Statuses:      []models.LedgerTransferStatus{models.LedgerTransferStatusPending},
})
// assert count == 0
```

This is the strongest possible statement: regardless of which terminal path was taken, no pending bet escrow transfer survives.

---

## Test Infrastructure Notes

All tests use `database.WithNewTestTx` — each test runs in a rolled-back transaction, so no cleanup is needed and tests are fully isolated from each other.

For expiry tests: game is created with `DurationSeconds: 1`, then `time.Sleep(2 * time.Second)`, then `rpsService.ExpireGamesAndRefundBets(ctx)`. This follows the pattern established in `TestDbRpsGameService_ExpireGamesAndRefundBets_RefundsHostEscrow`.

For the multi-game conservation test, all games are played sequentially within the same `WithNewTestTx` callback to ensure the global `CountTransfers` and balance sum see the full state.

---

## Success Criteria

1. All 8 new tests pass with no skips.
2. `TestBetting_GuestCanRetry_AfterInsufficientFunds` — validates the game remains `pending` after a failed bet response and can be completed after funding.
3. `TestBettingInvariant_MultiGame_TotalSystemConservation` — validates total conservation across all outcome types.
4. `TestBettingInvariant_NoOrphanPendingAfterAllPaths` — validates the ledger transfer table is clean after all terminal paths.
5. All existing tests continue to pass.

---

## Out of Scope

- HTTP-layer / API-level integration tests
- `PlaceHostBet` idempotency guard (production code change — separate ticket)
- Load/stress testing
- Worker scheduling integration (covered by existing `rps_game_expiry_worker_test.go`)
