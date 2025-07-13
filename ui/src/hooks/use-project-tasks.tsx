import { taskList } from "@/lib/queries";
import { useQuery } from "@tanstack/react-query";
import { useAuthProvider } from "./use-auth-provider";

/**
 *
 * @queryKey ["project-tasks", projectId]
 */
export const useProjectTasks = (projectId: string) => {
  const { user } = useAuthProvider();
  return useQuery({
    queryKey: ["project-tasks", projectId],
    queryFn: async () => {
      return await taskList(user!.tokens.access_token, projectId, {
        sort_by: "order",
        sort_order: "asc",
        per_page: 50,
      });
    },
    enabled: !!user?.tokens.access_token,
  });
};
