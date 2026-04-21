# Betting Feature — Design Spec

## Overview

Adds an optional points-based betting system to Rock Paper Scissors. Players wager points when creating a game; the backend holds stakes in escrow and settles on completion. This document covers the full feature: ledger architecture, betting service, RPS game integration, frontend UI, points purchase, and test coverage design.

---

## Backend — Ledger & Betting

### Ledger

Double-entry accounting with two-phase (pending) transfers.

**Schema:** `ledger.accounts`, `ledger.transfers`

**Key account codes:**
- `system:points_issuance` — source of all issued points; debited on issuance
- `system:game_escrow` — holds funds during active bets

**Transfer statuses:** `pending` → `posted` or `voided`

**Available balance:** `balance − sum(pending debits)`. Used everywhere bet eligibility is checked — never raw balance.

**Invariants enforced at DB level:**
- `CHECK (amount > 0)` on `ledger.transfers`
- `CHECK (balance >= 0)` on `ledger.accounts` (overdraft guard)
- `UNIQUE (reference_id, transfer_code)` on bet escrow transfers (prevents double-settlement)

### Betting Service

| Method | Description |
|---|---|
| `PlaceHostBet(ctx, gameID, hostUserID, amount)` | Creates pending escrow debit on host; stores `HostBetTransferID` on game |
| `EnsureGuestCanAffordBet(ctx, guestUserID, amount)` | Checks guest available balance ≥ amount; returns error if not |
| `PlaceGuestAndSettle(ctx, gameID, guestUserID, result)` | Creates guest escrow debit then immediately settles: posts winner's escrow + credits winnings, voids loser's escrow; or voids both on tie |
| `RefundHostBet(ctx, hostPendingTransferID)` | Voids host's pending escrow transfer; used when guest declines or game expires before guest accepts |
| `RefundBothBets(ctx, hostPendingTransferID, guestPendingTransferID)` | Voids both pending escrow transfers; used when game expires after both bets are placed |

**Money conservation:** `sum(all wallet balances) + escrow.Balance()` is constant across all operations.

### RPS Game Integration

- Host creates game with optional `bet_amount`; `PlaceHostBet` is called immediately
- Guest response calls `EnsureGuestCanAffordBet` before accepting; then `PlaceGuestAndSettle` on completion
- Expiry worker calls `RefundBothBets` if both host and guest bets are placed (guest accepted but game not completed), or `RefundHostBet` if only the host bet exists (guest never accepted)
- Game row stores `HostBetTransferID` and `GuestBetTransferID` for idempotency

**Guards:**
- Host user ID validated against requesting player to block self-play exploitation
- Row-level lock (`FOR UPDATE`) on game row during response to prevent concurrent double-settlement
- Double-settlement rejected via unique constraint on `(reference_id, transfer_code)`

---

## Frontend — RPS Betting UI

### Balance Query

```ts
getLedgerBalance({ token }) // GET /api/ledger/balance
// Returns: { available_balance: number }
```

Query key: `[{ key: "ledger-balance" }]` — shared cache across components.

### Balance Badge (RPS Dashboard Header)

- Shows `available_balance` as `🪙 N pts` pill in the page header
- When balance = 0: renders as a link to `/account/settings/points` with label `🪙 0 pts — Buy Points` (amber text)
- Skeleton while loading; hidden on error (non-blocking)

### Bet Toggle in Create-Game Dialog

- Appears only when playing with a registered player (unregistered email invite path does not support `bet_amount`)
- Switch: "Add a bet?" — off by default
- When enabled and balance > 0: number input, min=1, max=`available_balance`, helper text shows available balance
- When enabled and balance = 0: shows inline "You have 0 pts. [Buy Points →]" link (opens in new tab)
- `bet_amount` passed to `requestGame` mutation only; `requestGameEmail` is unaffected

### Bet Column in Games Table

| State | Display |
|---|---|
| No bet | `—` muted |
| Pending with bet | `N pts at stake` amber |
| Completed, won | `+N pts` green bold |
| Completed, lost | `−N pts` red bold |
| Completed / cancelled, tie or refund | `refunded` muted |

### Bet Outcome Card (GameResult)

Props: `betAmount?: number`, `betResult?: "win" | "lose" | "tie"`

Rendered below the moves card when `betAmount` is defined:
- **Win:** green background, `+N pts`, "Bet won"
- **Lose:** red background, `−N pts`, "Bet lost"
- **Tie:** neutral background, `N pts`, "Bet refunded"

### Guest Bet Banner (Public Invite Page)

When the invite token resolves to a game with `bet_amount` set, an amber warning banner is shown above the move cards: _"This game has a N pt bet — accepting will deduct N pts from your balance."_ Frontend does not check balance; backend enforces and returns an error on move submission if insufficient.

### SubmitMoveView — Insufficient Funds Guard

