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
import {
  acceptInvitation,
  declineInvitation,
  getTeamInvitationByToken,
} from "@/lib/api";
import { GetError } from "@/lib/get-error";
import { useMutation, useQuery } from "@tanstack/react-query";
import { ArrowRight, Check, Home } from "lucide-react";
import { useState } from "react";
import { Navigate, useNavigate } from "react-router";
import { toast } from "sonner";

export default function UserTeamInvitationRedirectPage() {
  const [disabled, setDisabled] = useState(false);
  const params = new URLSearchParams(window.location.search);
  const navigate = useNavigate();
  const { user } = useAuthProvider();
  const token = params.get("token");
  const { data, isLoading, error } = useQuery({
    queryKey: ["get-team-invitation-by-token"],
    queryFn: async () => {
      if (!token) {
        throw new Error("Missing session ID");
      }
      return getTeamInvitationByToken(token);
    },
  });

  const acceptMutation = useMutation({
    mutationFn: async (token?: string) => {
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token");
      }
      if (!token) {
        throw new Error("Missing invitation token");
      }
      const result = await acceptInvitation(user.tokens.access_token, token);
      if (!result) {
        throw new Error("Failed to accept invitation");
      }
      return result;
    },
    onSuccess: () => {
      toast.success("Invitation accepted successfully");
      console.log("navigating to team dashboard");
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
        throw new Error("Missing access token");
      }
      if (!token) {
        throw new Error("Missing invitation token");
      }
      const result = await declineInvitation(user.tokens.access_token, token);
      if (!result) {
        throw new Error("Failed to decline invitation");
      }
      return result;
    },
    onSuccess: () => {
      navigate(`/dashboard`);
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
  if (!user) {
    if (data) {
      params.set("email", data.email);
      return (
        <Navigate
          to={{
            pathname: "/signin",
            search: params.toString(),
          }}
        />
      );
    }
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
