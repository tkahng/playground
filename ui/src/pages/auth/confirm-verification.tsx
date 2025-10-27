import { Button } from "@/components/ui/button";
import { confirmVerification } from "@/lib/api";
import { useMutation } from "@tanstack/react-query";
import { useSearchParams } from "react-router";
// /payment/success?sessionId
export default function ConfirmVerification() {
  const [searchParams] = useSearchParams();
  const token = searchParams.get("token");
  const redirect = searchParams.get("redirect_to");

  const { isPending, isError, error } = useMutation({
    mutationKey: ["confirm-verification"],
    mutationFn: async () => {
      if (!token) {
        throw new Error("Missing token");
      }
      await confirmVerification(token);
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
      <p>Thank you for your verifying your email.</p>
      <Button asChild>
        <a href="/">Go Home</a>
      </Button>
    </div>
  );
}
