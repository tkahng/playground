import PageSectionLayout from "@/layouts/page-section";
import UserListPage from "@/pages/admin/users/user-list";
import { createFileRoute } from "@tanstack/react-router";

function AdminUsersPage() {
  return (
    <PageSectionLayout title="Users">
      <UserListPage />
    </PageSectionLayout>
  );
}

export const Route = createFileRoute(
  "/_authenticated/admin/_adminLayout/users/"
)({
  component: AdminUsersPage,
});
