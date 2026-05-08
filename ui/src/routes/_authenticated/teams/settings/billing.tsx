import DashboardLayout from "@/layouts/dashboard-layout";
import TeamSettingsRedirect from "@/pages/teams/team-settings-redirect";
import { createFileRoute } from "@tanstack/react-router";

function TeamSettingsBillingRedirect() {
  return (
    <DashboardLayout>
      <TeamSettingsRedirect />
    </DashboardLayout>
  );
}

export const Route = createFileRoute("/_authenticated/teams/settings/billing")({
  component: TeamSettingsBillingRedirect,
});
