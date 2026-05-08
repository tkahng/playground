import UserTeamInvitationRedirectPage from "@/pages/teams/user-team-invitation-redirect-page";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/_public/team-invitation")({
  component: UserTeamInvitationRedirectPage,
});
