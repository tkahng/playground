import { NexusAILandingHeader } from "@/components/nexus-landing-header";
import { NexusAIMinimalFooter } from "@/components/nexus-minimal-footer";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { ArrowLeft, Home, LogIn, ShieldAlert } from "lucide-react";
import { Link } from "react-router";

export default function NotAuthorizedPage() {
  return (
    <div className="flex flex-col min-h-screen">
      <NexusAILandingHeader />
      <main className="flex-1 flex items-center justify-center p-4">
        <Card className="max-w-md w-full">
          <CardHeader className="text-center">
            <div className="flex justify-center mb-4">
              <ShieldAlert className="h-12 w-12 text-primary" />
            </div>
            <CardTitle className="text-2xl">Access Denied</CardTitle>
            <CardDescription>
              You don't have permission to access this resource.
            </CardDescription>
          </CardHeader>
          <CardContent className="text-center space-y-4">
            <p className="text-muted-foreground">This might be because:</p>
            <ul className="text-sm text-muted-foreground space-y-2">
              <li>• You need to log in or sign up</li>
              <li>• Your account doesn't have the required permissions</li>
              <li>• The resource may have been moved or deleted</li>
              <li>• You've reached a subscription limit</li>
            </ul>
          </CardContent>
          <CardFooter className="flex flex-col space-y-2">
            <div className="flex flex-col sm:flex-row w-full gap-2">
              <Button variant="outline" className="flex-1" asChild>
                <Link to="javascript:history.back()">
                  <ArrowLeft className="mr-2 h-4 w-4" />
                  Go Back
                </Link>
              </Button>
              <Button variant="outline" className="flex-1" asChild>
                <Link to="/">
                  <Home className="mr-2 h-4 w-4" />
                  Home
                </Link>
              </Button>
            </div>
            <Button className="w-full" asChild>
              <Link to="/login">
                <LogIn className="mr-2 h-4 w-4" />
                Sign In
              </Link>
            </Button>
          </CardFooter>
        </Card>
      </main>
      <NexusAIMinimalFooter />
    </div>
  );
}
