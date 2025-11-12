import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useTeam } from "@/hooks/use-team";
import { getTeamMemberNotifications } from "@/lib/team-queries";
import { useQuery } from "@tanstack/react-query";

export function useTeamNotifications() {
  const { user } = useAuthProvider();
  const { teamMember } = useTeam();
  const {
    data: notifications,
    isLoading,
    error,
    isError,
  } = useQuery({
    queryKey: ["team-member-notifications", teamMember?.id, 0, 10],
    queryFn: async () => {
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token");
      }
      if (!teamMember?.id) {
        throw new Error("Current team member team ID is required");
      }
      const notifications = await getTeamMemberNotifications(
        user.tokens.access_token,
        teamMember!.id,
        0,
        10
      );
      const data = notifications.data?.map((n) => {
        const payload = JSON.parse(n.payload) as {
          notification: {
            title: string;
            body: string;
          };
          data: Record<string, unknown>;
        };

        return {
          ...n,
          payload,
        };
      });
      return {
        data,
        meta: notifications.meta,
      };
    },
  });
  return {
    notifications,
    notificationsLoading: isLoading,
    notificationsError: error,
    notificationsIsError: isError,
  };
}
