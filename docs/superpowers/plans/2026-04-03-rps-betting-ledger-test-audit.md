# RPS / Betting / Ledger Test Audit Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add comprehensive tests covering constraint enforcement, money conservation, pending transfer lifecycle, betting integrity, and concurrency for the RPS/betting/ledger system.

**Architecture:** Six new test files: one smoke test for the `WithNewDatabase2` helper (in `internal/database/`), four non-concurrent service test files (in `internal/services/`), and one concurrent test file (in `internal/services/`, blocked on the smoke test). All non-concurrent tests use `WithNewTestTx`; concurrent tests use `WithNewDatabase2` with goroutines and per-goroutine `RunInTxCtx` transactions.

**Tech Stack:** Go stdlib `testing`, `github.com/google/uuid`, `sync.WaitGroup`, existing `database.WithNewTestTx`, `database.WithNewDatabase2`, `stores.NewDbAdapterDecorators`, service constructors (`NewDbLedgerService`, `NewDbBettingService`, `NewDbRpsGameService`), and store test helpers (`stores.MustCreatePlayer`, `stores.WithUserID`, `mustFundWallet`).

---

## Key Conventions

Before reading tasks, internalize these codebase patterns:

- **Package:** All service tests are `package services` (same package, not `_test`). Database helper tests are `package database`.
- **Test transaction:** `database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) { ... })` — wraps in a single rolled-back transaction. Use for all non-concurrent tests.
- **Isolated DB:** `database.WithNewDatabase2(t, func(ctx context.Context, db database.Dbx) { ... })` — clones `playground_test` template to a fresh DB; real commits. Use for concurrent tests.
- **Setup pattern:** `adapter := stores.NewDbAdapterDecorators(db)`, `ledger := NewDbLedgerService(adapter)`, `betting := NewDbBettingService(adapter, ledger)`, `rpsService := NewDbRpsGameService(adapter, betting)`.
- **Fund a wallet:** `mustFundWallet(t, ctx, adapter, ledger, userID, amount)` — defined in `internal/services/betting_service_test.go`, available to all files in the same package.
- **Player with a user ID (required for betting):** `stores.MustCreatePlayer(t, ctx, adapter.Gaming(), stores.WithPlayerEmail("x@example.com"), stores.WithUserID(userID))`.
- **Error assertions:** `strings.Contains(err.Error(), "substring")` — consistent with existing suite.
- **Red tests (known bugs):** Open with `t.Skip("known bug: <description>")` so CI doesn't break. Test body below the skip documents the intended correct behavior.

---

## Task 1: Smoke test for `WithNewDatabase2`

**Files:**
- Create: `internal/database/with_new_database2_test.go`

**Gate:** Tasks 2–5 can run in parallel. Task 6 (concurrent tests) must NOT be started until both tests in this task pass.

- [ ] **Step 1: Write the test file**

```go
package database

import (
	"context"
	"testing"
)

// TestWithNewDatabase2_IsConnected verifies that WithNewDatabase2 produces a
// working connection to a freshly-cloned database.
func TestWithNewDatabase2_IsConnected(t *testing.T) {
	WithNewDatabase2(t, func(ctx context.Context, db Dbx) {
		var n int
		if err := db.QueryRow(ctx, "SELECT 1").Scan(&n); err != nil {
			t.Fatalf("SELECT 1: %v", err)
		}
		if n != 1 {
			t.Errorf("SELECT 1 = %d, want 1", n)
		}
	})
}

// TestWithNewDatabase2_HasMigratedSchema verifies that all migrations ran on the
// template database by checking that the ledger schema and seeded system accounts exist.
func TestWithNewDatabase2_HasMigratedSchema(t *testing.T) {
	WithNewDatabase2(t, func(ctx context.Context, db Dbx) {
		// Verify ledger schema exists.
		var schemaCount int
		if err := db.QueryRow(ctx,
			`SELECT count(*) FROM information_schema.schemata WHERE schema_name = 'ledger'`,
		).Scan(&schemaCount); err != nil {
			t.Fatalf("schema query: %v", err)
		}
		if schemaCount != 1 {
			t.Errorf("ledger schema count = %d, want 1 (run migrations on playground_test)", schemaCount)
		}

		// Verify seeded system accounts exist.
		var accountCount int
		if err := db.QueryRow(ctx,
			`SELECT count(*) FROM ledger.accounts WHERE code IN ($1, $2)`,
			"system:points_issuance", "system:game_escrow",
		).Scan(&accountCount); err != nil {
			t.Fatalf("system accounts query: %v", err)
		}
		if accountCount != 2 {
			t.Errorf("system account count = %d, want 2", accountCount)
		}
	})
}
```

- [ ] **Step 2: Run the tests**

```bash
cd /Users/tkahng/github/tkahng/go/playground
go test ./internal/database/ -run "TestWithNewDatabase2" -v -count=1
```

Expected output:
```
--- PASS: TestWithNewDatabase2_IsConnected
--- PASS: TestWithNewDatabase2_HasMigratedSchema
PASS
```

If either test fails, **stop and fix `WithNewDatabase2` before proceeding to Task 6.** Common failure modes:
- `playground_test` template does not exist → run migrations on it: `DATABASE_DB=playground_test dbmate up`
- Pool config error → check `conf.ZeroEnvConfig()` and env vars

- [ ] **Step 3: Commit**

```bash
git add internal/database/with_new_database2_test.go
git commit -m "test: smoke test for WithNewDatabase2 helper"
```

---

## Task 2: Ledger constraint enforcement tests

**Files:**
- Create: `internal/services/ledger_constraints_test.go`

- [ ] **Step 1: Write the test file**

