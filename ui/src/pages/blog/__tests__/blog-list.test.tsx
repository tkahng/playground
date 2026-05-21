import { describe, it, expect, vi } from "vitest";
import { screen } from "@testing-library/react";
import { render } from "@/test/test-utils";
import BlogListPage from "../blog-list";
import * as blogQueries from "@/lib/blog-queries";
import type { BlogPost } from "@/lib/blog-queries";

const meta = { page: 0, per_page: 10, total: 1, next_page: null, prev_page: null, has_more: false };

const basePost: BlogPost = {
  id: "post-1",
  slug: "hello",
  title: "Hello Post",
  content: "content",
  content_format: "markdown",
  status: "published",
  author_id: "user-1",
  published_at: "2026-01-15T00:00:00Z",
  featured_image_id: null,
  featured_image_url: null,
  seo_title: null,
  seo_description: "A great post",
  reading_time_minutes: 2,
  view_count: 5,
  tags: [],
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-01T00:00:00Z",
};

function mockListPosts(posts: BlogPost[]) {
  vi.spyOn(blogQueries, "listBlogPosts").mockResolvedValue({ data: posts, meta });
}

describe("BlogListPage", () => {
  it("renders post titles", async () => {
    mockListPosts([basePost]);
    render(<BlogListPage />);
    expect(await screen.findByText("Hello Post")).toBeInTheDocument();
  });

  it("renders seo_description as excerpt", async () => {
    mockListPosts([basePost]);
    render(<BlogListPage />);
    expect(await screen.findByText("A great post")).toBeInTheDocument();
  });

  it("does not render featured image when null", async () => {
    mockListPosts([{ ...basePost, featured_image_url: null }]);
    render(<BlogListPage />);
    await screen.findByText("Hello Post");
    const imgs = document.querySelectorAll("article img");
    expect(imgs).toHaveLength(0);
  });

  it("renders featured image in PostCard when present", async () => {
    mockListPosts([{ ...basePost, featured_image_url: "https://pub.example.com/media/card.jpg" }]);
    render(<BlogListPage />);
    await screen.findByText("Hello Post");
    const img = document.querySelector("img[src='https://pub.example.com/media/card.jpg']");
    expect(img).toBeInTheDocument();
  });

  it("shows 'No posts found' when list is empty", async () => {
    mockListPosts([]);
    render(<BlogListPage />);
    expect(await screen.findByText("No posts found.")).toBeInTheDocument();
  });
});
