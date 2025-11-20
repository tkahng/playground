import { CreateTeamDialog } from "@/components/create-team-dialog";
import { CreateTeamDisabledTooltip } from "@/components/create-team-disabled-tooltip";
import { DataTable } from "@/components/data-table";
import { RouteMap } from "@/components/route-map";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { GetError } from "@/lib/error";
import { getUserTeamMembers } from "@/lib/team-queries";
import { Team } from "@/schema.types";
import { useQuery } from "@tanstack/react-query";
import { PaginationState, Updater } from "@tanstack/react-table";
import { NavLink, useSearchParams } from "react-router";
import { toast } from "sonner";

export default function TeamSelect() {
  const { user } = useAuthProvider();
  const isVerified = !!user?.user?.email_verified_at;
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
    queryKey: [
      {
        key: "get-user-team-members",
        user_id: user?.user.id,
        page: pageIndex,
        per_page: pageSize,
      },
    ],
    queryFn: async () => {
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token");
      }
      const { data, meta } = await getUserTeamMembers({
        token: user.tokens.access_token,
        page: pageIndex,
        per_page: pageSize,
      });
      return { data: data, meta };
    },
  });

  if (isLoading) {
    return <div>Loading...</div>;
  }
  if (isError) {
    const err = GetError(error);
    if (err) {
      return <div>Error: {err.detail}</div>;
    }
    return <div>Error: {error?.message}</div>;
  }

  const handleSelectTeam = (team: Team) => {
    toast.success(`Selected team: ${team.name}`);
  };

  return (
    <div className="space-y-6 mx-auto px-8 py-8 justify-start items-stretch flex-1 max-w-[1200px]">
      <div className="flex items-center justify-between">
        <p>Create and manage Teams for your applications.</p>
        <div className="flex items-center space-x-2">
          <CreateTeamDialog />
          {!isVerified && <CreateTeamDisabledTooltip />}
        </div>
      </div>

      <DataTable
        columns={[
          {
            accessorKey: "name",
            header: "Name",
            cell: ({ row }) => {
              return (
                <NavLink
                  to={`${RouteMap.TEAM_LIST}/${row.original.team?.slug}/dashboard`}
                  className="hover:underline text-blue-500"
                  onClick={() => handleSelectTeam(row.original.team!)}
                >
                  {row.original.team?.name}
                </NavLink>
              );
            },
          },
          {
            accessorKey: "role",
            header: "Member Role",
            cell: ({ row }) => {
              return <span className="text-gray-500">{row.original.role}</span>;
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
