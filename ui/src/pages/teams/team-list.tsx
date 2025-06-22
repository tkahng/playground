import { CreateTeamDialog } from "@/components/create-team-dialog";
import { DataTable } from "@/components/data-table";
import { RouteMap } from "@/components/route-map";
import { useUserTeams } from "@/hooks/use-user-teams";
import { Team } from "@/schema.types";
import { PaginationState, Updater } from "@tanstack/react-table";
import { NavLink, useSearchParams } from "react-router";
import { toast } from "sonner";

export default function TeamListPage() {
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
  const { data, isLoading, isError, error } = useUserTeams();

  if (isLoading) {
    return <div>Loading...</div>;
  }
  if (isError) {
    return <div>Error: {error?.message}</div>;
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
        data={data?.data || []}
        rowCount={data?.meta.total || 0}
        paginationState={{ pageIndex, pageSize }}
        onPaginationChange={onPaginationChange}
        paginationEnabled
      />
    </div>
  );
}
