import TeamDashboardLayout from "@/layouts/team-dashboard-layout";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute(
  "/_authenticated/teams/$teamSlug/_teamLayout"
)({
  component: TeamDashboardLayout,
});
