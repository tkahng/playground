import { ChevronLeft, ChevronRight } from "lucide-react";
import { Pagination, PaginationContent, PaginationItem } from "./ui/pagination";
import { Button } from "./ui/button";
import { Table } from "@tanstack/react-table";

export function DataTablePagination<TData>({ table }: { table: Table<TData> }) {
  return (
    <>
      {table.getPageCount() > 1 && (
        <div className="mt-10">
          <Pagination>
            <PaginationContent>
              <PaginationItem>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => table.previousPage()}
                  disabled={!table.getCanPreviousPage()}
                  className="gap-1 bg-transparent"
                >
                  <ChevronLeft className="size-4" />
                  <span className="hidden sm:inline">Previous</span>
                </Button>
              </PaginationItem>

              <PaginationItem>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => table.nextPage()}
                  disabled={!table.getCanNextPage()}
                  className="gap-1 bg-transparent"
                >
                  <span className="hidden sm:inline">Next</span>
                  <ChevronRight className="size-4" />
                </Button>
              </PaginationItem>
            </PaginationContent>
          </Pagination>
        </div>
      )}
    </>
  );
}
