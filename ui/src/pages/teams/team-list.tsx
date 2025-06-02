// page with list of teams to select from
// import { CreateTeamDialog } from "@/components/create-team-dialog";
// import { TeamCard } from "@/components/team-card";
// import { useToast } from "@/components/ui/use-toast";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useTeamContext } from "@/hooks/use-team-context";
import { getUserTeams } from "@/lib/queries";
import { Team } from "@/schema.types";
import { useQuery } from "@tanstack/react-query";
import { toast } from "sonner";
import { CreateTeamDialog } from "./create-team-dialog";

export default function TeamListPage() {
  const { user } = useAuthProvider();
  const { setTeam } = useTeamContext();

  const { data, isLoading, isError, error } = useQuery({
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

  if (isLoading) {
    return <div>Loading...</div>;
  }
  if (isError) {
    return <div>Error: {error.message}</div>;
  }

  const handleSelectTeam = (team: Team) => {
    setTeam(team);
    toast.success(`Selected team: ${team.name}`);
  };

  return (
    <div className="p-4">
      <h1 className="text-2xl font-bold mb-4">Select a Team</h1>
      <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
        {data?.data.map((team) => (
          <div key={team.id} onSelect={() => handleSelectTeam(team)}>
            {team.name}
          </div>
        ))}
      </div>
      <CreateTeamDialog />
      {/* <CreateTeamDialog />
      <Button className="mt-4" onClick={() => navigate(RouteMap.CREATE_TEAM)}>
        Create New Team
      </Button> */}
    </div>
  );
}
