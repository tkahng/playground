import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useTeam } from "@/hooks/use-team";
import { taskQueries } from "@/lib/api";
import { getTeamTeamMembers } from "@/lib/team-queries";
import { useQuery } from "@tanstack/react-query";

export const useTeamTeamMembers = () => {
  const { user } = useAuthProvider();
  const { teamMember } = useTeam();
  return useQuery({
    queryKey: [
      {
        key: "team-team-members",
        team_id: teamMember?.team_id,
        page: 0,
        per_page: 20,
        active: true,
      },
    ],
    queryFn: async () => {
      return await getTeamTeamMembers({
        token: user!.tokens.access_token,
        teamId: teamMember!.team_id,
        page: 0,
        perPage: 20,
        active: true,
      });
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
