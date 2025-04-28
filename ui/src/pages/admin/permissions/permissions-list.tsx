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
import { deletePermission, permissionsPaginate } from "@/lib/queries";
import {
  keepPreviousData,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { PaginationState, Updater } from "@tanstack/react-table";
import { Ellipsis, Pencil, Trash } from "lucide-react";
import { useState } from "react";
import { NavLink, useNavigate, useSearchParams } from "react-router";
import { CreatePermissionDialog } from "./create-permission-dialog";

export default function PermissionListPage() {
  const { user, checkAuth } = useAuthProvider();
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
    queryKey: ["permissions-list", pageIndex, pageSize],
    queryFn: async () => {
      await checkAuth(); // Ensure user is authenticated
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token or role ID");
      }
      const data = await permissionsPaginate(user.tokens.access_token, {
        page: pageIndex,
        per_page: pageSize,
      });
      return data;
    },
    placeholderData: keepPreviousData,
  });
  const deletePermissionMutation = useMutation({
    mutationFn: async (permissionId: string) => {
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token or role ID");
      }
      await deletePermission(user.tokens.access_token, permissionId);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["permissions-list"] });
    },
    onError: (error) => {
      console.error(error);
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
        <h1 className="text-2xl font-bold">Permissions</h1>
        <CreatePermissionDialog />
      </div>
      <DataTable
        columns={[
          {
            accessorKey: "name",
            header: "Name",
            cell: ({ row }) => {
              return (
                <NavLink
                  to={`${RouteMap.ADMIN_PERMISSIONS}/${row.original.id}`}
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
                <div className="flex items-center gap-2 justify-end">
                  <PermissionEllipsisDropdown
                    permissionId={row.original.id}
                    onDelete={deletePermissionMutation.mutate}
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

function PermissionEllipsisDropdown({
  permissionId,
  onDelete,
}: {
  permissionId: string;
  onDelete: (permissionId: string) => void;
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
              navigate(`${RouteMap.ADMIN_PERMISSIONS}/${permissionId}`);
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
                  onDelete(permissionId);
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
