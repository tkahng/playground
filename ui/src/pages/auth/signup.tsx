import {
  ProviderConnectionForm,
  providerNames,
} from "@/components/connections";
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
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { SignupInput } from "@/schema.types";
import { Check } from "lucide-react";
import { useState } from "react";
import { Link, useNavigate } from "react-router";
import { toast } from "sonner";
export default function SignupPage() {
  const [input, setInput] = useState<SignupInput>({
    email: "",
    password: "",
    name: "",
  });
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate(); // Get navigation function
  const { signUp } = useAuthProvider();
  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    setLoading(true);

    // Simulate API call (replace with actual authentication)
    try {
      await signUp({
        email: input.email,
        password: input.password,
        name: input.name,
      });
      setLoading(false);
      navigate(RouteMap.DASHBOARD_HOME);
    } catch (error) {
      if (error instanceof Error) {
        toast.error(error.message, {
          description: "Please try again",
          action: {
            label: "Undo",
            onClick: () => console.log("Undo"),
          },
        });
      }
      console.error(error);
      toast.error("unknown error", {
        description: "Please try again",
        action: {
          label: "Undo",
          onClick: () => console.log("Undo"),
        },
      });
      setLoading(false);
    }
  };

  function handleChange(e: React.ChangeEvent<HTMLInputElement>) {
    const key = e.target.id;
    const value = e.target.value;
    setInput((values) => ({
      ...values,
      [key]: value,
    }));
  }
  //   const navigation = useNavigation();
  //   const isSubmitting = navigation.formAction === RouteMap.SIGNUP;
  return (
    // <div className="container px-4 md:px-6">
    //   <div className="mx-auto grid max-w-6xl items-center gap-6 lg:grid-cols-2 lg:gap-12">
    <div className="flex min-h-screen flex-col">
      <div className=" flex flex-1 items-center justify-center gap-16 px-6 py-4 lg:px-8 lg:py-4">
        <div className="flex flex-col justify-center space-y-4">
          <div className="space-y-2">
            <h1 className="text-3xl font-bold tracking-tighter sm:text-5xl">
              Join NexusAI and Unlock the Power of AI
            </h1>
            <p className="max-w-[600px] text-foreground dark:text-muted-foreground">
              Sign up today to access cutting-edge AI models, powerful APIs, and
              a suite of tools designed to revolutionize your development
              process.
            </p>
          </div>
          <ul className="grid grid-cols-1 gap-2 sm:grid-cols-2">
            {[
              "Access to state-of-the-art AI models",
              "Powerful and easy-to-use APIs",
              "Comprehensive documentation and support",
              "Scalable infrastructure for any project size",
            ].map((feature) => (
              <li key={feature} className="flex items-center gap-2">
                <Check className="h-4 w-4 text-primary" />
                <span className="text-sm">{feature}</span>
              </li>
            ))}
          </ul>
        </div>
        <Card>
          <CardHeader>
            <CardTitle>Create your account</CardTitle>
            <CardDescription>
              Enter your details to get started with NexusAI
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <form
              onSubmit={handleSubmit}
              method="post"
              action={RouteMap.SIGNUP}
              className="space-y-4"
            >
              <div className="space-y-2">
                <Label htmlFor="name">Full Name</Label>
                <Input
                  id="name"
                  name="name"
                  placeholder="John Doe"
                  onChange={handleChange}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="email">Email</Label>
                <Input
                  id="email"
                  name="email"
                  placeholder="john~example.com"
                  required
                  type="email"
                  onChange={handleChange}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="password">Password</Label>
                <Input
                  id="password"
                  name="password"
                  required
                  type="password"
                  onChange={handleChange}
                />
              </div>

              <div className="flex items-center space-x-2">
                <Checkbox id="terms" name="terms" />
                <label
                  htmlFor="terms"
                  className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                >
                  I agree to the{" "}
                  <Link to="#" className="text-primary underline">
                    terms of service
                  </Link>{" "}
                  and{" "}
                  <Link to="#" className="text-primary underline">
                    privacy policy
                  </Link>
                </label>
              </div>
              <Button className="w-full" type="submit" disabled={loading}>
                Sign Up
              </Button>
            </form>
          </CardContent>
          <CardFooter className="flex flex-col space-y-4">
            <div className="relative">
              <div className="absolute inset-0 flex items-center">
                <span className="w-full border-t" />
              </div>
              <div className="relative flex justify-center text-xs uppercase">
                <span className="bg-background px-2 text-muted-foreground">
                  Or continue with
                </span>
              </div>
            </div>
            <div className="flex flex-row justify-center space-x-4">
              <ul className="flex flex-row justify-center space-x-4">
                {providerNames.map((providerName) => (
                  <li key={providerName}>
                    <ProviderConnectionForm
                      type="Signup"
                      providerName={providerName}
                    />
                  </li>
                ))}
              </ul>
            </div>
            <p className="text-center text-xs text-gray-500 dark:text-gray-400">
              Already have an account?{" "}
              <Link to={RouteMap.SIGNIN} className="text-primary underline">
                Log in
              </Link>
            </p>
          </CardFooter>
        </Card>
      </div>
    </div>
  );
}

// export function ErrorBoundary() {
//   return <GeneralErrorBoundary />;
// }
