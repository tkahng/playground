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
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { GetError } from "@/lib/get-error";
import { getTeamInvitationByToken } from "@/lib/queries";
import { useQuery } from "@tanstack/react-query";
import { ArrowRight, Check, Home, Link } from "lucide-react";

export default function UserTeamInvitationRedirectPage() {
  const params = new URLSearchParams(window.location.search);
  const { user } = useAuthProvider();
  const token = params.get("token");
  const { data, isLoading, error } = useQuery({
    queryKey: ["get-team-invitation-by-token"],
    queryFn: async () => {
      if (!token) {
        throw new Error("Missing session ID");
      }
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token");
      }
      return getTeamInvitationByToken(user.tokens.access_token, token);
    },
  });

  if (isLoading) {
    return (
      <div>
        <p>Loading...</p>
      </div>
    );
  }

  if (error) {
    const err = GetError(error);
    return (
      <div>
        <p>Error: {err?.detail}</p>
      </div>
    );
  }
  if (!data?.team) {
    return (
      <div>
        <p>Error: Team not found</p>
      </div>
    );
  }

  return (
    <div className="flex">
      <Card className="max-w-md w-full">
        <CardHeader className="text-center">
          <div className="mx-auto rounded-full w-12 h-12 bg-green-100 dark:bg-green-900 flex items-center justify-center mb-4">
            <Check className="h-6 w-6 text-green-600 dark:text-green-300" />
          </div>
          <CardTitle className="text-2xl">Team Invitation</CardTitle>
          <CardDescription>
            You have been invited to join the team: {data.team.name}
          </CardDescription>
        </CardHeader>
        <CardContent className="text-center">
          <p className="text-muted-foreground">
            You have been invited to join the team: {data.team.name}
          </p>
        </CardContent>
        <CardFooter className="flex flex-col space-y-2">
          <Button className="w-full" asChild>
            <Link to={RouteMap.SIGNIN}>
              <ArrowRight className="mr-2 h-4 w-4" />
              Continue to Login
            </Link>
          </Button>
          <Button variant="outline" className="w-full" asChild>
            <Link to={RouteMap.HOME}>
              <Home className="mr-2 h-4 w-4" />
              Return to Home
            </Link>
          </Button>
        </CardFooter>
      </Card>
    </div>
  );
}
