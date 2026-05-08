import PageSectionLayout from "@/layouts/page-section";
import AccountDashboard from "@/pages/account/dashboard";
import { createFileRoute } from "@tanstack/react-router";

function AccountDashboardPage() {
  return (
    <PageSectionLayout title="Account Overview">
      <AccountDashboard />
    </PageSectionLayout>
  );
}

export const Route = createFileRoute(
  "/_authenticated/account/_accountLayout/dashboard"
)({
  component: AccountDashboardPage,
});
