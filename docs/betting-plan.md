# Betting Feature — Implementation Plan

## Goals

1. Double-entry ledger with pending (two-phase) transfers and overdraft protection
2. Betting service: escrow placement, guest eligibility check, settlement, refund
3. RPS game service integration: bet on create, check on accept, settle on complete, refund on expiry/cancel
4. Expiry sweep worker
5. Frontend: bet toggle in create-game dialog, balance badge, bet column, bet outcome card, guest banner, points purchase page
6. Comprehensive backend and frontend test coverage

## Files Changed

### Database / Migrations

| File | Change |
|---|---|
| `internal/database/migrations/20260228000001_add_ledger_tables.sql` | `ledger` schema, `accounts`, `transfers` tables |
| `internal/database/migrations/20260228000002_add_rps_betting.sql` | `bet_amount`, `host_bet_transfer_id`, `guest_bet_transfer_id` on `rps_games` |
| `internal/database/migrations/20260402000001_refactor_ledger_flags.sql` | Refactor transfer status flags |
| `internal/database/migrations/20260420000001_ledger_account_balance_check.sql` | DB-level `CHECK (balance >= 0)` on `ledger.accounts` |
| `internal/database/migrations/20260420000002_ledger_bet_transfer_unique.sql` | `UNIQUE (reference_id, transfer_code)` on bet escrow transfers |
| `internal/database/schema.sql` | Regenerated |

### Models

| File | Change |
|---|---|
| `internal/models/ledger.go` | `LedgerAccount`, `LedgerTransfer`, transfer status/code enums |
| `internal/models/rps_games.go` | `BetAmount`, `HostBetTransferID`, `GuestBetTransferID` fields |
| `internal/models/stripe.go` | `PointsAmount` metadata field |

### Stores

| File | Change |
|---|---|
| `internal/stores/ledger.go` | Base ledger store |
| `internal/stores/ledger_account.go` | Account CRUD, balance queries |
| `internal/stores/ledger_transfer.go` | Transfer CRUD, `CountTransfers`, `FindTransfers` with filters |
| `internal/stores/ledger_decorator.go` | Decorator wrapping ledger store |
| `internal/stores/gaming_rps_game.go` | `FindRpsGameForUpdate`, `FindExpiredPendingBetGames` |
| `internal/stores/gaming_decorator.go` | Updated decorator |
| `internal/stores/storage_adapter.go` | Ledger adapter registration |
| `internal/stores/storage_adapters_decorators.go` | Decorator wiring |
| `internal/stores/stripe_price.go` | `metadata_type` filter support |
| `internal/stores/stripe_product.go` | Product store updates |

### Services

| File | Change |
|---|---|
| `internal/services/ledger_service.go` | `PostTransfer`, `CreatePendingTransfer`, `PostPendingTransfer`, `VoidPendingTransfer`, `GetUserBalance`, `GetUserAvailableBalance` |
| `internal/services/betting_service.go` | `PlaceHostBet`, `EnsureGuestCanAffordBet`, `PlaceGuestAndSettle`, `RefundHostBet`, `RefundBothBets` |
| `internal/services/gaming_rps_game_service.go` | Bet placement on create; balance check + settlement on accept; `ExpireGamesAndRefundBets` |
| `internal/services/payment_service.go` | `CreatePointsCheckoutSession` |
| `internal/services/payment_client.go` | Points checkout client method |

### Workers

| File | Change |
|---|---|
| `internal/workers/rps_game_expiry_worker.go` | Sweep worker: finds expired pending bet games, calls `RefundBothBets`, continues on per-game error |

### APIs

| File | Change |
|---|---|
| `internal/apis/ledger_api.go` | `GET /api/ledger/balance`, `POST /api/ledger/points/checkout` |
| `internal/apis/game_rps.go` | Fix inverted token expiry check; expose `bet_amount` in response |
| `internal/apis/stripe_products.go` | `metadata_type` query param support |
| `internal/apis/stripe_webhook.go` | Return 200 for unhandled events; handle points fulfillment |
| `internal/apis/api_bind.go` | Register ledger API routes |
| `internal/apis/admin_stripe.go` | Minor update |

### Frontend