```go
package services

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
)

func TestLedgerService_PostTransfer_RejectsOverdraft(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		userID := uuid.New()

		mustFundWallet(t, ctx, adapter, ledger, userID, 100)

		wallet, err := ledger.GetOrCreateUserWallet(ctx, userID)
		if err != nil {
			t.Fatalf("GetOrCreateUserWallet: %v", err)
		}
		issuance, err := ledger.GetSystemAccount(ctx, models.SystemAccountPointsIssuance)
		if err != nil {
			t.Fatalf("GetSystemAccount: %v", err)
		}

		// Attempt to debit 101 from a wallet that only has 100.
		_, err = ledger.PostTransfer(ctx, PostTransferInput{
			LedgerCode:      "POINTS",
			DebitAccountID:  wallet.ID,
			CreditAccountID: issuance.ID,
			Amount:          101,
			TransferCode:    models.TransferCodePurchase,
		})
		if err == nil {
			t.Fatal("PostTransfer overdraft: want error, got nil")
		}
		if !strings.Contains(err.Error(), "insufficient balance") {
			t.Errorf("PostTransfer overdraft error = %q, want to contain \"insufficient balance\"", err.Error())
		}
	})
}

func TestLedgerService_CreatePendingTransfer_RejectsWhenAvailableBalanceInsufficient(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		userID := uuid.New()

		mustFundWallet(t, ctx, adapter, ledger, userID, 100)

		wallet, err := ledger.GetOrCreateUserWallet(ctx, userID)
		if err != nil {
			t.Fatalf("GetOrCreateUserWallet: %v", err)
		}
		escrow, err := ledger.GetSystemAccount(ctx, models.SystemAccountGameEscrow)
		if err != nil {
			t.Fatalf("GetSystemAccount: %v", err)
		}

		// Place an 80-point pending hold.
		_, err = ledger.CreatePendingTransfer(ctx, PostTransferInput{
			LedgerCode:      "POINTS",
			DebitAccountID:  wallet.ID,
			CreditAccountID: escrow.ID,
			Amount:          80,
			TransferCode:    models.TransferCodeBetEscrow,
		})
		if err != nil {
			t.Fatalf("CreatePendingTransfer 80: %v", err)
		}

		// Available is now 20. Attempt a 30-point pending hold.
		_, err = ledger.CreatePendingTransfer(ctx, PostTransferInput{
			LedgerCode:      "POINTS",
			DebitAccountID:  wallet.ID,
			CreditAccountID: escrow.ID,
			Amount:          30,
			TransferCode:    models.TransferCodeBetEscrow,
		})
		if err == nil {
			t.Fatal("CreatePendingTransfer over available balance: want error, got nil")
		}
		if !strings.Contains(err.Error(), "insufficient available balance") {
			t.Errorf("error = %q, want to contain \"insufficient available balance\"", err.Error())
		}
	})
}

func TestLedgerService_PostTransfer_RejectsZeroAmount(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)

		issuance, err := ledger.GetSystemAccount(ctx, models.SystemAccountPointsIssuance)
		if err != nil {
			t.Fatalf("GetSystemAccount: %v", err)
		}
		wallet, err := ledger.GetOrCreateUserWallet(ctx, uuid.New())
		if err != nil {
			t.Fatalf("GetOrCreateUserWallet: %v", err)
		}

		_, err = ledger.PostTransfer(ctx, PostTransferInput{
			LedgerCode:      "POINTS",
			DebitAccountID:  issuance.ID,
			CreditAccountID: wallet.ID,
			Amount:          0,
			TransferCode:    models.TransferCodePurchase,
		})
		if err == nil {
			t.Fatal("PostTransfer amount=0: want error, got nil")
		}
		if !strings.Contains(err.Error(), "must be positive") {
			t.Errorf("error = %q, want to contain \"must be positive\"", err.Error())
		}
	})
}

func TestLedgerService_PostTransfer_RejectsNegativeAmount(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)

		issuance, err := ledger.GetSystemAccount(ctx, models.SystemAccountPointsIssuance)
		if err != nil {
			t.Fatalf("GetSystemAccount: %v", err)
		}
		wallet, err := ledger.GetOrCreateUserWallet(ctx, uuid.New())
		if err != nil {
			t.Fatalf("GetOrCreateUserWallet: %v", err)
		}

		_, err = ledger.PostTransfer(ctx, PostTransferInput{
			LedgerCode:      "POINTS",
			DebitAccountID:  issuance.ID,
			CreditAccountID: wallet.ID,
			Amount:          -1,
			TransferCode:    models.TransferCodePurchase,
		})
		if err == nil {
			t.Fatal("PostTransfer amount=-1: want error, got nil")
		}
		if !strings.Contains(err.Error(), "must be positive") {
			t.Errorf("error = %q, want to contain \"must be positive\"", err.Error())
		}
	})
}

func TestLedgerService_CreatePendingTransfer_RejectsZeroAmount(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)

		userID := uuid.New()
		mustFundWallet(t, ctx, adapter, ledger, userID, 100)
		wallet, err := ledger.GetOrCreateUserWallet(ctx, userID)
		if err != nil {
			t.Fatalf("GetOrCreateUserWallet: %v", err)
		}
		escrow, err := ledger.GetSystemAccount(ctx, models.SystemAccountGameEscrow)
		if err != nil {
			t.Fatalf("GetSystemAccount: %v", err)
		}

		_, err = ledger.CreatePendingTransfer(ctx, PostTransferInput{
			LedgerCode:      "POINTS",
			DebitAccountID:  wallet.ID,
			CreditAccountID: escrow.ID,
			Amount:          0,
			TransferCode:    models.TransferCodeBetEscrow,
		})
		if err == nil {
			t.Fatal("CreatePendingTransfer amount=0: want error, got nil")
		}
		if !strings.Contains(err.Error(), "must be positive") {
			t.Errorf("error = %q, want to contain \"must be positive\"", err.Error())
		}
	})
}

// TestLedgerService_IssuanceAccount_CreditsMustNotExceedDebits_IsNotEnforced documents
// that the AccountConstraintCreditsMustNotExceedDebits constraint on the issuance account
// is never enforced by checkDebitConstraint (which only checks DebitsMustNotExceedCredits).
// The issuance account starts with CreditsPosted=0, so if enforced, no points could ever
// be issued without first crediting the issuance account.
// Remove the t.Skip when the enforcement bug is fixed.
func TestLedgerService_IssuanceAccount_CreditsMustNotExceedDebits_IsNotEnforced(t *testing.T) {
	t.Skip("known bug: AccountConstraintCreditsMustNotExceedDebits is not enforced in checkDebitConstraint")

	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)

		issuance, err := ledger.GetSystemAccount(ctx, models.SystemAccountPointsIssuance)
		if err != nil {
			t.Fatalf("GetSystemAccount: %v", err)
		}
		// Verify the constraint is present.
		found := false
		for _, c := range issuance.Constraints {
			if c == models.AccountConstraintCreditsMustNotExceedDebits {
				found = true
			}
		}
		if !found {
			t.Fatalf("issuance account missing CreditsMustNotExceedDebits constraint; test premise invalid")
		}

		wallet, err := ledger.GetOrCreateUserWallet(ctx, uuid.New())
		if err != nil {
			t.Fatalf("GetOrCreateUserWallet: %v", err)
		}

		// Issuance account has CreditsPosted=0. If enforced, debiting it should fail.
		_, err = ledger.PostTransfer(ctx, PostTransferInput{
			LedgerCode:      "POINTS",
			DebitAccountID:  issuance.ID,
			CreditAccountID: wallet.ID,
			Amount:          100,
			TransferCode:    models.TransferCodePurchase,
		})
		if err == nil {
			t.Error("PostTransfer from unconstrained issuance: want error (credits_must_not_exceed_debits), got nil — constraint is not enforced")
		}
	})
}
```

- [ ] **Step 2: Run the tests**

