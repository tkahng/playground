import { OTPForm } from "@/components/otp-form";
import { RouteMap } from "@/components/route-map";
import { Button } from "@/components/ui/button";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { confirmVerificationOtp } from "@/lib/api";
import { GetError } from "@/lib/error";
import { useMutation } from "@tanstack/react-query";
import { Link } from "react-router";
import { toast } from "sonner";

export default function VerifyEmailPage() {
  const { user, checkAuth } = useAuthProvider();
  // const navigate = useNavigate();
  const mutation = useMutation({
    mutationFn: async ({ otp }: { otp: string }) => {
      if (!user?.tokens.access_token) {
        throw new Error("User is not authenticated");
      }
      await confirmVerificationOtp(user.tokens.access_token, otp);
    },
    onSuccess: () => {
      checkAuth().finally(() => {
        toast.success("Email verified successfully!");
      });
    },
    onError: (error) => {
      const err = GetError(error);
      toast.error(`Failed to verify email: ${err.detail || err.title}`);
    },
  });

  if (!user) {
    return (
      <div className="flex min-h-svh w-full items-center justify-center p-6 md:p-10">
        <div className="w-full max-w-xs">
          <div>You are not logged in. Please sign in to verify your email.</div>
          <Button asChild className="mt-4">
            <Link to={RouteMap.SIGNIN}>Sign In</Link>
          </Button>
        </div>
      </div>
    );
  }
  if (user.user.email_verified_at) {
    return (
      <div className="flex min-h-svh w-full items-center justify-center p-6 md:p-10">
        <div className="w-full max-w-xs">
          <div>Your email is already verified.</div>
          <Button asChild className="mt-4">
            <Link to={RouteMap.ACCOUNT_DASHBOARD}>Go to Dashboard</Link>
          </Button>
        </div>
      </div>
    );
  }

  if (mutation.isPending) {
    return (
      <div className="flex min-h-svh w-full items-center justify-center p-6 md:p-10">
        <div className="w-full max-w-xs">
          <div>Verifying your email...</div>
        </div>
      </div>
    );
  }
  if (mutation.isError) {
    return (
      <div className="flex min-h-svh w-full items-center justify-center p-6 md:p-10">
        <div className="w-full max-w-xs">
          <div>Error: {mutation.error.message}</div>
          <Button onClick={() => mutation.reset()}>Try Again</Button>
        </div>
      </div>
    );
  }
  if (mutation.isSuccess) {
    return (
      <div className="flex min-h-svh w-full items-center justify-center p-6 md:p-10">
        <div className="w-full max-w-xs">
          <div>Email verified successfully!</div>
          <Button asChild>
            <Link to={RouteMap.ACCOUNT_DASHBOARD}>Continue</Link>
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="flex min-h-svh w-full items-center justify-center p-6 md:p-10">
      <div className="w-full max-w-xs">
        <OTPForm mutate={mutation.mutate} />
      </div>
    </div>
  );
}
