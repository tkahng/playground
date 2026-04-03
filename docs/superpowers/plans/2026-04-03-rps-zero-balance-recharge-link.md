# RPS Zero-Balance Recharge Link Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** When a user's point balance is 0 on the RPS page, make it obvious how to navigate to the points purchase page by linking the balance badge and showing an inline prompt in the create-game dialog bet section.

**Architecture:** Two targeted edits to existing components — the balance badge in the RPS page header conditionally renders as a react-router `<Link>` when balance is 0, and the bet input in the create-game dialog is replaced with an inline `<a>` prompt when balance is 0.

**Tech Stack:** React, react-router `<Link>`, TypeScript, Tailwind CSS

---

## Files

- Modify: `ui/src/pages/account/rock-paper-scissors/rock-paper-scissors.tsx` — balance badge conditional link
- Modify: `ui/src/pages/account/rock-paper-scissors/create-game-dialog.tsx` — bet section inline recharge prompt

---

### Task 1: Balance Badge — Link when 0 pts

**Files:**
- Modify: `ui/src/pages/account/rock-paper-scissors/rock-paper-scissors.tsx:111-120`

- [ ] **Step 1: Add the `Link` import**

In `rock-paper-scissors.tsx`, add `Link` to the react-router import at the top of the file. The existing imports already include `useSearchParams` from `"react-router"`:

```tsx
import { useSearchParams, Link } from "react-router";
```

Also add the `RouteMap` import:

```tsx
import { RouteMap } from "@/components/route-map";
```

- [ ] **Step 2: Replace the balance badge span with conditional rendering**

Find the existing balance badge block (lines ~115–119):

```tsx
{balanceData && (
  <span className="inline-flex items-center gap-1.5 rounded-full px-3 py-0.5 text-sm font-medium">
    🪙 {balanceData.available_balance} pts
  </span>
)}
```

Replace with:

```tsx
{balanceData && (
  balanceData.available_balance === 0 ? (
    <Link
      to={RouteMap.POINTS_SETTINGS}
      className="inline-flex items-center gap-1.5 rounded-full px-3 py-0.5 text-sm font-medium text-amber-600 underline underline-offset-2 hover:text-amber-700"
    >
      🪙 0 pts — Buy Points
    </Link>
  ) : (
    <span className="inline-flex items-center gap-1.5 rounded-full px-3 py-0.5 text-sm font-medium">
      🪙 {balanceData.available_balance} pts
    </span>
  )
)}
```

- [ ] **Step 3: Verify the page renders without TypeScript errors**

```bash
cd ui && npx tsc --noEmit
```

Expected: no errors related to the modified file.

- [ ] **Step 4: Commit**

```bash
git add ui/src/pages/account/rock-paper-scissors/rock-paper-scissors.tsx
git commit -m "feat(rps): link balance badge to points page when balance is 0"
```

---

### Task 2: Create Game Dialog — Inline recharge prompt when balance is 0

**Files:**
- Modify: `ui/src/pages/account/rock-paper-scissors/create-game-dialog.tsx:274-295`

- [ ] **Step 1: Add the `RouteMap` import**

At the top of `create-game-dialog.tsx`, add:

```tsx
import { RouteMap } from "@/components/route-map";
```

- [ ] **Step 2: Replace the bet input with a conditional block**

Find the existing `betEnabled` content block (lines ~274–295):

```tsx
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
      onChange={(e) => {
        const parsed = parseInt(e.target.value, 10);
        setBetAmount(
          e.target.value && Number.isFinite(parsed) ? parsed : undefined
        );
      }}
      placeholder="Enter bet amount"
    />
  </div>
)}
```

Replace with:

```tsx
{betEnabled && (
  <div className="space-y-1">
    <p className="text-xs text-muted-foreground">
      Available balance:{" "}
      {balanceData?.available_balance !== undefined
        ? `${balanceData.available_balance} pts`
        : "..."}
    </p>
    {balanceData?.available_balance === 0 ? (
      <p className="text-xs text-amber-600">
        You have 0 pts.{" "}
        <a
          href={RouteMap.POINTS_SETTINGS}
          target="_blank"
          rel="noopener noreferrer"
          className="underline hover:text-amber-700"
        >
          Buy Points →
        </a>
      </p>
    ) : (
      <Input
        type="number"
        min={1}
        max={balanceData?.available_balance}
        value={betAmount ?? ""}
        onChange={(e) => {
          const parsed = parseInt(e.target.value, 10);
          setBetAmount(
            e.target.value && Number.isFinite(parsed) ? parsed : undefined
          );
        }}
        placeholder="Enter bet amount"
      />
    )}
  </div>
)}
```

- [ ] **Step 3: Verify no TypeScript errors**

```bash
cd ui && npx tsc --noEmit
```

Expected: no errors related to the modified file.

- [ ] **Step 4: Commit**

```bash
git add ui/src/pages/account/rock-paper-scissors/create-game-dialog.tsx
git commit -m "feat(rps): show buy points link in bet section when balance is 0"
```
