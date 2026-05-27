import { describe, it, expect, vi, afterEach } from "vitest";
import {
  listTaskComments,
  createTaskComment,
  updateTaskComment,
  deleteTaskComment,
  type TaskComment,
} from "../task-comment-queries";

// Mock the openapi client so URL parsing never runs (the client constructs
// absolute URLs internally, which fails in Node.js with relative-path bases).
vi.mock("@/lib/client", () => ({
  client: {
    GET: vi.fn(),
    POST: vi.fn(),
    PUT: vi.fn(),
    DELETE: vi.fn(),
  },
}));

// Import the mock after vi.mock() is hoisted
import { client } from "@/lib/client";

const TOKEN = "test-token";
const TASK_ID = "task-uuid-1";
const COMMENT_ID = "comment-uuid-1";

const mockComment: TaskComment = {
  id: COMMENT_ID,
  task_id: TASK_ID,
  created_by_member_id: "member-uuid-1",
  content: "hello world",
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-01T00:00:00Z",
};

afterEach(() => vi.clearAllMocks());

describe("listTaskComments", () => {
  it("calls GET /api/tasks/{task-id}/comments with auth header", async () => {
    vi.mocked((client as any).GET).mockResolvedValue({
      data: [mockComment],
      error: null,
    });

    const result = await listTaskComments(TOKEN, TASK_ID);

    expect((client as any).GET).toHaveBeenCalledWith(
      `/api/tasks/{task-id}/comments`,
      expect.objectContaining({
        headers: expect.objectContaining({ Authorization: `Bearer ${TOKEN}` }),
        params: { path: { "task-id": TASK_ID } },
      }),
    );
    expect(result).toHaveLength(1);
    expect(result[0]!.content).toBe("hello world");
  });

  it("returns empty array when data is null", async () => {
    vi.mocked((client as any).GET).mockResolvedValue({ data: null, error: null });
    const result = await listTaskComments(TOKEN, TASK_ID);
    expect(result).toEqual([]);
  });

  it("throws ApiError on server error", async () => {
    vi.mocked((client as any).GET).mockResolvedValue({
      data: null,
      error: { title: "Not Found", status: 404 },
    });
    await expect(listTaskComments(TOKEN, TASK_ID)).rejects.toThrow();
  });
});

describe("createTaskComment", () => {
  it("calls POST /api/tasks/{task-id}/comments with body", async () => {
    vi.mocked((client as any).POST).mockResolvedValue({
      data: mockComment,
      error: null,
    });

    const result = await createTaskComment(TOKEN, TASK_ID, {
      content: "hello world",
    });

    expect((client as any).POST).toHaveBeenCalledWith(
      `/api/tasks/{task-id}/comments`,
      expect.objectContaining({
        headers: expect.objectContaining({ Authorization: `Bearer ${TOKEN}` }),
        params: { path: { "task-id": TASK_ID } },
        body: { content: "hello world" },
      }),
    );
    expect(result.id).toBe(COMMENT_ID);
    expect(result.content).toBe("hello world");
  });

  it("throws ApiError on server error", async () => {
    vi.mocked((client as any).POST).mockResolvedValue({
      data: null,
      error: { title: "Unprocessable Entity", status: 422 },
    });
    await expect(
      createTaskComment(TOKEN, TASK_ID, { content: "" }),
    ).rejects.toThrow();
  });
});

describe("updateTaskComment", () => {
  it("calls PUT /api/tasks/{task-id}/comments/{comment-id}", async () => {
    const updated = { ...mockComment, content: "edited" };
    vi.mocked((client as any).PUT).mockResolvedValue({
      data: updated,
      error: null,
    });

    const result = await updateTaskComment(TOKEN, TASK_ID, COMMENT_ID, {
      content: "edited",
    });

    expect((client as any).PUT).toHaveBeenCalledWith(
      `/api/tasks/{task-id}/comments/{comment-id}`,
      expect.objectContaining({
        headers: expect.objectContaining({ Authorization: `Bearer ${TOKEN}` }),
        params: { path: { "task-id": TASK_ID, "comment-id": COMMENT_ID } },
        body: { content: "edited" },
      }),
    );
    expect(result.content).toBe("edited");
  });

  it("throws ApiError when not the author", async () => {
    vi.mocked((client as any).PUT).mockResolvedValue({
      data: null,
      error: { title: "Forbidden", status: 403 },
    });
    await expect(
      updateTaskComment(TOKEN, TASK_ID, COMMENT_ID, { content: "x" }),
    ).rejects.toThrow();
  });
});

describe("deleteTaskComment", () => {
  it("calls DELETE /api/tasks/{task-id}/comments/{comment-id}", async () => {
    vi.mocked((client as any).DELETE).mockResolvedValue({
      data: null,
      error: null,
    });

    await deleteTaskComment(TOKEN, TASK_ID, COMMENT_ID);

    expect((client as any).DELETE).toHaveBeenCalledWith(
      `/api/tasks/{task-id}/comments/{comment-id}`,
      expect.objectContaining({
        headers: expect.objectContaining({ Authorization: `Bearer ${TOKEN}` }),
        params: { path: { "task-id": TASK_ID, "comment-id": COMMENT_ID } },
      }),
    );
  });

  it("throws ApiError on 403", async () => {
    vi.mocked((client as any).DELETE).mockResolvedValue({
      data: null,
      error: { title: "Forbidden", status: 403 },
    });
    await expect(
      deleteTaskComment(TOKEN, TASK_ID, COMMENT_ID),
    ).rejects.toThrow();
  });
});

describe("ParseMentionedMemberIDs (mention syntax in content)", () => {
  it("mention markup format @[Name](uuid) is preserved in content", () => {
    const memberId = "550e8400-e29b-41d4-a716-446655440000";
    const content = `hello @[Alice](${memberId}) please review`;
    expect(content).toContain(`@[Alice](${memberId})`);
  });
});