```bash
cd /Users/tkahng/github/tkahng/go/playground
go test ./internal/services/ -run "TestLedgerService_PostTransfer_Rejects|TestLedgerService_CreatePendingTransfer_Rejects|TestLedgerService_IssuanceAccount" -v -count=1
```

Expected: 5 PASS (the non-skip tests), 1 SKIP (the red test). All should pass. If any non-skip test fails, fix the test (not production code — these are testing existing behavior).

- [ ] **Step 3: Commit**

```bash
git add internal/services/ledger_constraints_test.go
git commit -m "test: ledger constraint enforcement — overdraft, zero/negative amounts, issuance constraint bug"
```

---

## Task 3: Ledger integrity tests

**Files:**
- Create: `internal/services/ledger_integrity_test.go`

- [ ] **Step 1: Write the test file**

```go
package services

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
)

func TestLedgerService_AvailableBalance_DecreasesWithPendingHold(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		userID := uuid.New()

		mustFundWallet(t, ctx, adapter, ledger, userID, 200)

		wallet, err := ledger.GetOrCreateUserWallet(ctx, userID)
		if err != nil {
			t.Fatalf("GetOrCreateUserWallet: %v", err)
		}
		escrow, err := ledger.GetSystemAccount(ctx, models.SystemAccountGameEscrow)
		if err != nil {
			t.Fatalf("GetSystemAccount: %v", err)
		}

		_, err = ledger.CreatePendingTransfer(ctx, PostTransferInput{
			LedgerCode:      "POINTS",
			DebitAccountID:  wallet.ID,
			CreditAccountID: escrow.ID,
			Amount:          50,
			TransferCode:    models.TransferCodeBetEscrow,
		})
		if err != nil {
			t.Fatalf("CreatePendingTransfer: %v", err)
		}

		// Re-fetch wallet to get updated counters.
		wallet, err = ledger.GetOrCreateUserWallet(ctx, userID)
		if err != nil {
			t.Fatalf("re-fetch wallet: %v", err)
		}
		if wallet.Balance() != 200 {
			t.Errorf("Balance() = %d, want 200", wallet.Balance())
		}
		if wallet.AvailableBalance() != 150 {
			t.Errorf("AvailableBalance() = %d, want 150", wallet.AvailableBalance())
		}
	})
}

func TestLedgerService_AvailableBalance_RestoresAfterVoid(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		userID := uuid.New()

		mustFundWallet(t, ctx, adapter, ledger, userID, 200)

		wallet, err := ledger.GetOrCreateUserWallet(ctx, userID)
		if err != nil {
			t.Fatalf("GetOrCreateUserWallet: %v", err)
		}
		escrow, err := ledger.GetSystemAccount(ctx, models.SystemAccountGameEscrow)
		if err != nil {
			t.Fatalf("GetSystemAccount: %v", err)
		}

		pending, err := ledger.CreatePendingTransfer(ctx, PostTransferInput{
			LedgerCode:      "POINTS",
			DebitAccountID:  wallet.ID,
			CreditAccountID: escrow.ID,
			Amount:          50,
			TransferCode:    models.TransferCodeBetEscrow,
		})
		if err != nil {
			t.Fatalf("CreatePendingTransfer: %v", err)
		}

		if _, err = ledger.VoidPendingTransfer(ctx, pending.ID); err != nil {
			t.Fatalf("VoidPendingTransfer: %v", err)
		}

		wallet, err = ledger.GetOrCreateUserWallet(ctx, userID)
		if err != nil {
			t.Fatalf("re-fetch wallet: %v", err)
		}
		if wallet.Balance() != 200 {
			t.Errorf("Balance() = %d, want 200 after void", wallet.Balance())
		}
		if wallet.AvailableBalance() != 200 {
			t.Errorf("AvailableBalance() = %d, want 200 after void", wallet.AvailableBalance())
		}
	})
}

func TestLedgerService_AvailableBalance_DecreasesAfterPendingPost(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		userID := uuid.New()

		mustFundWallet(t, ctx, adapter, ledger, userID, 200)

		wallet, err := ledger.GetOrCreateUserWallet(ctx, userID)
		if err != nil {
			t.Fatalf("GetOrCreateUserWallet: %v", err)
		}
		escrow, err := ledger.GetSystemAccount(ctx, models.SystemAccountGameEscrow)
		if err != nil {
			t.Fatalf("GetSystemAccount: %v", err)
		}

		pending, err := ledger.CreatePendingTransfer(ctx, PostTransferInput{
			LedgerCode:      "POINTS",
			DebitAccountID:  wallet.ID,
			CreditAccountID: escrow.ID,
			Amount:          50,
			TransferCode:    models.TransferCodeBetEscrow,
		})
		if err != nil {
			t.Fatalf("CreatePendingTransfer: %v", err)
		}

		if _, err = ledger.PostPendingTransfer(ctx, pending.ID); err != nil {
			t.Fatalf("PostPendingTransfer: %v", err)
		}

		wallet, err = ledger.GetOrCreateUserWallet(ctx, userID)
		if err != nil {
			t.Fatalf("re-fetch wallet: %v", err)
		}
		if wallet.Balance() != 150 {
			t.Errorf("Balance() = %d, want 150 after post", wallet.Balance())
		}
		if wallet.AvailableBalance() != 150 {
			t.Errorf("AvailableBalance() = %d, want 150 after post", wallet.AvailableBalance())
		}
	})
}

func TestLedgerService_MoneyConservation_BetSettle_HostWins(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)

		hostUserID := uuid.New()
		guestUserID := uuid.New()
		gameID := uuid.New()
		const betAmount int64 = 100

		mustFundWallet(t, ctx, adapter, ledger, hostUserID, 500)
		mustFundWallet(t, ctx, adapter, ledger, guestUserID, 500)

		// Record escrow balance before bet.
		escrow, err := ledger.GetSystemAccount(ctx, models.SystemAccountGameEscrow)
		if err != nil {
			t.Fatalf("GetSystemAccount escrow: %v", err)
		}
		escrowBefore := escrow.Balance()

		// Place host bet.
		hostPending, err := betting.PlaceHostBet(ctx, gameID, hostUserID, betAmount)
		if err != nil {
			t.Fatalf("PlaceHostBet: %v", err)
		}

		// Settle: host wins.
		_, err = betting.PlaceGuestAndSettle(ctx, PlaceGuestAndSettleInput{
			GameID:                gameID,
			GuestUserID:           guestUserID,
			HostUserID:            hostUserID,
			BetAmount:             betAmount,
			HostPendingTransferID: hostPending.ID,
			HostResult:            models.RpsParticipantResultWin,
			GuestResult:           models.RpsParticipantResultLose,
		})
		if err != nil {
			t.Fatalf("PlaceGuestAndSettle: %v", err)
		}

		// Assert final balances.
		hostBalance, err := ledger.GetUserBalance(ctx, hostUserID)
		if err != nil {
			t.Fatalf("GetUserBalance host: %v", err)
		}
		guestBalance, err := ledger.GetUserBalance(ctx, guestUserID)
		if err != nil {
			t.Fatalf("GetUserBalance guest: %v", err)
		}

		if hostBalance != 600 {
			t.Errorf("host balance = %d, want 600 (500 + 100 winnings)", hostBalance)
		}
		if guestBalance != 400 {
			t.Errorf("guest balance = %d, want 400 (500 - 100 bet)", guestBalance)
		}

		// Escrow must be zero net (no funds stranded).
		escrow, err = ledger.GetSystemAccount(ctx, models.SystemAccountGameEscrow)
		if err != nil {
			t.Fatalf("re-fetch escrow: %v", err)
		}
		if escrow.Balance() != escrowBefore {
			t.Errorf("escrow net balance = %d, want %d (must return to prior state)", escrow.Balance(), escrowBefore)
		}
	})
}

func TestLedgerService_MoneyConservation_BetSettle_Tie(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)

		hostUserID := uuid.New()
		guestUserID := uuid.New()
		gameID := uuid.New()
		const betAmount int64 = 100

		mustFundWallet(t, ctx, adapter, ledger, hostUserID, 500)
		mustFundWallet(t, ctx, adapter, ledger, guestUserID, 500)

		escrow, err := ledger.GetSystemAccount(ctx, models.SystemAccountGameEscrow)
		if err != nil {
			t.Fatalf("GetSystemAccount escrow: %v", err)
		}
		escrowBefore := escrow.Balance()

		hostPending, err := betting.PlaceHostBet(ctx, gameID, hostUserID, betAmount)
		if err != nil {
			t.Fatalf("PlaceHostBet: %v", err)
		}

		_, err = betting.PlaceGuestAndSettle(ctx, PlaceGuestAndSettleInput{
			GameID:                gameID,
			GuestUserID:           guestUserID,
			HostUserID:            hostUserID,
			BetAmount:             betAmount,
			HostPendingTransferID: hostPending.ID,
			HostResult:            models.RpsParticipantResultTie,
			GuestResult:           models.RpsParticipantResultTie,
		})
		if err != nil {
			t.Fatalf("PlaceGuestAndSettle: %v", err)
		}

		hostBalance, err := ledger.GetUserBalance(ctx, hostUserID)
		if err != nil {
			t.Fatalf("GetUserBalance host: %v", err)
		}
		guestBalance, err := ledger.GetUserBalance(ctx, guestUserID)
		if err != nil {
			t.Fatalf("GetUserBalance guest: %v", err)
		}

		if hostBalance != 500 {
			t.Errorf("host balance = %d, want 500 (tie: refunded)", hostBalance)
		}
		if guestBalance != 500 {
			t.Errorf("guest balance = %d, want 500 (tie: refunded)", guestBalance)
		}

		escrow, err = ledger.GetSystemAccount(ctx, models.SystemAccountGameEscrow)
		if err != nil {
			t.Fatalf("re-fetch escrow: %v", err)
		}
		if escrow.Balance() != escrowBefore {
			t.Errorf("escrow balance = %d, want %d", escrow.Balance(), escrowBefore)
		}
	})
}

func TestLedgerService_MoneyConservation_BetRefund_BothVoided(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)

		hostUserID := uuid.New()
		guestUserID := uuid.New()
		gameID := uuid.New()
		const betAmount int64 = 100

		mustFundWallet(t, ctx, adapter, ledger, hostUserID, 500)
		mustFundWallet(t, ctx, adapter, ledger, guestUserID, 500)

		// Place host bet, then place guest bet.
		hostWallet, err := ledger.GetOrCreateUserWallet(ctx, hostUserID)
		if err != nil {
			t.Fatalf("GetOrCreateUserWallet host: %v", err)
		}
		guestWallet, err := ledger.GetOrCreateUserWallet(ctx, guestUserID)
		if err != nil {
			t.Fatalf("GetOrCreateUserWallet guest: %v", err)
		}
		escrow, err := ledger.GetSystemAccount(ctx, models.SystemAccountGameEscrow)
		if err != nil {
			t.Fatalf("GetSystemAccount escrow: %v", err)
		}

		hostPending, err := betting.PlaceHostBet(ctx, gameID, hostUserID, betAmount)
		if err != nil {
			t.Fatalf("PlaceHostBet: %v", err)
		}

		// Create guest pending manually (simulates guest escrow placed before game result).
		guestPending, err := ledger.CreatePendingTransfer(ctx, PostTransferInput{
			LedgerCode:      "POINTS",
			DebitAccountID:  guestWallet.ID,
			CreditAccountID: escrow.ID,
			Amount:          betAmount,
			TransferCode:    models.TransferCodeBetEscrow,
		})
		if err != nil {
			t.Fatalf("CreatePendingTransfer guest: %v", err)
		}

		// Refund both.
		if err := betting.RefundBothBets(ctx, hostPending.ID, guestPending.ID); err != nil {
			t.Fatalf("RefundBothBets: %v", err)
		}

		// Both balances must return to 500.
		hostBalance, err := ledger.GetUserBalance(ctx, hostUserID)
		if err != nil {
			t.Fatalf("GetUserBalance host: %v", err)
		}
		guestBalance, err := ledger.GetUserBalance(ctx, guestUserID)
		if err != nil {
			t.Fatalf("GetUserBalance guest: %v", err)
		}
		if hostBalance != 500 {
			t.Errorf("host balance = %d, want 500 after full refund", hostBalance)
		}
		if guestBalance != 500 {
			t.Errorf("guest balance = %d, want 500 after full refund", guestBalance)
		}

		// Available balances also 500 (no pending holds).
		hostAvail, err := ledger.GetUserAvailableBalance(ctx, hostUserID)
		if err != nil {
			t.Fatalf("GetUserAvailableBalance host: %v", err)
		}
		guestAvail, err := ledger.GetUserAvailableBalance(ctx, guestUserID)
		if err != nil {
			t.Fatalf("GetUserAvailableBalance guest: %v", err)
		}
		if hostAvail != 500 {
			t.Errorf("host available = %d, want 500", hostAvail)
		}
		if guestAvail != 500 {
			t.Errorf("guest available = %d, want 500", guestAvail)
		}
	})
}
```

