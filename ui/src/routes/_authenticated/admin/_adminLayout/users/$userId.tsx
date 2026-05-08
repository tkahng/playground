import PageSectionLayout from "@/layouts/page-section";
import UserEdit from "@/pages/admin/users/user-edit";
import { createFileRoute } from "@tanstack/react-router";

function AdminUserEditPage() {
  return (
    <PageSectionLayout title="Users">
      <UserEdit />
    </PageSectionLayout>
  );
}

export const Route = createFileRoute(
  "/_authenticated/admin/_adminLayout/users/$userId"
)({
  component: AdminUserEditPage,
});
