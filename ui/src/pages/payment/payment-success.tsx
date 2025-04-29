import { NexusAILandingHeader } from "@/components/nexus-landing-header";
import { RouteMap } from "@/components/route-map";
import { Button } from "@/components/ui/button";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { getCheckoutSession } from "@/lib/queries";
import { useQuery } from "@tanstack/react-query";
import { CheckCircle, Home, Settings } from "lucide-react";
import { Link, useSearchParams } from "react-router";
// /payment/success?sessionId

// interface PaymentDetails {
//   confirmationNumber: string;
//   date: string;
//   time: string;
//   plan: string;
//   nextBillingDate: string;
//   paymentMethod: string;
// }
export default function PaymentSuccessPage() {
  const { user, checkAuth } = useAuthProvider();
  //   const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const sessionId = searchParams.get("sessionId");
  const { data, isPending, isError, error } = useQuery({
    queryKey: ["subscription-by-session-id", sessionId],
    queryFn: async () => {
      await checkAuth(); // Ensure user is authenticated
      if (!sessionId) {
        throw new Error("Missing session ID");
      }
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token");
      }
      const sub = await getCheckoutSession(user.tokens.access_token, sessionId);
      return sub;
    },
  });

  if (isPending) {
    return <div>Loading...</div>;
  }

  if (isError) {
    return <div>Error: {error.message}</div>;
  }

  if (!data) {
    return <div>No data</div>;
  }

  return (
    <div className="flex flex-col min-h-screen">
      <NexusAILandingHeader />
      <main className="flex-1 flex items-center justify-center p-4">
        <div className="max-w-md w-full space-y-8">
          <div className="text-center">
            <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-green-100 dark:bg-green-900 mb-4">
              <CheckCircle className="h-8 w-8 text-green-600 dark:text-green-300" />
            </div>
            <h1 className="text-3xl font-bold">Payment Successful!</h1>
            <p className="text-muted-foreground mt-2">
              Thank you for your payment. Your transaction has been completed.
            </p>
          </div>

          <div className="flex flex-col sm:flex-row gap-2">
            <Button variant="outline" className="flex-1" asChild>
              <Link to={RouteMap.DASHBOARD}>
                <Home className="mr-2 h-4 w-4" />
                Dashboard
              </Link>
            </Button>
            <Button className="flex-1" asChild>
              <Link to={RouteMap.BILLING_SETTINGS}>
                <Settings className="mr-2 h-4 w-4" />
                Manage Subscription
              </Link>
            </Button>
          </div>
        </div>
      </main>
      <footer className="border-t">
        <div className="container px-4 md:px-6 py-8">
          <p className="text-xs text-center">
            © 2023 NexusAI. All rights reserved.
          </p>
        </div>
      </footer>
    </div>
  );
}
