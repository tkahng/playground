import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { getTeamInvitationByToken } from "@/lib/queries";
import { useQuery } from "@tanstack/react-query";

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
    return (
      <div>
        <p>Error: {error.message}</p>
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
      <Card>
        <CardHeader>
          <CardTitle>Card Title</CardTitle>
          <CardDescription>Card Description</CardDescription>
          <CardAction>Card Action</CardAction>
        </CardHeader>
        <CardContent>
          <p>Card Content</p>
        </CardContent>
        <CardFooter>
          <p>Card Footer</p>
        </CardFooter>
      </Card>
    </div>
  );
}
