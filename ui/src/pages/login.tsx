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
import { Label } from "@radix-ui/react-label";
import { Lock } from "lucide-react";
import { Form, Link, useNavigation } from "react-router";

export default function LoginPage() {
  const navigation = useNavigation();
  const isSubmitting = navigation.formAction === RouteMap.SIGNIN;
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
          <CardContent className="space-y-4">
            <Form method="post" action={RouteMap.SIGNIN}>
              <div className="space-y-2">
                <Label htmlFor="email">Email</Label>
                <Input
                  id="email"
                  placeholder="tkahng+0@gmail.com"
                  required
                  name="email"
                  type="email"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="password">Password</Label>
                <Input id="password" required name="password" type="password" />
              </div>
              <Button className="w-full" type="submit" disabled={isSubmitting}>
                <Lock className="mr-2 h-4 w-4" /> Login
              </Button>
            </Form>
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
