import { useSearchParams } from "@/hooks/use-search-params";
import { CenteredSpinner } from "@/components/centered-spinner";
import { CreateTeamDialog } from "@/components/create-team-dialog";
import { DataTable } from "@/components/data-table";
import { RouteMap } from "@/components/route-map";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { getUserTeamMembers } from "@/lib/team-queries";
import { useQuery } from "@tanstack/react-query";
import { PaginationState, Updater } from "@tanstack/react-table";
import { Link } from "@tanstack/react-router";

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
    return <CenteredSpinner />;
  }
  if (isError) {
    return <div>Error: {error?.message}</div>;
  }

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
                <Link
                  to={`${RouteMap.TEAM_LIST}/${row.original.team?.slug}/dashboard`}
                  className="hover:underline text-blue-500"
                >
                  {row.original.team?.name}
                </Link>
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
