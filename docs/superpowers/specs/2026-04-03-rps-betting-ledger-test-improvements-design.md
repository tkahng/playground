# RPS / Betting / Ledger — Test Audit Design

**Date:** 2026-04-03  
**Status:** Approved  
**Scope:** Full audit — constraints, integrity, pending lifecycle, betting pot conservation, concurrency

---

## Background

A codebase exploration identified significant gaps in test coverage across three related systems:

- **Ledger service** — double-entry accounting with pending (two-phase) transfers
- **Betting service** — escrow placement and settlement for RPS wagers
- **RPS game service** — game lifecycle including concurrent access paths

The gaps fall into five categories: constraint enforcement, money conservation, pending transfer lifecycle, betting integrity, and concurrency. This spec defines the tests to address all five.

---

## Decisions

- **Bug-exposing tests are written red.** When a test reveals a missing guard or enforcement gap, the test is written to assert the correct behavior (which currently fails). The test body opens with `t.Skip("known bug: <description>")` so CI does not break. The skip is removed when the corresponding bug is fixed. Fixing bugs is out of scope for this work.
- **Concurrency tests use `WithNewDatabase2`**, which creates an isolated Postgres DB from the `playground_test` template. This helper is unvetted; a smoke test must pass before concurrent tests are written.
- **All non-concurrent tests use `WithNewTestTx`** (single rolled-back transaction), consistent with the existing test suite.
- **All new files live in `internal/services/`**, the same package as existing service tests.

---

## Architecture

No new production code. Five new test files and one smoke test file.

```
internal/
  database/
    with_new_database2_test.go        ← smoke test for WithNewDatabase2
  services/
    ledger_constraints_test.go        ← overdraft, zero/negative amounts
    ledger_integrity_test.go          ← money conservation, available balance math
    ledger_pending_lifecycle_test.go  ← double-post, double-void, wrong-state transitions
    betting_integrity_test.go         ← pot conservation, double-settlement guard
    rps_concurrent_test.go            ← goroutine races (blocked on smoke test)
```

---

## Test Inventory

### Smoke test — `internal/database/with_new_database2_test.go`

| Test | What it verifies |
|------|-----------------|
| `TestWithNewDatabase2_CommitsAndReadsBack` | Creates a row inside the callback, reads it back in the same callback — confirms real committed connection, not a rolled-back tx |
| `TestWithNewDatabase2_IsolatedFromOtherTests` | Two parallel calls each write distinct rows; neither sees the other's data — confirms true DB isolation |

**Gate:** concurrent tests in `rps_concurrent_test.go` must not be written until both smoke tests pass. If they fail, `WithNewDatabase2` must be fixed first.

---

### `ledger_constraints_test.go`

| Test | Setup | Assert |
|------|-------|--------|
| `TestLedgerService_PostTransfer_RejectsOverdraft` | Fund wallet 100, attempt debit 101 | Returns error containing "insufficient balance" |
| `TestLedgerService_CreatePendingTransfer_RejectsWhenAvailableBalanceInsufficient` | Fund 100, create 80 pending hold, attempt 30 more pending | Returns error containing "insufficient available balance" |
| `TestLedgerService_PostTransfer_RejectsZeroAmount` | Call PostTransfer with Amount=0 | Returns error containing "must be positive" |
| `TestLedgerService_PostTransfer_RejectsNegativeAmount` | Call PostTransfer with Amount=-1 | Returns error containing "must be positive" |
| `TestLedgerService_CreatePendingTransfer_RejectsZeroAmount` | Call CreatePendingTransfer with Amount=0 | Returns error containing "must be positive" |
| `TestLedgerService_IssuanceAccount_CreditsMustNotExceedDebits_IsNotEnforced` | *(red test)* Attempt to over-issue from issuance account beyond its own debits | Currently succeeds (no guard); test asserts it should return error — will fail until enforcement is added |

---

### `ledger_integrity_test.go`

| Test | Setup | Assert |
|------|-------|--------|
| `TestLedgerService_AvailableBalance_DecreasesWithPendingHold` | Fund 200, create 50 pending hold | `AvailableBalance()=150`, `Balance()=200` |
| `TestLedgerService_AvailableBalance_RestoresAfterVoid` | Fund 200, create 50 pending, void it | `AvailableBalance()=200`, `Balance()=200` |
| `TestLedgerService_AvailableBalance_DecreasesAfterPendingPost` | Fund 200, create 50 pending, post it | `Balance()=150`, `AvailableBalance()=150` |
| `TestLedgerService_MoneyConservation_BetSettle_HostWins` | Fund both users 500, run full host-wins bet settle cycle | Winner balance +betAmount, loser balance -betAmount, escrow net=0, total system unchanged |
| `TestLedgerService_MoneyConservation_BetSettle_Tie` | Fund both users 500, run tie settle cycle | Both balances unchanged, escrow net=0 |
| `TestLedgerService_MoneyConservation_BetRefund_BothVoided` | Fund both users 500, place host bet, then void both | Both balances unchanged, escrow net=0 |

