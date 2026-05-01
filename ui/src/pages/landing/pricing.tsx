import { CenteredSpinner } from "@/components/centered-spinner";
import PricingTeam from "@/components/pricing/pricing-team";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useOnboardingProgress } from "@/hooks/use-onboarding-progress";
import { getProductsWithPrices } from "@/lib/api";
import { useQuery } from "@tanstack/react-query";
import { useEffect } from "react";

export default function PricingPage() {
  const { user } = useAuthProvider();
  const { markStep } = useOnboardingProgress();
  useEffect(() => {
    markStep("visitedPricing");
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);
  const {
    data: products,
    isPending: isPendingProducts,
    isError: isErrorProducts,
    error: errorProducts,
  } = useQuery({
    queryKey: ["stripe-products-with-prices"],
    queryFn: async () => {
      // let userSubs = null;
      // if (user) {
      //   userSubs = await getUserSubscriptions(user.tokens.access_token);
      // }
      const products = await getProductsWithPrices(undefined, "subscription");
      return { products, userSubs: null };
    },
  });
  if (isPendingProducts) {
    return <CenteredSpinner />;
  }
  if (isErrorProducts) {
    return <div>Error: {errorProducts.message}</div>;
  }
  return (
    <PricingTeam
      user={user?.user}
      products={products?.products.data || []}
      subscription={products?.userSubs}
    />
  );
}
