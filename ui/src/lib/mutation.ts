import { useAuthProvider } from "@/hooks/use-auth-provider";
import { TaskCreateParams } from "@/schema.types";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { createTask, taskPositionStatus } from "./queries";

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
      await taskPositionStatus(user?.tokens.access_token, taskId, {
        status: status,
        position: position,
      });
      return;
    },
    onSuccess: async (_, variables) => {
      await queryClient.invalidateQueries({
        queryKey: ["project-with-tasks", variables.projectId],
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
        queryKey: ["project-tasks", projectId],
      });
      toast.success("Task created successfully");
    },
    onError: (error) => {
      toast.error(`Failed to create task: ${error.message}`);
    },
  });
}
