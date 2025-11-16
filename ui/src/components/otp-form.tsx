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
import { zodResolver } from "@hookform/resolvers/zod";
import { Controller, useForm } from "react-hook-form";
import z from "zod";

type OTPFormProps = {
  mutate: (data: z.infer<typeof otpSchema>) => void;
};

const otpSchema = z.object({
  otp: z.string().length(6, "OTP must be 6 digits"),
});

export function OTPForm({
  mutate,
  ...props
}: React.ComponentProps<typeof Card> & OTPFormProps) {
  const form = useForm<z.infer<typeof otpSchema>>({
    resolver: zodResolver(otpSchema),
    defaultValues: {
      otp: "",
    },
  });
  function onUpdateSubmit(data: z.infer<typeof otpSchema>) {
    mutate(data);
  }
  return (
    <Card {...props}>
      <CardHeader>
        <CardTitle>Enter verification code</CardTitle>
        <CardDescription>We sent a 6-digit code to your email.</CardDescription>
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
                    <FieldLabel htmlFor="otp">Verification code</FieldLabel>
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
                    <Button type="submit" form="otp-form">
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
  );
}
