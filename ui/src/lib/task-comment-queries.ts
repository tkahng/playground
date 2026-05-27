import { client } from "@/lib/client";
import { ApiError } from "@/lib/error";
import { components } from "@/schema";

// After running `npm run generate:schema` these types will be generated.
// components["schemas"]["TaskComment"] and the operation types below will be available.
export type TaskComment = {
  id: string;
  task_id: string;
  created_by_member_id: string;
  content: string;
  created_at: string;
  updated_at: string;
  created_by_member?: components["schemas"]["TeamMember"];
};

export type CreateTaskCommentBody = {
  content: string;
};

export type UpdateTaskCommentBody = {
  content: string;
};

export const listTaskComments = async (
  token: string,
  taskId: string,
): Promise<TaskComment[]> => {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const { data, error } = await (client as any).GET(
    `/api/tasks/{task-id}/comments`,
    {
      headers: { Authorization: `Bearer ${token}` },
      params: { path: { "task-id": taskId } },
    },
  );
  if (error) throw ApiError.fromErrorModel(error);
  return data ?? [];
};

export const createTaskComment = async (
  token: string,
  taskId: string,
  body: CreateTaskCommentBody,
): Promise<TaskComment> => {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const { data, error } = await (client as any).POST(
    `/api/tasks/{task-id}/comments`,
    {
      headers: { Authorization: `Bearer ${token}` },
      params: { path: { "task-id": taskId } },
      body,
    },
  );
  if (error) throw ApiError.fromErrorModel(error);
  return data;
};

export const updateTaskComment = async (
  token: string,
  taskId: string,
  commentId: string,
  body: UpdateTaskCommentBody,
): Promise<TaskComment> => {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const { data, error } = await (client as any).PUT(
    `/api/tasks/{task-id}/comments/{comment-id}`,
    {
      headers: { Authorization: `Bearer ${token}` },
      params: { path: { "task-id": taskId, "comment-id": commentId } },
      body,
    },
  );
  if (error) throw ApiError.fromErrorModel(error);
  return data;
};

export const deleteTaskComment = async (
  token: string,
  taskId: string,
  commentId: string,
): Promise<void> => {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const { error } = await (client as any).DELETE(
    `/api/tasks/{task-id}/comments/{comment-id}`,
    {
      headers: { Authorization: `Bearer ${token}` },
      params: { path: { "task-id": taskId, "comment-id": commentId } },
    },
  );
  if (error) throw ApiError.fromErrorModel(error);
};
