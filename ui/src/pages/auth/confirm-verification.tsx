import { confirmVerification } from "@/lib/queries";
import { useQuery } from "@tanstack/react-query";
import { useSearchParams } from "react-router";
// /payment/success?sessionId
export default function ConfirmVerification() {
  const [searchParams] = useSearchParams();
  const token = searchParams.get("token");
  const type = searchParams.get("type");
  const redirect = searchParams.get("redirect_to");

  const { isPending, isError, error } = useQuery({
    queryKey: ["confirm-verification"],
    queryFn: async () => {
      if (!token) {
        throw new Error("Missing session ID");
      }
      if (!type) {
        throw new Error("Missing access token");
      }
      await confirmVerification(token, type);
      if (redirect) {
        window.location.href = redirect;
      }
    },
  });

  if (isPending) {
    return <div>Loading...</div>;
  }

  if (isError) {
    return <div>Error: {error.message}</div>;
  }

  return (
    <div className="flex w-full flex-col items-center justify-center">
      <h2>Email Confirm Success</h2>
      <p>Thank you for your payment.</p>
    </div>
  );
}
