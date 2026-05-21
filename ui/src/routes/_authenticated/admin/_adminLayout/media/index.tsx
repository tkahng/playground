import PageSectionLayout from "@/layouts/page-section";
import AdminMediaListPage from "@/pages/admin/media/media-list";
import { createFileRoute } from "@tanstack/react-router";

function AdminMediaPage() {
  return (
    <PageSectionLayout title="Media">
      <AdminMediaListPage />
    </PageSectionLayout>
  );
}

export const Route = createFileRoute(
  "/_authenticated/admin/_adminLayout/media/"
)({
  component: AdminMediaPage,
});