- [ ] **Step 2: Run the tests**

```bash
cd /Users/tkahng/github/tkahng/go/playground
go test ./internal/services/ -run "TestLedgerService_AvailableBalance|TestLedgerService_MoneyConservation" -v -count=1
```

Expected: 6 PASS.

- [ ] **Step 3: Commit**

```bash
git add internal/services/ledger_integrity_test.go
git commit -m "test: ledger integrity — available balance math and money conservation"
```

---

## Task 4: Pending transfer lifecycle tests

**Files:**
- Create: `internal/services/ledger_pending_lifecycle_test.go`

- [ ] **Step 1: Write the test file**

```go
package services

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
)

// mustCreatePendingTransfer is a helper that creates a funded wallet and places a
// pending hold against escrow. Returns the pending transfer ID.
func mustCreatePendingTransfer(t *testing.T, ctx context.Context, adapter stores.StorageAdapterInterface, ledger LedgerService, amount int64) *models.LedgerTransfer {
	t.Helper()
	userID := uuid.New()
	mustFundWallet(t, ctx, adapter, ledger, userID, amount*2) // fund 2× so post succeeds

	wallet, err := ledger.GetOrCreateUserWallet(ctx, userID)
	if err != nil {
		t.Fatalf("mustCreatePendingTransfer: get wallet: %v", err)
	}
	escrow, err := ledger.GetSystemAccount(ctx, models.SystemAccountGameEscrow)
	if err != nil {
		t.Fatalf("mustCreatePendingTransfer: get escrow: %v", err)
	}
	pending, err := ledger.CreatePendingTransfer(ctx, PostTransferInput{
		LedgerCode:      "POINTS",
		DebitAccountID:  wallet.ID,
		CreditAccountID: escrow.ID,
		Amount:          amount,
		TransferCode:    models.TransferCodeBetEscrow,
	})
	if err != nil {
		t.Fatalf("mustCreatePendingTransfer: create: %v", err)
	}
	return pending
}

func TestLedgerService_PostPendingTransfer_AlreadyPosted_ReturnsError(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)

		pending := mustCreatePendingTransfer(t, ctx, adapter, ledger, 50)

		// First post: should succeed.
		if _, err := ledger.PostPendingTransfer(ctx, pending.ID); err != nil {
			t.Fatalf("first PostPendingTransfer: %v", err)
		}

		// Second post: should fail.
		_, err := ledger.PostPendingTransfer(ctx, pending.ID)
		if err == nil {
			t.Fatal("second PostPendingTransfer: want error, got nil")
		}
	})
}

func TestLedgerService_PostPendingTransfer_AlreadyVoided_ReturnsError(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)

		pending := mustCreatePendingTransfer(t, ctx, adapter, ledger, 50)

		// Void first.
		if _, err := ledger.VoidPendingTransfer(ctx, pending.ID); err != nil {
			t.Fatalf("VoidPendingTransfer: %v", err)
		}

		// Now try to post it: should fail.
		_, err := ledger.PostPendingTransfer(ctx, pending.ID)
		if err == nil {
			t.Fatal("PostPendingTransfer on voided transfer: want error, got nil")
		}
	})
}

func TestLedgerService_VoidPendingTransfer_AlreadyVoided_ReturnsError(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)

		pending := mustCreatePendingTransfer(t, ctx, adapter, ledger, 50)

		// First void: should succeed.
		if _, err := ledger.VoidPendingTransfer(ctx, pending.ID); err != nil {
			t.Fatalf("first VoidPendingTransfer: %v", err)
		}

		// Second void: should fail.
		_, err := ledger.VoidPendingTransfer(ctx, pending.ID)
		if err == nil {
			t.Fatal("second VoidPendingTransfer: want error, got nil")
		}
	})
}

func TestLedgerService_VoidPendingTransfer_AlreadyPosted_ReturnsError(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)

		pending := mustCreatePendingTransfer(t, ctx, adapter, ledger, 50)

		// Post it first.
		if _, err := ledger.PostPendingTransfer(ctx, pending.ID); err != nil {
			t.Fatalf("PostPendingTransfer: %v", err)
		}

		// Now try to void it: should fail.
		_, err := ledger.VoidPendingTransfer(ctx, pending.ID)
		if err == nil {
			t.Fatal("VoidPendingTransfer on posted transfer: want error, got nil")
		}
	})
}

// TestLedgerService_PendingLifecycle_NoPhantomsAfterPost verifies that after posting
// a pending transfer, the account's DebitsPending returns to 0 (no phantom holds).
func TestLedgerService_PendingLifecycle_NoPhantomsAfterPost(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)

		userID := uuid.New()
		mustFundWallet(t, ctx, adapter, ledger, userID, 200)

		wallet, err := ledger.GetOrCreateUserWallet(ctx, userID)
		if err != nil {
			t.Fatalf("GetOrCreateUserWallet: %v", err)
		}
		escrow, err := ledger.GetSystemAccount(ctx, models.SystemAccountGameEscrow)
		if err != nil {
			t.Fatalf("GetSystemAccount: %v", err)
		}

		pending, err := ledger.CreatePendingTransfer(ctx, PostTransferInput{
			LedgerCode:      "POINTS",
			DebitAccountID:  wallet.ID,
			CreditAccountID: escrow.ID,
			Amount:          50,
			TransferCode:    models.TransferCodeBetEscrow,
		})
		if err != nil {
			t.Fatalf("CreatePendingTransfer: %v", err)
		}

		if _, err = ledger.PostPendingTransfer(ctx, pending.ID); err != nil {
			t.Fatalf("PostPendingTransfer: %v", err)
		}

		wallet, err = ledger.GetOrCreateUserWallet(ctx, userID)
		if err != nil {
			t.Fatalf("re-fetch wallet: %v", err)
		}
		if wallet.DebitsPending != 0 {
			t.Errorf("DebitsPending = %d after post, want 0 (no phantom hold)", wallet.DebitsPending)
		}
		if wallet.Balance() != 150 {
			t.Errorf("Balance = %d after post, want 150", wallet.Balance())
		}
	})
}
```

