import { describe, it, expect, vi, beforeEach } from "vitest";
import { screen } from "@testing-library/react";
import { render } from "@/test/test-utils";
import BlogEditor from "../blog-editor";
import * as blogQueries from "@/lib/blog-queries";
import type { BlogPost } from "@/lib/blog-queries";

vi.mock("sonner", () => ({ toast: { success: vi.fn(), error: vi.fn() } }));
vi.mock("@tanstack/react-router", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@tanstack/react-router")>();
  return { ...actual, useNavigate: () => vi.fn() };
});
vi.mock("@/lib/media-queries", () => ({
  listMedia: vi.fn().mockResolvedValue({ data: [], meta: { page: 0, per_page: 20, total: 0, next_page: null, prev_page: null, has_more: false } }),
  uploadMedia: vi.fn(),
}));

const draftPost: BlogPost = {
  id: "post-1",
  slug: "my-draft",
  title: "Draft Post",
  content: "content",
  content_format: "markdown",
  status: "draft",
  author_id: "user-1",
  published_at: null,
  featured_image_id: null,
  featured_image_url: null,
  seo_title: null,
  seo_description: null,
  reading_time_minutes: null,
  view_count: 0,
  tags: [],
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-01T00:00:00Z",
};

const publishedPost: BlogPost = { ...draftPost, status: "published", published_at: "2026-01-15T00:00:00Z" };
const archivedPost: BlogPost = { ...draftPost, status: "archived" };

function setupMocks(post: BlogPost) {
  vi.spyOn(blogQueries, "getBlogPost").mockResolvedValue(post);
  vi.spyOn(blogQueries, "listBlogTags").mockResolvedValue([]);
}

describe("BlogEditor — new post", () => {
  it("renders 'New Post' heading", () => {
    render(<BlogEditor />);
    expect(screen.getByText("New Post")).toBeInTheDocument();
  });

  it("shows only Save button (no status actions)", () => {
    render(<BlogEditor />);
    expect(screen.getByRole("button", { name: /save/i })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /publish/i })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /unpublish/i })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /archive/i })).not.toBeInTheDocument();
  });

  it("shows 'Choose image' featured image picker button", () => {
    render(<BlogEditor />);
    expect(screen.getByRole("button", { name: /choose image/i })).toBeInTheDocument();
  });
});

describe("BlogEditor — draft post", () => {
  beforeEach(() => setupMocks(draftPost));

  it("shows Publish and Archive buttons, not Unpublish", async () => {
    render(<BlogEditor postId="post-1" />);
    expect(await screen.findByRole("button", { name: /publish/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /archive/i })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /unpublish/i })).not.toBeInTheDocument();
  });

  it("shows read-only slug field", async () => {
    render(<BlogEditor postId="post-1" />);
    const slugInput = await screen.findByDisplayValue("my-draft");
    expect(slugInput).toHaveAttribute("readonly");
  });
});

describe("BlogEditor — published post", () => {
  beforeEach(() => setupMocks(publishedPost));

  it("shows Unpublish and Archive buttons, not Publish", async () => {
    render(<BlogEditor postId="post-1" />);
    expect(await screen.findByRole("button", { name: /unpublish/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /archive/i })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /^publish$/i })).not.toBeInTheDocument();
  });

  it("shows 'published' status badge", async () => {
    render(<BlogEditor postId="post-1" />);
    expect(await screen.findByText("published")).toBeInTheDocument();
  });
});

describe("BlogEditor — archived post", () => {
  beforeEach(() => setupMocks(archivedPost));

  it("shows no Publish, Unpublish, or Archive buttons", async () => {
    render(<BlogEditor postId="post-1" />);
    await screen.findByText("archived");
    expect(screen.queryByRole("button", { name: /^publish$/i })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /unpublish/i })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /archive/i })).not.toBeInTheDocument();
  });
});
