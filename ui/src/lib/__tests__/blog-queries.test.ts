import { describe, it, expect, vi, afterEach } from "vitest";
import {
  getBlogPost,
  createBlogPost,
  updateBlogPost,
  publishBlogPost,
  listBlogPosts,
  type BlogPost,
} from "../blog-queries";

const mockPost: BlogPost = {
  id: "post-1",
  slug: "test-post",
  title: "Test Post",
  content: "Hello",
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

function mockFetch(body: unknown, status = 200) {
  return vi.spyOn(globalThis, "fetch").mockResolvedValue(
    new Response(JSON.stringify(body), {
      status,
      headers: { "Content-Type": "application/json" },
    }),
  );
}

describe("BlogPost type", () => {
  it("has featured_image_id and featured_image_url (not featured_image_key)", () => {
    // TypeScript compile-time check — if featured_image_key existed this would fail tsc
    expect("featured_image_id" in mockPost).toBe(true);
    expect("featured_image_url" in mockPost).toBe(true);
    expect("featured_image_key" in mockPost).toBe(false);
  });

  it("featured_image_id can be null or string", () => {
    const withImage: BlogPost = {
      ...mockPost,
      featured_image_id: "media-uuid",
      featured_image_url: "https://pub.example.com/media/image.jpg",
    };
    expect(withImage.featured_image_id).toBe("media-uuid");
    expect(withImage.featured_image_url).toBe(
      "https://pub.example.com/media/image.jpg",
    );
  });
});

describe("getBlogPost", () => {
  afterEach(() => vi.restoreAllMocks());

  it("calls GET /api/blog/posts/:slug and returns data", async () => {
    const spy = mockFetch({ data: mockPost });
    const result = await getBlogPost("test-post");
    expect(spy).toHaveBeenCalledWith(
      "/api/blog/posts/test-post",
      expect.any(Object),
    );
    expect(result.slug).toBe("test-post");
  });
});

describe("listBlogPosts", () => {
  afterEach(() => vi.restoreAllMocks());

  it("calls GET /api/blog/posts without token for public list", async () => {
    const spy = mockFetch({
      data: [mockPost],
      meta: {
        page: 0,
        per_page: 10,
        total: 1,
        next_page: null,
        prev_page: null,
        has_more: false,
      },
    });
    await listBlogPosts();
    expect(spy).toHaveBeenCalledWith(
      "/api/blog/posts",
      expect.objectContaining({ method: "GET" }),
    );
  });

  it("appends status query params when provided", async () => {
    const spy = mockFetch({
      data: [],
      meta: {
        page: 0,
        per_page: 10,
        total: 0,
        next_page: null,
        prev_page: null,
        has_more: false,
      },
    });
    await listBlogPosts({ status: ["draft", "published"] }, "token");
    const url = spy.mock.calls[0]![0] as string;
    expect(url).toContain("status=draft");
    expect(url).toContain("status=published");
  });
});

describe("createBlogPost", () => {
  afterEach(() => vi.restoreAllMocks());

  it("sends featured_image_media_id not featured_image_key", async () => {
    const spy = mockFetch({ data: mockPost });
    await createBlogPost("token", {
      title: "Test",
      featured_image_media_id: "media-uuid",
    });
    const body = JSON.parse(spy.mock.calls[0]![1]!.body as string);
    expect(body.featured_image_media_id).toBe("media-uuid");
    expect(body.featured_image_key).toBeUndefined();
  });
});

describe("updateBlogPost", () => {
  afterEach(() => vi.restoreAllMocks());

  it("sends featured_image_media_id not featured_image_key", async () => {
    const spy = mockFetch({ data: mockPost });
    await updateBlogPost("token", "post-1", {
      featured_image_media_id: "media-uuid",
    });
    const body = JSON.parse(spy.mock.calls[0]![1]!.body as string);
    expect(body.featured_image_media_id).toBe("media-uuid");
    expect(body.featured_image_key).toBeUndefined();
  });

  it("sends PATCH to /api/blog/posts/:id", async () => {
    const spy = mockFetch({ data: mockPost });
    await updateBlogPost("token", "post-1", { title: "New title" });
    expect(spy).toHaveBeenCalledWith(
      "/api/blog/posts/post-1",
      expect.objectContaining({ method: "PATCH" }),
    );
  });
});

describe("publishBlogPost", () => {
  afterEach(() => vi.restoreAllMocks());

  it("sends POST to /api/blog/posts/:id/publish with auth header", async () => {
    const spy = mockFetch({ data: { ...mockPost, status: "published" } });
    await publishBlogPost("my-token", "post-1");
    expect(spy).toHaveBeenCalledWith(
      "/api/blog/posts/post-1/publish",
      expect.objectContaining({
        method: "POST",
        headers: expect.objectContaining({ Authorization: "Bearer my-token" }),
      }),
    );
  });
});
