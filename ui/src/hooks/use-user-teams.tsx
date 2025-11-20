import { getUserTeamMembers } from "@/lib/team-queries";
import { useQuery } from "@tanstack/react-query";
import { useAuthProvider } from "./use-auth-provider";

export const useUserTeams = () => {
  const { user } = useAuthProvider();
  const { data, isLoading, error, isError } = useQuery({
    queryKey: [
      {
        key: "get-user-teams",
        user_id: user?.user.id,
        page: 0,
        per_page: 20,
      },
    ] as const,
    queryFn: async () => {
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token");
      }
      const { data, meta } = await getUserTeamMembers({
        token: user.tokens.access_token,
        page: 0,
        per_page: 20,
      });

      return { data: data || [], meta };
    },
    enabled: !!user?.tokens.access_token,
  });
  if (!user?.tokens.access_token) {
    return {
      data: null,
      isError: true,
      isLoading: false,
      error: new Error("User is not authenticated"),
    };
  }
  return {
    data: data,
    isError,
    isLoading,
    error,
  };
};
