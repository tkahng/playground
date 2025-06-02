// page with list of teams to select from
// import { CreateTeamDialog } from "@/components/create-team-dialog";
// import { TeamCard } from "@/components/team-card";
// import { useToast } from "@/components/ui/use-toast";
import { DataTable } from "@/components/data-table";
import { RouteMap } from "@/components/route-map";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { getUserTeams } from "@/lib/queries";
import { Team } from "@/schema.types";
import { useQuery } from "@tanstack/react-query";
import { PaginationState, Updater } from "@tanstack/react-table";
import { NavLink, useSearchParams } from "react-router";
import { toast } from "sonner";
import { CreateTeamDialog } from "./create-team-dialog";

export default function TeamListPage() {
  const { user } = useAuthProvider();
  const [searchParams, setSearchParams] = useSearchParams();
  const pageIndex = parseInt(searchParams.get("page") || "0", 10);
  const pageSize = parseInt(searchParams.get("per_page") || "10", 10);

  const onPaginationChange = (updater: Updater<PaginationState>) => {
    const newState =
      typeof updater === "function"
        ? updater({ pageIndex, pageSize })
        : updater;
    setSearchParams({
      page: String(newState.pageIndex),
      per_page: String(newState.pageSize),
    });
  };
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
    toast.success(`Selected team: ${team.name}`);
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <p>Create and manage Teams for your applications.</p>
        <CreateTeamDialog />
      </div>

      <DataTable
        columns={[
          {
            accessorKey: "name",
            header: "Name",
            cell: ({ row }) => {
              return (
                <NavLink
                  to={`${RouteMap.TEAM_LIST}/${row.original.slug}/dashboard`}
                  className="hover:underline text-blue-500"
                  onClick={() => handleSelectTeam(row.original)}
                >
                  {row.original.name}
                </NavLink>
              );
            },
          },
        ]}
        data={data?.data || []}
        rowCount={data?.meta.total || 0}
        paginationState={{ pageIndex, pageSize }}
        onPaginationChange={onPaginationChange}
        paginationEnabled
      />
    </div>
  );
}
