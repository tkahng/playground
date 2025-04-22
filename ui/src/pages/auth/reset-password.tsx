"use client";

import { RouteMap } from "@/components/route-map";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { requestPasswordReset } from "@/lib/queries";
import { useMutation } from "@tanstack/react-query";
import {
  AlertCircle,
  ArrowLeft,
  Brain,
  CheckCircle,
  Loader2,
  Mail,
} from "lucide-react";
import { useState } from "react";
import { Link } from "react-router";
import { z } from "zod";
// Form validation schema
const formSchema = z.object({
  email: z.string().email({ message: "Please enter a valid email address" }),
});

export default function ResetPasswordRequestPage() {
  const [email, setEmail] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [formState, setFormState] = useState<{
    status: "idle" | "success" | "error";
    message: string;
  }>({
    status: "idle",
    message: "",
  });
  const [validationError, setValidationError] = useState("");

  const mutation = useMutation({
    mutationFn: async (email: string) => {
      await requestPasswordReset(email);
    },
    // onSuccess: () => {
    //   setFormState({
    //     status: "success",
    //     message: "Password reset link sent! Please check your email inbox.",
    //   });
    // },
    // onError: () => {
    //   setFormState({
    //     status: "error",
    //     message: "Failed to send reset email. Please try again.",
    //   });
    // },
  });
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    // Reset states
    setValidationError("");
    setFormState({ status: "idle", message: "" });

    // Validate email
    try {
      formSchema.parse({ email });
    } catch (error) {
      if (error instanceof z.ZodError) {
        setValidationError(error.errors[0].message);
        return;
      }
    }

    setIsLoading(true);

    try {
      mutation.mutate(email);
      setFormState({
        status: "success",
        message: "Password reset link sent! Please check your email inbox.",
      });
    } catch (error) {
      setFormState({
        status: "error",
        message:
          error instanceof Error
            ? error.message
            : "Failed to send reset email. Please try again.",
      });
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="flex flex-col min-h-screen bg-gray-50 dark:bg-gray-900">
      <header className="px-4 lg:px-6 h-14 flex items-center border-b bg-white dark:bg-gray-800">
        <Link className="flex items-center justify-center" to={RouteMap.HOME}>
          <Brain className="h-6 w-6 text-primary" />
          <span className="ml-2 text-2xl font-bold text-primary">NexusAI</span>
        </Link>
      </header>
      <main className="flex-1 flex items-center justify-center p-4">
        <Card className="w-full max-w-md">
          <CardHeader className="space-y-1">
            <CardTitle className="text-2xl font-bold">
              Reset your password
            </CardTitle>
            <CardDescription>
              Enter your email address and we'll send you a link to reset your
              password.
            </CardDescription>
          </CardHeader>
          <CardContent>
            {formState.status === "success" ? (
              <Alert className="bg-green-50 border-green-200 dark:bg-green-900/20 dark:border-green-900">
                <CheckCircle className="h-4 w-4 text-green-600 dark:text-green-400" />
                <AlertTitle>Email sent</AlertTitle>
                <AlertDescription>{formState.message}</AlertDescription>
              </Alert>
            ) : formState.status === "error" ? (
              <Alert variant="destructive">
                <AlertCircle className="h-4 w-4" />
                <AlertTitle>Error</AlertTitle>
                <AlertDescription>{formState.message}</AlertDescription>
              </Alert>
            ) : (
              <form onSubmit={handleSubmit} className="space-y-4">
                <div className="space-y-2">
                  <label
                    htmlFor="email"
                    className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                  >
                    Email
                  </label>
                  <div className="relative">
                    <Mail className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                    <Input
                      id="email"
                      type="email"
                      placeholder="name@example.com"
                      className="pl-10"
                      value={email}
                      onChange={(e) => setEmail(e.target.value)}
                      disabled={isLoading}
                      required
                    />
                  </div>
                  {validationError && (
                    <p className="text-sm text-destructive">
                      {validationError}
                    </p>
                  )}
                </div>
                <Button type="submit" className="w-full" disabled={isLoading}>
                  {isLoading ? (
                    <>
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      Sending...
                    </>
                  ) : (
                    "Send Reset Link"
                  )}
                </Button>
              </form>
            )}
          </CardContent>
          <CardFooter className="flex flex-col space-y-4">
            <div className="text-sm text-center text-muted-foreground">
              {formState.status === "success" ? (
                <p>
                  Didn't receive an email? Check your spam folder or{" "}
                  <Button
                    variant="link"
                    className="p-0 h-auto"
                    onClick={() =>
                      setFormState({ status: "idle", message: "" })
                    }
                  >
                    try again
                  </Button>
                </p>
              ) : (
                <p>
                  Remember your password?{" "}
                  <Link
                    to={RouteMap.SIGNIN}
                    className="text-primary hover:underline"
                  >
                    Back to login
                  </Link>
                </p>
              )}
            </div>
            <Button variant="outline" size="sm" className="w-full" asChild>
              <Link to={RouteMap.SIGNIN}>
                <ArrowLeft className="mr-2 h-4 w-4" />
                Back to Login
              </Link>
            </Button>
          </CardFooter>
        </Card>
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
