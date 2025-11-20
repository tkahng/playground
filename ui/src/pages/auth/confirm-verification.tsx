import { RouteMap } from "@/components/route-map";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Field, FieldGroup } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { confirmVerification } from "@/lib/api";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { useState } from "react";
import { Controller, useForm } from "react-hook-form";
import { Link, useSearchParams } from "react-router";
import { toast } from "sonner";
import z from "zod";

const formSchema = z.object({
  token: z.string().nonempty(),
});
export default function ConfirmVerification() {
  const [searchParams] = useSearchParams();
  const token = searchParams.get("token");
  const { user } = useAuthProvider();
  const [confirmed, setConfirmed] = useState(false);
  const mutation = useMutation({
    mutationFn: async ({ token }: z.infer<typeof formSchema>) => {
      await confirmVerification(token);
    },
    onSuccess: () => {
      setConfirmed(true);
    },
    onError: (error) => {
      console.error(error);
      toast.error("Error confirming email", {
        description: "Please try again",
      });
    },
  });
  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      token: token || "",
    },
  });
  const onSubmit = (values: z.infer<typeof formSchema>) => {
    mutation.mutate(values);
  };
  if (!token) {
    return (
      <div className="flex w-full flex-col items-center justify-center">
        <Card>
          <CardHeader>
            <CardTitle>Invalid Token</CardTitle>
          </CardHeader>
          <CardContent>
            <p>The provided token is invalid.</p>
          </CardContent>
          <CardFooter>
            <Button asChild>
              <Link to={RouteMap.HOME}>Go Home</Link>
            </Button>
          </CardFooter>
        </Card>
      </div>
    );
  }
  return (
    <div className="flex-row">
      <div className="flex min-h-svh w-full items-center justify-center p-6 md:p-10">
        <div className="w-full max-w-xs">
          {!confirmed && (
            <Card>
              <CardHeader>
                <CardTitle>Confirm Email</CardTitle>
                <CardDescription>
                  Click the button below to confirm your email.
                </CardDescription>
              </CardHeader>
              <CardContent>
                <form
                  id="form-rhf-demo"
                  onSubmit={form.handleSubmit(onSubmit)}
                  hidden
                >
                  <FieldGroup hidden>
                    <Controller
                      name="token"
                      control={form.control}
                      render={({ field, fieldState }) => (
                        <Field data-invalid={fieldState.invalid} hidden>
                          <Input
                            {...field}
                            id="form-rhf-demo-title"
                            type="text"
                            defaultValue={token}
                            hidden
                            aria-invalid={fieldState.invalid}
                          />
                        </Field>
                      )}
                    />
                  </FieldGroup>
                </form>
              </CardContent>
              <CardFooter>
                <Field
                  orientation="horizontal"
                  className="flex items-center justify-center"
                >
                  <Button type="submit" form="form-rhf-demo">
                    Submit
                  </Button>
                </Field>
              </CardFooter>
            </Card>
          )}
          {confirmed && (
            <Card>
              <CardHeader>
                <CardTitle>Email Confirm Success</CardTitle>
              </CardHeader>
              <CardContent>
                <p>Thank you for your verifying your email.</p>
              </CardContent>
              <CardFooter className="flex items-center justify-center">
                {user && (
                  <Button asChild>
                    <Link to={RouteMap.ACCOUNT_DASHBOARD}>Go Home</Link>
                  </Button>
                )}
                {!user && (
                  <Button asChild>
                    <Link to={RouteMap.SIGNIN}>Sign In</Link>
                  </Button>
                )}
              </CardFooter>
            </Card>
          )}
        </div>
      </div>
    </div>
  );
}
