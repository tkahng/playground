import { describe, it, expect, vi, afterEach } from "vitest";
import {
  listMedia,
  getMedia,
  uploadMedia,
  updateMediaAltText,
  deleteMedia,
  type MediaItem,
} from "../media-queries";

const mockItem: MediaItem = {
  id: "media-1",
  storage_key: "media/uuid.jpg",
  url: "https://pub.example.com/media/uuid.jpg",
  mime_type: "image/jpeg",
  size: 102400,
  original_filename: "photo.jpg",
  alt_text: null,
  width: 1200,
  height: 800,
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-01T00:00:00Z",
};

const mockMeta = { page: 0, per_page: 10, total: 1, next_page: null, prev_page: null, has_more: false };

function mockFetch(body: unknown, status = 200) {
  return vi.spyOn(globalThis, "fetch").mockResolvedValue(
    new Response(JSON.stringify(body), {
      status,
      headers: { "Content-Type": "application/json" },
    }),
  );
}

function mockFetchEmpty(status = 204) {
  return vi.spyOn(globalThis, "fetch").mockResolvedValue(
    new Response(null, { status }),
  );
}

afterEach(() => vi.restoreAllMocks());

describe("listMedia", () => {
  it("calls GET /api/media with auth header", async () => {
    const spy = mockFetch({ data: [mockItem], meta: mockMeta });
    await listMedia("my-token");
    expect(spy).toHaveBeenCalledWith(
      "/api/media",
      expect.objectContaining({
        method: "GET",
        headers: expect.objectContaining({ Authorization: "Bearer my-token" }),
      }),
    );
  });

  it("appends page and per_page query params", async () => {
    const spy = mockFetch({ data: [], meta: mockMeta });
    await listMedia("token", { page: 2, per_page: 20 });
    const url = spy.mock.calls[0]![0] as string;
    expect(url).toContain("page=2");
    expect(url).toContain("per_page=20");
  });

  it("appends q search param", async () => {
    const spy = mockFetch({ data: [], meta: mockMeta });
    await listMedia("token", { q: "photo" });
    const url = spy.mock.calls[0]![0] as string;
    expect(url).toContain("q=photo");
  });

  it("returns MediaListResponse", async () => {
    mockFetch({ data: [mockItem], meta: mockMeta });
    const res = await listMedia("token");
    expect(res.data).toHaveLength(1);
    expect(res.data[0]!.storage_key).toBe("media/uuid.jpg");
    expect(res.meta.total).toBe(1);
  });
});

describe("getMedia", () => {
  it("calls GET /api/media/:id", async () => {
    const spy = mockFetch(mockItem);
    await getMedia("token", "media-1");
    expect(spy).toHaveBeenCalledWith(
      "/api/media/media-1",
      expect.objectContaining({ method: "GET" }),
    );
  });
});

describe("uploadMedia", () => {
  it("calls POST /api/media with FormData", async () => {
    const spy = mockFetchEmpty();
    const file = new File(["data"], "photo.jpg", { type: "image/jpeg" });
    await uploadMedia("token", [file]);
    expect(spy).toHaveBeenCalledWith(
      "/api/media",
      expect.objectContaining({ method: "POST" }),
    );
    const body = spy.mock.calls[0]![1]!.body;
    expect(body).toBeInstanceOf(FormData);
  });

  it("appends each file under 'files' field", async () => {
    const spy = mockFetchEmpty();
    const f1 = new File(["a"], "a.jpg", { type: "image/jpeg" });
    const f2 = new File(["b"], "b.jpg", { type: "image/jpeg" });
    await uploadMedia("token", [f1, f2]);
    const form = spy.mock.calls[0]![1]!.body as FormData;
    expect(form.getAll("files")).toHaveLength(2);
  });
});

describe("updateMediaAltText", () => {
  it("calls PATCH /api/media/:id with alt_text body", async () => {
    const spy = mockFetch({ ...mockItem, alt_text: "A beautiful photo" });
    await updateMediaAltText("token", "media-1", "A beautiful photo");
    expect(spy).toHaveBeenCalledWith(
      "/api/media/media-1",
      expect.objectContaining({ method: "PATCH" }),
    );
    const body = JSON.parse(spy.mock.calls[0]![1]!.body as string);
    expect(body.alt_text).toBe("A beautiful photo");
  });

  it("can clear alt_text by passing null", async () => {
    const spy = mockFetch({ ...mockItem, alt_text: null });
    await updateMediaAltText("token", "media-1", null);
    const body = JSON.parse(spy.mock.calls[0]![1]!.body as string);
    expect(body.alt_text).toBeNull();
  });
});

describe("deleteMedia", () => {
  it("calls DELETE /api/media/:id", async () => {
    const spy = mockFetchEmpty();
    await deleteMedia("token", "media-1");
    expect(spy).toHaveBeenCalledWith(
      "/api/media/media-1",
      expect.objectContaining({ method: "DELETE" }),
    );
  });

  it("resolves without error on 204", async () => {
    mockFetchEmpty(204);
    await expect(deleteMedia("token", "media-1")).resolves.toBeUndefined();
  });
});
