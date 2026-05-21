import { ApiError } from "@/lib/error";

export type BlogPostStatus = "draft" | "published" | "archived";
export type BlogContentFormat = "tiptap" | "markdown";

export interface BlogTag {
  id: string;
  name: string;
  slug: string;
  created_at: string;
}

export interface BlogPost {
  id: string;
  slug: string;
  title: string;
  content: string;
  content_format: BlogContentFormat;
  status: BlogPostStatus;
  author_id: string;
  published_at: string | null;
  featured_image_id: string | null;
  featured_image_url: string | null;
  seo_title: string | null;
  seo_description: string | null;
  reading_time_minutes: number | null;
  view_count: number;
  tags?: BlogTag[];
  created_at: string;
  updated_at: string;
}

export interface BlogMeta {
  page: number;
  per_page: number;
  total: number;
  next_page: number | null;
  prev_page: number | null;
  has_more: boolean;
}

export interface BlogPostListResponse {
  data: BlogPost[];
  meta: BlogMeta;
}

export interface CreateBlogPostInput {
  title: string;
  content?: string;
  content_format?: BlogContentFormat;
  featured_image_media_id?: string;
  seo_title?: string;
  seo_description?: string;
  tag_ids?: string[];
}

export interface UpdateBlogPostInput {
  title?: string;
  content?: string;
  content_format?: BlogContentFormat;
  /** null = clear the featured image; omit = leave unchanged. */
  featured_image_media_id?: string | null;
  /** null = clear; omit = leave unchanged. */
  seo_title?: string | null;
  /** null = clear; omit = leave unchanged. */
  seo_description?: string | null;
  tag_ids?: string[];
}

export interface CreateBlogTagInput {
  name: string;
}

async function blogFetch<T>(
  method: string,
  path: string,
  token?: string,
  body?: unknown,
): Promise<T> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }
  const res = await fetch(`/api${path}`, {
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });
  if (!res.ok) {
    const errBody = await res.json().catch(() => ({ detail: res.statusText }));
    throw new ApiError(errBody.detail ?? res.statusText, undefined, res.status);
  }
  if (res.status === 204) return undefined as T;
  const json = await res.json();
  return json;
}

// ── Public ───────────────────────────────────────────────────────────────────

export const listBlogPosts = async (
  params?: {
    page?: number;
    per_page?: number;
    sort_by?: string;
    sort_order?: "asc" | "desc";
    status?: BlogPostStatus[];
    author_id?: string;
    q?: string;
  },
  token?: string,
): Promise<BlogPostListResponse> => {
  const query = new URLSearchParams();
  if (params?.page !== undefined) query.set("page", String(params.page));
  if (params?.per_page !== undefined) query.set("per_page", String(params.per_page));
  if (params?.sort_by) query.set("sort_by", params.sort_by);
  if (params?.sort_order) query.set("sort_order", params.sort_order);
  if (params?.q) query.set("q", params.q);
  if (params?.author_id) query.set("author_id", params.author_id);
  if (params?.status?.length) {
    params.status.forEach((s) => query.append("status", s));
  }
  const qs = query.toString() ? `?${query.toString()}` : "";
  const res = await blogFetch<{ data: BlogPost[]; meta: BlogMeta }>(
    "GET",
    `/blog/posts${qs}`,
    token,
  );
  return res;
};

export const getBlogPost = async (slug: string, token?: string): Promise<BlogPost> => {
  const res = await blogFetch<{ data: BlogPost }>("GET", `/blog/posts/${slug}`, token);
  return res.data;
};

export const listBlogTags = async (): Promise<BlogTag[]> => {
  const res = await blogFetch<{ data: BlogTag[] }>("GET", "/blog/tags");
  return res.data;
};

// ── Admin ────────────────────────────────────────────────────────────────────

export const createBlogPost = async (
  token: string,
  input: CreateBlogPostInput,
): Promise<BlogPost> => {
  const res = await blogFetch<{ data: BlogPost }>("POST", "/blog/posts", token, input);
  return res.data;
};

export const updateBlogPost = async (
  token: string,
  postId: string,
  input: UpdateBlogPostInput,
): Promise<BlogPost> => {
  const res = await blogFetch<{ data: BlogPost }>("PATCH", `/blog/posts/${postId}`, token, input);
  return res.data;
};

export const publishBlogPost = async (token: string, postId: string): Promise<BlogPost> => {
  const res = await blogFetch<{ data: BlogPost }>("POST", `/blog/posts/${postId}/publish`, token);
  return res.data;
};

export const unpublishBlogPost = async (token: string, postId: string): Promise<BlogPost> => {
  const res = await blogFetch<{ data: BlogPost }>("POST", `/blog/posts/${postId}/unpublish`, token);
  return res.data;
};

export const archiveBlogPost = async (token: string, postId: string): Promise<BlogPost> => {
  const res = await blogFetch<{ data: BlogPost }>("POST", `/blog/posts/${postId}/archive`, token);
  return res.data;
};

export const deleteBlogPost = async (token: string, postId: string): Promise<void> => {
  await blogFetch<void>("DELETE", `/blog/posts/${postId}`, token);
};

export const createBlogTag = async (
  token: string,
  input: CreateBlogTagInput,
): Promise<BlogTag> => {
  const res = await blogFetch<{ data: BlogTag }>("POST", "/blog/tags", token, input);
  return res.data;
};
