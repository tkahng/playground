import { RouteMap } from "@/components/route-map";
import { Button } from "@/components/ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { permissionsPaginate } from "@/lib/api";
import { Permission } from "@/schema.types";
import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  getPaginationRowModel,
  PaginationState,
  Updater,
  useReactTable,
} from "@tanstack/react-table";
import { useEffect, useState } from "react";
import { useNavigate, useSearchParams } from "react-router";
export const columns: ColumnDef<Permission>[] = [
  {
    accessorKey: "id",
    header: "Id",
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

export default function PermissionListPage() {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const { user } = useAuthProvider();
  const [permissions, setPermissions] = useState<Permission[]>([]);
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
        const { data, meta } = await permissionsPaginate(
          user.tokens.access_token,
          {
            page: pageIndex + 1,
            per_page: pageSize,
          }
        );
        if (!data) {
          setLoading(false);
          return;
        }
        console.log(data);
        setLoading(false);
        setPermissions(data);
        setRowCount(meta.total);
      } catch (error) {
        console.error("Error fetching users:", error);
        setLoading(false);
      }
    };
    fetchUsers();
  }, [pageIndex, pageSize]);
  const table = useReactTable({
    onPaginationChange,
    data: permissions,
    columns,
    rowCount,
    // pageCount: Math.ceil(rowCount || 0 / (pagination?.pageSize ?? 10)),
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    manualPagination: true,
    state: {
      pagination: { pageIndex, pageSize },
    },
  });
  if (loading) {
    return <div>Loading...</div>;
  }

  return (
    <div className="flex w-full flex-col items-center justify-center">
      <h1>Users</h1>
      {permissions.length && user && !loading && (
        // <DataTable
        //   columns={columns}
        //   data={roles}
        //   onClick={(row) => {
        //     // @ts-ignore
        //     navigate(RouteMap.ROLE_EDIT.replace(":roleId", row.original.id));
        //   }}
        //   rowCount={rowCount}
        //   pagination={{ pageIndex, pageSize }}
        //   onPaginationChange={onPaginationChange}
        // />
        <>
          <div>
            <div className="rounded-md border">
              <Table>
                <TableHeader>
                  {table.getHeaderGroups().map((headerGroup) => (
                    <TableRow key={headerGroup.id}>
                      {headerGroup.headers.map((header) => {
                        return (
                          <TableHead key={header.id}>
                            {header.isPlaceholder
                              ? null
                              : flexRender(
                                  header.column.columnDef.header,
                                  header.getContext()
                                )}
                          </TableHead>
                        );
                      })}
                    </TableRow>
                  ))}
                </TableHeader>
                <TableBody>
                  {table.getRowModel().rows?.length ? (
                    table.getRowModel().rows.map((row) => (
                      <TableRow
                        key={row.id}
                        data-state={row.getIsSelected() && "selected"}
                        onClick={() => {
                          navigate(`/dashboard/permissions/${row.original.id}`);
                        }}
                      >
                        {row.getVisibleCells().map((cell) => (
                          <TableCell key={cell.id}>
                            {flexRender(
                              cell.column.columnDef.cell,
                              cell.getContext()
                            )}
                          </TableCell>
                        ))}
                      </TableRow>
                    ))
                  ) : (
                    <TableRow>
                      <TableCell
                        colSpan={columns.length}
                        className="h-24 text-center"
                      >
                        No results.
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </div>
            <div className="flex items-center justify-end space-x-2 py-4">
              <Button
                variant="outline"
                size="sm"
                onClick={() => table.previousPage()}
                disabled={!table.getCanPreviousPage()}
              >
                Previous
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => table.nextPage()}
                disabled={!table.getCanNextPage()}
              >
                Next
              </Button>
            </div>
          </div>
        </>
      )}
    </div>
  );
}
