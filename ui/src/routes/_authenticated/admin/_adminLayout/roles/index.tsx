import PageSectionLayout from "@/layouts/page-section";
import RolesListPage from "@/pages/admin/roles/roles-list";
import { createFileRoute } from "@tanstack/react-router";

function AdminRolesPage() {
  return (
    <PageSectionLayout title="Roles">
      <RolesListPage />
    </PageSectionLayout>
  );
}

export const Route = createFileRoute(
  "/_authenticated/admin/_adminLayout/roles/"
)({
  component: AdminRolesPage,
});
