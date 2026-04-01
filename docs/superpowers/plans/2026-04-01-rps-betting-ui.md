# RPS Betting UI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Surface the existing backend betting system in the RPS frontend: bet toggle when creating a game, balance badge on the RPS page, a Bet column in the games table, bet outcome card in results, and a warning banner for guests on the invite page.

**Architecture:** Direct edits to 6 existing files plus one small addition to `move.tsx` to support a children slot. No new files. No new shared components. All balance data flows from `GET /api/ledger/balance` via a React Query cached under `[{ key: "ledger-balance" }]`.

**Tech Stack:** React 19, TanStack React Query v5, React Hook Form, Zod, shadcn/ui (Switch, Input, Card), openapi-fetch, TypeScript

---

## File Map

| File | Change |
|---|---|
| `internal/apis/game.go` | Add `BetAmount` field to `RpsGame` response struct; map it in `toApiRpsGame` |
| `ui/src/schema.d.ts` | Regenerated from backend OpenAPI spec (do not edit manually) |
| `ui/src/lib/rps-game-queries.tsx` | Add `getLedgerBalance` and update `requestGame` to accept optional `betAmount` |
| `ui/src/pages/account/rock-paper-scissors/move.tsx` | Add optional `children` prop rendered above the submit button |
| `ui/src/pages/account/rock-paper-scissors/create-game-dialog.tsx` | Add balance query, bet toggle (Switch), bet amount input; pass `bet_amount` in `requestGameMutation` |
| `ui/src/pages/account/rock-paper-scissors/rock-paper-scissors.tsx` | Add balance query, balance badge in header, Bet column in DataTable |
| `ui/src/pages/account/rock-paper-scissors/game-result.tsx` | Add optional `betAmount`/`betResult` props; render bet outcome card |
| `ui/src/pages/account/rock-paper-scissors/selected-game-dialog.tsx` | Pass `betAmount`/`betResult` to `GameResult` |
| `ui/src/pages/rock-paper-scissors/rock-paper-scissors.tsx` | Amber bet banner before move selection; pass `betAmount`/`betResult` to `GameResult` |

---

## Task 0: Expose `bet_amount` in the API response and regenerate the TypeScript schema

**Files:**
- Modify: `internal/apis/game.go`
- Regenerate: `ui/src/schema.d.ts` (via `npm run generate:schema` in the `ui/` directory)

The `RpsGame` API response struct in `game.go` currently does not include `BetAmount`. The model (`internal/models/rps_games.go`) has it; the API response layer does not. Without it, game responses won't include the bet amount and the frontend TypeScript types won't know it exists.

- [ ] **Step 1: Add `BetAmount` to the `RpsGame` API struct**

In `internal/apis/game.go`, add `BetAmount` to the `RpsGame` struct (after `UpdatedAt` around line 56):

