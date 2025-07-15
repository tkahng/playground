import { PlaygroundLandingHeader } from "@/components/nexus-landing-header";
import { PlaygroundMinimalFooter } from "@/components/nexus-minimal-footer";
import { Button } from "@/components/ui/button";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { GetError } from "@/lib/get-error";
import { getCheckoutSession } from "@/lib/queries";
import { useQuery } from "@tanstack/react-query";
import { CheckCircle, Home, Settings } from "lucide-react";
import { Link, useSearchParams } from "react-router";

export default function PaymentSuccessPage() {
  const { user } = useAuthProvider();
  const [searchParams] = useSearchParams();
  const sessionId = searchParams.get("sessionId");
  const { data, isPending, isError, error } = useQuery({
    queryKey: ["subscription-by-session-id", sessionId],
    queryFn: async () => {
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
    const err = GetError(error);
    return <div>Error: {err?.detail}</div>;
  }

  if (!data) {
    return <div>No data</div>;
  }
  const team = data?.stripe_customer?.team;
  if (!team) {
    return <div>No team</div>;
  }
  return (
    <div className="flex flex-col min-h-screen ">
      <div className="px-4 md:px-6 lg:px-8 py-2 items-center sticky top-0 z-50 w-full bg-background shadow-sm border-b">
        <PlaygroundLandingHeader />
      </div>
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
              <Link to={`/teams/${team?.slug}/dashboard`}>
                <Home className="mr-2 h-4 w-4" />
                Dashboard
              </Link>
            </Button>
            <Button className="flex-1" asChild>
              <Link to={`/teams/${team?.slug}/settings/billing`}>
                <Settings className="mr-2 h-4 w-4" />
                Manage Subscription
              </Link>
            </Button>
          </div>
        </div>
      </main>
      <PlaygroundMinimalFooter />
    </div>
  );
}
