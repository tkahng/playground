import PageSectionLayout from "@/layouts/page-section";
import PlanFeaturesListPage from "@/pages/admin/plan-features/plan-features-list";
import { createFileRoute } from "@tanstack/react-router";

function AdminPlanFeaturesPage() {
  return (
    <PageSectionLayout title="Plan Features">
      <PlanFeaturesListPage />
    </PageSectionLayout>
  );
}

export const Route = createFileRoute(
  "/_authenticated/admin/_adminLayout/plan-features/"
)({
  component: AdminPlanFeaturesPage,
});
