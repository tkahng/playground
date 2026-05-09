import { useSearchParams } from "@/hooks/use-search-params";
import { CenteredSpinner } from "@/components/centered-spinner";
import { DataTable } from "@/components/data-table";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { adminStripeProducts } from "@/lib/api";
import { useQuery } from "@tanstack/react-query";
import { PaginationState, Updater } from "@tanstack/react-table";
import { Ellipsis, Pencil } from "lucide-react";
import { useState } from "react";
import { Link, useNavigate } from "@tanstack/react-router";
// import { CreateUserDialog } from "./create-user-dialog";
export default function ProductsListPage() {
  const { user } = useAuthProvider();

  const [searchParams, setSearchParams] = useSearchParams();
  const pageIndex = parseInt(searchParams.get("page") || "0", 10);
  const pageSize = parseInt(searchParams.get("per_page") || "10", 10);
  const sortBy = searchParams.get("sort_by") || "updated_at";
  const sortOrder = (searchParams.get("sort_order") || "desc") as "asc" | "desc";

  const onPaginationChange = (updater: Updater<PaginationState>) => {
    const newState =
      typeof updater === "function"
        ? updater({ pageIndex, pageSize })
        : updater;
    setSearchParams({
      page: String(newState.pageIndex),
      per_page: String(newState.pageSize),
      sort_by: sortBy,
      sort_order: sortOrder,
    });
  };
  const { data, isLoading, isError, error } = useQuery({
    queryKey: ["products-list", pageIndex, pageSize, sortBy, sortOrder],
    queryFn: async () => {
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token");
      }
      const data = await adminStripeProducts(user.tokens.access_token, {
        page: pageIndex,
        per_page: pageSize,
        sort_by: sortBy,
        sort_order: sortOrder,
        expand: ["prices", "permissions"],
      });
      return data;
    },
  });

  if (isLoading) {
    return <CenteredSpinner />;
  }
  if (isError) {
    return <div>Error: {error.message}</div>;
  }

  return (
    <div className="space-y-6">
      <p>
        Edit product roles and permissions. For more information, visit the
        stripe dashboard.
      </p>
      <div className="flex items-center gap-2">
        <Select
          value={sortBy}
          onValueChange={(v) =>
            setSearchParams({ page: "0", per_page: String(pageSize), sort_by: v, sort_order: sortOrder })
          }
        >
          <SelectTrigger className="w-[160px]">
            <SelectValue placeholder="Sort by" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="updated_at">Updated At</SelectItem>
            <SelectItem value="created_at">Created At</SelectItem>
          </SelectContent>
        </Select>
        <Select
          value={sortOrder}
          onValueChange={(v) =>
            setSearchParams({ page: "0", per_page: String(pageSize), sort_by: sortBy, sort_order: v })
          }
        >
          <SelectTrigger className="w-[120px]">
            <SelectValue placeholder="Order" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="desc">Descending</SelectItem>
            <SelectItem value="asc">Ascending</SelectItem>
          </SelectContent>
        </Select>
      </div>

      <DataTable
        columns={[
          {
            accessorKey: "id",
            header: "ID",
            cell: ({ row }) => {
              return (
                <Link
                  to='/admin/products/$productId' params={{ productId: row.original.id }}
                  className="hover:underline text-blue-500"
                >
                  {row.original.id}
                </Link>
              );
            },
          },
          {
            accessorKey: "active",
            header: "Active",
          },
          {
            id: "permissions",
            header: "Permissions",
            cell: ({ row }) => {
              return row.original.permissions?.map((r) => r.name).join(",");
            },
          },
          {
            accessorKey: "created_at",
            header: "Created At",
            cell: ({ row }) => {
              return new Date(row.original.created_at).toLocaleDateString();
            },
          },
          {
            accessorKey: "updated_at",
            header: "Updated At",
            cell: ({ row }) => {
              return new Date(row.original.updated_at).toLocaleDateString();
            },
          },
          {
            id: "actions",
            cell: ({ row }) => {
              return (
                <div className="flex flex-row gap-2 justify-end">
                  <ProductEllipsisDropdown productId={row.original.id} />
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

function ProductEllipsisDropdown({ productId }: { productId: string }) {
  // const editDialog = useDialog();
  const navigate = useNavigate();
  const [dropdownOpen, setDropdownOpen] = useState(false);
  return (
    <>
      <DropdownMenu open={dropdownOpen} onOpenChange={setDropdownOpen}>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" size="icon">
            <Ellipsis className="h-4 w-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent>
          <DropdownMenuItem
            onSelect={() => {
              setDropdownOpen(false);
              navigate({ to: '/admin/products/$productId', params: { productId } });
            }}
          >
            <Button variant="ghost" size="sm">
              <Pencil className="h-4 w-4" />
              <span>Edit</span>
            </Button>
          </DropdownMenuItem>
          <DropdownMenuItem
            onSelect={() => {
              setDropdownOpen(false);
              navigate({ to: '/admin/products/$productId', params: { productId } });
            }}
          >
            <Button variant="ghost" size="sm">
              <Pencil className="h-4 w-4" />
              <span>Assign Roles</span>
            </Button>
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </>
  );
}
