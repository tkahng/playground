import PageSectionLayout from "@/layouts/page-section";
import PermissionListPage from "@/pages/admin/permissions/permissions-list";
import { createFileRoute } from "@tanstack/react-router";

function AdminPermissionsPage() {
  return (
    <PageSectionLayout title="Permissions">
      <PermissionListPage />
    </PageSectionLayout>
  );
}

export const Route = createFileRoute(
  "/_authenticated/admin/_adminLayout/permissions/"
)({
  component: AdminPermissionsPage,
});
