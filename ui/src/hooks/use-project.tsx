import { taskProjectGet } from "@/lib/queries";
import { useQuery } from "@tanstack/react-query";
import { useParams } from "react-router";
import { useAuthProvider } from "./use-auth-provider";

/**
 *
 * @queryKey ["project", projectId]
 */
export function useProject() {
  const { user } = useAuthProvider();
  const { projectId } = useParams<{
    projectId: string;
    teamSlug: string;
  }>();
  return useQuery({
    queryKey: ["project", projectId],
    queryFn: async () => {
      return await taskProjectGet(user!.tokens.access_token, projectId!);
    },
    enabled: !!user?.tokens.access_token && !!projectId,
  });
}
