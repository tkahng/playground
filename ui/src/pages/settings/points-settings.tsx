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

  // Client-side guard: the API already filters by metadata_type=points,
  // but this ensures stale cache data doesn't include non-points products.
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
          <p className="text-muted-foreground">
            No points packages available.{" "}
            <a href="/admin/products" className="underline">
              Manage in admin
            </a>
          </p>
        ) : (
          <div className="grid grid-cols-3 gap-4">
            {prices.map((price, index) => {
              const label = CARD_LABELS[index] ?? `Package ${index + 1}`;
              const isPopular = index === 1;
              const parsed = parseInt(price.metadata?.points_amount ?? "", 10);
              const pointsAmount = isNaN(parsed) ? null : parsed;
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
