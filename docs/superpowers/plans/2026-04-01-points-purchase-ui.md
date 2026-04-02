# Points Purchase UI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a `/account/settings/points` page where authenticated users can view their balance and purchase one of three points packages via Stripe Checkout.

**Architecture:** Four tasks, no new shared components. Task 1 extends the API layer (`api.ts`). Task 2 wires up routing constants and sidebar navigation. Task 3 creates the page. Task 4 registers the route in `App.tsx`. Each task type-checks cleanly before the next begins.

**Tech Stack:** React 19, TanStack React Query v5, openapi-fetch, shadcn/ui (Button), Sonner toasts, Zod (not needed here — no forms), TypeScript strict mode.

---

### Task 1: Extend `api.ts` — add `metadata_type` param and `createPointsCheckoutSession`

**Files:**
- Modify: `ui/src/lib/api.ts:675-691` (`getProductsWithPrices`)
- Modify: `ui/src/lib/api.ts:769` (append after `createCheckoutSession`)

- [ ] **Step 1: Update `getProductsWithPrices` to accept optional `metadata_type`**

Replace the current function (lines 675–691) with:

```typescript
export const getProductsWithPrices = async (
  token?: string,
  metadata_type?: "subscription" | "points"
) => {
  const { data, error } = await client.GET("/api/stripe/products", {
    headers: token
      ? {
          Authorization: `Bearer ${token}`,
        }
      : {},
    params: { query: { metadata_type } },
  });
  if (error) {
    throw ApiError.fromErrorModel(error);
  }
  if (!data) {
    throw new Error("No data");
  }

  return data;
};
```

Existing callers (`pricing.tsx`, `team-customer-form.tsx`) call `getProductsWithPrices()` or `getProductsWithPrices(token)` with no `metadata_type` — passing `undefined` in the query param is equivalent to omitting it, so they remain unaffected.

- [ ] **Step 2: Add `createPointsCheckoutSession` after `createCheckoutSession`**

Add this function after the closing brace of `createCheckoutSession` (currently ending at line 769):

```typescript
export const createPointsCheckoutSession = async (
  token: string,
  { price_id }: { price_id: string }
) => {
  const { data, error } = await client.POST("/api/ledger/points/checkout", {
    headers: {
      Authorization: `Bearer ${token}`,
    },
    body: {
      price_id,
    },
  });
  if (error) {
    throw ApiError.fromErrorModel(error);
  }
  if (!data) {
    throw new Error("No data");
  }
  return data;
};
```

The return type is inferred as `{ url: string }` from `components["schemas"]["Create-points-checkoutResponse"]` in `schema.d.ts`.

- [ ] **Step 3: Type-check**

```bash
cd ui && npm run typecheck
```

Expected: no errors.

- [ ] **Step 4: Commit**

```bash
git add ui/src/lib/api.ts
git commit -m "feat(ui): add metadata_type param and createPointsCheckoutSession to api"
```

---

### Task 2: Add `POINTS_SETTINGS` to route map and sidebar navigation

**Files:**
- Modify: `ui/src/components/route-map.ts:18`
- Modify: `ui/src/components/links.tsx:78` (add RouteLink entry)
- Modify: `ui/src/components/links.tsx:164-167` (update `settingsSidebarLinks`)

- [ ] **Step 1: Add `POINTS_SETTINGS` to `RouteMap`**

In `ui/src/components/route-map.ts`, add `POINTS_SETTINGS` immediately after `BILLING_SETTINGS` on line 18:

```typescript
  BILLING_SETTINGS: "/account/settings/billing",
  POINTS_SETTINGS: "/account/settings/points",
```

- [ ] **Step 2: Add `POINTS_SETTINGS` RouteLink in `links.tsx`**

In `ui/src/components/links.tsx`, line 78 currently reads:

```typescript
  BILLING_SETTINGS: { to: RouteMap.BILLING_SETTINGS, title: "Billing" },
```

Add `POINTS_SETTINGS` on the next line:

```typescript
  BILLING_SETTINGS: { to: RouteMap.BILLING_SETTINGS, title: "Billing" },
  POINTS_SETTINGS: { to: RouteMap.POINTS_SETTINGS, title: "Points" },
```

