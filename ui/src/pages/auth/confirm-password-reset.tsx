import { Link, useSearchParams } from "react-router";

import { Alert, AlertDescription } from "@/components/ui/alert";
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
import { Label } from "@/components/ui/label";
import { Progress } from "@/components/ui/progress";
import { checkPasswordReset, confirmPasswordReset } from "@/lib/queries";
import { useMutation, useQuery } from "@tanstack/react-query";
import {
  AlertCircle,
  ArrowRight,
  Brain,
  Check,
  Eye,
  EyeOff,
  Home,
  KeyRound,
} from "lucide-react";
import { useEffect, useState } from "react";
import { z } from "zod";

const formSchema = z.object({
  password: z.string().min(8),
  confirmPassword: z.string().min(8),
  token: z.string().min(1),
});

export default function ResetPasswordPage() {
  const [searchParams] = useSearchParams();
  const token = searchParams.get("token");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isSuccess, setIsSuccess] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const {
    isPending: isCheckPasswordResetPending,
    isError: isCheckPasswordResetError,
    error: checkPasswordResetError,
  } = useQuery({
    queryKey: ["check-password-reset", token],
    queryFn: async () => {
      if (!token) {
        throw new Error("Missing token");
      }
      await checkPasswordReset(token);
      return true;
    },
  });
  // Password strength calculation
  const calculateStrength = (password: string): number => {
    let strength = 0;
    if (password.length >= 8) strength += 25;
    if (/[A-Z]/.test(password)) strength += 25;
    if (/[0-9]/.test(password)) strength += 25;
    if (/[^A-Za-z0-9]/.test(password)) strength += 25;
    return strength;
  };

  const passwordStrength = calculateStrength(password);

  const getStrengthText = (strength: number): string => {
    if (strength === 0) return "No password entered";
    if (strength <= 25) return "Weak";
    if (strength <= 50) return "Fair";
    if (strength <= 75) return "Good";
    return "Strong";
  };

  const getStrengthColor = (strength: number): string => {
    if (strength <= 25) return "bg-destructive";
    if (strength <= 50) return "bg-amber-500";
    if (strength <= 75) return "bg-yellow-500";
    return "bg-green-500";
  };

  // const form = useForm<z.infer<typeof formSchema>>({
  //   resolver: zodResolver(formSchema),
  //   defaultValues: {
  //     password: "",
  //     confirmPassword: "",
  //     token: token || "",
  //   },
  // });

  const mutation = useMutation({
    mutationFn: async (data: z.infer<typeof formSchema>) => {
      return await confirmPasswordReset(
        data.token,
        data.password,
        data.confirmPassword
      );
    },
    onSuccess: () => {
      setIsSuccess(true);
      setIsSubmitting(false);
    },
    onError: (error) => {
      setError(error.message);
      setIsSubmitting(false);
    },
  });
  // function onSubmit(values: z.infer<typeof formSchema>) {
  //   setIsSubmitting(true);
  //   mutation.mutate(values);
  //   setIsSubmitting(false);
  // }
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    // Validate passwords
    if (password.length < 8) {
      setError("Password must be at least 8 characters long");
      return;
    }

    if (password !== confirmPassword) {
      setError("Passwords do not match");
      return;
    }

    if (passwordStrength < 75) {
      setError("Please choose a stronger password");
      return;
    }

    if (!token) {
      setError(
        "Invalid or expired reset link. Please request a new password reset."
      );
      return;
    }

    setIsSubmitting(true);

    try {
      // This would be replaced with your actual API call
      // await resetPassword(token, password)

      // Simulate API call
      // await new Promise((resolve) => setTimeout(resolve, 1500));
      mutation.mutate({
        token,
        password,
        confirmPassword,
      });

      // setIsSuccess(true);
    } catch (err) {
      setError(
        "Failed to reset password. Please try again or request a new reset link."
      );
    } finally {
      setIsSubmitting(false);
    }
  };

  if (isSuccess) {
    return (
      <div className="flex flex-col min-h-screen bg-gray-50 dark:bg-gray-900">
        <header className="px-4 lg:px-6 h-14 flex items-center border-b bg-white dark:bg-gray-800">
          <Link className="flex items-center justify-center" to="/">
            <Brain className="h-6 w-6 text-primary" />
            <span className="ml-2 text-2xl font-bold text-primary">
              NexusAI
            </span>
          </Link>
        </header>
        <main className="flex-1 flex items-center justify-center p-4">
          <Card className="max-w-md w-full">
            <CardHeader className="text-center">
              <div className="mx-auto rounded-full w-12 h-12 bg-green-100 dark:bg-green-900 flex items-center justify-center mb-4">
                <Check className="h-6 w-6 text-green-600 dark:text-green-300" />
              </div>
              <CardTitle className="text-2xl">
                Password Reset Successful
              </CardTitle>
              <CardDescription>
                Your password has been successfully updated.
              </CardDescription>
            </CardHeader>
            <CardContent className="text-center">
              <p className="text-muted-foreground">
                You can now log in with your new password.
              </p>
            </CardContent>
            <CardFooter className="flex flex-col space-y-2">
              <Button className="w-full" asChild>
                <Link to="/login">
                  <ArrowRight className="mr-2 h-4 w-4" />
                  Continue to Login
                </Link>
              </Button>
              <Button variant="outline" className="w-full" asChild>
                <Link to="/">
                  <Home className="mr-2 h-4 w-4" />
                  Return to Home
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
  useEffect(() => {
    if (isCheckPasswordResetError) {
      setError(checkPasswordResetError?.message || "An error occurred");
    }
  }, [isCheckPasswordResetError, checkPasswordResetError]);
  if (isCheckPasswordResetPending) {
    return <div>Loading...</div>;
  }
  return (
    <div className="flex flex-col min-h-screen bg-gray-50 dark:bg-gray-900">
      <header className="px-4 lg:px-6 h-14 flex items-center border-b bg-white dark:bg-gray-800">
        <Link className="flex items-center justify-center" to="/">
          <Brain className="h-6 w-6 text-primary" />
          <span className="ml-2 text-2xl font-bold text-primary">NexusAI</span>
        </Link>
      </header>
      <main className="flex-1 flex items-center justify-center p-4">
        <Card className="max-w-md w-full">
          <CardHeader>
            <div className="mx-auto rounded-full w-12 h-12 bg-primary/10 flex items-center justify-center mb-4">
              <KeyRound className="h-6 w-6 text-primary" />
            </div>
            <CardTitle className="text-2xl text-center">
              Reset Your Password
            </CardTitle>
            <CardDescription className="text-center">
              Create a new password for your NexusAI account
            </CardDescription>
          </CardHeader>
          <CardContent>
            {error && (
              <Alert variant="destructive" className="mb-4">
                <AlertCircle className="h-4 w-4" />
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            )}

            {!token && (
              <Alert variant="destructive" className="mb-4">
                <AlertCircle className="h-4 w-4" />
                <AlertDescription>
                  Invalid or expired reset link. Please request a new password
                  reset.
                </AlertDescription>
              </Alert>
            )}
            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="password">New Password</Label>
                <div className="relative">
                  <Input
                    id="password"
                    type={showPassword ? "text" : "password"}
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    placeholder="Enter your new password"
                    disabled={isSubmitting || !token}
                    className="pr-10"
                  />
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    className="absolute right-0 top-0 h-full"
                    onClick={() => setShowPassword(!showPassword)}
                  >
                    {showPassword ? (
                      <EyeOff className="h-4 w-4 text-muted-foreground" />
                    ) : (
                      <Eye className="h-4 w-4 text-muted-foreground" />
                    )}
                  </Button>
                </div>

                <div className="space-y-1">
                  <div className="flex justify-between text-xs">
                    <span>Password strength:</span>
                    <span>{getStrengthText(passwordStrength)}</span>
                  </div>
                  <Progress
                    value={passwordStrength}
                    className={getStrengthColor(passwordStrength)}
                  />
                </div>

                <ul className="text-xs text-muted-foreground space-y-1 mt-2">
                  <li
                    className={
                      password.length >= 8
                        ? "text-green-500 dark:text-green-400"
                        : ""
                    }
                  >
                    • At least 8 characters
                  </li>
                  <li
                    className={
                      /[A-Z]/.test(password)
                        ? "text-green-500 dark:text-green-400"
                        : ""
                    }
                  >
                    • At least one uppercase letter
                  </li>
                  <li
                    className={
                      /[0-9]/.test(password)
                        ? "text-green-500 dark:text-green-400"
                        : ""
                    }
                  >
                    • At least one number
                  </li>
                  <li
                    className={
                      /[^A-Za-z0-9]/.test(password)
                        ? "text-green-500 dark:text-green-400"
                        : ""
                    }
                  >
                    • At least one special character
                  </li>
                </ul>
              </div>

              <div className="space-y-2">
                <Label htmlFor="confirm-password">Confirm Password</Label>
                <div className="relative">
                  <Input
                    id="confirm-password"
                    type={showPassword ? "text" : "password"}
                    value={confirmPassword}
                    onChange={(e) => setConfirmPassword(e.target.value)}
                    placeholder="Confirm your new password"
                    disabled={isSubmitting || !token}
                    className="pr-10"
                  />
                </div>
                {password &&
                  confirmPassword &&
                  password !== confirmPassword && (
                    <p className="text-xs text-destructive">
                      Passwords do not match
                    </p>
                  )}
              </div>

              <Button
                type="submit"
                className="w-full"
                disabled={
                  isSubmitting || !token || !password || !confirmPassword
                }
              >
                {isSubmitting ? "Resetting..." : "Reset Password"}
              </Button>
            </form>
          </CardContent>
          <CardFooter className="flex justify-center">
            <p className="text-xs text-muted-foreground">
              Remember your password?{" "}
              <Link to="/login" className="text-primary hover:underline">
                Log in
              </Link>
            </p>
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
// export default function ConfirmPasswordReset() {
//   const [searchParams] = useSearchParams();
//   const token = searchParams.get("token");
//   const { isPending, isError, error } = useQuery({
//     queryKey: ["check-password-reset", token],
//     queryFn: async () => {
//       if (!token) {
//         throw new Error("Missing token");
//       }
//       return await checkPasswordReset(token);
//     },
//   });
//   if (isPending) {
//     return <div>Loading...</div>;
//   }
//   if (isError) {
//     return <div>Error: {error.message}</div>;
//   }
//   return <div>ConfirmPasswordReset</div>;
// }
