import { DataTable } from "@/components/data-table";
import { RouteMap } from "@/components/route-map";
import { Button } from "@/components/ui/button";
import {
  DialogClose,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { ConfirmDialog, useDialog } from "@/hooks/use-dialog";
import { deleteRole, rolesPaginate } from "@/lib/queries";
import { CreateRoleDialog } from "@/pages/admin/roles/create-role-dialog";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { PaginationState, Updater } from "@tanstack/react-table";
import { Ellipsis, Pencil, Trash } from "lucide-react";
import { useState } from "react";
import { NavLink, useNavigate, useSearchParams } from "react-router";
import { toast } from "sonner";

export default function RolesListPage() {
  const { user } = useAuthProvider();
  const [searchParams, setSearchParams] = useSearchParams();
  const pageIndex = parseInt(searchParams.get("page") || "0", 10);
  const pageSize = parseInt(searchParams.get("per_page") || "10", 10);
  const queryClient = useQueryClient();
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

  const { data, isLoading, isError, error } = useQuery({
    queryKey: ["roles-list"],
    queryFn: async () => {
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token or role ID");
      }
      const data = await rolesPaginate(user.tokens.access_token, {
        page: pageIndex,
        per_page: pageSize,
      });
      if (!data.data) {
        throw new Error("No data returned from rolesPaginate");
      }
      return data;
    },
  });
  const mutation = useMutation({
    mutationFn: async (roleId: string) => {
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token or role ID");
      }
      await deleteRole(user.tokens.access_token, roleId);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["roles-list"] });
      toast.success("Role deleted successfully");
    },
    onError: (error) => {
      console.error(error);
      toast.error("Failed to delete role");
    },
  });
  if (isLoading) {
    return <div>Loading...</div>;
  }
  if (isError) {
    return <div>Error: {error?.message}</div>;
  }
  const roles = data?.data || [];
  const rowCount = data?.meta.total || 0;

  return (
    // <div className="flex w-full flex-col items-center justify-center">
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Roles</h1>
        <CreateRoleDialog />
      </div>
      <p>
        Create and manage Roles for your applications. Roles contain collections
        of Permissions and can be assigned to Users.
      </p>
      <DataTable
        columns={[
          {
            accessorKey: "name",
            header: "Name",
            cell: ({ row }) => {
              return (
                <NavLink
                  to={`${RouteMap.ADMIN_DASHBOARD_ROLES}/${row.original.id}`}
                  className="hover:underline text-blue-500"
                >
                  {row.original.name}
                </NavLink>
              );
            },
          },
          {
            accessorKey: "description",
            header: "Description",
          },
          {
            id: "actions",
            cell: ({ row }) => {
              return (
                <div className="flex flex-row gap-2 justify-end">
                  <RoleEllipsisDropdown
                    roleId={row.original.id}
                    onDelete={(roleId) => {
                      mutation.mutate(roleId);
                    }}
                  />
                </div>
              );
            },
          },
        ]}
        data={roles}
        rowCount={rowCount}
        paginationState={{ pageIndex, pageSize }}
        onPaginationChange={onPaginationChange}
        paginationEnabled
      />
    </div>
  );
}

function RoleEllipsisDropdown({
  roleId,
  onDelete,
}: {
  roleId: string;
  onDelete: (roleId: string) => void;
}) {
  const editDialog = useDialog();
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
              navigate(`${RouteMap.ADMIN_DASHBOARD_ROLES}/${roleId}`);
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
              editDialog.trigger();
            }}
          >
            <Button variant="ghost" size="sm">
              <Trash className="h-4 w-4" />
              <span>Remove</span>
            </Button>
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
      <ConfirmDialog dialogProps={editDialog.props}>
        <>
          <DialogHeader>
            <DialogTitle>Are you absolutely sure?</DialogTitle>
          </DialogHeader>
          {/* Dialog Content */}
          <DialogDescription>This action cannot be undone.</DialogDescription>
          <DialogFooter>
            <DialogClose asChild>
              <Button
                variant="outline"
                onClick={() => {
                  console.log("cancel");
                  // editDialog.props.onOpenChange(false);
                }}
              >
                Cancel
              </Button>
            </DialogClose>
            <DialogClose asChild>
              <Button
                variant="destructive"
                onClick={() => {
                  console.log("delete");
                  // editDialog.props.onOpenChange(false);
                  onDelete(roleId);
                }}
              >
                Delete
              </Button>
            </DialogClose>
          </DialogFooter>
        </>
      </ConfirmDialog>
    </>
  );
}
