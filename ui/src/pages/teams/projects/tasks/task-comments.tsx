import { MentionText } from "@/components/mention-text";
import { MentionTextarea } from "@/components/mention-textarea";
import { Button } from "@/components/ui/button";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { ApiError } from "@/lib/error";
import {
  createTaskComment,
  deleteTaskComment,
  listTaskComments,
  updateTaskComment,
  type TaskComment,
} from "@/lib/task-comment-queries";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { formatDistanceToNow } from "date-fns";
import { Pencil, Trash2 } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

interface TaskCommentsProps {
  taskId: string;
  currentMemberId: string;
  isOwner?: boolean;
}

function CommentItem({
  comment,
  currentMemberId,
  isOwner,
  onDelete,
}: {
  comment: TaskComment;
  currentMemberId: string;
  isOwner?: boolean;
  onDelete: (id: string) => void;
}) {
  const [editing, setEditing] = useState(false);
  const [editContent, setEditContent] = useState(comment.content);
  const { user } = useAuthProvider();
  const queryClient = useQueryClient();

  const isAuthor = comment.created_by_member_id === currentMemberId;
  const canEdit = isAuthor;
  const canDelete = isAuthor || isOwner;

  const updateMutation = useMutation({
    mutationFn: async () => {
      return updateTaskComment(
        user!.tokens.access_token,
        comment.task_id,
        comment.id,
        { content: editContent },
      );
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: [{ key: "task-comments", task_id: comment.task_id }],
      });
      setEditing(false);
      toast.success("Comment updated");
    },
    onError: (err) => {
      toast.error(err instanceof ApiError ? err.message : "Failed to update comment");
    },
  });

  const authorName =
    comment.created_by_member?.user?.name ??
    comment.created_by_member?.user?.email ??
    "Unknown";

  return (
    <div className="flex flex-col gap-1 rounded-md border p-3">
      <div className="flex items-center justify-between gap-2">
        <div className="flex items-center gap-2 text-sm">
          <span className="font-medium">{authorName}</span>
          <span className="text-muted-foreground">
            {formatDistanceToNow(new Date(comment.created_at), { addSuffix: true })}
            {comment.updated_at !== comment.created_at && " (edited)"}
          </span>
        </div>
        <div className="flex items-center gap-1">
          {canEdit && !editing && (
            <Button
              variant="ghost"
              size="icon"
              className="h-6 w-6"
              onClick={() => {
                setEditContent(comment.content);
                setEditing(true);
              }}
            >
              <Pencil className="h-3 w-3" />
            </Button>
          )}
          {canDelete && (
            <Button
              variant="ghost"
              size="icon"
              className="h-6 w-6 text-destructive hover:text-destructive"
              onClick={() => onDelete(comment.id)}
            >
              <Trash2 className="h-3 w-3" />
            </Button>
          )}
        </div>
      </div>

      {editing ? (
        <div className="flex flex-col gap-2">
          <MentionTextarea
            value={editContent}
            onChange={setEditContent}
            placeholder="Edit your comment…"
          />
          <div className="flex gap-2">
            <Button
              size="sm"
              onClick={() => updateMutation.mutate()}
              disabled={updateMutation.isPending || editContent.trim() === ""}
            >
              Save
            </Button>
            <Button
              size="sm"
              variant="outline"
              onClick={() => setEditing(false)}
            >
              Cancel
            </Button>
          </div>
        </div>
      ) : (
        <MentionText content={comment.content} />
      )}
    </div>
  );
}

export function TaskComments({
  taskId,
  currentMemberId,
  isOwner,
}: TaskCommentsProps) {
  const { user } = useAuthProvider();
  const queryClient = useQueryClient();
  const [newContent, setNewContent] = useState("");

  const { data: comments = [], isLoading } = useQuery({
    queryKey: [{ key: "task-comments", task_id: taskId }],
    queryFn: () => listTaskComments(user!.tokens.access_token, taskId),
    enabled: !!taskId && !!user?.tokens.access_token,
  });

  const createMutation = useMutation({
    mutationFn: () =>
      createTaskComment(user!.tokens.access_token, taskId, {
        content: newContent,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: [{ key: "task-comments", task_id: taskId }],
      });
      setNewContent("");
      toast.success("Comment added");
    },
    onError: (err) => {
      toast.error(err instanceof ApiError ? err.message : "Failed to add comment");
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (commentId: string) =>
      deleteTaskComment(user!.tokens.access_token, taskId, commentId),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: [{ key: "task-comments", task_id: taskId }],
      });
      toast.success("Comment deleted");
    },
    onError: (err) => {
      toast.error(err instanceof ApiError ? err.message : "Failed to delete comment");
    },
  });

  return (
    <div className="flex flex-col gap-4 pt-4">
      <h3 className="text-sm font-semibold">Comments</h3>

      {isLoading ? (
        <p className="text-sm text-muted-foreground">Loading comments…</p>
      ) : comments.length === 0 ? (
        <p className="text-sm text-muted-foreground">No comments yet.</p>
      ) : (
        <div className="flex flex-col gap-2">
          {comments.map((comment) => (
            <CommentItem
              key={comment.id}
              comment={comment}
              currentMemberId={currentMemberId}
              isOwner={isOwner}
              onDelete={(id) => deleteMutation.mutate(id)}
            />
          ))}
        </div>
      )}

      <div className="flex flex-col gap-2">
        <MentionTextarea
          value={newContent}
          onChange={setNewContent}
          placeholder="Add a comment… Type @ to mention someone"
        />
        <Button
          size="sm"
          className="self-end"
          disabled={createMutation.isPending || newContent.trim() === ""}
          onClick={() => createMutation.mutate()}
        >
          Comment
        </Button>
      </div>
    </div>
  );
}
