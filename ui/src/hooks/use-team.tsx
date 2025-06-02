import { getTeamBySlug } from "@/lib/queries";
import { Team } from "@/schema.types";
import { useQuery } from "@tanstack/react-query";
import { useParams } from "react-router";
import { useAuthProvider } from "./use-auth-provider";
import { useTeamContext } from "./use-team-context";

export function useTeamBySlug(slug?: string) {
  const { user } = useAuthProvider();
  return useQuery({
    queryKey: ["team-by-slug", slug],
    queryFn: async () => {
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token");
      }
      if (!slug) {
        throw new Error("Team slug is required");
      }
      const response = await getTeamBySlug(user.tokens.access_token, slug);

      return response;
    },
    enabled: !!user?.tokens.access_token && !!slug,
  });
}
export function useTeam(): {
  team: Team | null;
  isLoading: boolean;
  error: Error | null;
} {
  const { user } = useAuthProvider();
  const { teamSlug } = useParams<{ teamSlug: string }>();
  const { setTeam, team } = useTeamContext();
  const { data, isLoading, error } = useQuery({
    queryKey: ["team-by-slug"],
    queryFn: async () => {
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token");
      }
      if (!teamSlug) {
        throw new Error("Team slug is required");
      }
      const response = await getTeamBySlug(user.tokens.access_token, teamSlug);
      if (!team) {
        setTeam(response.team);
      }
      return response;
    },
  });
  return {
    team: data?.team || team,
    isLoading,
    error: error as Error | null,
  };
}
