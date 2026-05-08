import PageSectionLayout from "@/layouts/page-section";
import PermissionEdit from "@/pages/admin/permissions/permissions-edit";
import { createFileRoute } from "@tanstack/react-router";

function AdminPermissionEditPage() {
  return (
    <PageSectionLayout title="Permissions">
      <PermissionEdit />
    </PageSectionLayout>
  );
}

export const Route = createFileRoute(
  "/_authenticated/admin/_adminLayout/permissions/$permissionId"
)({
  component: AdminPermissionEditPage,
});
