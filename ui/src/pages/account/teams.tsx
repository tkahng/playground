import { CreateTeamDialog } from "@/components/create-team-dialog";
import { DashboardSidebar } from "@/components/dashboard-sidebar";
import { DataTable } from "@/components/data-table";
import { accountSidebarLinks } from "@/components/links";
import { RouteMap } from "@/components/route-map";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useTeam } from "@/hooks/use-team";
import { useUserTeams } from "@/hooks/use-user-teams";
import { getUserSubscriptions, getUserTeams } from "@/lib/api";
import { GetError } from "@/lib/get-error";
import { Team } from "@/schema.types";
import { useQuery } from "@tanstack/react-query";
import { PaginationState, Updater } from "@tanstack/react-table";
import { NavLink, useNavigate, useSearchParams } from "react-router";
import { toast } from "sonner";

export default function AccountTeamsPage() {
  const navigate = useNavigate();
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
  const { user } = useAuthProvider();
  const { data, error, isError, isLoading } = useQuery({
    queryKey: ["stats"],
    queryFn: async () => {
      if (!user) {
        throw new Error("User not found");
      }

      // const stats = await getStats(user.tokens.access_token);
      const subs = await getUserSubscriptions(user.tokens.access_token);
      const teams = await getUserTeams(user.tokens.access_token);
      return {
        // ...stats,
        sub: subs,
        teams: teams.data,
      };
    },
  });

  const {
    data: teams,
    isLoading: teamsIsLoading,
    isError: teamsIsError,
    error: teamsError,
  } = useUserTeams();
  const { setTeam } = useTeam();
  const handleSelectTeam = (team: Team) => {
    toast.success(`Selected team: ${team.name}`);
    setTeam(team);
    navigate(`/teams/${team.slug}/dashboard`);
  };

  if (isLoading) {
    return <div>Loading...</div>;
  }
  if (isError) {
    return <div>Error: {error?.message}</div>;
  }
  if (!data) {
    return <div>No data</div>;
  }
  if (teamsIsLoading) {
    return <div>Loading...</div>;
  }
  if (teamsIsError) {
    const err = GetError(teamsError);
    return <div>Error: {err?.detail}</div>;
  }
  return (
    <div className="flex">
      <DashboardSidebar links={accountSidebarLinks} />
      <div className="flex-1 space-y-6 p-12 w-full">
        <div className="mx-auto px-8 py-8 justify-start items-stretch flex-1 max-w-[1200px]">
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
              {
                accessorKey: "role",
                header: "Member Role",
                cell: ({ row }) => {
                  const members = row.original.members;
                  if (!members || members.length === 0) {
                    return <span className="text-gray-500">No members</span>;
                  }
                  return (
                    <span className="text-gray-500">
                      {members[0].role || "Member"}
                    </span>
                  );
                },
              },
            ]}
            data={teams?.data || []}
            rowCount={teams?.meta.total || 0}
            paginationState={{ pageIndex, pageSize }}
            onPaginationChange={onPaginationChange}
            paginationEnabled
          />
        </div>
      </div>
    </div>
  );
}
