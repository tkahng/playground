import PageSectionLayout from "@/layouts/page-section";
import BlogEditor from "@/pages/admin/blog/blog-editor";
import { createFileRoute } from "@tanstack/react-router";

function AdminBlogEditPage() {
  const { postId } = Route.useParams();
  return (
    <PageSectionLayout title="Edit Post">
      <BlogEditor postId={postId} />
    </PageSectionLayout>
  );
}

export const Route = createFileRoute(
  "/_authenticated/admin/_adminLayout/blog/$postId/edit"
)({
  component: AdminBlogEditPage,
});
