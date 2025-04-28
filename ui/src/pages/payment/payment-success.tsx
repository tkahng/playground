import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { getCheckoutSession } from "@/lib/queries";
import { Separator } from "@radix-ui/react-dropdown-menu";
import { useQuery } from "@tanstack/react-query";
import {
  Brain,
  CheckCircle,
  Download,
  Home,
  MailOpen,
  Settings,
} from "lucide-react";
import { useState } from "react";
import { Link, useSearchParams } from "react-router";
// /payment/success?sessionId

interface PaymentDetails {
  confirmationNumber: string;
  date: string;
  time: string;
  plan: string;
  nextBillingDate: string;
  paymentMethod: string;
}
export default function PaymentSuccessPage() {
  const { user, checkAuth } = useAuthProvider();
  //   const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const sessionId = searchParams.get("sessionId");
  const [paymentDetails] = useState<PaymentDetails | null>(null);
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
    <div className="flex flex-col min-h-screen bg-gray-50 dark:bg-gray-900">
      <header className="px-4 lg:px-6 h-14 flex items-center border-b bg-white dark:bg-gray-800">
        <Link className="flex items-center justify-center" to="/">
          <Brain className="h-6 w-6 text-primary" />
          <span className="ml-2 text-2xl font-bold text-primary">NexusAI</span>
        </Link>
      </header>
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

          <Card>
            <CardHeader>
              <CardTitle>Payment Details</CardTitle>
              <CardDescription>
                Confirmation #{paymentDetails?.confirmationNumber}
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex justify-between">
                <span className="text-muted-foreground">Date</span>
                <span className="font-medium">{paymentDetails?.date}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Time</span>
                <span className="font-medium">{paymentDetails?.time}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Plan</span>
                <span className="font-medium">{paymentDetails?.plan}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Next Billing Date</span>
                <span className="font-medium">
                  {paymentDetails?.nextBillingDate}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Payment Method</span>
                <span className="font-medium">
                  {paymentDetails?.paymentMethod}
                </span>
              </div>

              <Separator className="my-2" />

              <div className="bg-muted p-3 rounded-md">
                <p className="text-sm">
                  A receipt has been sent to your email address. Please check
                  your inbox.
                </p>
              </div>
            </CardContent>
            <CardFooter className="flex flex-col space-y-2">
              <div className="flex flex-col sm:flex-row w-full gap-2">
                <Button variant="outline" className="flex-1">
                  <Download className="mr-2 h-4 w-4" />
                  Download Receipt
                </Button>
                <Button variant="outline" className="flex-1">
                  <MailOpen className="mr-2 h-4 w-4" />
                  Email Receipt
                </Button>
              </div>
            </CardFooter>
          </Card>

          <div className="flex flex-col sm:flex-row gap-2">
            <Button variant="outline" className="flex-1" asChild>
              <Link to="/dashboard">
                <Home className="mr-2 h-4 w-4" />
                Dashboard
              </Link>
            </Button>
            <Button className="flex-1" asChild>
              <Link to="/account/subscription">
                <Settings className="mr-2 h-4 w-4" />
                Manage Subscription
              </Link>
            </Button>
          </div>
        </div>
      </main>
      <footer className="border-t bg-gray-100 dark:bg-gray-800">
        <div className="container px-4 md:px-6 py-8">
          <p className="text-xs text-center text-gray-500 dark:text-gray-400">
            © 2023 NexusAI. All rights reserved.
          </p>
        </div>
      </footer>
    </div>
  );
}
