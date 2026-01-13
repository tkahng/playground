import { DataTable } from "@/components/data-table";
import { RouteMap } from "@/components/route-map";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { deletePermission, permissionsPaginate } from "@/lib/api";
import {
  keepPreviousData,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { PaginationState, Updater } from "@tanstack/react-table";
import { NavLink, useSearchParams } from "react-router";
import { CreatePermissionDialog } from "./create-permission-dialog";
import { PermissionsActionDropdown } from "./permissions-action-dropdown";

export default function PermissionListPage() {
  const { user } = useAuthProvider();
  const [searchParams, setSearchParams] = useSearchParams();
  const pageIndex = parseInt(searchParams.get("page") || "0", 10);
  const pageSize = parseInt(searchParams.get("per_page") || "10", 10);
  const queryClient = useQueryClient();
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
    queryKey: ["permissions-list", pageIndex, pageSize],
    queryFn: async () => {
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token or role ID");
      }
      const data = await permissionsPaginate(user.tokens.access_token, {
        page: pageIndex,
        per_page: pageSize,
      });
      return data;
    },
    placeholderData: keepPreviousData,
  });
  const deletePermissionMutation = useMutation({
    mutationFn: async (permissionId: string) => {
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token or role ID");
      }
      await deletePermission(user.tokens.access_token, permissionId);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["permissions-list"] });
    },
    onError: (error) => {
      console.error(error);
    },
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
        <p>
          Create and manage permissions for your applications. Permissions and
          can be assigned to Users.
        </p>
        <CreatePermissionDialog />
      </div>
      <DataTable
        columns={[
          {
            accessorKey: "name",
            header: "Name",
            cell: ({ row }) => {
              return (
                <NavLink
                  to={`${RouteMap.ADMIN_PERMISSIONS}/${row.original.id}`}
                  className="hover:underline text-blue-500"
                >
                  {row.original.name}
                </NavLink>
              );
            },
          },
          {
            accessorKey: "description",
            header: "Description",
          },
          {
            id: "actions",
            cell: ({ row }) => {
              return (
                <div className="flex items-center gap-2 justify-end">
                  <PermissionsActionDropdown
                    permissionId={row.original.id}
                    onDelete={deletePermissionMutation.mutate}
                  />
                </div>
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
