import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useTeam } from "@/hooks/use-team";
import { taskQueries } from "@/lib/api";
import { getTeamTeamMembers } from "@/lib/team-queries";
import { useQuery } from "@tanstack/react-query";

export const useTeamTeamMembers = () => {
  const { user } = useAuthProvider();
  const { teamMember } = useTeam();
  return useQuery({
    queryKey: ["team-members", teamMember?.team_id],
    queryFn: async () => {
      return await getTeamTeamMembers(
        user!.tokens.access_token,
        teamMember!.team_id,
        0,
        50
      );
    },
    enabled: !!teamMember?.team_id && !!user?.tokens.access_token,
  });
};

export const useTaskQuery = (taskId?: string) => {
  const { user } = useAuthProvider();
  return useQuery({
    queryKey: ["task", taskId],
    queryFn: async () => {
      return await taskQueries.findTaskById(user!.tokens.access_token, taskId!);
    },
    enabled: !!taskId && !!user?.tokens.access_token,
  });
};