When `bet_amount > 0` and `available_balance < bet_amount`:
- Warning message: "You need N pts to accept this bet but only have M pts."
- Link: "Buy points" → `/account/settings/points`
- Submit button disabled regardless of move selection

---

## Frontend — Points Purchase UI

### Wallet Creation on Startup

On app startup (inside the authenticated layout), `createLedgerWallet` is called to ensure a wallet exists for the current user before any balance queries run. This is a no-op if the wallet already exists.

### Route

`/account/settings/points` — authenticated, inside settings layout with sidebar.

### API

```ts
getProductsWithPrices(token?, metadata_type?: "subscription" | "points")
// GET /api/stripe/products?metadata_type=points

createPointsCheckoutSession(token, { price_id })
// POST /api/ledger/points/checkout
// Returns: { url: string }
```

### Page

- Three cards side-by-side showing available points packages (fetched from Stripe with `metadata_type=points`)
- Each card: package name, price in USD, point count from `metadata.points_amount`
- "Buy" → calls `createPointsCheckoutSession`, then `window.location.href = url` (same tab redirect to Stripe Checkout)
- After payment Stripe redirects to `/payment/points-success` (success page confirms purchase and links back to the game)
- Balance badge top-right (same `ledger-balance` query key, served from cache)
- Settings sidebar includes "Points" link; "Billing" link also uncommented

---

## Test Coverage Design

### Backend — Ledger Constraints (`ledger_constraints_test.go`)

| Test | Asserts |
|---|---|
| Overdraft rejected | `PostTransfer` with amount > balance returns error |
| Available balance overdraft rejected | `CreatePendingTransfer` rejected when pending hold would exceed available balance |
| Zero amount rejected | `PostTransfer` / `CreatePendingTransfer` with amount=0 returns error |
| Negative amount rejected | Same with amount=-1 |

### Backend — Ledger Integrity (`ledger_integrity_test.go`)

| Test | Asserts |
|---|---|
| Pending hold decreases available balance | `AvailableBalance` = balance − pending |
| Void restores available balance | After void: `AvailableBalance` = original balance |
| Post decreases both balances | After post: `Balance` = `AvailableBalance` = original − amount |
| Money conservation, bet settle (host wins) | Winner +bet, loser -bet, escrow net=0, total unchanged |
| Money conservation, bet settle (tie) | Both balances unchanged, escrow net=0 |
| Money conservation, refund | Both balances unchanged, escrow net=0 |

### Backend — Pending Lifecycle (`ledger_pending_lifecycle_test.go`)

All four invalid state transitions rejected:
- Post → post again
- Void → post
- Post → void
- Void → void again

### Backend — Betting Integrity (`betting_integrity_test.go`)

| Test | Asserts |
|---|---|
| Pot conservation (host wins / guest wins / tie) | Correct balance changes, escrow net=0 |
| `EnsureGuestCanAffordBet` uses available balance | Rejects when pending hold makes available < bet |
| Double-settlement rejected | Second `PlaceGuestAndSettle` on same game returns error |

### Backend — Comprehensive Integration (`betting_lifecycle_test.go`, `betting_invariants_test.go`)

**Per-game pending-transfer assertions** — after every terminal path, `CountTransfers(status=pending, reference=gameID) == 0`:
- After cancel
- After expiry (sweep runs)
- After complete, host wins
- After complete, tie
- After guest retry following insufficient funds

**System-wide invariants:**
- `TestBettingInvariant_MultiGame_TotalSystemConservation` — 6 games (all terminal paths), asserts `sum(all wallets) + escrow == total_issued` and `escrow == escrowStart`
- `TestBettingInvariant_EscrowNetsZeroAfterEachGame` — checks escrow returns to baseline after each individual game
- `TestBettingInvariant_NoOrphanPendingAfterAllPaths` — `CountTransfers(code=bet_escrow, status=pending) == 0` after running cancel + complete + expiry

### Frontend (`ui/src/pages/account/rock-paper-scissors/__tests__/`)

| File | Tests |
|---|---|
| `game-result.test.tsx` | Win/lose/tie headers; bet outcome card labels and amounts; hidden when no bet |
| `move.test.tsx` | Move selection, submit disabled until selected, disabled prop, handleSubmit callback, children slot |
| `create-game-dialog-bet.test.tsx` | Bet toggle visibility, amount input show/hide, zero-balance → buy-points link, max capped at balance |
| `submit-move-view.test.tsx` | Insufficient funds warning, disabled submit, no warning when no bet |

---

## Error Handling

| Scenario | Behavior |
|---|---|
| Insufficient balance on game creation | Backend error → `toast.error` via mutation `onError` |
| Insufficient balance on guest accept | Backend error → `toast.error` on move submission |
| Balance query failure | Badge hidden; bet toggle shows without max constraint |
| Unhandled Stripe webhook events | 200 returned (not 400) to prevent Stripe retries |
