import { DataTable } from "@/components/data-table";
import { RouteMap } from "@/components/route-map";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { adminJobQueries } from "@/lib/queries";
import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { PaginationState, Updater } from "@tanstack/react-table";
import { NavLink, useSearchParams } from "react-router";

export default function JobsListPage() {
  const { user } = useAuthProvider();
  const [searchParams, setSearchParams] = useSearchParams();
  const pageIndex = parseInt(searchParams.get("page") || "0", 10);
  const pageSize = parseInt(searchParams.get("per_page") || "10", 10);
  // const queryClient = useQueryClient();
  const onPaginationChange = (updater: Updater<PaginationState>) => {
    const newState =
      typeof updater === "function"
        ? updater({ pageIndex, pageSize })
        : updater;
    if (newState.pageIndex !== pageIndex || newState.pageSize !== pageSize) {
      setSearchParams({
        page: String(newState.pageIndex),
        per_page: String(newState.pageSize),
      });
    }
  };

  const { data, isLoading, isError, error } = useQuery({
    queryKey: ["jobs-list", pageIndex, pageSize],
    queryFn: async () => {
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token or role ID");
      }
      const data = await adminJobQueries.getJobs(user.tokens.access_token, {
        page: pageIndex,
        per_page: pageSize,
      });
      return data;
    },
    placeholderData: keepPreviousData,
  });

  if (isLoading) {
    return <div>Loading...</div>;
  }
  if (isError) {
    return <div>Error: {error.message}</div>;
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <p>Manage Jobs for your applications.</p>
      </div>
      <DataTable
        columns={[
          {
            accessorKey: "id",
            header: "ID",
            cell: ({ row }) => {
              return (
                <NavLink
                  to={`${RouteMap.ADMIN_JOBS}/${row.original.id}`}
                  className="hover:underline text-blue-500"
                >
                  {row.original.id}
                </NavLink>
              );
            },
          },
          {
            accessorKey: "kind",
            header: "Kind",
            cell: ({ row }) => {
              return row.original.kind;
            },
          },
          {
            accessorKey: "unique_key",
            header: "UniqueKey",
            cell: ({ row }) => {
              return row.original.unique_key;
            },
          },
          {
            accessorKey: "status",
            header: "Status",
            cell: ({ row }) => {
              return row.original.status;
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
