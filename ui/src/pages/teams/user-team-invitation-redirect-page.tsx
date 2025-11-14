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
import { ApiError, GetError, isErrorModel } from "@/lib/error";
import {
  acceptInvitation,
  declineInvitation,
  getTeamInvitationByToken,
} from "@/lib/team-queries";
import { useMutation, useQuery } from "@tanstack/react-query";
import { ArrowRight, Check, Home } from "lucide-react";
import { useState } from "react";
import {
  createSearchParams,
  Navigate,
  useLocation,
  useNavigate,
} from "react-router";
import { toast } from "sonner";

export default function UserTeamInvitationRedirectPage() {
  const [disabled, setDisabled] = useState(false);
  const { search } = useLocation();
  const params = new URLSearchParams(search);
  const navigate = useNavigate();
  const { user } = useAuthProvider();
  const token = params.get("token");
  const { data, isLoading, error } = useQuery({
    queryKey: ["get-team-invitation-by-token"],
    queryFn: async () => {
      if (!token) {
        throw new ApiError("Missing token");
      }
      return getTeamInvitationByToken(token);
    },
  });

  const acceptMutation = useMutation({
    mutationFn: async (token?: string) => {
      if (!user?.tokens.access_token) {
        throw new ApiError("Missing access token");
      }
      if (!token) {
        throw new ApiError("Missing invitation token");
      }
      const result = await acceptInvitation(user.tokens.access_token, token);
      if (!result) {
        throw new ApiError("Failed to accept invitation");
      }
      return result;
    },
    onSuccess: () => {
      toast.success("Invitation accepted successfully");
      navigate(`/teams/${data?.team?.slug}/dashboard`);
    },
    onError: (err) => {
      const error = GetError(err);
      toast.error(`Failed to update role: ${error?.detail}`);
    },
  });
  const declineMutation = useMutation({
    mutationFn: async (token?: string) => {
      if (!user?.tokens.access_token) {
        throw new ApiError("Missing access token");
      }
      if (!token) {
        throw new ApiError("Missing invitation token");
      }
      const result = await declineInvitation(user.tokens.access_token, token);
      if (!result) {
        throw new ApiError("Failed to decline invitation");
      }
      return result;
    },
    onSuccess: () => {
      toast.success("Invitation declined successfully");
      navigate(RouteMap.ACCOUNT_OVERVIEW_TEAMS_INVITATION);
    },
    onError: (err) => {
      toast.error(`Failed to decline role: ${err.message}`);
    },
  });
  function onAccept() {
    setDisabled(true);
    acceptMutation.mutateAsync(token!);
    setDisabled(false);
  }
  function onDecline() {
    setDisabled(true);
    declineMutation.mutateAsync(token!);
    setDisabled(false);
  }
  if (isLoading) {
    return (
      <div>
        <p>Loading...</p>
      </div>
    );
  }

  if (error) {
    if (isErrorModel(error)) {
      return (
        <div>
          <p>Error: {error.detail}</p>
        </div>
      );
    }
    return (
      <div>
        <p>Error: {error?.message}</p>
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
  if (!user) {
    return (
      <Navigate
        to={{
          pathname: "/signin",
          search: createSearchParams({
            redirect_to: encodeURIComponent(window.location.href),
            email: data.email,
          }).toString(),
        }}
      />
    );
  }

  return (
    <div className="flex min-h-screen flex-col">
      <div className="flex flex-1 items-center justify-center">
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
              by {data.inviter_member?.user?.email}
            </p>
          </CardContent>
          <CardFooter className="flex flex-col space-y-2">
            <Button className="w-full" disabled={disabled} onClick={onAccept}>
              <ArrowRight className="mr-2 h-4 w-4" />
              Accept
            </Button>
            <Button
              variant="outline"
              className="w-full"
              disabled={disabled}
              onClick={onDecline}
            >
              <Home className="mr-2 h-4 w-4" />
              Decline
            </Button>
          </CardFooter>
        </Card>
      </div>
    </div>
  );
}
