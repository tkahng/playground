import { getUserTeams } from "@/lib/api";
import { useQuery } from "@tanstack/react-query";
import { useAuthProvider } from "./use-auth-provider";

export const useUserTeams = () => {
  const { user } = useAuthProvider();
  const { data, isLoading, error, isError } = useQuery({
    queryKey: ["user-teams-list"],
    queryFn: async () => {
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token");
      }
      const { data, meta } = await getUserTeams(user.tokens.access_token);

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
