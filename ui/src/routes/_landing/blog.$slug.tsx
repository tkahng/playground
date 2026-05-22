import BlogPostPage from "@/pages/blog/blog-post";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/_landing/blog/$slug")({
  component: function BlogPostRoute() {
    const { slug } = Route.useParams();
    return <BlogPostPage slug={slug} />;
  },
});
