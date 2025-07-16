import { taskList } from "@/lib/api";
import { TaskStatus } from "@/schema.types";
import { useQuery } from "@tanstack/react-query";
import { useAuthProvider } from "./use-auth-provider";

/**
 *
 * @queryKey ["project-tasks", projectId]
 */
export const useProjectTasks = (
  projectId?: string,
  status?: TaskStatus,
  q?: string
) => {
  const { user } = useAuthProvider();
  return useQuery({
    queryKey: ["project-tasks", projectId],
    queryFn: async () => {
      return await taskList(user!.tokens.access_token, projectId!, {
        sort_by: "order",
        sort_order: "asc",
        per_page: 50,
        status: status ? [status] : undefined,
        q,
      });
    },
    enabled: !!user?.tokens.access_token && !!projectId,
  });
};
