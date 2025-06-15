import { getUserTeams } from "@/lib/queries";
import { useQuery } from "@tanstack/react-query";
import { useAuthProvider } from "./use-auth-provider";

export const useUserTeams = () => {
  const { user } = useAuthProvider();
  return useQuery({
    queryKey: ["user-teams-list"],
    queryFn: async () => {
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token");
      }
      const { data, meta } = await getUserTeams(user.tokens.access_token);
      if (!data) {
        throw new Error("No teams found");
      }
      return { data, meta };
    },
  });
};
