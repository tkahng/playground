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
import { userPaginate } from "@/lib/queries";
import { useQuery } from "@tanstack/react-query";
import { PaginationState, Updater } from "@tanstack/react-table";
import { Ellipsis, Pencil } from "lucide-react";
import { useState } from "react";
import { NavLink, useNavigate, useSearchParams } from "react-router";
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
    queryKey: ["users-list"],
    queryFn: async () => {
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token");
      }
      const data = await userPaginate(user.tokens.access_token, {
        page: pageIndex + 1,
        per_page: pageSize,
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
      </div>
      <p>
        Create and manage Users for your applications. Users contain collections
        of Roles and can be assigned to Applications.
      </p>

      <DataTable
        columns={[
          {
            accessorKey: "email",
            header: "Email",
            cell: ({ row }) => {
              return (
                <NavLink
                  to={`${RouteMap.ADMIN_DASHBOARD_USERS}/${row.original.id}`}
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
            id: "actions",
            cell: ({ row }) => {
              return (
                <div className="flex flex-row gap-2 justify-end">
                  <UserEllipsisDropdown userId={row.original.id} />
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

function UserEllipsisDropdown({ userId }: { userId: string }) {
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
              navigate(`${RouteMap.ADMIN_DASHBOARD_USERS}/${userId}`);
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
              navigate(`${RouteMap.ADMIN_DASHBOARD_USERS}/${userId}?tab=roles`);
            }}
          >
            <Button variant="ghost" size="sm">
              <Pencil className="h-4 w-4" />
              <span>Assign Roles</span>
            </Button>
          </DropdownMenuItem>
          {/* <DropdownMenuItem
            onSelect={() => {
              setDropdownOpen(false);
              editDialog.trigger();
            }}
          >
            <Button variant="ghost" size="sm">
              <Trash className="h-4 w-4" />
              <span>Remove</span>
            </Button>
          </DropdownMenuItem> */}
        </DropdownMenuContent>
      </DropdownMenu>
    </>
  );
}
