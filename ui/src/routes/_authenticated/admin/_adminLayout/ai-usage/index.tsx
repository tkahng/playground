import PageSectionLayout from "@/layouts/page-section";
import AiUsageListPage from "@/pages/admin/ai-usage/ai-usage-list";
import { createFileRoute } from "@tanstack/react-router";

function AdminAiUsagePage() {
  return (
    <PageSectionLayout title="AI Usage">
      <AiUsageListPage />
    </PageSectionLayout>
  );
}

export const Route = createFileRoute(
  "/_authenticated/admin/_adminLayout/ai-usage/"
)({
  component: AdminAiUsagePage,
});
