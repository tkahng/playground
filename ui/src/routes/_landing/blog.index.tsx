import BlogListPage from "@/pages/blog/blog-list";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/_landing/blog/")({
  component: BlogListPage,
});