| File | Change |
|---|---|
| `ui/src/lib/rps-game-queries.tsx` | `getLedgerBalance` query |
| `ui/src/lib/api.ts` | `createPointsCheckoutSession`; `metadata_type` param on `getProductsWithPrices` |
| `ui/src/pages/account/rock-paper-scissors/rock-paper-scissors.tsx` | Balance badge; bet column in games table; query invalidation |
| `ui/src/pages/account/rock-paper-scissors/create-game-dialog.tsx` | Bet toggle, amount input, zero-balance buy-points link |
| `ui/src/pages/account/rock-paper-scissors/game-result.tsx` | Bet outcome card |
| `ui/src/pages/account/rock-paper-scissors/selected-game-dialog.tsx` | Insufficient-funds guard in `SubmitMoveView`; bet amount in `PendingGameView` |
| `ui/src/pages/account/rock-paper-scissors/move.tsx` | `disabled` prop; `children` slot |
| `ui/src/pages/rock-paper-scissors/rock-paper-scissors.tsx` | Guest bet warning banner |
| `ui/src/pages/settings/points-settings.tsx` | New — points purchase page |
| `ui/src/pages/payment/points-payment-success.tsx` | New — post-checkout success page |
| `ui/src/components/route-map.ts` | `POINTS_SETTINGS` route |
| `ui/src/components/links.tsx` | Points + Billing sidebar links |
| `ui/src/layouts/authenticated-layout-outlet.tsx` | Wallet creation on startup |
| `ui/src/App.tsx` | Points settings and payment success routes |
| `ui/src/schema.d.ts` | Regenerated from OpenAPI |

### Tests

| File | What it covers |
|---|---|
| `internal/database/with_new_database_test.go` | Smoke test for `WithNewDatabase` helper |
| `internal/services/ledger_constraints_test.go` | Overdraft, zero/negative amount rejections |
| `internal/services/ledger_integrity_test.go` | Available balance math, money conservation |
| `internal/services/ledger_pending_lifecycle_test.go` | All invalid pending state transitions |
| `internal/services/ledger_service_test.go` | Ledger service unit tests |
| `internal/services/betting_service_test.go` | Betting service unit tests |
| `internal/services/betting_integrity_test.go` | Pot conservation, double-settlement guard |
| `internal/services/betting_lifecycle_test.go` | Per-game pending-transfer count after every terminal path |
| `internal/services/betting_invariants_test.go` | Multi-game system conservation, escrow baseline, no orphan pending |
| `internal/services/gaming_rps_game_service_test.go` | Game service with betting integration |
| `internal/services/rps_concurrent_test.go` | Concurrent response races, expiry races, ledger balance consistency |
| `internal/stores/ledger_test.go` | Ledger store |
| `internal/stores/gaming_rps_game_test.go` | RPS game store |
| `internal/stores/gaming_rps_participant_test.go` | Participant store |
| `internal/workers/rps_game_expiry_worker_test.go` | Expiry sweep worker |
| `internal/apis/game_rps_test.go` | API-level RPS betting tests |
| `internal/apis/ledger_api_test.go` | Ledger API endpoints |
| `ui/src/pages/account/rock-paper-scissors/__tests__/game-result.test.tsx` | Bet outcome rendering |
| `ui/src/pages/account/rock-paper-scissors/__tests__/move.test.tsx` | Move selection behavior |
| `ui/src/pages/account/rock-paper-scissors/__tests__/create-game-dialog-bet.test.tsx` | Bet toggle and balance validation |
| `ui/src/pages/account/rock-paper-scissors/__tests__/submit-move-view.test.tsx` | Insufficient funds guard |

## Key Architectural Decisions

**Two-phase transfers for escrow.** Pending transfers deduct from `available_balance` immediately, preventing double-spend before a game settles. Posting/voiding transitions are enforced by the service layer and cannot be repeated (state machine).

**DB-level overdraft guard.** `CHECK (balance >= 0)` on `ledger.accounts` means a concurrent race that bypasses the service-layer balance check still cannot corrupt balances.

**Unique constraint prevents double-settlement.** `UNIQUE (reference_id, transfer_code)` on bet escrow transfers means calling `PlaceGuestAndSettle` twice on the same game fails at the DB level, not just the service layer.

**Row-level lock on game during response.** `FindRpsGameForUpdate` acquires `FOR UPDATE` before the guest responds, serializing concurrent accept attempts and expiry races for the same game.

**Expiry worker continues on error.** Per-game errors are logged and the sweep continues to the next game rather than aborting, preventing a single bad game from blocking all refunds.

**Frontend balance from `available_balance`.** Raw `balance` is never displayed; only `available_balance` (balance minus pending holds) is shown and used for bet validation, matching the backend's eligibility logic.

**Query cache shared across components.** All components use `[{ key: "ledger-balance" }]` as the query key, so a balance fetch on the RPS page is served from cache in the create-game dialog and points settings page.