- [ ] **Step 3: Update `settingsSidebarLinks` in `links.tsx`**

Lines 164–167 currently read:

```typescript
export const settingsSidebarLinks: LinkDto[] = [
  RouteLinks.GENERAL_SETTINGS,
  // RouteLinks.BILLING_SETTINGS,
];
```

Replace with:

```typescript
export const settingsSidebarLinks: LinkDto[] = [
  RouteLinks.GENERAL_SETTINGS,
  RouteLinks.BILLING_SETTINGS,
  RouteLinks.POINTS_SETTINGS,
];
```

- [ ] **Step 4: Type-check**

```bash
cd ui && npm run typecheck
```

Expected: no errors.

- [ ] **Step 5: Commit**

```bash
git add ui/src/components/route-map.ts ui/src/components/links.tsx
git commit -m "feat(ui): add POINTS_SETTINGS route and sidebar link"
```

---

### Task 3: Create `points-settings.tsx`

**Files:**
- Create: `ui/src/pages/settings/points-settings.tsx`

- [ ] **Step 1: Create the file**

Create `ui/src/pages/settings/points-settings.tsx` with this content:

```typescript
import { CenteredSpinner } from "@/components/centered-spinner";
import { DashboardSidebar } from "@/components/dashboard-sidebar";
import { settingsSidebarLinks } from "@/components/links";
import { Button } from "@/components/ui/button";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import {
  createPointsCheckoutSession,
  getProductsWithPrices,
} from "@/lib/api";
import { rpsGameQueries } from "@/lib/rps-game-queries";
import { useMutation, useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { toast } from "sonner";

const CARD_LABELS = ["Starter", "Popular", "Value"];

export default function PointsSettingsPage() {
  const { user } = useAuthProvider();
  const token = user?.tokens.access_token;

  const productsQuery = useQuery({
    queryKey: ["points-products"],
    queryFn: () => getProductsWithPrices(token, "points"),
    enabled: !!token,
  });

  const balanceQuery = useQuery({
    queryKey: [{ key: "ledger-balance" }],
    queryFn: () => rpsGameQueries.getLedgerBalance({ token: token! }),
    enabled: !!token,
  });

  const [loadingPriceId, setLoadingPriceId] = useState<string | null>(null);

  const checkoutMutation = useMutation({
    mutationFn: ({ price_id }: { price_id: string }) =>
      createPointsCheckoutSession(token!, { price_id }),
    onSuccess: (data) => {
      window.location.href = data.url;
    },
    onError: (error: Error) => {
      setLoadingPriceId(null);
      toast.error(error.message);
    },
  });

  if (productsQuery.isPending) {
    return (
      <div className="flex">
        <DashboardSidebar links={settingsSidebarLinks} />
        <div className="flex-1 flex items-center justify-center p-12">
          <CenteredSpinner />
        </div>
      </div>
    );
  }

  if (productsQuery.isError) {
    return (
      <div className="flex">
        <DashboardSidebar links={settingsSidebarLinks} />
        <div className="flex-1 p-12">
          <p className="text-destructive">Failed to load points packages.</p>
        </div>
      </div>
    );
  }

  const pointsProducts =
    productsQuery.data?.data?.filter(
      (p) => p.metadata?.metadata_type === "points"
    ) ?? [];

  const prices = pointsProducts
    .flatMap((p) => p.prices ?? [])
    .filter((price) => price.unit_amount != null)
    .sort((a, b) => (a.unit_amount ?? 0) - (b.unit_amount ?? 0));

  return (
    <div className="flex">
      <DashboardSidebar links={settingsSidebarLinks} />
      <div className="flex-1 space-y-6 p-12 w-full">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-2xl font-bold">Buy Points</h2>
            <p className="text-muted-foreground text-sm mt-1">
              Points are used to place bets in games
            </p>
          </div>
          {balanceQuery.data != null && (
            <div className="bg-indigo-100 text-indigo-700 rounded-full px-3 py-1 text-sm font-medium">
              🪙 {balanceQuery.data.available_balance} pts
            </div>
          )}
        </div>

        {prices.length === 0 ? (
          <p className="text-muted-foreground">No points packages available.</p>
        ) : (
          <div className="grid grid-cols-3 gap-4">
            {prices.map((price, index) => {
              const label = CARD_LABELS[index] ?? `Package ${index + 1}`;
              const isPopular = index === 1;
              const pointsAmount = price.metadata?.points_amount
                ? parseInt(price.metadata.points_amount, 10)
                : null;
              const isLoading = loadingPriceId === price.id;

              return (
                <div
                  key={price.id}
                  className={`border-2 rounded-xl p-6 text-center flex flex-col items-center gap-2 ${
                    isPopular
                      ? "border-indigo-500 bg-indigo-50"
                      : "border-border"
                  }`}
                >
                  <div
                    className={`text-xs font-semibold uppercase tracking-wide ${
                      isPopular ? "text-indigo-600" : "text-muted-foreground"
                    }`}
                  >
                    {isPopular && (
                      <span className="bg-indigo-100 text-indigo-600 rounded px-1.5 py-0.5 mr-1">
                        POPULAR
                      </span>
                    )}
                    {label}
                  </div>
                  <div className="text-3xl font-extrabold">
                    ${((price.unit_amount ?? 0) / 100).toFixed(0)}
                  </div>
                  {pointsAmount != null && (
                    <div className="text-muted-foreground text-sm">
                      {pointsAmount} pts
                    </div>
                  )}
                  <Button
                    className="w-full mt-2"
                    disabled={checkoutMutation.isPending}
                    onClick={() => {
                      setLoadingPriceId(price.id);
                      checkoutMutation.mutate({ price_id: price.id });
                    }}
                  >
                    {isLoading ? "Loading..." : "Buy"}
                  </Button>
                </div>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}
```

