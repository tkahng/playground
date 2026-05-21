import { describe, it, expect, vi, beforeEach } from "vitest";
import { screen, fireEvent, waitFor } from "@testing-library/react";
import { render } from "@/test/test-utils";
import AdminMediaListPage from "../media-list";
import * as mediaQueries from "@/lib/media-queries";
import type { MediaItem } from "@/lib/media-queries";

vi.mock("sonner", () => ({ toast: { success: vi.fn(), error: vi.fn() } }));

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

const meta = { page: 0, per_page: 20, total: 1, next_page: null, prev_page: null, has_more: false };

function mockList(items: MediaItem[], total?: number) {
  vi.spyOn(mediaQueries, "listMedia").mockResolvedValue({
    data: items,
    meta: { ...meta, total: total ?? items.length },
  });
}

describe("AdminMediaListPage", () => {
  beforeEach(() => vi.restoreAllMocks());

  it("shows empty state when no media", async () => {
    mockList([]);
    render(<AdminMediaListPage />);
    expect(await screen.findByText(/no media yet/i)).toBeInTheDocument();
  });

  it("renders media items with filename and size", async () => {
    mockList([mockItem]);
    render(<AdminMediaListPage />);
    expect(await screen.findByText("photo.jpg")).toBeInTheDocument();
    expect(screen.getByText(/100\.0 KB/)).toBeInTheDocument();
  });

  it("renders image preview for image mime types", async () => {
    mockList([mockItem]);
    render(<AdminMediaListPage />);
    await screen.findByText("photo.jpg");
    const img = document.querySelector("img[src='https://pub.example.com/media/uuid.jpg']");
    expect(img).toBeInTheDocument();
  });

  it("shows total file count", async () => {
    mockList([mockItem]);
    render(<AdminMediaListPage />);
    expect(await screen.findByText("1 file")).toBeInTheDocument();
  });

  it("shows 'Add alt text' prompt when alt_text is null", async () => {
    mockList([{ ...mockItem, alt_text: null }]);
    render(<AdminMediaListPage />);
    await screen.findByText("photo.jpg");
    expect(screen.getByText("Add alt text…")).toBeInTheDocument();
  });

  it("shows existing alt text when set", async () => {
    mockList([{ ...mockItem, alt_text: "A mountain landscape" }]);
    render(<AdminMediaListPage />);
    await screen.findByText("photo.jpg");
    expect(screen.getByText("A mountain landscape")).toBeInTheDocument();
  });

  it("clicking alt text enters edit mode with input", async () => {
    mockList([{ ...mockItem, alt_text: "Old text" }]);
    render(<AdminMediaListPage />);
    await screen.findByText("photo.jpg");
    fireEvent.click(screen.getByText("Old text"));
    expect(screen.getByRole("textbox")).toHaveValue("Old text");
  });

  it("shows delete button for each item", async () => {
    mockList([mockItem]);
    render(<AdminMediaListPage />);
    await screen.findByText("photo.jpg");
    expect(screen.getByRole("button", { name: /delete media/i })).toBeInTheDocument();
  });

  it("shows dimensions when width/height available", async () => {
    mockList([mockItem]);
    render(<AdminMediaListPage />);
    await screen.findByText("photo.jpg");
    expect(screen.getByText(/1200×800/)).toBeInTheDocument();
  });

  it("shows Upload button", async () => {
    mockList([]);
    render(<AdminMediaListPage />);
    await screen.findByText(/no media yet/i);
    expect(screen.getByRole("button", { name: /upload/i })).toBeInTheDocument();
  });

  it("calls uploadMedia when files selected", async () => {
    mockList([]);
    const uploadSpy = vi.spyOn(mediaQueries, "uploadMedia").mockResolvedValue(undefined);
    render(<AdminMediaListPage />);
    await screen.findByText(/no media yet/i);
    const input = document.querySelector("input[type='file']") as HTMLInputElement;
    const file = new File(["img"], "test.jpg", { type: "image/jpeg" });
    fireEvent.change(input, { target: { files: [file] } });
    await waitFor(() => expect(uploadSpy).toHaveBeenCalledWith("test-access-token", [file]));
  });

  it("shows pagination when total exceeds page size", async () => {
    mockList(Array.from({ length: 20 }, (_, i) => ({ ...mockItem, id: `m-${i}`, original_filename: `f${i}.jpg` })), 50);
    render(<AdminMediaListPage />);
    await screen.findByText(/50 files/);
    expect(screen.getByRole("button", { name: /next/i })).toBeInTheDocument();
  });
});
