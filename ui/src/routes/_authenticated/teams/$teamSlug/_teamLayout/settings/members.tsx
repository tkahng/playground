import PageSectionLayout from "@/layouts/page-section";
import TeamMembersSettingPage from "@/pages/teams/settings/team-members-settings";
import { createFileRoute } from "@tanstack/react-router";

function TeamSettingsMembersPage() {
  return (
    <PageSectionLayout title="Settings">
      <TeamMembersSettingPage />
    </PageSectionLayout>
  );
}

export const Route = createFileRoute(
  "/_authenticated/teams/$teamSlug/_teamLayout/settings/members"
)({
  component: TeamSettingsMembersPage,
});
