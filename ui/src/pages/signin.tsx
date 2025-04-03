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
import { AuthContext } from "@/context/auth-context";
import { SigninInput } from "@/schema.types";
import { Label } from "@radix-ui/react-label";
import { Lock } from "lucide-react";
import { useContext, useState } from "react";
import { Link, useNavigate } from "react-router";
import { toast } from "sonner";

export default function SigninPage() {
  const [input, setInput] = useState<SigninInput>({ email: "", password: "" });
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate(); // Get navigation function
  const { login } = useContext(AuthContext);

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    setLoading(true);

    // Simulate API call (replace with actual authentication)
    try {
      await login({ email: input.email, password: input.password });
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
  return (
    <div className="flex min-h-screen flex-col">
      <div className="flex flex-1 items-center justify-center">
        <Card className="w-full max-w-md">
          <CardHeader className="space-y-1">
            <CardTitle className="text-center text-2xl font-bold">
              Login to NexusAI
            </CardTitle>
            <CardDescription className="text-center">
              Enter your email and password to access your account
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-1">
            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="space-y-2">
                <div className="space-y-4">
                  <Label htmlFor="email">Email</Label>
                  <Input
                    id="email"
                    placeholder="tkahng+0@gmail.com"
                    required
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
                    type="password"
                    onChange={handleChange}
                  />
                </div>
              </div>
              <Button className="w-full" type="submit" disabled={loading}>
                <Lock className="mr-2 h-4 w-4" /> Login
              </Button>
            </form>
          </CardContent>
          <CardFooter className="flex flex-col space-y-4">
            <div className="text-center text-sm text-gray-500 dark:text-gray-400">
              Don't have an account?{" "}
              <Link
                className="text-primary underline-offset-4 hover:underline"
                to={RouteMap.SIGNUP}
              >
                Sign up
              </Link>
            </div>
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
            {/* <div className="flex justify-center space-x-4">
              <ul>
                {providerNames.map((providerName) => (
                  <li key={providerName}>
                    <ProviderConnectionForm
                      type="Login"
                      providerName={providerName}
                    />
                  </li>
                ))}
              </ul>
            </div> */}
            {/* <div className="flex justify-center space-x-4">
              <Button variant="outline" size="icon">
                <Twitter className="h-4 w-4" />
              </Button>
              <Button variant="outline" size="icon">
                <Facebook className="h-4 w-4" />
              </Button>
              <Button variant="outline" size="icon">
                <Linkedin className="h-4 w-4" />
              </Button>
              <Button variant="outline" size="icon">
                <Github className="h-4 w-4" />
              </Button>
            </div> */}
          </CardFooter>
        </Card>
      </div>
    </div>
  );
}
