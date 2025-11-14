import { OTPForm } from "@/components/otp-form";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { confirmVerificationOtp } from "@/lib/api";
import { GetError } from "@/lib/get-error";
import { useMeQuery } from "@/lib/queries";
import { useMutation } from "@tanstack/react-query";
import { useState } from "react";
import { useNavigate } from "react-router";
import { toast } from "sonner";

export default function VerifyEmail() {
  const { user } = useAuthProvider();
  const [success, setSuccess] = useState(false);
  const navigate = useNavigate();
  const { isPending, isError, error } = useMeQuery();
  const mutation = useMutation({
    mutationFn: async ({ otp }: { otp: string }) => {
      if (!user?.tokens.access_token) {
        throw new Error("User is not authenticated");
      }
      await confirmVerificationOtp(user.tokens.access_token, otp);
    },
    onSuccess: () => {
      setSuccess(true);
      toast.success("Email verified successfully!");
    },
    onError: (error: Error) => {
      const err = GetError(error);
      toast.error(`Failed to verify email: ${err.detail || err.title}`);
    },
  });

  if (!user) {
    navigate("/");
  }
  if (isPending) {
    return <div>Loading...</div>;
  }

  if (isError) {
    return <div>Error: {error.message}</div>;
  }

  return (
    <div className="flex min-h-svh w-full items-center justify-center p-6 md:p-10">
      <div className="w-full max-w-xs">
        {success ? (
          <OTPForm mutate={mutation.mutate} />
        ) : (
          <div>Email verified successfully!</div>
        )}
      </div>
    </div>
  );
}
