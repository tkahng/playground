import { taskProjectGet } from "@/lib/task-queries";
import { useQuery } from "@tanstack/react-query";
import { useParams } from "@tanstack/react-router";
import { useAuthProvider } from "./use-auth-provider";

/**
 *
 * @queryKey ["project", projectId]
 */
export function useProject() {
  const { user } = useAuthProvider();
  const { projectId } = useParams({ strict: false });
  return useQuery({
    queryKey: [{ key: "project", project_id: projectId }],
    queryFn: async () => {
      return await taskProjectGet(user!.tokens.access_token, projectId!);
    },
    enabled: !!user?.tokens.access_token && !!projectId,
  });
}
