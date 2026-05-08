import PageSectionLayout from "@/layouts/page-section";
import RoleEdit from "@/pages/admin/roles/role-edit";
import { createFileRoute } from "@tanstack/react-router";

function AdminRoleEditPage() {
  return (
    <PageSectionLayout title="Roles">
      <RoleEdit />
    </PageSectionLayout>
  );
}

export const Route = createFileRoute(
  "/_authenticated/admin/_adminLayout/roles/$roleId"
)({
  component: AdminRoleEditPage,
});
