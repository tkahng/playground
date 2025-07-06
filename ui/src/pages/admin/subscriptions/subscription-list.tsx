import { DataTable } from "@/components/data-table";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { adminStripeSubscriptions } from "@/lib/queries";
import { useQuery } from "@tanstack/react-query";
import { PaginationState, Updater } from "@tanstack/react-table";
import { useSearchParams } from "react-router";
export default function SubscriptionsListPage() {
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
  const { data, isLoading, isError, error } = useQuery({
    queryKey: ["subscription-list", pageIndex, pageSize],
    queryFn: async () => {
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token");
      }
      const data = await adminStripeSubscriptions(user.tokens.access_token, {
        page: pageIndex,
        per_page: pageSize,
        sort_by: "updated_at",
        sort_order: "desc",
        expand: ["price", "product", "user"],
      });
      return data;
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
      <p>
        This is a list of subscriptions. For more details, visit the stripe
        dashboard.
      </p>

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
