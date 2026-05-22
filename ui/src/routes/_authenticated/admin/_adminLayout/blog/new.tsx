import PageSectionLayout from "@/layouts/page-section";
import BlogEditor from "@/pages/admin/blog/blog-editor";
import { createFileRoute } from "@tanstack/react-router";

function AdminBlogNewPage() {
  return (
    <PageSectionLayout title="New Post">
      <BlogEditor />
    </PageSectionLayout>
  );
}

export const Route = createFileRoute(
  "/_authenticated/admin/_adminLayout/blog/new"
)({
  component: AdminBlogNewPage,
});
