import PageSectionLayout from "@/layouts/page-section";
import PlanFeaturesEditPage from "@/pages/admin/plan-features/plan-features-edit";
import { createFileRoute } from "@tanstack/react-router";

function AdminPlanFeaturesEditPage() {
  return (
    <PageSectionLayout title="Plan Features">
      <PlanFeaturesEditPage />
    </PageSectionLayout>
  );
}

export const Route = createFileRoute(
  "/_authenticated/admin/_adminLayout/plan-features/$productId"
)({
  component: AdminPlanFeaturesEditPage,
});
