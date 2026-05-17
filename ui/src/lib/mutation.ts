import { useAuthProvider } from "@/hooks/use-auth-provider";
import { ApiError } from "@/lib/error";
import { updateTeamMemberLastSelectedAt } from "@/lib/team-queries";
import { TaskCreateParams } from "@/schema.types";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { createTask, updateTaskPositionStatus } from "./task-queries";

export function useUpdateTaskPosition() {
  const { user } = useAuthProvider();
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({
      taskId,
      workflowStatusId,
      position,
    }: {
      projectId: string;
      taskId: string;
      workflowStatusId: string;
      position: number;
    }) => {
      if (!user?.tokens.access_token) return;
      await updateTaskPositionStatus(user?.tokens.access_token, taskId, {
        workflow_status_id: workflowStatusId,
        position: position,
      });
      return;
    },
    onSuccess: async (_, variables) => {
      await queryClient.invalidateQueries({
        queryKey: [{ key: "project-tasks", project_id: variables.projectId }],
      });
      toast.success("Task updated");
    },
    onError: (error) => {
      toast.error("Failed to update task", {
        description: error.message,
      });
    },
  });
}

export function useCreateProjectTask(projectId: string, onSuccess: () => void) {
  const { user } = useAuthProvider();
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (values: TaskCreateParams) => {
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token");
      }
      await createTask(user.tokens.access_token, projectId, values);
    },
    onSuccess: async () => {
      onSuccess();
      await queryClient.invalidateQueries({
        queryKey: [{ key: "project-tasks", project_id: projectId }],
      });
      toast.success("Task created successfully");
    },
    onError: (error) => {
      toast.error(`Failed to create task: ${error.message}`);
    },
  });
}

export function useUpdateMemberLastSelectedAt() {
  const { user } = useAuthProvider();
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ teamId }: { teamId: string }) => {
      if (!user?.tokens.access_token) {
        throw new ApiError("Missing access token");
      }
      await updateTeamMemberLastSelectedAt({
        token: user.tokens.access_token,
        teamId,
      });
    },
    onSuccess: async () => {
      queryClient.invalidateQueries({
        queryKey: [
          { key: "get-user-team-members", sort_by: "last_selected_at" },
        ],
      });
      toast.success("Member last selected at updated");
    },
    onError: (error) => {
      queryClient.invalidateQueries({
        queryKey: [
          { key: "get-user-team-members", sort_by: "last_selected_at" },
        ],
      });
      toast.error("Failed to update member last selected at", {
        description: error.message,
      });
    },
  });
}
