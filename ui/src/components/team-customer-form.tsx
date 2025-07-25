import { Button } from "@/components/ui/button";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useTeam } from "@/hooks/use-team";
import {
  createTeamBillingPortalSession,
  getProductsWithPrices,
  getUserSubscriptions,
} from "@/lib/api";
import { SubscriptionWithPrice } from "@/schema.types";
import { useQuery } from "@tanstack/react-query";
import { ReactNode, useState } from "react";
import { Link } from "react-router";
import { toast } from "sonner";
import PricingMini from "./pricing/pricing-mini";

type SubscriptionWithPriceAndProduct = SubscriptionWithPrice;

interface Props {
  subscription: SubscriptionWithPriceAndProduct | null;
}

export default function TeamCustomerForm({ subscription }: Props) {
  //   const router = useRouter();
  const { user } = useAuthProvider();
  const { team, teamMember } = useTeam();
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
  // const { pathname: currentPath } = useLocation();
  const [isSubmitting, setIsSubmitting] = useState(false);

  const subscriptionPrice =
    subscription &&
    new Intl.NumberFormat("en-US", {
      style: "currency",
      currency: subscription?.price?.currency,
      minimumFractionDigits: 0,
    }).format((subscription?.price?.unit_amount || 0) / 100);

  const handleStripePortalRequest = async () => {
    setIsSubmitting(true);
    if (!user) {
      setIsSubmitting(false);
      toast.error("Please login to open the customer portal.");
      return;
    }
    if (!team) {
      setIsSubmitting(false);
      toast.error("Please join a team to open the customer portal.");
      return;
    }
    const redirectUrl = await createTeamBillingPortalSession(
      user.tokens.access_token,
      team?.id
    );
    window.location.href = redirectUrl;
    setIsSubmitting(false);
  };
  if (isPendingProducts) {
    return <div>Loading...</div>;
  }
  if (isErrorProducts) {
    return <div>{errorProducts.message}</div>;
  }
  return (
    <div>
      <PricingMini
        user={user?.user}
        products={products?.products.data || []}
        subscription={products?.userSubs}
      />
      <Card
        title="Your Plan"
        description={
          subscription
            ? `You are currently on the ${subscription?.price?.product?.name} plan.`
            : "You are not currently subscribed to any plan."
        }
        footer={
          subscription && (
            <div className="flex flex-col items-start justify-between sm:flex-row sm:items-center">
              <p className="pb-4 sm:pb-0">
                Manage your subscription on Stripe.
              </p>
              <Button
                // variant="slim"
                onClick={handleStripePortalRequest}
                // loading={isSubmitting}
                disabled={isSubmitting || teamMember?.role !== "owner"}
              >
                Open customer portal
              </Button>
            </div>
          )
        }
      >
        <div className="mt-8 mb-4 text-xl font-semibold">
          {subscription ? (
            `${subscriptionPrice}/${subscription?.price?.interval}`
          ) : (
            <Link to="/pricing">Choose your plan</Link>
          )}
        </div>
      </Card>
    </div>
  );
}

interface CardProps {
  title: string;
  description?: string;
  footer?: ReactNode;
  children: ReactNode;
}

export function Card({ title, description, footer, children }: CardProps) {
  return (
    <div className="w-full max-w-3xl m-auto my-8 border rounded-md p border-zinc-700">
      <div className="px-5 py-4">
        <h3 className="mb-1 text-2xl font-medium">{title}</h3>
        <p className="text-zinc-300">{description}</p>
        {children}
      </div>
      {footer && (
        <div className="p-4 border-t rounded-b-md border-zinc-700 bg-zinc-900 text-zinc-500">
          {footer}
        </div>
      )}
    </div>
  );
}
