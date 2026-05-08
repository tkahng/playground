import PageSectionLayout from "@/layouts/page-section";
import SubscriptionsListPage from "@/pages/admin/subscriptions/subscription-list";
import { createFileRoute } from "@tanstack/react-router";

function AdminSubscriptionsPage() {
  return (
    <PageSectionLayout title="Subscriptions">
      <SubscriptionsListPage />
    </PageSectionLayout>
  );
}

export const Route = createFileRoute(
  "/_authenticated/admin/_adminLayout/subscriptions/"
)({
  component: AdminSubscriptionsPage,
});
