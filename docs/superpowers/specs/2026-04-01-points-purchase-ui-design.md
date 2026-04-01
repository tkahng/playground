# Points Purchase UI Design

**Date:** 2026-04-01  
**Status:** Approved

## Overview

Add a Points settings page at `/account/settings/points` where authenticated users can view their current balance and purchase one of three points packages via Stripe Checkout. The backend endpoint, webhook fulfillment, and ledger are already fully implemented. This spec covers the frontend only.

## Scope

One new file, four modified files.

### Files

| File | Change |
|---|---|
| `ui/src/pages/settings/points-settings.tsx` | **New** — Points settings page |
| `ui/src/lib/api.ts` | Add `createPointsCheckoutSession`; update `getProductsWithPrices` to accept optional `metadata_type` query param |
| `ui/src/components/route-map.ts` | Add `POINTS_SETTINGS: "/account/settings/points"` |
| `ui/src/components/links.tsx` | Add `POINTS_SETTINGS` to `settingsSidebarLinks`; uncomment `BILLING_SETTINGS` |
| `ui/src/App.tsx` | Add route for the new points page |

## API Layer

### `getProductsWithPrices` — update signature

Add optional `metadata_type` param to filter products by type:

```typescript
getProductsWithPrices(token?: string, metadata_type?: "subscription" | "points")
// GET /api/stripe/products?metadata_type=points
```

Existing callers pass no `metadata_type` — backward compatible (omitting the param returns all products as before).

### `createPointsCheckoutSession` — new function

```typescript
createPointsCheckoutSession(token: string, { price_id }: { price_id: string })
// POST /api/ledger/points/checkout
// Body: { price_id: string }
// Returns: { url: string }
```

On success, the caller redirects to the returned URL via `window.location.href`. After payment, Stripe redirects to the existing `/payment` success page.

## Points Settings Page

### Route

`/account/settings/points` — authenticated, rendered inside the account dashboard layout with the settings sidebar.

### Layout

Three cards side by side (matches the existing `pricing-mini.tsx` aesthetic). Balance badge top-right of the page header.

```
┌─────────────────────────────────────────┐
│ Buy Points                  🪙 250 pts  │
│ Points are used to place bets in games  │
│                                         │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐│
│  │ STARTER  │ │ POPULAR  │ │  VALUE   ││
│  │   $2     │ │   $5     │ │   $10    ││
│  │ 100 pts  │ │ 300 pts  │ │ 700 pts  ││
│  │  [Buy]   │ │  [Buy]   │ │  [Buy]   ││
│  └──────────┘ └──────────┘ └──────────││
└─────────────────────────────────────────┘
```

### Data

Two queries on mount:

1. `getProductsWithPrices(token, "points")` — fetches points packages from Stripe, keyed as `["points-products"]`
2. `rpsGameQueries.getLedgerBalance({ token })` — fetches current balance, keyed as `[{ key: "ledger-balance" }]` (shared cache with RPS page)

### Card rendering

Products are filtered to `metadata_type === "points"`. Prices within the product are sorted ascending by `unit_amount`. Each price renders as one card showing:
- Price nickname or index-based label (e.g. "Starter", "Popular", "Value") — middle card gets an indigo "POPULAR" badge
- `unit_amount / 100` formatted as USD currency
- `metadata.points_amount` as the point count
- "Buy" button

### Interaction

- Clicking "Buy" calls `createPointsCheckoutSession({ price_id })` with a loading spinner on the button
- On success: `window.location.href = url` (redirects to Stripe Checkout in same tab)
- On error: `toast.error(error.message)` via existing Sonner toast

### States

- **Loading**: spinner centered
- **Error fetching products**: error message
- **No points products**: message "No points packages available" with a link to the admin Stripe dashboard
- **Balance query fails**: badge hidden silently (non-blocking)

## Navigation

`settingsSidebarLinks` in `links.tsx`:

```typescript
export const settingsSidebarLinks: LinkDto[] = [
  RouteLinks.GENERAL_SETTINGS,   // "General"
  RouteLinks.BILLING_SETTINGS,   // "Billing" (uncommented)
  RouteLinks.POINTS_SETTINGS,    // "Points" (new)
];
```

`RouteLinks.POINTS_SETTINGS` added alongside `BILLING_SETTINGS`:
```typescript
POINTS_SETTINGS: { to: RouteMap.POINTS_SETTINGS, title: "Points" },
```

## Out of Scope

- Transaction history view
- Points balance top-up notifications
- Admin points management
- Refunds
