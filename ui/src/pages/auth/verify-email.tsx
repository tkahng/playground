import { RouteMap } from "@/components/route-map";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import {
  InputOTP,
  InputOTPGroup,
  InputOTPSlot,
} from "@/components/ui/input-otp";
import { Spinner } from "@/components/ui/spinner";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { confirmVerificationOtp } from "@/lib/api";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { useState } from "react";
import { Controller, useForm } from "react-hook-form";
import { Link, useNavigate } from "react-router";
import { toast } from "sonner";
import z from "zod";

const otpSchema = z.object({
  otp: z.string().length(6, "OTP must be 6 digits"),
});
export default function VerifyEmailPage() {
  const { user, checkAuth } = useAuthProvider();
  const [isPending, setIsPending] = useState(false);
  const navigate = useNavigate();
  const form = useForm<z.infer<typeof otpSchema>>({
    resolver: zodResolver(otpSchema),
    defaultValues: {
      otp: "",
    },
  });
  const mutation = useMutation({
    mutationFn: async ({ otp }: { otp: string }) => {
      setIsPending(true);
      if (!user?.tokens.access_token) {
        throw new Error("User is not authenticated");
      }
      await confirmVerificationOtp(user.tokens.access_token, otp);
    },
    onSuccess: () => {
      checkAuth().finally(() => {
        setIsPending(false);
        toast.success("Email verified successfully!");
      });
    },
    onError: (error) => {
      setIsPending(false);
      form.reset();
      toast.error(`Failed to verify email: ${error.message}`);
    },
  });

  function onUpdateSubmit(data: z.infer<typeof otpSchema>) {
    mutation.mutate(data);
  }
  if (!user) {
    navigate(RouteMap.SIGNIN);
    return;
  }

  if (user.user.email_verified_at) {
    navigate(RouteMap.ACCOUNT_DASHBOARD);
    return;
  }

  return (
    <div className="flex-row">
      <div className="flex min-h-svh w-full items-center justify-center p-6 md:p-10">
        <div className="w-full max-w-xs">
          <Card>
            <CardHeader>
              <CardTitle>Enter verification code</CardTitle>
              <CardDescription>
                We sent a 6-digit code to your email.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <form id="otp-form" onSubmit={form.handleSubmit(onUpdateSubmit)}>
                <Controller
                  name="otp"
                  control={form.control}
                  render={({ field }) => {
                    return (
                      <FieldGroup>
                        <Field>
                          <FieldLabel htmlFor="otp">
                            Verification code
                          </FieldLabel>
                          <InputOTP
                            maxLength={6}
                            id="otp"
                            required
                            name={field.name}
                            value={field.value}
                            onChange={field.onChange}
                          >
                            <InputOTPGroup className="gap-2.5 *:data-[slot=input-otp-slot]:rounded-md *:data-[slot=input-otp-slot]:border">
                              <InputOTPSlot index={0} />
                              <InputOTPSlot index={1} />
                              <InputOTPSlot index={2} />
                              <InputOTPSlot index={3} />
                              <InputOTPSlot index={4} />
                              <InputOTPSlot index={5} />
                            </InputOTPGroup>
                          </InputOTP>
                          <FieldDescription>
                            Enter the 6-digit code sent to your email.
                          </FieldDescription>
                        </Field>
                        <FieldGroup>
                          <Button
                            type="submit"
                            form="otp-form"
                            disabled={isPending}
                          >
                            {isPending && <Spinner />}
                            Verify
                          </Button>
                          <FieldDescription className="text-center">
                            Didn&apos;t receive the code? <a href="#">Resend</a>
                          </FieldDescription>
                        </FieldGroup>
                      </FieldGroup>
                    );
                  }}
                />
              </form>
            </CardContent>
          </Card>
          <div className="flex justify-end px-4 md:px-6 lg:px-8 py-2">
            <Link
              className="text-muted-foreground"
              to={RouteMap.ACCOUNT_DASHBOARD}
            >
              skip
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
}
