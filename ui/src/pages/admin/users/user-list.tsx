import { DataTable } from "@/components/data-table";
import { RouteMap } from "@/components/route-map";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { userPaginate } from "@/lib/api";
import { useQuery } from "@tanstack/react-query";
import { PaginationState, Updater } from "@tanstack/react-table";
import { NavLink, useSearchParams } from "react-router";
import { CreateUserDialog } from "./create-user-dialog";
import { UserActionDropdown } from "./user-action-dropdown";
export default function UserListPage() {
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
    queryKey: ["users-list", pageIndex, pageSize],
    queryFn: async () => {
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token");
      }
      const data = await userPaginate(user.tokens.access_token, {
        page: pageIndex,
        per_page: pageSize,
        sort_by: "updated_at",
        sort_order: "desc",
        expand: ["roles", "accounts"],
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
        <p>
          Create and manage Users for your applications. Users contain
          collections of Roles and can be assigned to Applications.
        </p>
        <CreateUserDialog />
      </div>

      <DataTable
        columns={[
          {
            accessorKey: "email",
            header: "Email",
            cell: ({ row }) => {
              return (
                <NavLink
                  to={`${RouteMap.ADMIN_USERS}/${row.original.id}`}
                  className="hover:underline text-blue-500"
                >
                  {row.original.email}
                </NavLink>
              );
            },
          },
          {
            accessorKey: "name",
            header: "Name",
          },
          {
            accessorKey: "accounts",
            header: "Accounts",
            cell: ({ row }) => {
              return (
                row.original.accounts
                  ?.map((account) => account.provider)
                  .join(", ") || "None"
              );
            },
          },
          {
            accessorKey: "roles",
            header: "Roles",
            cell: ({ row }) => {
              return (
                row.original.roles?.map((role) => role.name).join(", ") ||
                "None"
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
            accessorKey: "email_verified_at",
            header: "Email Verified At",
            cell: ({ row }) => {
              return row.original.email_verified_at
                ? new Date(row.original.email_verified_at).toLocaleDateString()
                : "Not Verified";
            },
          },
          {
            id: "actions",
            cell: ({ row }) => {
              return (
                <div className="flex flex-row gap-2 justify-end">
                  <UserActionDropdown userId={row.original.id} />
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