```go
type RpsGame struct {
	_            struct{}          `db:"rps_games" schema:"gaming" json:"-"`
	ID           uuid.UUID         `db:"id,pk" json:"id"`
	CompletedAt  *time.Time        `db:"completed_at" json:"completed_at,omitempty"`
	ExpiresAt    time.Time         `db:"expires_at" json:"expires_at"`
	Status       RpsGameStatus     `db:"status" json:"status" default:"pending" enum:"pending,cancelled,completed"`
	Metadata     []byte            `db:"metadata" json:"metadata"`
	CreatedAt    time.Time         `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time         `db:"updated_at" json:"updated_at"`
	BetAmount    *int64            `json:"bet_amount,omitempty" doc:"Optional points wager. If set, the host must have sufficient balance."`
	Participants []*RpsParticipant `db:"rps_participants" src:"id" dest:"game_id" table:"gaming.rps_participants" json:"participants,omitempty"`
}
```

- [ ] **Step 2: Map `BetAmount` in `toApiRpsGame`**

In `internal/apis/game.go`, update the `toApiRpsGame` function (currently around line 73):

```go
func toApiRpsGame(game *models.RpsGame) *RpsGame {
	if game == nil {
		return nil
	}
	return &RpsGame{
		ID:           game.ID,
		CompletedAt:  game.CompletedAt,
		ExpiresAt:    game.ExpiresAt,
		Status:       toApiRpsGameStatus(game.Status),
		Metadata:     game.Metadata,
		CreatedAt:    game.CreatedAt,
		UpdatedAt:    game.UpdatedAt,
		BetAmount:    game.BetAmount,
		Participants: mapper.Map(game.Participants, ToApiRpsParticipant),
	}
}
```

- [ ] **Step 3: Verify the backend compiles**

```bash
cd /Users/tkahng/github/tkahng/go/playground && go build ./...
```

Expected: no errors.

- [ ] **Step 4: Regenerate the TypeScript schema**

Start the backend server (it must be running at `localhost:8080`) then run:

```bash
cd /Users/tkahng/github/tkahng/go/playground/ui && npm run generate:schema
```

Expected: `ui/src/schema.d.ts` is updated. Verify `bet_amount` now appears in the `RpsGame` schema section:

```bash
grep -A2 "bet_amount" /Users/tkahng/github/tkahng/go/playground/ui/src/schema.d.ts | head -10
```

Expected output includes something like:
```
bet_amount?: number;
```
inside the `RpsGame` type definition.

- [ ] **Step 5: Commit**

```bash
git add internal/apis/game.go ui/src/schema.d.ts
git commit -m "feat(api): expose bet_amount in RpsGame response and regenerate schema"
```

---

## Task 1: Extend `rps-game-queries.tsx` — add `getLedgerBalance` and update `requestGame`

**Files:**
- Modify: `ui/src/lib/rps-game-queries.tsx`

> No test infrastructure exists for the UI. Verification is TypeScript compilation (`cd ui && npx tsc --noEmit`).

- [ ] **Step 1: Add `getLedgerBalance` method to the `RpsGameQueries` class**

In `ui/src/lib/rps-game-queries.tsx`, add this method inside the `RpsGameQueries` class after `findPlayerByEmail`:

```typescript
async getLedgerBalance({ token }: { token: string }) {
  const { data, error } = await client.GET(`/api/ledger/balance`, {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
  if (error) {
    throw ApiError.fromErrorModel(error);
  }
  return data;
}
```

- [ ] **Step 2: Update `requestGame` to accept optional `betAmount`**

Replace the existing `requestGame` method:

```typescript
async requestGame({
  token,
  move,
  playerId,
  betAmount,
}: {
  token: string;
  move: "rock" | "paper" | "scissors";
  playerId: string;
  betAmount?: number;
}) {
  const { data, error } = await client.POST(`/api/games/rps/requests`, {
    headers: {
      Authorization: `Bearer ${token}`,
    },
    body: {
      inviting_player_id: playerId,
      move,
      ...(betAmount ? { bet_amount: betAmount } : {}),
    },
  });
  if (error) {
    throw ApiError.fromErrorModel(error);
  }
  return data;
}
```

- [ ] **Step 3: Verify TypeScript compiles**

```bash
cd /Users/tkahng/github/tkahng/go/playground/ui && npx tsc --noEmit
```

Expected: no errors.

- [ ] **Step 4: Commit**

```bash
git add ui/src/lib/rps-game-queries.tsx
git commit -m "feat(ui): add getLedgerBalance query and betAmount to requestGame"
```

---

## Task 2: Add `children` slot to `MoveSelection`

**Files:**
- Modify: `ui/src/pages/account/rock-paper-scissors/move.tsx`

This allows callers to inject content (the bet toggle) between the move cards and the submit button.

- [ ] **Step 1: Add `children` to `MoveSelectionProps` and render it above the submit button**

Replace the `MoveSelectionProps` type and the submit section in `move.tsx`:

```typescript
export type MoveSelectionProps = {
  handleSubmit: (move: Move) => void;
  opponentPlayer?: Player | null;
  children?: React.ReactNode;
};
```

In the `MoveSelection` function, replace the `{/* Submit Button */}` section (currently starts at line 127):

```tsx
{/* Optional slot — rendered between move cards and submit button */}
{children}

{/* Submit Button */}
<div className="flex justify-center">
  <Button
    size="lg"
    className="min-w-64 text-lg h-12"
    disabled={!selectedMove}
    onClick={() => handleSubmit(selectedMove || "rock")}
  >
    {selectedMove
      ? `Play ${selectedMove.charAt(0).toUpperCase() + selectedMove.slice(1)}`
      : "Select a Move"}
  </Button>
</div>
```

- [ ] **Step 2: Verify TypeScript compiles**

```bash
cd /Users/tkahng/github/tkahng/go/playground/ui && npx tsc --noEmit
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add ui/src/pages/account/rock-paper-scissors/move.tsx
git commit -m "feat(ui): add children slot to MoveSelection component"
```

---

## Task 3: Add bet toggle + field to `create-game-dialog.tsx`

**Files:**
- Modify: `ui/src/pages/account/rock-paper-scissors/create-game-dialog.tsx`

- [ ] **Step 1: Add imports**

Add to the import block at the top of the file:

```typescript
import { useQuery } from "@tanstack/react-query";
import { Switch } from "@/components/ui/switch";
```

- [ ] **Step 2: Add bet state and balance query inside `CreateGameDialog`**

After the existing `useState` declarations (after line 35 `const [email, setEmail] = useState...`), add:

```typescript
const [betEnabled, setBetEnabled] = useState(false);
const [betAmount, setBetAmount] = useState<number | undefined>(undefined);

const { data: balanceData } = useQuery({
  queryKey: [{ key: "ledger-balance" }],
  enabled: !!user?.tokens.access_token,
  queryFn: async () => {
    if (!user?.tokens.access_token) throw new Error("No access token");
    return rpsGameQueries.getLedgerBalance({ token: user.tokens.access_token });
  },
});
```

- [ ] **Step 3: Update `requestGameMutation` to pass `bet_amount`**

Replace the `requestGame` call inside `requestGameMutation.mutationFn`:

```typescript
return rpsGameQueries.requestGame({
  token: user?.tokens.access_token,
  move: data.move,
  playerId: player.id,
  betAmount: betEnabled ? betAmount : undefined,
});
```

- [ ] **Step 4: Reset bet state on dialog close**

In `DialogContent`'s `onCloseAutoFocus`, add resets alongside the existing ones:

```typescript
onCloseAutoFocus={() => {
  setPlayer(null);
  setEmailRequest(false);
  setSearched(false);
  setBetEnabled(false);
  setBetAmount(undefined);
  searchForm.reset();
}}
```

- [ ] **Step 5: Add bet toggle as `children` of `MoveSelection` in the registered player branch**

Replace the registered player `MoveSelection` call (currently at line 229–238):

```tsx
{searched &&
  player &&
  !findPlayerMutation.isPending &&
  !findPlayerMutation.isError && (
    <MoveSelection
      handleSubmit={(move: Move) =>
        requestGameMutation.mutate({ move })
      }
    >
      <div className="border-t pt-3 mt-2 space-y-2">
        <div className="flex items-center justify-between">
          <label htmlFor="bet-toggle" className="text-sm cursor-pointer">
            Add a bet?
          </label>
          <Switch
            id="bet-toggle"
            checked={betEnabled}
            onCheckedChange={(checked) => {
              setBetEnabled(checked);
              if (!checked) setBetAmount(undefined);
            }}
          />
        </div>
        {betEnabled && (
          <div className="space-y-1">
            <p className="text-xs text-muted-foreground">
              Available balance:{" "}
              {balanceData?.available_balance !== undefined
                ? `${balanceData.available_balance} pts`
                : "..."}
            </p>
            <Input
              type="number"
              min={1}
              max={balanceData?.available_balance}
              value={betAmount ?? ""}
              onChange={(e) =>
                setBetAmount(
                  e.target.value ? parseInt(e.target.value, 10) : undefined
                )
              }
              placeholder="Enter bet amount"
            />
          </div>
        )}
      </div>
    </MoveSelection>
  )}
```

- [ ] **Step 6: Verify TypeScript compiles**

```bash
cd /Users/tkahng/github/tkahng/go/playground/ui && npx tsc --noEmit
```

Expected: no errors.

- [ ] **Step 7: Commit**

```bash
git add ui/src/pages/account/rock-paper-scissors/create-game-dialog.tsx
git commit -m "feat(ui): add optional bet toggle and amount field to create-game dialog"
```

---

## Task 4: Balance badge + Bet column in authenticated RPS page

**Files:**
- Modify: `ui/src/pages/account/rock-paper-scissors/rock-paper-scissors.tsx`

- [ ] **Step 1: Add balance query**

After the existing `useQuery` for games (after line 80 `return { data: playerGames, meta: games.meta }`), add a second `useQuery`:

```typescript
const { data: balanceData } = useQuery({
  queryKey: [{ key: "ledger-balance" }],
  queryFn: async () => {
    if (!userInfo.user?.tokens.access_token) {
      throw new Error("No access token");
    }
    return rpsGameQueries.getLedgerBalance({
      token: userInfo.user.tokens.access_token,
    });
  },
});
```

- [ ] **Step 2: Add balance badge to the page header**

Replace:

```tsx
<h1>Rock Paper Scissors</h1>
```

With:

```tsx
<div className="flex items-center gap-3 mb-1">
  <h1>Rock Paper Scissors</h1>
  {balanceData && (
    <span className="inline-flex items-center gap-1.5 bg-indigo-100 text-indigo-700 rounded-full px-3 py-0.5 text-sm font-medium">
      🪙 {balanceData.available_balance} pts
    </span>
  )}
</div>
```

- [ ] **Step 3: Add Bet column to the DataTable**

Add a new column object after the `"Created At"` column entry:

```typescript
{
  header: "Bet",
  cell: ({ row }) => {
    const betAmount = row.original.rpsGame.bet_amount;
    if (!betAmount) {
      return <span className="text-muted-foreground">—</span>;
    }
    const state = CalculateGameState(row.original);
    if (state === GameState.Win) {
      return (
        <span className="text-green-600 font-semibold">+{betAmount} pts</span>
      );
    }
    if (state === GameState.Lose) {
      return (
        <span className="text-red-600 font-semibold">−{betAmount} pts</span>
      );
    }
    if (
      state === GameState.Tie ||
      state === GameState.Cancelled ||
      state === GameState.Expired
    ) {
      return <span className="text-muted-foreground">refunded</span>;
    }
    // Pending or Submitted
    return (
      <span className="text-amber-600">{betAmount} pts at stake</span>
    );
  },
},
```

- [ ] **Step 4: Verify TypeScript compiles**

```bash
cd /Users/tkahng/github/tkahng/go/playground/ui && npx tsc --noEmit
```

Expected: no errors.

- [ ] **Step 5: Commit**

```bash
git add ui/src/pages/account/rock-paper-scissors/rock-paper-scissors.tsx
git commit -m "feat(ui): add balance badge to RPS page header and Bet column to games table"
```

---

## Task 5: Add bet outcome card to `game-result.tsx`

**Files:**
- Modify: `ui/src/pages/account/rock-paper-scissors/game-result.tsx`

- [ ] **Step 1: Add optional bet props to `GameResultProps`**

Replace the `GameResultProps` interface:

```typescript
interface GameResultProps {
  result: Result;
  opponent: string;
  playerMove: Move;
  opponentMove: Move;
  betAmount?: number;
  betResult?: "win" | "lose" | "tie";
}
```

- [ ] **Step 2: Update the function signature to destructure the new props**

Replace:

```typescript
export function GameResult({
  result,
  opponent,
  playerMove,
  opponentMove,
}: GameResultProps) {
```

With:

```typescript
export function GameResult({
  result,
  opponent,
  playerMove,
  opponentMove,
  betAmount,
  betResult,
}: GameResultProps) {
```

- [ ] **Step 3: Add bet outcome card after the moves Card**

After the closing `</Card>` tag (currently the last element before the closing `</div>` of the component, around line 106), add:

```tsx
{/* Bet Outcome */}
{betAmount !== undefined && betResult !== undefined && (
  <Card
    className={`p-6 text-center ${
      betResult === "win"
        ? "bg-success/10 border-success/30"
        : betResult === "lose"
          ? "bg-destructive/10 border-destructive/30"
          : "bg-muted/50"
    }`}
  >
    <p className="text-sm text-muted-foreground mb-1">
      {betResult === "win"
        ? "Bet won"
        : betResult === "lose"
          ? "Bet lost"
          : "Bet refunded"}
    </p>
    <p
      className={`text-3xl font-bold ${
        betResult === "win"
          ? "text-success"
          : betResult === "lose"
            ? "text-destructive"
            : "text-muted-foreground"
      }`}
    >
      {betResult === "win"
        ? `+${betAmount} pts`
        : betResult === "lose"
          ? `−${betAmount} pts`
          : `${betAmount} pts`}
    </p>
  </Card>
)}
```

- [ ] **Step 4: Verify TypeScript compiles**

```bash
cd /Users/tkahng/github/tkahng/go/playground/ui && npx tsc --noEmit
```

Expected: no errors.

- [ ] **Step 5: Commit**

```bash
git add ui/src/pages/account/rock-paper-scissors/game-result.tsx
git commit -m "feat(ui): add bet outcome card to GameResult component"
```

---

## Task 6: Wire bet outcome into `selected-game-dialog.tsx`

**Files:**
- Modify: `ui/src/pages/account/rock-paper-scissors/selected-game-dialog.tsx`

- [ ] **Step 1: Pass `betAmount` and `betResult` to `GameResult`**

Find the `GameResult` call (around line 140–148) and replace it:

```tsx
{selectedGame?.data &&
  !expired &&
  selectedGame.data.rpsGame.status === "completed" && (
    <GameResult
      result={selectedGame.data.player.result}
      opponent={selectedGame.data.opponent.player?.email || ""}
      playerMove={selectedGame.data.player.move}
      opponentMove={selectedGame.data.opponent.move}
      betAmount={selectedGame.data.rpsGame.bet_amount}
      betResult={selectedGame.data.player.result}
    />
  )}
```

- [ ] **Step 2: Verify TypeScript compiles**

```bash
cd /Users/tkahng/github/tkahng/go/playground/ui && npx tsc --noEmit
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add ui/src/pages/account/rock-paper-scissors/selected-game-dialog.tsx
git commit -m "feat(ui): pass bet outcome to GameResult in selected-game-dialog"
```

---

## Task 7: Amber bet banner + bet outcome on public invite page

**Files:**
- Modify: `ui/src/pages/rock-paper-scissors/rock-paper-scissors.tsx`

- [ ] **Step 1: Add amber bet banner above `MoveSelection`**

Replace the `{isSelection && ...}` block (currently around lines 65–72):

```tsx
{isSelection && (
  <div>
    {rpsGame.data.rps_game.bet_amount ? (
      <div className="mb-4 rounded-lg border border-amber-300 bg-amber-50 p-4 text-center">
        <p className="font-semibold text-amber-800">
          🪙 This game has a {rpsGame.data.rps_game.bet_amount} pt bet
        </p>
        <p className="mt-1 text-sm text-amber-700">
          Accepting will deduct {rpsGame.data.rps_game.bet_amount} pts from
          your balance
        </p>
      </div>
    ) : null}
    <MoveSelection
      handleSubmit={(move) => mutation.mutate({ token: token!, move })}
      opponentPlayer={rpsGame?.data.requesting_participant?.player}
    />
  </div>
)}
```

- [ ] **Step 2: Pass bet props to `GameResult` in the result display**

Replace the `{isResult && ...}` block (currently around lines 73–84):

```tsx
{isResult && (
  <div>
    <GameResult
      result={game.invited_participant.result}
      opponent={game.requesting_participant.player?.email || ""}
      playerMove={game.invited_participant.move}
      opponentMove={game.requesting_participant.move}
      betAmount={game.rps_game.bet_amount}
      betResult={game.invited_participant.result}
    />
  </div>
)}
```

- [ ] **Step 3: Verify TypeScript compiles**

```bash
cd /Users/tkahng/github/tkahng/go/playground/ui && npx tsc --noEmit
```

Expected: no errors.

- [ ] **Step 4: Commit**

```bash
git add ui/src/pages/rock-paper-scissors/rock-paper-scissors.tsx
git commit -m "feat(ui): add bet warning banner and bet outcome to public RPS invite page"
```

---

## Browser Verification Checklist

After all tasks complete, verify manually in the browser:

- [ ] RPS page header shows a `🪙 N pts` badge
- [ ] Games table has a "Bet" column; non-bet games show `—`
- [ ] Creating a game with a registered player shows the bet toggle below move cards
- [ ] Toggling bet on shows available balance and a number input
- [ ] Toggling bet off hides the input and resets the amount
- [ ] Submitting a game with a bet sends `bet_amount` (check Network tab)
- [ ] Completed games with bets show correct `+N pts` / `−N pts` / `refunded` in the Bet column
- [ ] Opening a completed game dialog with a bet shows the bet outcome card
- [ ] Visiting a public invite link for a game with a bet shows the amber banner
- [ ] After playing a bet game via public invite, the result screen shows the bet outcome card
