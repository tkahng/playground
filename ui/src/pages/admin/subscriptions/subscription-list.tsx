import { useSearchParams } from "@/hooks/use-search-params";
import { CenteredSpinner } from "@/components/centered-spinner";
import { DataTable } from "@/components/data-table";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { adminStripeSubscriptions } from "@/lib/api";
import { useQuery } from "@tanstack/react-query";
import { PaginationState, Updater } from "@tanstack/react-table";

export default function SubscriptionsListPage() {
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
    queryKey: ["subscription-list", pageIndex, pageSize, sortBy, sortOrder],
    queryFn: async () => {
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token");
      }
      const data = await adminStripeSubscriptions(user.tokens.access_token, {
        page: pageIndex,
        per_page: pageSize,
        sort_by: sortBy,
        sort_order: sortOrder,
        expand: ["price", "product", "user"],
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
        This is a list of subscriptions. For more details, visit the stripe
        dashboard.
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
            <SelectItem value="status">Status</SelectItem>
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
          },
          // {
          //   id: "user",
          //   header: "User",
          //   cell: ({ row }) => {
          //     return row.original.user?.email;
          //   },
          // },
          {
            id: "product",
            header: "Product",
            cell: ({ row }) => {
              return row.original.price?.product?.name;
            },
          },
          {
            id: "price",
            header: "Price",
            cell: ({ row }) => {
              return row.original.price?.unit_amount
                ? `$${(row.original.price.unit_amount / 100).toFixed(2)}`
                : "Free";
            },
          },
          {
            accessorKey: "status",
            header: "Status",
            cell: ({ row }) => {
              return (
                row.original.status.charAt(0).toUpperCase() +
                row.original.status.slice(1)
              );
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
