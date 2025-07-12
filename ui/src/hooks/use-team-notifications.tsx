import { teamQueries } from "@/lib/queries";
import { useQuery } from "@tanstack/react-query";
import { useAuthProvider } from "./use-auth-provider";
import { useTeam } from "./use-team";

export function useTeamNotifications() {
  const { user } = useAuthProvider();
  const { teamMember } = useTeam();
  const {
    data: notifications,
    isLoading,
    error,
    isError,
  } = useQuery({
    queryKey: ["team-notifications"],
    queryFn: async () => {
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token");
      }
      if (!teamMember?.id) {
        throw new Error("Current team member team ID is required");
      }
      const notifications = await teamQueries.getTeamMemberNotifications(
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
