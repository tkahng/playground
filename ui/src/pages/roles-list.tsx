import { DataTable } from "@/components/data-table";
import { RouteMap } from "@/components/route-map";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { rolesPaginate } from "@/lib/api";
import { RoleWithPermissions } from "@/schema.types";
import { ColumnDef, PaginationState, Updater } from "@tanstack/react-table";
import { useEffect, useState } from "react";
import { Link, useNavigate, useSearchParams } from "react-router";
export const columns: ColumnDef<RoleWithPermissions>[] = [
  {
    accessorKey: "id",
    header: "Id",
    cell: ({ row }) => (
      <Link to={`/dashboard/roles/${row.original.id}`}>
        {row.original.name}
      </Link>
    ),
  },
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
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const { user } = useAuthProvider();
  const [roles, setRoles] = useState<RoleWithPermissions[]>([]);
  const [rowCount, setRowCount] = useState(0);

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

  useEffect(() => {
    const fetchUsers = async () => {
      setLoading(true);
      if (!user) {
        navigate(RouteMap.SIGNIN);
        setLoading(false);

        return;
      }
      try {
        const { data, meta } = await rolesPaginate(user.tokens.access_token, {
          page: pageIndex + 1,
          per_page: pageSize,
        });
        if (!data) {
          setLoading(false);
          return;
        }
        console.log(data);
        setLoading(false);
        setRoles(data);
        setRowCount(meta.total);
      } catch (error) {
        console.error("Error fetching users:", error);
        setLoading(false);
      }
    };
    fetchUsers();
  }, [pageIndex, pageSize]);
  if (loading) {
    return <div>Loading...</div>;
  }

  return (
    <div className="flex w-full flex-col items-center justify-center">
      <h1>Users</h1>
      {roles.length && user && !loading && (
        <DataTable
          columns={columns}
          data={roles}
          onClick={(row) => {
            // @ts-ignore
            navigate(RouteMap.ROLE_EDIT.replace(":roleId", row.original.id));
          }}
          rowCount={rowCount}
          pagination={{ pageIndex, pageSize }}
          onPaginationChange={onPaginationChange}
        />
      )}
    </div>
  );
}
