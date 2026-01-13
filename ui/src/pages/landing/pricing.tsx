import PricingTeam from "@/components/pricing/pricing-team";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { getProductsWithPrices } from "@/lib/api";
import { useQuery } from "@tanstack/react-query";

export default function PricingPage() {
  const { user } = useAuthProvider();
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
      const products = await getProductsWithPrices();
      return { products, userSubs: null };
    },
  });
  if (isPendingProducts) {
    return <div>Loading...</div>;
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
