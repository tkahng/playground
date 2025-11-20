import { useAuthProvider } from "@/hooks/use-auth-provider";
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
      status,
      position,
    }: {
      projectId: string;
      taskId: string;
      status: "todo" | "in_progress" | "done";
      position: number;
    }) => {
      if (!user?.tokens.access_token) return;
      await updateTaskPositionStatus(user?.tokens.access_token, taskId, {
        status: status,
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
  return useMutation({
    mutationFn: async ({
      token,
      teamId,
    }: {
      token: string;
      teamId: string;
    }) => {
      await updateTeamMemberLastSelectedAt({
        token,
        teamId,
      });
    },
    onError: (error) => {
      toast.error("Failed to update last selected at", {
        description: error.message,
      });
    },
    onSuccess: () =>
      toast.success("Last selected at updated", { description: "Success" }),
  });
}
