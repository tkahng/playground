import { describe, it, expect, vi, beforeEach } from "vitest";
import { screen, fireEvent, waitFor } from "@testing-library/react";
import { render } from "@/test/test-utils";
import { MediaPickerDialog } from "../media-picker-dialog";
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
  alt_text: "A photo",
  width: 1200,
  height: 800,
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-01T00:00:00Z",
};

const meta = { page: 0, per_page: 20, total: 1, next_page: null, prev_page: null, has_more: false };

function mockList(items: MediaItem[]) {
  vi.spyOn(mediaQueries, "listMedia").mockResolvedValue({
    data: items,
    meta: { ...meta, total: items.length },
  });
}

describe("MediaPickerDialog", () => {
  beforeEach(() => vi.restoreAllMocks());

  it("renders default trigger button", () => {
    mockList([]);
    render(<MediaPickerDialog onSelect={vi.fn()} />);
    expect(screen.getByRole("button", { name: /choose image/i })).toBeInTheDocument();
  });

  it("renders custom trigger when provided", () => {
    mockList([]);
    render(
      <MediaPickerDialog onSelect={vi.fn()} trigger={<button>Pick</button>} />,
    );
    expect(screen.getByRole("button", { name: "Pick" })).toBeInTheDocument();
  });

  it("opens dialog on trigger click", async () => {
    mockList([]);
    render(<MediaPickerDialog onSelect={vi.fn()} />);
    fireEvent.click(screen.getByRole("button", { name: /choose image/i }));
    expect(await screen.findByRole("dialog")).toBeInTheDocument();
    expect(screen.getByText("Select media")).toBeInTheDocument();
  });

  it("shows 'No media found' when list is empty", async () => {
    mockList([]);
    render(<MediaPickerDialog onSelect={vi.fn()} />);
    fireEvent.click(screen.getByRole("button", { name: /choose image/i }));
    expect(await screen.findByText(/no media found/i)).toBeInTheDocument();
  });

  it("renders image thumbnails for image items", async () => {
    mockList([mockItem]);
    render(<MediaPickerDialog onSelect={vi.fn()} />);
    fireEvent.click(screen.getByRole("button", { name: /choose image/i }));
    await screen.findByRole("dialog");
    const img = await screen.findByAltText("A photo");
    expect(img).toHaveAttribute("src", "https://pub.example.com/media/uuid.jpg");
  });

  it("calls onSelect with item when thumbnail clicked and closes dialog", async () => {
    mockList([mockItem]);
    const onSelect = vi.fn();
    render(<MediaPickerDialog onSelect={onSelect} />);
    fireEvent.click(screen.getByRole("button", { name: /choose image/i }));
    const btn = await screen.findByRole("button", { name: /select photo\.jpg/i });
    fireEvent.click(btn);
    await waitFor(() => expect(onSelect).toHaveBeenCalledWith(mockItem));
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("shows search input inside dialog", async () => {
    mockList([]);
    render(<MediaPickerDialog onSelect={vi.fn()} />);
    fireEvent.click(screen.getByRole("button", { name: /choose image/i }));
    await screen.findByRole("dialog");
    expect(screen.getByPlaceholderText("Search…")).toBeInTheDocument();
  });

  it("shows Upload button inside dialog", async () => {
    mockList([]);
    render(<MediaPickerDialog onSelect={vi.fn()} />);
    fireEvent.click(screen.getByRole("button", { name: /choose image/i }));
    await screen.findByRole("dialog");
    expect(screen.getByRole("button", { name: /upload/i })).toBeInTheDocument();
  });
});
