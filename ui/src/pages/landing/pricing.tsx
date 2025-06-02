import Pricing from "@/components/pricing/pricing";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { getProductsWithPrices, getUserSubscriptions } from "@/lib/queries";
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
      let userSubs = null;
      if (user) {
        userSubs = await getUserSubscriptions(user.tokens.access_token);
      }
      const products = await getProductsWithPrices();
      return { products, userSubs };
    },
  });
  if (isPendingProducts) {
    return <div>Loading...</div>;
  }
  if (isErrorProducts) {
    return <div>Error: {errorProducts.message}</div>;
  }
  return (
    <Pricing
      user={user?.user}
      products={products?.products.data || []}
      subscription={products?.userSubs}
    />
  );
}
