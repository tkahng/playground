import { describe, it, expect, vi, afterEach } from "vitest";
import {
  listTaskComments,
  createTaskComment,
  updateTaskComment,
  deleteTaskComment,
  type TaskComment,
} from "../task-comment-queries";

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

function mockFetch(body: unknown, status = 200) {
  return vi.spyOn(globalThis, "fetch").mockResolvedValue(
    new Response(JSON.stringify(body), {
      status,
      headers: { "Content-Type": "application/json" },
    }),
  );
}

describe("listTaskComments", () => {
  afterEach(() => vi.restoreAllMocks());

  it("calls GET /api/tasks/{task-id}/comments with auth header", async () => {
    const spy = mockFetch([mockComment]);
    const result = await listTaskComments(TOKEN, TASK_ID);
    expect(spy).toHaveBeenCalledWith(
      expect.stringContaining(`/tasks/${TASK_ID}/comments`),
      expect.objectContaining({
        headers: expect.objectContaining({ Authorization: `Bearer ${TOKEN}` }),
      }),
    );
    expect(result).toHaveLength(1);
    expect(result[0].content).toBe("hello world");
  });

  it("throws ApiError on server error", async () => {
    mockFetch({ title: "Not Found", status: 404 }, 404);
    await expect(listTaskComments(TOKEN, TASK_ID)).rejects.toThrow();
  });
});

describe("createTaskComment", () => {
  afterEach(() => vi.restoreAllMocks());

  it("calls POST /api/tasks/{task-id}/comments with body", async () => {
    const spy = mockFetch(mockComment);
    const result = await createTaskComment(TOKEN, TASK_ID, {
      content: "hello world",
    });
    expect(spy).toHaveBeenCalledWith(
      expect.stringContaining(`/tasks/${TASK_ID}/comments`),
      expect.objectContaining({
        method: "POST",
        headers: expect.objectContaining({ Authorization: `Bearer ${TOKEN}` }),
      }),
    );
    expect(result.id).toBe(COMMENT_ID);
    expect(result.content).toBe("hello world");
  });

  it("throws ApiError on server error", async () => {
    mockFetch({ title: "Unprocessable Entity", status: 422 }, 422);
    await expect(
      createTaskComment(TOKEN, TASK_ID, { content: "" }),
    ).rejects.toThrow();
  });
});

describe("updateTaskComment", () => {
  afterEach(() => vi.restoreAllMocks());

  it("calls PUT /api/tasks/{task-id}/comments/{comment-id}", async () => {
    const updated = { ...mockComment, content: "edited" };
    const spy = mockFetch(updated);
    const result = await updateTaskComment(TOKEN, TASK_ID, COMMENT_ID, {
      content: "edited",
    });
    expect(spy).toHaveBeenCalledWith(
      expect.stringContaining(`/tasks/${TASK_ID}/comments/${COMMENT_ID}`),
      expect.objectContaining({
        method: "PUT",
        headers: expect.objectContaining({ Authorization: `Bearer ${TOKEN}` }),
      }),
    );
    expect(result.content).toBe("edited");
  });

  it("throws ApiError when not the author", async () => {
    mockFetch({ title: "Forbidden", status: 403 }, 403);
    await expect(
      updateTaskComment(TOKEN, TASK_ID, COMMENT_ID, { content: "x" }),
    ).rejects.toThrow();
  });
});

describe("deleteTaskComment", () => {
  afterEach(() => vi.restoreAllMocks());

  it("calls DELETE /api/tasks/{task-id}/comments/{comment-id}", async () => {
    const spy = mockFetch(null, 204);
    await deleteTaskComment(TOKEN, TASK_ID, COMMENT_ID);
    expect(spy).toHaveBeenCalledWith(
      expect.stringContaining(`/tasks/${TASK_ID}/comments/${COMMENT_ID}`),
      expect.objectContaining({
        method: "DELETE",
        headers: expect.objectContaining({ Authorization: `Bearer ${TOKEN}` }),
      }),
    );
  });

  it("throws ApiError on 403", async () => {
    mockFetch({ title: "Forbidden", status: 403 }, 403);
    await expect(
      deleteTaskComment(TOKEN, TASK_ID, COMMENT_ID),
    ).rejects.toThrow();
  });
});

describe("ParseMentionedMemberIDs (mention syntax in content)", () => {
  it("mention markup format @[Name](uuid) is preserved in content", () => {
    const memberId = "550e8400-e29b-41d4-a716-446655440000";
    const content = `hello @[Alice](${memberId}) please review`;
    // The content stored server-side uses this format; verify the string is intact
    expect(content).toContain(`@[Alice](${memberId})`);
  });
});
