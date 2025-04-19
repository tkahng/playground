import { Button } from "@/components/ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  getPaginationRowModel,
  OnChangeFn,
  PaginationState,
  Table as ReactTable,
  Row,
  useReactTable,
} from "@tanstack/react-table";

interface DataTableProps<TData, TValue> {
  onClick?: (row: Row<TData>) => void;
  columns: ColumnDef<TData, TValue>[];
  data: TData[];
  rowCount?: number;
  paginationState?: PaginationState;
  onPaginationChange?: OnChangeFn<PaginationState>;
  paginationEnabled?: boolean;
}

export function DataTable<TData, TValue>({
  columns,
  data,
  paginationState: pagination,
  rowCount,
  onClick,
  onPaginationChange,
  paginationEnabled = false,
}: DataTableProps<TData, TValue>) {
  const table = useReactTable({
    onPaginationChange,
    data,
    columns,
    rowCount,
    // pageCount: Math.ceil(rowCount || 0 / (pagination?.pageSize ?? 10)),
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    manualPagination: true,
    state: {
      pagination,
    },
  });

  return (
    <div>
      <div>
        <DataTableBody table={table} onClick={onClick} columns={columns} />
      </div>
      <div className="flex items-center justify-end space-x-2 py-4">
        {paginationEnabled && <DataTableFooter table={table} />}
      </div>
    </div>
  );
}

export function DataTableFooter<TData>({
  table,
}: {
  table: ReactTable<TData>;
}) {
  return (
    <>
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
    </>
  );
}
export function DataTableBody<TData, TValue>({
  table,
  onClick,
  columns,
}: {
  table: ReactTable<TData>;
  onClick: ((row: Row<TData>) => void) | undefined;
  columns: ColumnDef<TData, TValue>[];
}) {
  return (
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
      <TableBody className="border-b">
        {table.getRowModel().rows?.length ? (
          table.getRowModel().rows.map((row) => (
            <TableRow
              key={row.id}
              data-state={row.getIsSelected() && "selected"}
              onClick={() => {
                if (onClick) {
                  onClick(row);
                }
              }}
            >
              {row.getVisibleCells().map((cell) => (
                <TableCell key={cell.id}>
                  {flexRender(cell.column.columnDef.cell, cell.getContext())}
                </TableCell>
              ))}
            </TableRow>
          ))
        ) : (
          <TableRow>
            <TableCell colSpan={columns.length} className="h-24 text-center">
              No results.
            </TableCell>
          </TableRow>
        )}
      </TableBody>
    </Table>
  );
}