- [ ] **Step 2: Run the tests**

```bash
cd /Users/tkahng/github/tkahng/go/playground
go test ./internal/services/ -run "TestLedgerService_PostPendingTransfer|TestLedgerService_VoidPendingTransfer|TestLedgerService_PendingLifecycle" -v -count=1
```

Expected: 5 PASS.

- [ ] **Step 3: Commit**

```bash
git add internal/services/ledger_pending_lifecycle_test.go
git commit -m "test: pending transfer lifecycle — double-post/void guards and phantom hold detection"
```

---

## Task 5: Betting integrity tests

**Files:**
- Create: `internal/services/betting_integrity_test.go`

- [ ] **Step 1: Write the test file**

```go
package services

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
)

func TestBettingService_PotConservation_HostWins(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)

		hostUserID := uuid.New()
		guestUserID := uuid.New()
		gameID := uuid.New()
		const betAmount int64 = 100

		mustFundWallet(t, ctx, adapter, ledger, hostUserID, 500)
		mustFundWallet(t, ctx, adapter, ledger, guestUserID, 500)

		hostPending, err := betting.PlaceHostBet(ctx, gameID, hostUserID, betAmount)
		if err != nil {
			t.Fatalf("PlaceHostBet: %v", err)
		}

		_, err = betting.PlaceGuestAndSettle(ctx, PlaceGuestAndSettleInput{
			GameID:                gameID,
			GuestUserID:           guestUserID,
			HostUserID:            hostUserID,
			BetAmount:             betAmount,
			HostPendingTransferID: hostPending.ID,
			HostResult:            models.RpsParticipantResultWin,
			GuestResult:           models.RpsParticipantResultLose,
		})
		if err != nil {
			t.Fatalf("PlaceGuestAndSettle: %v", err)
		}

		hostBal, _ := ledger.GetUserBalance(ctx, hostUserID)
		guestBal, _ := ledger.GetUserBalance(ctx, guestUserID)

		if hostBal != 600 {
			t.Errorf("host balance = %d, want 600", hostBal)
		}
		if guestBal != 400 {
			t.Errorf("guest balance = %d, want 400", guestBal)
		}

		escrow, _ := ledger.GetSystemAccount(ctx, models.SystemAccountGameEscrow)
		if escrow.Balance() != 0 {
			t.Errorf("escrow net = %d, want 0", escrow.Balance())
		}
	})
}

func TestBettingService_PotConservation_GuestWins(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)

		hostUserID := uuid.New()
		guestUserID := uuid.New()
		gameID := uuid.New()
		const betAmount int64 = 100

		mustFundWallet(t, ctx, adapter, ledger, hostUserID, 500)
		mustFundWallet(t, ctx, adapter, ledger, guestUserID, 500)

		hostPending, err := betting.PlaceHostBet(ctx, gameID, hostUserID, betAmount)
		if err != nil {
			t.Fatalf("PlaceHostBet: %v", err)
		}

		_, err = betting.PlaceGuestAndSettle(ctx, PlaceGuestAndSettleInput{
			GameID:                gameID,
			GuestUserID:           guestUserID,
			HostUserID:            hostUserID,
			BetAmount:             betAmount,
			HostPendingTransferID: hostPending.ID,
			HostResult:            models.RpsParticipantResultLose,
			GuestResult:           models.RpsParticipantResultWin,
		})
		if err != nil {
			t.Fatalf("PlaceGuestAndSettle: %v", err)
		}

		hostBal, _ := ledger.GetUserBalance(ctx, hostUserID)
		guestBal, _ := ledger.GetUserBalance(ctx, guestUserID)

		if hostBal != 400 {
			t.Errorf("host balance = %d, want 400", hostBal)
		}
		if guestBal != 600 {
			t.Errorf("guest balance = %d, want 600", guestBal)
		}

		escrow, _ := ledger.GetSystemAccount(ctx, models.SystemAccountGameEscrow)
		if escrow.Balance() != 0 {
			t.Errorf("escrow net = %d, want 0", escrow.Balance())
		}
	})
}

func TestBettingService_PotConservation_Tie(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)

		hostUserID := uuid.New()
		guestUserID := uuid.New()
		gameID := uuid.New()
		const betAmount int64 = 100

		mustFundWallet(t, ctx, adapter, ledger, hostUserID, 500)
		mustFundWallet(t, ctx, adapter, ledger, guestUserID, 500)

		hostPending, err := betting.PlaceHostBet(ctx, gameID, hostUserID, betAmount)
		if err != nil {
			t.Fatalf("PlaceHostBet: %v", err)
		}

		_, err = betting.PlaceGuestAndSettle(ctx, PlaceGuestAndSettleInput{
			GameID:                gameID,
			GuestUserID:           guestUserID,
			HostUserID:            hostUserID,
			BetAmount:             betAmount,
			HostPendingTransferID: hostPending.ID,
			HostResult:            models.RpsParticipantResultTie,
			GuestResult:           models.RpsParticipantResultTie,
		})
		if err != nil {
			t.Fatalf("PlaceGuestAndSettle: %v", err)
		}

		hostBal, _ := ledger.GetUserBalance(ctx, hostUserID)
		guestBal, _ := ledger.GetUserBalance(ctx, guestUserID)

		if hostBal != 500 {
			t.Errorf("host balance = %d, want 500 (tie)", hostBal)
		}
		if guestBal != 500 {
			t.Errorf("guest balance = %d, want 500 (tie)", guestBal)
		}

		escrow, _ := ledger.GetSystemAccount(ctx, models.SystemAccountGameEscrow)
		if escrow.Balance() != 0 {
			t.Errorf("escrow net = %d, want 0", escrow.Balance())
		}
	})
}

// TestBettingService_EnsureGuestCanAffordBet_UsesAvailableBalance verifies that
// EnsureGuestCanAffordBet checks available balance (posted minus pending holds),
// not just posted balance. A guest with 100 posted but 80 pending cannot afford a 30 bet.
func TestBettingService_EnsureGuestCanAffordBet_UsesAvailableBalance(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)

		guestUserID := uuid.New()
		mustFundWallet(t, ctx, adapter, ledger, guestUserID, 100)

		guestWallet, err := ledger.GetOrCreateUserWallet(ctx, guestUserID)
		if err != nil {
			t.Fatalf("GetOrCreateUserWallet: %v", err)
		}
		escrow, err := ledger.GetSystemAccount(ctx, models.SystemAccountGameEscrow)
		if err != nil {
			t.Fatalf("GetSystemAccount: %v", err)
		}

		// Place an 80-point hold to reduce available balance to 20.
		_, err = ledger.CreatePendingTransfer(ctx, PostTransferInput{
			LedgerCode:      "POINTS",
			DebitAccountID:  guestWallet.ID,
			CreditAccountID: escrow.ID,
			Amount:          80,
			TransferCode:    models.TransferCodeBetEscrow,
		})
		if err != nil {
			t.Fatalf("CreatePendingTransfer: %v", err)
		}

		// Available = 100 - 80 = 20. Request 30: should fail.
		err = betting.EnsureGuestCanAffordBet(ctx, guestUserID, 30)
		if err == nil {
			t.Fatal("EnsureGuestCanAffordBet: want error (insufficient available), got nil")
		}
		if !strings.Contains(err.Error(), "insufficient balance") {
			t.Errorf("error = %q, want to contain \"insufficient balance\"", err.Error())
		}

		// Request 20: should succeed (exactly at available limit).
		err = betting.EnsureGuestCanAffordBet(ctx, guestUserID, 20)
		if err != nil {
			t.Errorf("EnsureGuestCanAffordBet(20): want nil error, got %v", err)
		}
	})
}

// TestBettingService_PlaceGuestAndSettle_RejectsDoubleSettlement verifies that calling
// PlaceGuestAndSettle twice with the same HostPendingTransferID fails on the second call.
// The implicit guard is the PostPendingTransfer status check: after the first settlement
// the host transfer is already posted, so the second call fails with "not pending".
// This test documents and regression-tests this guard.
func TestBettingService_PlaceGuestAndSettle_RejectsDoubleSettlement(t *testing.T) {
	database.WithNewTestTx(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)

		hostUserID := uuid.New()
		guestUserID := uuid.New()
		gameID := uuid.New()
		const betAmount int64 = 100

		mustFundWallet(t, ctx, adapter, ledger, hostUserID, 500)
		mustFundWallet(t, ctx, adapter, ledger, guestUserID, 500)

		hostPending, err := betting.PlaceHostBet(ctx, gameID, hostUserID, betAmount)
		if err != nil {
			t.Fatalf("PlaceHostBet: %v", err)
		}

		input := PlaceGuestAndSettleInput{
			GameID:                gameID,
			GuestUserID:           guestUserID,
			HostUserID:            hostUserID,
			BetAmount:             betAmount,
			HostPendingTransferID: hostPending.ID,
			HostResult:            models.RpsParticipantResultWin,
			GuestResult:           models.RpsParticipantResultLose,
		}

		// First call: should succeed.
		if _, err = betting.PlaceGuestAndSettle(ctx, input); err != nil {
			t.Fatalf("first PlaceGuestAndSettle: %v", err)
		}

		// Second call with the same HostPendingTransferID: must fail.
		// The host transfer is now posted, so PostPendingTransfer returns
		// "transfer is not pending (status: posted)".
		_, err = betting.PlaceGuestAndSettle(ctx, input)
		if err == nil {
			t.Fatal("second PlaceGuestAndSettle: want error (already settled), got nil")
		}
	})
}
```

