import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useTeam } from "@/hooks/use-team";
import { getTeamMemberUnreadNotificationCount } from "@/lib/team-queries";
import { useQuery } from "@tanstack/react-query";

export function useUnreadNotificationCount() {
  const { user } = useAuthProvider();
  const { teamMember } = useTeam();
  const { data, isLoading, isError } = useQuery({
    queryKey: ["team-member-notifications-unread-count", teamMember?.id],
    queryFn: async () => {
      if (!user?.tokens.access_token) throw new Error("Missing access token");
      if (!teamMember?.id) throw new Error("Missing team member id");
      return getTeamMemberUnreadNotificationCount(
        user.tokens.access_token,
        teamMember.id
      );
    },
    refetchInterval: 60_000,
    enabled: !!user?.tokens.access_token && !!teamMember?.id,
  });
  return {
    unreadCount: data?.count ?? 0,
    isLoading,
    isError,
  };
}
