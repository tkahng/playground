# RPS Zero-Balance Recharge Link Design

**Date:** 2026-04-03  
**Status:** Approved

## Problem

On the Rock Paper Scissors page, when a user has 0 points there is no clear path to the points purchase page. Users are left stranded with no actionable next step.

## Goal

Make it obvious and easy to navigate to `/account/settings/points` when the user's balance is zero, at the two most relevant moments: viewing the page header and enabling a bet in the create-game dialog.

## Scope

Two touch points, both in the RPS account section:

1. `ui/src/pages/account/rock-paper-scissors/rock-paper-scissors.tsx` — balance badge in header
2. `ui/src/pages/account/rock-paper-scissors/create-game-dialog.tsx` — bet section inside create game dialog

## Design

### 1. Balance Badge (RPS Page Header)

**File:** `rock-paper-scissors.tsx`, lines 115–119

- When `balanceData.available_balance > 0`: render the existing plain `<span>` unchanged.
- When `balanceData.available_balance === 0`: replace the `<span>` with a `<Link>` (react-router) pointing to `RouteMap.POINTS_SETTINGS` (`/account/settings/points`).
  - Label: `🪙 0 pts — Buy Points`
  - Style: amber text color, underline, underline-offset-2, hover darkens amber

No banner or callout is added to the page — the badge link is the sole affordance.

### 2. Bet Section in Create Game Dialog

**File:** `create-game-dialog.tsx`, lines 274–295

- When `betEnabled` is true and `balanceData?.available_balance === 0`: hide the number `<Input>` and show an inline `<p>` instead:
  - Text: `You have 0 pts.`
  - Followed by an `<a>` link: `Buy Points →`
  - Link opens `/account/settings/points` in a **new tab** (`target="_blank" rel="noopener noreferrer"`) so the dialog/game flow is not interrupted
  - Style: amber text color, underline
- When balance > 0: existing input renders normally, no change.
- The bet toggle (`Switch`) remains visible in both cases so users understand the section's purpose.

## What is NOT changing

- No banner on the main RPS page
- No changes to the points settings page itself
- No changes to the bet toggle behavior or validation logic
- Badge is not linked when balance > 0 (avoids distracting users who are fine)

## Routes

- Points settings page: `RouteMap.POINTS_SETTINGS` = `/account/settings/points`
- Already defined in `ui/src/components/route-map.ts` and `ui/src/components/links.tsx`
