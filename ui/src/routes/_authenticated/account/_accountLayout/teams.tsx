import PageSectionLayout from "@/layouts/page-section";
import AccountTeamsPage from "@/pages/account/teams";
import { createFileRoute } from "@tanstack/react-router";

function AccountTeams() {
  return (
    <PageSectionLayout title="Teams Overview">
      <AccountTeamsPage />
    </PageSectionLayout>
  );
}

export const Route = createFileRoute(
  "/_authenticated/account/_accountLayout/teams"
)({
  component: AccountTeams,
});
