import { Button } from "@/components/ui/button";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { createCheckoutSession } from "@/lib/queries";
import { cn } from "@/lib/utils";

import { ProductWithPrices, SubscriptionWithPrice, User } from "@/schema.types";
import { useMutation } from "@tanstack/react-query";

import { useState } from "react";
import { createSearchParams, useNavigate } from "react-router";
import { z } from "zod";

interface Props {
  user: User | null | undefined;
  products: ProductWithPrices[];
  subscription: SubscriptionWithPrice | null;
}

type BillingInterval = "lifetime" | "year" | "month";

export const formSchema = z.object({
  price_id: z.string().min(2, {
    message: "name must be at least 2 characters.",
  }),
});

export default function Pricing({ products, subscription }: Props) {
  const { user } = useAuthProvider();
  const intervals = Array.from(
    new Set(
      products.flatMap((product) =>
        product?.prices?.map((price) => price?.interval)
      )
    )
  );
  //   const router = useRouter();
  const navigate = useNavigate();
  const [billingInterval, setBillingInterval] =
    useState<BillingInterval>("month");
  const [priceIdLoading, setPriceIdLoading] = useState<string>();
  // const { pathname: currentPath } = useLocation();

  const mutation = useMutation({
    mutationFn: async (values: z.infer<typeof formSchema>) => {
      setPriceIdLoading(values.price_id);
      if (!user) {
        setPriceIdLoading(undefined);
        return navigate({
          pathname: "/signin",
          search: createSearchParams({
            redirect_to: window.location.pathname + window.location.search,
          }).toString(),
        });
      }
      const { url } = await createCheckoutSession(
        user.tokens.access_token,
        values
      );
      setPriceIdLoading(undefined);
      window.location.href = url;
    },
  });
  function onSubmit(values: z.infer<typeof formSchema>) {
    mutation.mutate(values);
  }

  if (!products.length) {
    return (
      <section className="">
        <div className="max-w-6xl px-4 py-8 mx-auto sm:py-24 sm:px-6 lg:px-8">
          <div className="sm:flex sm:flex-col sm:align-center"></div>
          <p className="text-4xl font-extrabold sm:text-center sm:text-6xl">
            No subscription pricing plans found. Create them in your{" "}
            <a
              className="text-pink-500 underline"
              href="https://dashboard.stripe.com/products"
              rel="noopener noreferrer"
              target="_blank"
            >
              Stripe Dashboard
            </a>
            .
          </p>
        </div>
        {/* <LogoCloud /> */}
      </section>
    );
  } else {
    return (
      <section className="">
        <div className="max-w-6xl px-4 py-8 mx-auto sm:py-24 sm:px-6 lg:px-8">
          <div className="sm:flex sm:flex-col sm:align-center">
            <h1 className="text-4xl font-extrabold text-primary sm:text-center sm:text-6xl">
              Pricing Plans
            </h1>
            <p className="max-w-2xl m-auto mt-5 text-xl text-secondary-foreground sm:text-center sm:text-2xl">
              Start building for free, then add a site plan to go live. Account
              plans unlock additional features.
            </p>
            <div className="relative self-center mt-6 bg-primary-foreground rounded-lg p-0.5 flex sm:mt-8 border">
              {intervals.includes("month") && (
                <button
                  onClick={() => setBillingInterval("month")}
                  type="button"
                  className={`${
                    billingInterval === "month"
                      ? "relative w-1/2 bg-zinc-700 border-zinc-800 shadow-sm text-white"
                      : "ml-0.5 relative w-1/2 border border-transparent text-zinc-400"
                  } rounded-md m-1 py-2 text-sm font-medium whitespace-nowrap focus:outline-none focus:ring-2 focus:ring-pink-500 focus:ring-opacity-50 focus:z-10 sm:w-auto sm:px-8`}
                >
                  Monthly billing
                </button>
              )}
              {intervals.includes("year") && (
                <button
                  onClick={() => setBillingInterval("year")}
                  type="button"
                  className={`${
                    billingInterval === "year"
                      ? "relative w-1/2 bg-zinc-700 border-zinc-800 shadow-sm text-white"
                      : "ml-0.5 relative w-1/2 border border-transparent text-zinc-400"
                  } rounded-md m-1 py-2 text-sm font-medium whitespace-nowrap focus:outline-none focus:ring-2 focus:ring-pink-500 focus:ring-opacity-50 focus:z-10 sm:w-auto sm:px-8`}
                >
                  Yearly billing
                </button>
              )}
            </div>
          </div>
          <div className="mt-12 space-y-0 sm:mt-16 flex flex-wrap justify-center gap-6 lg:max-w-4xl lg:mx-auto xl:max-w-none xl:mx-0">
            {products.map((product) => {
              const price = product?.prices?.find(
                (price) => price.interval === billingInterval
              );
              if (!price) return null;
              const priceString = new Intl.NumberFormat("en-US", {
                style: "currency",
                currency: price.currency!,
                minimumFractionDigits: 0,
              }).format((price?.unit_amount || 0) / 100);
              return (
                <div
                  key={product.id}
                  className={cn(
                    "flex flex-col rounded-lg shadow-sm divide-y divide-zinc-600",
                    {
                      "border border-pink-500": subscription
                        ? product.name === subscription?.price?.product?.name
                        : product.name === "Freelancer",
                    },
                    "flex-1", // This makes the flex item grow to fill the space
                    "basis-1/3", // Assuming you want each card to take up roughly a third of the container's width
                    "max-w-xs" // Sets a maximum width to the cards to prevent them from getting too large
                  )}
                >
                  <div className="p-6">
                    <h2 className="text-2xl font-semibold leading-6 ">
                      {product.name}
                    </h2>
                    <p className="mt-4 text-muted-foreground">
                      {product.description}
                    </p>
                    <p className="mt-8">
                      <span className="text-5xl font-extrabold">
                        {priceString}
                      </span>
                      <span className="text-base font-medium">
                        /{billingInterval}
                      </span>
                    </p>
                    <Button
                      //   variant="slim"
                      type="submit"
                      disabled={priceIdLoading === price.id}
                      // loading={priceIdLoading === price.id}
                      onClick={() => onSubmit({ price_id: price.id })}
                      className="block w-full py-2 mt-8 text-sm font-semibold text-center rounded-md"
                    >
                      {subscription ? "Manage" : "Subscribe"}
                    </Button>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      </section>
    );
  }
}
