import DashboardLayout from "@/layouts/dashboard-layout";
import { authenticatedSubHeaderLinks } from "@/components/links";
import PageSectionLayout from "@/layouts/page-section";
import ProtectedRouteLayout from "@/pages/protected-routes/protected-layout";
import { createFileRoute } from "@tanstack/react-router";

function ProtectedLayout() {
  return (
    <DashboardLayout headerLinks={authenticatedSubHeaderLinks}>
      <PageSectionLayout title="Protected">
        <ProtectedRouteLayout />
      </PageSectionLayout>
    </DashboardLayout>
  );
}

export const Route = createFileRoute(
  "/_authenticated/protected/_protectedLayout"
)({
  component: ProtectedLayout,
});
