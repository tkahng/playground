import { useAuthProvider } from "@/hooks/use-auth-provider";
import { getCheckoutSession } from "@/lib/queries";
import { useQuery } from "@tanstack/react-query";
import { useSearchParams } from "react-router";
// /payment/success?sessionId
export default function PaymentSuccessPage() {
  const { user } = useAuthProvider();
  //   const navigate = useNavigate();
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
    return <div>Error: {error.message}</div>;
  }

  if (!data) {
    return <div>No data</div>;
  }

  return (
    <div className="flex w-full flex-col items-center justify-center">
      <h2>Payment Success</h2>
      <p>Thank you for your payment.</p>
    </div>
  );
}
