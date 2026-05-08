import PageSectionLayout from "@/layouts/page-section";
import InvitationsPage from "@/pages/account/invitations";
import { createFileRoute } from "@tanstack/react-router";

function TeamsInvitations() {
  return (
    <PageSectionLayout title="Teams Overview">
      <InvitationsPage />
    </PageSectionLayout>
  );
}

export const Route = createFileRoute(
  "/_authenticated/account/_accountLayout/teams-invitations"
)({
  component: TeamsInvitations,
});