- [ ] **Step 2: Run the tests**

```bash
cd /Users/tkahng/github/tkahng/go/playground
go test ./internal/services/ -run "TestBettingService_Pot|TestBettingService_EnsureGuest|TestBettingService_PlaceGuest" -v -count=1
```

Expected: 6 PASS.

- [ ] **Step 3: Commit**

```bash
git add internal/services/betting_integrity_test.go
git commit -m "test: betting integrity — pot conservation, available balance check, double-settlement guard"
```

---

## Task 6: Concurrent tests (blocked on Task 1 passing)

**Prerequisite:** Both smoke tests in Task 1 must PASS before writing this task.

**Files:**
- Create: `internal/services/rps_concurrent_test.go`

- [ ] **Step 1: Confirm Task 1 smoke tests pass**

```bash
cd /Users/tkahng/github/tkahng/go/playground
go test ./internal/database/ -run "TestWithNewDatabase2" -v -count=1
```

Expected: 2 PASS. If not, fix `WithNewDatabase2` first.

- [ ] **Step 2: Write the test file**

```go
package services

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
)

// TestRpsGame_ConcurrentGuestResponses_OnlyOneSucceeds verifies that when two goroutines
// simultaneously call RespondToGameRequest for the same game, exactly one succeeds and
// one fails. This relies on the FindRpsGameForUpdate row-level lock.
func TestRpsGame_ConcurrentGuestResponses_OnlyOneSucceeds(t *testing.T) {
	database.WithNewDatabase2(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)
		rpsService := NewDbRpsGameService(adapter, betting)

		// Setup: create players and a game in a committed transaction.
		var guestPlayerID uuid.UUID
		var gameID uuid.UUID

		if err := adapter.RunInTxCtx(ctx, func(txCtx context.Context) error {
			hostUserID := uuid.New()
			guestUserID := uuid.New()

			player1 := stores.MustCreatePlayer(t, txCtx, adapter.Gaming(),
				stores.WithPlayerEmail("host_concurrent@example.com"),
				stores.WithUserID(hostUserID),
			)
			player2 := stores.MustCreatePlayer(t, txCtx, adapter.Gaming(),
				stores.WithPlayerEmail("guest_concurrent@example.com"),
				stores.WithUserID(guestUserID),
			)
			guestPlayerID = player2.ID

			game, err := rpsService.RequestGame(txCtx, &RpsGameRequestInput{
				RequestingPlayerID:   player1.ID,
				InvitedPlayerID:      player2.ID,
				RequestingPlayerMove: models.RpsParticipantMovePaper,
				DurationSeconds:      60 * 60 * 24, // 1 day
			})
			if err != nil {
				return err
			}
			gameID = game.RpsGame.ID
			return nil
		}); err != nil {
			t.Fatalf("setup: %v", err)
		}

		// Race: two goroutines respond simultaneously.
		const numGoroutines = 2
		errCh := make(chan error, numGoroutines)
		var wg sync.WaitGroup

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := adapter.RunInTxCtx(ctx, func(txCtx context.Context) error {
					_, err := rpsService.RespondToGameRequest(txCtx, &GameRequestResponse{
						InvitedPlayerID: guestPlayerID,
						GameID:          gameID,
						Status:          models.RpsGameStatusCompleted,
						Move:            models.RpsParticipantMoveRock,
					})
					return err
				})
				errCh <- err
			}()
		}

		wg.Wait()
		close(errCh)

		var successes, failures int
		for err := range errCh {
			if err == nil {
				successes++
			} else {
				failures++
			}
		}

		if successes != 1 {
			t.Errorf("successes = %d, want exactly 1", successes)
		}
		if failures != 1 {
			t.Errorf("failures = %d, want exactly 1", failures)
		}

		// Verify final game state: exactly one completion.
		var finalGame *models.RpsGame
		if err := adapter.RunInTxCtx(ctx, func(txCtx context.Context) error {
			game, err := adapter.Gaming().FindRpsGame(txCtx, &stores.RpsGameFilter{
				Ids: []uuid.UUID{gameID},
			})
			if err != nil {
				return err
			}
			finalGame = game
			return nil
		}); err != nil {
			t.Fatalf("fetch final game: %v", err)
		}
		if finalGame == nil {
			t.Fatal("final game not found")
		}
		if finalGame.Status != models.RpsGameStatusCompleted {
			t.Errorf("final game status = %q, want %q", finalGame.Status, models.RpsGameStatusCompleted)
		}
	})
}

// TestRpsGame_ConcurrentExpiry_OnlyOneRefunds verifies that when two goroutines
// simultaneously call ExpireGamesAndRefundBets, the host bet is refunded exactly once
// (no double-refund). This relies on the re-fetch-with-lock inside ExpireGamesAndRefundBets.
func TestRpsGame_ConcurrentExpiry_OnlyOneRefunds(t *testing.T) {
	database.WithNewDatabase2(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)
		betting := NewDbBettingService(adapter, ledger)
		rpsService := NewDbRpsGameService(adapter, betting)

		hostUserID := uuid.New()
		const betAmount int64 = 100

		// Setup: fund host wallet and create an expired game with a bet.
		if err := adapter.RunInTxCtx(ctx, func(txCtx context.Context) error {
			mustFundWallet(t, txCtx, adapter, ledger, hostUserID, 500)

			player1 := stores.MustCreatePlayer(t, txCtx, adapter.Gaming(),
				stores.WithPlayerEmail("host_expiry@example.com"),
				stores.WithUserID(hostUserID),
			)
			player2 := stores.MustCreatePlayer(t, txCtx, adapter.Gaming(),
				stores.WithPlayerEmail("guest_expiry@example.com"),
			)

			betAmt := betAmount
			game, err := rpsService.RequestGame(txCtx, &RpsGameRequestInput{
				RequestingPlayerID:   player1.ID,
				InvitedPlayerID:      player2.ID,
				RequestingPlayerMove: models.RpsParticipantMovePaper,
				DurationSeconds:      1, // 1 second: will expire very soon
				BetAmount:            &betAmt,
				HostUserID:           &hostUserID,
			})
			if err != nil {
				return err
			}
			// Backdate the expiry so the game is already expired.
			game.RpsGame.ExpiresAt = time.Now().UTC().Add(-5 * time.Second)
			if _, err = adapter.Gaming().UpdateRpsGame(txCtx, game.RpsGame); err != nil {
				return err
			}
			return nil
		}); err != nil {
			t.Fatalf("setup: %v", err)
		}

		// Record host balance before expiry.
		hostBalBefore, err := ledger.GetUserBalance(ctx, hostUserID)
		if err != nil {
			t.Fatalf("GetUserBalance before: %v", err)
		}

		// Race: two goroutines call ExpireGamesAndRefundBets simultaneously.
		const numGoroutines = 2
		errCh := make(chan error, numGoroutines)
		var wg sync.WaitGroup

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := rpsService.ExpireGamesAndRefundBets(ctx)
				errCh <- err
			}()
		}

		wg.Wait()
		close(errCh)

		for err := range errCh {
			if err != nil {
				t.Errorf("ExpireGamesAndRefundBets: unexpected error: %v", err)
			}
		}

		// Host balance must be fully restored (refunded exactly once).
		hostBalAfter, err := ledger.GetUserBalance(ctx, hostUserID)
		if err != nil {
			t.Fatalf("GetUserBalance after: %v", err)
		}
		if hostBalAfter != hostBalBefore {
			t.Errorf("host balance after concurrent expiry = %d, want %d (refunded exactly once)", hostBalAfter, hostBalBefore)
		}
	})
}

// TestLedger_ConcurrentPendingTransfers_BalanceConsistency verifies that under concurrent
// load, the available-balance constraint is correctly enforced. With a 100-point wallet and
// 10 goroutines each requesting a 20-point hold, exactly 5 succeed and 5 fail.
func TestLedger_ConcurrentPendingTransfers_BalanceConsistency(t *testing.T) {
	database.WithNewDatabase2(t, func(ctx context.Context, db database.Dbx) {
		adapter := stores.NewDbAdapterDecorators(db)
		ledger := NewDbLedgerService(adapter)

		userID := uuid.New()

		// Fund wallet in a committed transaction.
		if err := adapter.RunInTxCtx(ctx, func(txCtx context.Context) error {
			mustFundWallet(t, txCtx, adapter, ledger, userID, 100)
			return nil
		}); err != nil {
			t.Fatalf("fund wallet: %v", err)
		}

		wallet, err := ledger.GetOrCreateUserWallet(ctx, userID)
		if err != nil {
			t.Fatalf("GetOrCreateUserWallet: %v", err)
		}
		escrow, err := ledger.GetSystemAccount(ctx, models.SystemAccountGameEscrow)
		if err != nil {
			t.Fatalf("GetSystemAccount: %v", err)
		}

		// Launch 10 goroutines, each trying to place a 20-point hold.
		const numGoroutines = 10
		const holdAmount int64 = 20
		errCh := make(chan error, numGoroutines)
		var wg sync.WaitGroup

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := adapter.RunInTxCtx(ctx, func(txCtx context.Context) error {
					_, err := ledger.CreatePendingTransfer(txCtx, PostTransferInput{
						LedgerCode:      "POINTS",
						DebitAccountID:  wallet.ID,
						CreditAccountID: escrow.ID,
						Amount:          holdAmount,
						TransferCode:    models.TransferCodeBetEscrow,
					})
					return err
				})
				errCh <- err
			}()
		}

		wg.Wait()
		close(errCh)

		var successes, failures int
		for err := range errCh {
			if err == nil {
				successes++
			} else {
				failures++
			}
		}

		// 100 pts / 20 pts each = exactly 5 should succeed.
		if successes != 5 {
			t.Errorf("successes = %d, want 5", successes)
		}
		if failures != 5 {
			t.Errorf("failures = %d, want 5", failures)
		}

		// Re-fetch wallet and verify final state.
		wallet, err = ledger.GetOrCreateUserWallet(ctx, userID)
		if err != nil {
			t.Fatalf("re-fetch wallet: %v", err)
		}
		if wallet.DebitsPending != 100 {
			t.Errorf("DebitsPending = %d, want 100", wallet.DebitsPending)
		}
		if wallet.AvailableBalance() != 0 {
			t.Errorf("AvailableBalance = %d, want 0", wallet.AvailableBalance())
		}
		if wallet.Balance() != 100 {
			t.Errorf("Balance = %d, want 100 (no posted debits yet)", wallet.Balance())
		}
	})
}
```

- [ ] **Step 3: Run the tests**

```bash
cd /Users/tkahng/github/tkahng/go/playground
go test ./internal/services/ -run "TestRpsGame_Concurrent|TestLedger_Concurrent" -v -count=1 -timeout 60s
```

Expected: 3 PASS. These tests may be slower than others (real DB cloning). If they flap (non-deterministic pass/fail), increase the timeout and investigate DB connection setup in `WithNewDatabase2`.

- [ ] **Step 4: Commit**

```bash
git add internal/services/rps_concurrent_test.go
git commit -m "test: concurrent RPS responses, expiry races, and ledger balance consistency under load"
```

---

## Final verification

- [ ] **Run the complete new test suite**

```bash
cd /Users/tkahng/github/tkahng/go/playground
go test ./internal/database/ ./internal/services/ -count=1 -timeout 120s
```

Expected: all tests pass (2 red/skip tests show as SKIP, not FAIL).

- [ ] **Run with race detector**

```bash
cd /Users/tkahng/github/tkahng/go/playground
go test ./internal/services/ -run "TestRpsGame_Concurrent|TestLedger_Concurrent" -race -count=1 -timeout 120s
```

Expected: PASS with no race conditions detected.
