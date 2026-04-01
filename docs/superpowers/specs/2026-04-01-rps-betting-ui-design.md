# RPS Betting UI Design

**Date:** 2026-04-01  
**Branch:** feat/betting  
**Status:** Approved

## Overview

Add optional betting UI to the Rock Paper Scissors game. The backend betting system (ledger, escrow, settlement) is fully implemented. This spec covers the frontend only: surfacing bet controls when creating a game, displaying bet status in the games list, notifying guests of a bet before they respond, and showing bet outcomes on the result screen.

## Scope

No new files. No new shared components. Direct edits to 5 existing files.

### Files Modified

| File | Change summary |
|---|---|
| `ui/src/lib/rps-game-queries.tsx` | Add `getLedgerBalance` query |
| `ui/src/pages/account/rock-paper-scissors/rock-paper-scissors.tsx` | Balance badge in header; Bet column in games table |
| `ui/src/pages/account/rock-paper-scissors/create-game-dialog.tsx` | Bet toggle + amount field; pass `bet_amount` to mutations |
| `ui/src/pages/account/rock-paper-scissors/game-result.tsx` | Optional bet outcome card below moves card |
| `ui/src/pages/rock-paper-scissors/rock-paper-scissors.tsx` | Amber banner when game has a bet (public token page) |

## Data

### New query: `getLedgerBalance`

```ts
getLedgerBalance({ token }) // GET /api/ledger/balance
// Returns: { balance: number, available_balance: number }
```

Query key: `[{ key: "ledger-balance" }]`

`available_balance` is used everywhere balance is displayed (settled balance minus pending holds).

### Existing API fields used

- `RpsGame.bet_amount?: number` — set on game when host placed a bet
- `RpsGameRequestInput.bet_amount?: number` — passed when host creates game with a bet

## Components

### 1. Balance badge — `rock-paper-scissors.tsx` (authenticated page)

- `getLedgerBalance` query runs alongside the games query
- Page header gains a balance pill alongside the title: `🪙 250 pts`
  - Styled: indigo background, white text, rounded pill
  - Shows `available_balance`
  - Displays while loading as a skeleton; hidden on error (non-blocking)

### 2. Bet toggle in create-game dialog — `create-game-dialog.tsx`

The move selection step gains an optional bet section below the move cards:

- A labelled Switch: "Add a bet?" — off by default
- **Only shown when playing with a registered player** (`player` is set). The unregistered email invite endpoint (`UnregisteredPlayerInput`) does not support `bet_amount`, so the bet toggle is hidden in the email invite flow.
- When toggled on, a number input appears below:
  - Helper text: "Available balance: N pts"
  - Min: 1, Max: `available_balance`
  - Zod validation: `z.number().int().min(1).optional()`
- Only `requestGameMutation` passes `bet_amount` (registered player flow). `emailRequestMutation` is unaffected.
- Balance is fetched via `getLedgerBalance` in this component (reuses the same query key, served from cache if already fetched on the page)

### 3. Bet column — `rock-paper-scissors.tsx` (authenticated page)

A "Bet" column is added to the games data table, positioned after "Created At":

| Game state | Display |
|---|---|
| No bet (`bet_amount` is null/undefined) | `—` (muted dash) |
| Pending with bet | `N pts at stake` — amber text |
| Completed, player won | `+N pts` — green text, bold |
| Completed, player lost | `−N pts` — red text, bold |
| Completed, tie | `refunded` — muted text |
| Cancelled/expired with bet | `refunded` — muted text |

### 4. Bet outcome card — `game-result.tsx`

`GameResult` gains two new optional props:

```ts
betAmount?: number
betResult?: "win" | "lose" | "tie"
```

When `betAmount` is defined, a card is rendered below the existing moves card:

- **Win**: green background, `+N pts` large text, "Bet won"
- **Lose**: red background, `−N pts` large text, "Bet lost"  
- **Tie**: neutral background, `N pts`, "Bet refunded"

`betResult` mirrors the game result for simplicity (win → won points, lose → lost points, tie → refunded).

### 5. Guest bet banner — `rock-paper-scissors.tsx` (public token page)

When `verifyRpsGameInvite` returns a game with `bet_amount` set:

- An amber warning banner is rendered above the move cards
- Content: "This game has a N pt bet — accepting will deduct N pts from your balance"
- The move cards remain interactable below it (selecting a move = accepting the bet)
- No balance check on the frontend; backend enforces sufficient funds and returns an error if the guest cannot cover the bet

## Error handling

- Insufficient balance when creating a game: backend returns an error; displayed via existing `toast.error` in mutation `onError`
- Insufficient balance when guest accepts: backend returns an error on move submission; displayed via existing `toast.error`
- Balance query failure: balance badge is hidden; bet toggle still appears but without the helper text max constraint (backend still enforces)

## Out of scope

- Transaction history / ledger view
- Stripe purchase flow (already exists separately)
- Admin bet management
- Balance display in global nav/header
