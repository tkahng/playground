import { DataTable } from "@/components/data-table";
import { RouteMap } from "@/components/route-map";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { adminStripeSubscriptions } from "@/lib/queries";
import { useQuery } from "@tanstack/react-query";
import { PaginationState, Updater } from "@tanstack/react-table";
import { Ellipsis, Pencil } from "lucide-react";
import { useState } from "react";
import { NavLink, useNavigate, useSearchParams } from "react-router";
// import { CreateUserDialog } from "./create-user-dialog";
export default function SubscriptionsListPage() {
  const { user, checkAuth } = useAuthProvider();

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
      await checkAuth(); // Ensure user is authenticated
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
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Users</h1>
        {/* <CreateUserDialog /> */}
      </div>
      <p>
        Create and manage Users for your applications. Users contain collections
        of Roles and can be assigned to Applications.
      </p>

      <DataTable
        columns={[
          {
            accessorKey: "id",
            header: "ID",
            cell: ({ row }) => {
              return (
                <NavLink
                  to={`${RouteMap.ADMIN_SUBSCRIPTIONS}/${row.original.id}`}
                  className="hover:underline text-blue-500"
                >
                  {row.original.id}
                </NavLink>
              );
            },
          },
          {
            id: "user",
            header: "User",
            cell: ({ row }) => {
              return row.original.user?.email;
            },
          },
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
          {
            id: "actions",
            cell: ({ row }) => {
              return (
                <div className="flex flex-row gap-2 justify-end">
                  <SubscriptionEllipsisDropdown
                    subscriptionId={row.original.id}
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

function SubscriptionEllipsisDropdown({
  subscriptionId,
}: {
  subscriptionId: string;
}) {
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
              navigate(`${RouteMap.ADMIN_SUBSCRIPTIONS}/${subscriptionId}`);
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
              navigate(
                `${RouteMap.ADMIN_SUBSCRIPTIONS}/${subscriptionId}?tab=roles`
              );
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
