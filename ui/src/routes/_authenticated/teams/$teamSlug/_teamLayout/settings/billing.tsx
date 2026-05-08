import PageSectionLayout from "@/layouts/page-section";
import TeamBillingSettingPage from "@/pages/teams/settings/team-billing-settings";
import { createFileRoute } from "@tanstack/react-router";

function TeamSettingsBillingPage() {
  return (
    <PageSectionLayout title="Settings">
      <TeamBillingSettingPage />
    </PageSectionLayout>
  );
}

export const Route = createFileRoute(
  "/_authenticated/teams/$teamSlug/_teamLayout/settings/billing"
)({
  component: TeamSettingsBillingPage,
});
