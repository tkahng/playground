import PageSectionLayout from "@/layouts/page-section";
import PointsSettingsPage from "@/pages/settings/points-settings";
import { createFileRoute } from "@tanstack/react-router";

function AccountSettingsPoints() {
  return (
    <PageSectionLayout title="Account Settings">
      <PointsSettingsPage />
    </PageSectionLayout>
  );
}

export const Route = createFileRoute(
  "/_authenticated/account/_accountLayout/settings/points"
)({
  component: AccountSettingsPoints,
});
