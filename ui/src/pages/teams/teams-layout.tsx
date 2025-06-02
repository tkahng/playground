import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useTeamContext } from "@/hooks/use-team-context";
import { getTeamBySlug } from "@/lib/queries";
import { useQuery } from "@tanstack/react-query";
import { Navigate, Outlet, useParams } from "react-router";

export default function TeamsLayout() {
  const { user } = useAuthProvider();
  const { teamSlug } = useParams<{ teamSlug: string }>();
  const { team, setTeam } = useTeamContext();
  const { data, isLoading, isError, error } = useQuery({
    queryKey: ["team-by-slug"],
    queryFn: async () => {
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token");
      }
      if (!teamSlug) {
        throw new Error("Team slug is required");
      }
      const response = await getTeamBySlug(user.tokens.access_token, teamSlug);
      setTeam(response.team);
      return response;
    },
  });
  if (!teamSlug) {
    <Navigate
      to={{
        pathname: "/teams",
      }}
    />;
  }
  if (!user) {
    return (
      <Navigate
        to={{
          pathname: "/signin",
          search: `?redirect_to=/teams/${teamSlug}`,
        }}
      />
    );
  }
  if (isLoading) {
    return <div>Loading...</div>;
  }
  if (isError) {
    return <div>Error: {error.message}</div>;
  }
  if (!data) {
    return <div>No team found with slug: {teamSlug}</div>;
  }

  return (
    // <TeamContext.Provider value={{ team: data || null }}>
    <Outlet context={{ team: team }} />
    // </TeamContext.Provider>
  );
}
