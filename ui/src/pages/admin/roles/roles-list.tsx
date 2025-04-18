import { DataTable } from "@/components/data-table";
import { RouteMap } from "@/components/route-map";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { rolesPaginate } from "@/lib/queries";
import { CreateRoleDialog } from "@/pages/admin/roles/create-role-dialog";
import { RoleWithPermissions } from "@/schema.types";
import { useQuery } from "@tanstack/react-query";
import { ColumnDef, PaginationState, Updater } from "@tanstack/react-table";
import { useNavigate, useSearchParams } from "react-router";
export const columns: ColumnDef<RoleWithPermissions>[] = [
  {
    accessorKey: "name",
    header: "Name",
  },
  {
    accessorKey: "description",
    header: "Description",
  },
];

export default function RolesListPage() {
  const { user } = useAuthProvider();
  const navigate = useNavigate();

  const [searchParams, setSearchParams] = useSearchParams();
  const pageIndex = parseInt(searchParams.get("page") || "0", 10);
  const pageSize = parseInt(searchParams.get("per_page") || "10", 10);

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
    queryKey: ["roles-list"],
    queryFn: async () => {
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token or role ID");
      }
      const data = await rolesPaginate(user.tokens.access_token, {
        page: pageIndex + 1,
        per_page: pageSize,
      });
      if (!data.data) {
        throw new Error("No data returned from rolesPaginate");
      }
      return data;
    },
  });

  if (isLoading) {
    return <div>Loading...</div>;
  }
  if (isError) {
    return <div>Error: {error?.message}</div>;
  }
  const roles = data?.data || [];
  const rowCount = data?.meta.total || 0;

  return (
    // <div className="flex w-full flex-col items-center justify-center">
    <div className="h-full px-4 py-6 lg:px-8 space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Roles</h1>
        <CreateRoleDialog />
      </div>
      <p>
        Create and manage Roles for your applications. Roles contain collections
        of Permissions and can be assigned to Users.
      </p>
      <DataTable
        columns={columns}
        data={roles}
        onClick={(row) => {
          navigate(`${RouteMap.ADMIN_DASHBOARD_ROLES}/${row.original.id}`);
        }}
        rowCount={rowCount}
        paginationState={{ pageIndex, pageSize }}
        onPaginationChange={onPaginationChange}
        paginationEnabled
      />
    </div>
  );
}
