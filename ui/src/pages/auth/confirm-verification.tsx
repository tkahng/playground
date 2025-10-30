import { Button } from "@/components/ui/button";
import { confirmVerification } from "@/lib/api";
import { useQuery } from "@tanstack/react-query";
import { useSearchParams } from "react-router";
// /payment/success?sessionId
export default function ConfirmVerification() {
  const [searchParams] = useSearchParams();
  const token = searchParams.get("token");

  const { isPending, isError, error } = useQuery({
    queryKey: ["confirm-verification"],
    queryFn: async () => {
      if (!token) {
        throw new Error("Missing token");
      }
      await confirmVerification(token);
      return true;
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
      <p>Thank you for your verifying your email.</p>
      <Button asChild>
        <a href="/">Go Home</a>
      </Button>
    </div>
  );
}
