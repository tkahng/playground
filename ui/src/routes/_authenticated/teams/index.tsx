import DashboardLayout from "@/layouts/dashboard-layout";
import TeamSelect from "@/pages/teams";
import { createFileRoute } from "@tanstack/react-router";

function TeamsPage() {
  return (
    <DashboardLayout>
      <TeamSelect />
    </DashboardLayout>
  );
}

export const Route = createFileRoute("/_authenticated/teams/")({
  component: TeamsPage,
});