Money conservation formula checked in each test:
```
sum(all account balances) before == sum(all account balances) after
```
Accounts to sum: host wallet, guest wallet, escrow account.

---

### `ledger_pending_lifecycle_test.go`

| Test | Steps | Assert |
|------|-------|--------|
| `TestLedgerService_PostPendingTransfer_AlreadyPosted_ReturnsError` | Create pending → post → post again | Second post returns error |
| `TestLedgerService_PostPendingTransfer_AlreadyVoided_ReturnsError` | Create pending → void → post | Post returns error |
| `TestLedgerService_VoidPendingTransfer_AlreadyVoided_ReturnsError` | Create pending → void → void again | Second void returns error |
| `TestLedgerService_VoidPendingTransfer_AlreadyPosted_ReturnsError` | Create pending → post → void | Void returns error |

All four tests also assert the final account balances are self-consistent (no phantom pending amounts lingering).

---

### `betting_integrity_test.go`

| Test | Setup | Assert |
|------|-------|--------|
| `TestBettingService_PotConservation_HostWins` | Fund host+guest 500 each, betAmount=100, PlaceHostBet + PlaceGuestAndSettle(host wins) | Host balance=600, guest balance=400, escrow net=0 |
| `TestBettingService_PotConservation_GuestWins` | Same, guest wins | Host balance=400, guest balance=600, escrow net=0 |
| `TestBettingService_PotConservation_Tie` | Same, tie | Both 500, escrow net=0 |
| `TestBettingService_EnsureGuestCanAffordBet_UsesAvailableBalance` | Fund guest 100, place 80 pending hold on guest, call EnsureGuestCanAffordBet(30) | Returns error — available=20, not 100 |
| `TestBettingService_PlaceGuestAndSettle_RejectsDoubleSettlement` | *(red test)* PlaceHostBet then call PlaceGuestAndSettle twice | Second call returns error — documents missing double-settlement guard |

---

### `rps_concurrent_test.go`

**Prerequisite:** Both smoke tests in `with_new_database2_test.go` must pass.

Uses `WithNewDatabase2` for each test (real DB, real commits, goroutines can race).

| Test | Setup | Assert |
|------|-------|--------|
| `TestRpsGame_ConcurrentGuestResponses_OnlyOneSucceeds` | Create game, launch 2 goroutines both calling `RespondToGameRequest` for the same game | Exactly 1 succeeds, 1 returns error; game status is `completed` exactly once |
| `TestRpsGame_ConcurrentExpiryAndResponse_ConsistentFinalState` | Create expired game with bet, launch goroutine A (`RespondToGameRequest`) and goroutine B (`ExpireGamesAndRefundBets`) simultaneously | Game ends in exactly one terminal state (`completed` or `cancelled`); escrow account net=0 (no orphaned funds) |
| `TestLedger_ConcurrentPendingTransfers_BalanceConsistency` | Fund wallet 100, launch 10 goroutines each attempting a 20-point pending hold | Exactly 5 succeed (100÷20=5), final `debits_pending`=100, `AvailableBalance()=0`, no errors from successful goroutines |

Each concurrent test runs with `t.Parallel()` and uses `sync.WaitGroup` + error channel pattern to collect goroutine results.

---

## Error Message Conventions

Tests that assert on error strings use `strings.Contains(err.Error(), ...)` rather than `errors.Is`, consistent with the existing test suite pattern.

---

## Dependencies & Sequencing

```
smoke test passes
    └── rps_concurrent_test.go (all 3 tests)

WithNewTestTx tests (no dependency on smoke test):
    ledger_constraints_test.go
    ledger_integrity_test.go
    ledger_pending_lifecycle_test.go
    betting_integrity_test.go
```

Red tests (expected to fail until bugs are fixed):
- `TestLedgerService_IssuanceAccount_CreditsMustNotExceedDebits_IsNotEnforced`
- `TestBettingService_PlaceGuestAndSettle_RejectsDoubleSettlement`

These will be marked with a `t.Skip` + explanatory comment so CI doesn't break, and the skip can be removed when the corresponding bug is fixed.

---

## Out of Scope

- Fixing the bugs exposed by red tests (separate work)
- Stripe integration edge cases (non-UUID session IDs, different amounts on duplicate)
- Metadata edge cases (NULL, oversized JSON)
- Timeout/TTL enforcement on pending transfers
- HTTP API-layer tests (covered separately in `apis/`)
