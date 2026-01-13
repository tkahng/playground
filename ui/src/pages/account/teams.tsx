import { CreateTeamDialog } from "@/components/create-team-dialog";
import { CreateTeamDisabledTooltip } from "@/components/create-team-disabled-tooltip";
import { DashboardSidebar } from "@/components/dashboard-sidebar";
import { DataTable } from "@/components/data-table";
import { accountSidebarLinks } from "@/components/links";
import { RouteMap } from "@/components/route-map";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { GetError } from "@/lib/error";
import { getUserTeamMembers } from "@/lib/team-queries";
import { useQuery } from "@tanstack/react-query";
import { PaginationState, Updater } from "@tanstack/react-table";
import { NavLink, useSearchParams } from "react-router";

export default function AccountTeamsPage() {
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
  const isUserVerified = !!user?.user?.email_verified_at;
  const { data, error, isError, isLoading } = useQuery({
    queryKey: [
      {
        key: "get-user-team-members",
        user_id: user?.user.id,
        page: pageIndex,
        per_page: pageSize,
      },
    ],
    queryFn: async () => {
      if (!user) {
        throw new Error("User not found");
      }

      // const stats = await getStats(user.tokens.access_token);
      const teams = await getUserTeamMembers({
        token: user.tokens.access_token,
        page: pageIndex,
        per_page: pageSize,
      });
      console.log({ teams });
      return teams;
    },
  });

  if (isLoading) {
    return <div>Loading...</div>;
  }
  if (isError) {
    const err = GetError(error);
    return <div>Error: {err?.detail}</div>;
  }
  if (!data) {
    return <div>No data</div>;
  }

  return (
    <div className="flex">
      <DashboardSidebar links={accountSidebarLinks} />
      <div className="flex-1 space-y-6 p-12 w-full">
        <div className="mx-auto px-8 py-8 justify-start items-stretch flex-1 max-w-[1200px]">
          <div className="flex items-center justify-between">
            <p>Create and manage Teams for your applications.</p>
            <div className="flex items-center space-x-2">
              <CreateTeamDialog />
              {!isUserVerified && <CreateTeamDisabledTooltip />}
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
                  return (
                    <span className="text-gray-500">{row.original.role}</span>
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
      </div>
    </div>
  );
}