- [ ] **Step 2: Type-check**

```bash
cd ui && npm run typecheck
```

Expected: no errors. If `rpsGameQueries` is not a named export from `rps-game-queries.tsx`, check with:

```bash
grep "export" ui/src/lib/rps-game-queries.tsx | head -5
```

The export should be `export const rpsGameQueries = { ... }`. If instead `getLedgerBalance` is exported as a standalone function, import it directly:

```typescript
import { getLedgerBalance } from "@/lib/rps-game-queries";
// and call: getLedgerBalance({ token: token! })
```

- [ ] **Step 3: Commit**

```bash
git add ui/src/pages/settings/points-settings.tsx
git commit -m "feat(ui): add points settings page"
```

---

### Task 4: Register routes in `App.tsx`

**Files:**
- Modify: `ui/src/App.tsx:53` (add two imports)
- Modify: `ui/src/App.tsx:214-220` (settings route block)

- [ ] **Step 1: Add imports for `BillingSettingPage` and `PointsSettingsPage`**

After line 53 (`import AccountSettingsPage from "./pages/settings/general-settings";`), add:

```typescript
import AccountSettingsPage from "./pages/settings/general-settings";
import BillingSettingPage from "./pages/settings/billing-settings";
import PointsSettingsPage from "./pages/settings/points-settings";
```

- [ ] **Step 2: Update the settings route block**

Lines 212–220 currently read:

```tsx
                {/* <Route path="billing" element={<BillingSettingPage />} /> */}

                <Route element={<PageSectionLayout title="Account Settings" />}>
                  <Route path="settings" element={<AccountSettingsPage />} />
                  {/* <Route
                    path="settings/billing"
                    element={<BillingSettingPage />}
                  /> */}
                </Route>
```

Replace with:

```tsx
                <Route element={<PageSectionLayout title="Account Settings" />}>
                  <Route path="settings" element={<AccountSettingsPage />} />
                  <Route
                    path="settings/billing"
                    element={<BillingSettingPage />}
                  />
                  <Route
                    path="settings/points"
                    element={<PointsSettingsPage />}
                  />
                </Route>
```

(The orphaned `{/* <Route path="billing" ... */}` comment on line 212 is removed as part of this replacement.)

- [ ] **Step 3: Type-check**

```bash
cd ui && npm run typecheck
```

Expected: no errors.

- [ ] **Step 4: Commit**

```bash
git add ui/src/App.tsx
git commit -m "feat(ui): register billing and points settings routes"
```
