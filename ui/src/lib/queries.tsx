import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useTeam } from "@/hooks/use-team";
import { getMe } from "@/lib/api";
import { getTeamBySlug, getTeamTeamMembers } from "@/lib/team-queries";
import { TeamWithMember } from "@/schema.types";
import { useQuery } from "@tanstack/react-query";
import { useParams } from "@tanstack/react-router";
import { findTaskById } from "./task-queries";

export const useTeamTeamMembers = () => {
  const { user } = useAuthProvider();
  const { teamMember } = useTeam();
  return useQuery({
    queryKey: [
      {
        key: "team-team-members",
        team_id: teamMember?.team_id,
        page: 0,
        per_page: 100,
        active: true,
      },
    ],
    queryFn: async () => {
      return await getTeamTeamMembers({
        token: user!.tokens.access_token,
        teamId: teamMember!.team_id,
        page: 0,
        perPage: 100,
        active: true,
      });
    },
    enabled: !!teamMember?.team_id && !!user?.tokens.access_token,
  });
};

export const useTaskQuery = (taskId?: string) => {
  const { user } = useAuthProvider();
  return useQuery({
    queryKey: [{ key: "task", task_id: taskId }],
    queryFn: async () => {
      return await findTaskById(user!.tokens.access_token, taskId!);
    },
    enabled: !!taskId && !!user?.tokens.access_token,
  });
};

export const useMeQuery = () => {
  const { user } = useAuthProvider();
  return useQuery({
    queryKey: [{ key: "me" }],
    queryFn: async () => {
      if (!user?.tokens.access_token) {
        throw new Error("No access token");
      }
      return getMe(user!.tokens.access_token);
    },
    enabled: !!user?.tokens.access_token,
  });
};

export const useTeamBySlugQuery = () => {
  const { user } = useAuthProvider();
  const { teamSlug } = useParams({ strict: false });
  return useQuery({
    select: (data): TeamWithMember => {
      return {
        ...data.team,
        member: data.member,
      };
    },
    queryKey: [{ key: "team-by-slug", teamSlug }],
    queryFn: async () => {
      if (!user?.tokens.access_token) {
        throw new Error("No access token");
      }
      if (!teamSlug) {
        throw new Error("No team slug");
      }
      return getTeamBySlug(user!.tokens.access_token, teamSlug);
    },
    enabled: !!user?.tokens?.access_token && !!teamSlug,
  });
};
