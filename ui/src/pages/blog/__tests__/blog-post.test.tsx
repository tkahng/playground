import { describe, it, expect, vi } from "vitest";
import { screen } from "@testing-library/react";
import { render } from "@/test/test-utils";
import BlogPostPage from "../blog-post";
import * as blogQueries from "@/lib/blog-queries";
import type { BlogPost } from "@/lib/blog-queries";

const basePost: BlogPost = {
  id: "post-1",
  slug: "test-post",
  title: "Test Post",
  content: "Hello world",
  content_format: "markdown",
  status: "published",
  author_id: "user-1",
  published_at: "2026-01-15T00:00:00Z",
  featured_image_id: null,
  featured_image_url: null,
  seo_title: null,
  seo_description: null,
  reading_time_minutes: 3,
  view_count: 42,
  tags: [],
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-01T00:00:00Z",
};

function mockGetBlogPost(post: BlogPost) {
  vi.spyOn(blogQueries, "getBlogPost").mockResolvedValue(post);
}

describe("BlogPostPage", () => {
  it("renders the post title", async () => {
    mockGetBlogPost(basePost);
    render(<BlogPostPage slug="test-post" />);
    expect(await screen.findByRole("heading", { level: 1, name: "Test Post" })).toBeInTheDocument();
  });

  it("does not render featured image when null", async () => {
    mockGetBlogPost({ ...basePost, featured_image_url: null });
    render(<BlogPostPage slug="test-post" />);
    await screen.findByRole("heading", { level: 1 });
    const imgs = document.querySelectorAll("article > img");
    expect(imgs).toHaveLength(0);
  });

  it("renders featured image when present", async () => {
    mockGetBlogPost({
      ...basePost,
      featured_image_url: "https://pub.example.com/media/hero.jpg",
    });
    render(<BlogPostPage slug="test-post" />);
    await screen.findByRole("heading", { level: 1 });
    const img = document.querySelector("img[src='https://pub.example.com/media/hero.jpg']");
    expect(img).toBeInTheDocument();
    expect(img).toHaveAttribute("alt", "Test Post");
  });

  it("renders tags", async () => {
    mockGetBlogPost({
      ...basePost,
      tags: [{ id: "t1", name: "Go", slug: "go", created_at: "2026-01-01T00:00:00Z" }],
    });
    render(<BlogPostPage slug="test-post" />);
    expect(await screen.findByText("Go")).toBeInTheDocument();
  });

  it("renders reading time", async () => {
    mockGetBlogPost(basePost);
    render(<BlogPostPage slug="test-post" />);
    expect(await screen.findByText("3 min read")).toBeInTheDocument();
  });

  it("shows 'Post not found' on error", async () => {
    vi.spyOn(blogQueries, "getBlogPost").mockRejectedValue(new Error("404"));
    render(<BlogPostPage slug="missing" />);
    expect(await screen.findByText("Post not found")).toBeInTheDocument();
  });
});
