import PageSectionLayout from "@/layouts/page-section";
import TeamDashboard from "@/pages/teams/dashboard";
import { createFileRoute } from "@tanstack/react-router";

function TeamDashboardPage() {
  return (
    <PageSectionLayout>
      <TeamDashboard />
    </PageSectionLayout>
  );
}

export const Route = createFileRoute(
  "/_authenticated/teams/$teamSlug/_teamLayout/dashboard"
)({
  component: TeamDashboardPage,
});
