import PageSectionLayout from "@/layouts/page-section";
import AccountSettingsPage from "@/pages/settings/general-settings";
import { createFileRoute } from "@tanstack/react-router";

function AccountSettings() {
  return (
    <PageSectionLayout title="Account Settings">
      <AccountSettingsPage />
    </PageSectionLayout>
  );
}

export const Route = createFileRoute(
  "/_authenticated/account/_accountLayout/settings"
)({
  component: AccountSettings,
});
