import PageSectionLayout from "@/layouts/page-section";
import TeamNotifications from "@/pages/teams/settings/team-notifications";
import { createFileRoute } from "@tanstack/react-router";

function TeamNotificationsPage() {
  return (
    <PageSectionLayout title="Notifications">
      <TeamNotifications />
    </PageSectionLayout>
  );
}

export const Route = createFileRoute(
  "/_authenticated/teams/$teamSlug/_teamLayout/notifications/"
)({
  component: TeamNotificationsPage,
});
