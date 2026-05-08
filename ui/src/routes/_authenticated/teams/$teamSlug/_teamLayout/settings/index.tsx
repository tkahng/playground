import PageSectionLayout from "@/layouts/page-section";
import TeamSettingsPage from "@/pages/teams/settings/team-general-settings";
import { createFileRoute } from "@tanstack/react-router";

function TeamSettingsIndexPage() {
  return (
    <PageSectionLayout title="Settings">
      <TeamSettingsPage />
    </PageSectionLayout>
  );
}

export const Route = createFileRoute(
  "/_authenticated/teams/$teamSlug/_teamLayout/settings/"
)({
  component: TeamSettingsIndexPage,
});
