import PageSectionLayout from "@/layouts/page-section";
import BillingSettingPage from "@/pages/settings/billing-settings";
import { createFileRoute } from "@tanstack/react-router";

function AccountSettingsBilling() {
  return (
    <PageSectionLayout title="Account Settings">
      <BillingSettingPage />
    </PageSectionLayout>
  );
}

export const Route = createFileRoute(
  "/_authenticated/account/_accountLayout/settings/billing"
)({
  component: AccountSettingsBilling,
});
