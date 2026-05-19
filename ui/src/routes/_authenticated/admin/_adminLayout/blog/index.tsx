import PageSectionLayout from "@/layouts/page-section";
import AdminBlogListPage from "@/pages/admin/blog/blog-list";
import { createFileRoute } from "@tanstack/react-router";

function AdminBlogPage() {
  return (
    <PageSectionLayout title="Blog">
      <AdminBlogListPage />
    </PageSectionLayout>
  );
}

export const Route = createFileRoute(
  "/_authenticated/admin/_adminLayout/blog/"
)({
  component: AdminBlogPage,
});
