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
import { Input } from "@/components/ui/input";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { SignupInput } from "@/schema.types";
import { Label } from "@radix-ui/react-label";
import { Lock } from "lucide-react";
import { useState } from "react";
import { Link, useLocation, useNavigate } from "react-router";
import { toast } from "sonner";

export default function SignupPage() {
  const [input, setInput] = useState<SignupInput>({
    email: "",
    password: "",
    name: "",
  });
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();
  const { signUp } = useAuthProvider();
  const { search } = useLocation();
  const params = new URLSearchParams(search);
  const token = params.get("token");
  const redirectTo = params.get("redirect_to");
  const email = params.get("email");
  const navigateTo =
    email && token ? `/team-invitation` : redirectTo ? redirectTo : "?";
  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    setLoading(true);

    try {
      await signUp({
        email: email || input.email,
        password: input.password,
        name: input.name,
      });
      setLoading(false);
      navigate({ pathname: navigateTo, search: search });
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
  return (
    <div className="flex min-h-screen flex-col">
      <div className="flex flex-1 items-center justify-center">
        <Card className="w-full max-w-md">
          <CardHeader className="space-y-1">
            <CardTitle className="text-center text-2xl font-bold">
              Create your account
            </CardTitle>
            <CardDescription className="text-center">
              Enter your details to get started with Playground
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-1">
            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="space-y-2">
                <div className="space-y-4">
                  <Label htmlFor="name">Full Name</Label>
                  <Input
                    id="name"
                    placeholder="John Doe"
                    required
                    name="name"
                    type="text"
                    onChange={handleChange}
                  />
                </div>
                <div className="space-y-4">
                  <Label htmlFor="email">Email</Label>
                  <Input
                    id="email"
                    placeholder={email || "tkahng+01@gmail.com"}
                    required
                    disabled={!!email}
                    name="email"
                    type="email"
                    onChange={handleChange}
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="password">Password</Label>
                  <Input
                    id="password"
                    required
                    name="password"
                    placeholder="Password123!"
                    type="password"
                    onChange={handleChange}
                  />
                </div>
              </div>
              <Button className="w-full" type="submit" disabled={loading}>
                <Lock className="mr-2 h-4 w-4" /> Create account
              </Button>
            </form>
          </CardContent>
          <CardFooter className="flex flex-col space-y-4">
            <div className="text-center text-sm text-gray-500 dark:text-gray-400">
              Already have an account?{" "}
              <Link
                className="text-primary underline-offset-4 hover:underline"
                to={{
                  pathname: RouteMap.SIGNIN,
                  search: params.toString(),
                }}
              >
                Sign in
              </Link>
            </div>
            <div className="relative">
              <div className="relative flex justify-center text-xs uppercase">
                <span className="px-2 text-muted-foreground">
                  Or continue with
                </span>
              </div>
            </div>
            <div className="flex flex-row justify-center space-x-4">
              <ul className="flex flex-row justify-center space-x-4">
                {providerNames.map((providerName) => (
                  <li key={providerName}>
                    <ProviderConnectionForm
                      type="Login"
                      providerName={providerName}
                    />
                  </li>
                ))}
              </ul>
            </div>
          </CardFooter>
        </Card>
      </div>
    </div>
  );
}
