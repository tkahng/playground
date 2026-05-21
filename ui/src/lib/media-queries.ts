import { ApiError } from "@/lib/error";

export interface MediaItem {
  id: string;
  storage_key: string;
  url: string;
  mime_type: string;
  size: number;
  original_filename: string;
  alt_text: string | null;
  width: number | null;
  height: number | null;
  created_at: string;
  updated_at: string;
}

export interface MediaMeta {
  page: number;
  per_page: number;
  total: number;
  next_page: number | null;
  prev_page: number | null;
  has_more: boolean;
}

export interface MediaListResponse {
  data: MediaItem[];
  meta: MediaMeta;
}

async function mediaFetch<T>(
  method: string,
  path: string,
  token: string,
  body?: unknown,
): Promise<T> {
  const isForm = body instanceof FormData;
  const headers: Record<string, string> = {
    Authorization: `Bearer ${token}`,
  };
  if (!isForm && body !== undefined) {
    headers["Content-Type"] = "application/json";
  }
  const res = await fetch(`/api${path}`, {
    method,
    headers,
    body: isForm ? body : body !== undefined ? JSON.stringify(body) : undefined,
  });
  if (!res.ok) {
    const errBody = await res.json().catch(() => ({ detail: res.statusText }));
    throw new ApiError(errBody.detail ?? res.statusText, undefined, res.status);
  }
  if (res.status === 204) return undefined as T;
  return res.json() as Promise<T>;
}

export const listMedia = async (
  token: string,
  params?: { page?: number; per_page?: number; q?: string },
): Promise<MediaListResponse> => {
  const query = new URLSearchParams();
  if (params?.page !== undefined) query.set("page", String(params.page));
  if (params?.per_page !== undefined) query.set("per_page", String(params.per_page));
  if (params?.q) query.set("q", params.q);
  const qs = query.toString() ? `?${query.toString()}` : "";
  return mediaFetch<MediaListResponse>("GET", `/media${qs}`, token);
};

export const getMedia = async (token: string, id: string): Promise<MediaItem> => {
  return mediaFetch<MediaItem>("GET", `/media/${id}`, token);
};

export const uploadMedia = async (token: string, files: File[]): Promise<void> => {
  const form = new FormData();
  for (const file of files) form.append("files", file);
  await mediaFetch<void>("POST", "/media", token, form);
};

export const updateMediaAltText = async (
  token: string,
  id: string,
  altText: string | null,
): Promise<MediaItem> => {
  const res = await mediaFetch<{ id: string; storage_key: string; url: string; mime_type: string; size: number; original_filename: string; alt_text: string | null; width: number | null; height: number | null; created_at: string; updated_at: string }>(
    "PATCH",
    `/media/${id}`,
    token,
    { alt_text: altText },
  );
  return res;
};

export const deleteMedia = async (token: string, id: string): Promise<void> => {
  await mediaFetch<void>("DELETE", `/media/${id}`, token);
};
