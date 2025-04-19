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
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { UserDetailContext } from "@/context/user-detail-context";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { ConfirmDialog, useDialog } from "@/hooks/use-dialog";
import { useTabs } from "@/hooks/use-tabs";
import {
  getUserInfo,
  removeUserPermission,
  removeUserRole,
} from "@/lib/queries";
import { UserPermissionDialog } from "@/pages/admin/users/user-permissions-dialog";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { ChevronLeft, Trash } from "lucide-react";
import { Link, useParams } from "react-router";
import { toast } from "sonner";
import { UserRolesDialog } from "./user-roles-dialog";

export default function UserEdit() {
  // const navigate = useNavigate();
  const { tab, onClick } = useTabs("profile");
  const queryClient = useQueryClient();
  const { user } = useAuthProvider();
  const { userId } = useParams<{ userId: string }>();
  const {
    data: userInfo,
    isLoading: loading,
    error,
  } = useQuery({
    queryKey: ["userInfo", userId],
    queryFn: async () => {
      if (!user?.tokens.access_token || !userId) {
        throw new Error("Missing access token or role ID");
      }
      return getUserInfo(user.tokens.access_token, userId);
    },
  });
  const deleteUserRoleMutation = useMutation({
    mutationFn: async (roleId: string) => {
      if (!user?.tokens.access_token || !userId) {
        throw new Error("Missing access token or role ID");
      }
      return removeUserRole(user.tokens.access_token, userId, roleId);
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: ["userInfo", userId],
      });
      toast.success("Role deleted");
    },
    onError: (error) => {
      console.error(error);
      toast.error("Failed to delete role");
    },
  });
  const deleteUserPermissionMutation = useMutation({
    mutationFn: async (permissionId: string) => {
      if (!user?.tokens.access_token || !userId) {
        throw new Error("Missing access token or role ID");
      }
      return removeUserPermission(
        user.tokens.access_token,
        userId,
        permissionId
      );
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: ["userInfo", userId],
      });
      toast.success("Permission deleted");
    },
    onError: (error) => {
      console.error(error);
      toast.error("Failed to delete permission");
    },
  });

  if (loading) {
    return <div>Loading...</div>;
  }
  if (error) {
    return <div>Error: {error.message}</div>;
  }
  if (!userInfo) {
    return <div>User not found</div>;
  }
  return (
    <UserDetailContext.Provider value={userInfo}>
      <div className="space-y-6">
        <Link
          to={RouteMap.ADMIN_DASHBOARD_USERS}
          className="flex items-center gap-2 text-sm text-muted-foreground"
        >
          <ChevronLeft className="h-4 w-4" />
          Back to Users
        </Link>
        <h1 className="text-2xl font-bold">{userInfo.email}</h1>
        <Tabs value={tab} onValueChange={onClick} className="h-full space-y-6">
          <TabsList>
            <TabsTrigger value="profile">Account</TabsTrigger>
            <TabsTrigger value="roles">roles</TabsTrigger>
            <TabsTrigger value="permissions">permissions</TabsTrigger>
          </TabsList>

          <TabsContent value="profile">
            <div>
              <h1>Account</h1>
              <h3>
                Make changes to your account here. Click save when you're done.
              </h3>
              <div className="space-y-2">
                <div className="space-y-1">
                  <Label htmlFor="name">email</Label>
                  <Input id="name" defaultValue={userInfo.email} />
                </div>
                <div className="space-y-1">
                  <Label htmlFor="username">Username</Label>
                  <Input id="username" defaultValue={userInfo.name || ""} />
                </div>
              </div>
              <div>
                <Button>Save changes</Button>
              </div>
            </div>
          </TabsContent>
          <TabsContent value="roles">
            <div className="space-y-4 flex flex-row space-x-16">
              <p className="flex-1">
                Add Permissions to this Role. Users who have this Role will
                receive all Permissions below that match the API of their login
                request.
              </p>
              <UserRolesDialog userDetail={userInfo} />
            </div>
            <DataTable
              columns={[
                {
                  header: "Role",
                  accessorKey: "name",
                },
                {
                  header: "Description",
                  accessorKey: "description",
                },
                {
                  id: "actions",
                  cell: ({ row }) => {
                    return (
                      <div className="flex flex-row gap-2 justify-end">
                        <DeleteButton
                          onDelete={() => {
                            deleteUserRoleMutation.mutate(row.original.id);
                          }}
                          disabled={false}
                        />
                      </div>
                    );
                  },
                },
              ]}
              data={userInfo.roles || []}
            />
          </TabsContent>
          <TabsContent value="permissions">
            <div className="space-y-4 flex flex-row space-x-16">
              <p className="flex-1">
                Add Permissions to this Role. Users who have this Role will
                receive all Permissions below that match the API of their login
                request.
              </p>
              <UserPermissionDialog userDetail={userInfo} />
            </div>
            <DataTable
              columns={[
                {
                  header: "Permission",
                  accessorKey: "name",
                },
                {
                  header: "Description",
                  accessorKey: "description",
                },
                {
                  header: "Assignment",
                  cell: ({ row }) => {
                    return (
                      <p>
                        {row.original.is_directly_assigned && "DIRECT"},{" "}
                        {row.original.roles.length &&
                          row.original.roles
                            .map((role) => role.name)
                            .join(", ")}
                      </p>
                    );
                  },
                },
                {
                  id: "actions",
                  cell: ({ row }) => {
                    return (
                      <div className="flex flex-row gap-2 justify-end">
                        <DeleteButton
                          onDelete={() => {
                            deleteUserPermissionMutation.mutate(
                              row.original.id
                            );
                          }}
                          disabled={!row.original.is_directly_assigned}
                        />
                      </div>
                    );
                  },
                },
              ]}
              data={userInfo.permissions || []}
            />
          </TabsContent>
        </Tabs>
      </div>
    </UserDetailContext.Provider>
  );
}

function DeleteButton({
  onDelete,
  disabled,
}: {
  onDelete: () => void;
  disabled: boolean;
}) {
  const editDialog = useDialog();
  return (
    <>
      <Button
        variant="outline"
        size="icon"
        onClick={editDialog.trigger}
        disabled={disabled}
      >
        <Trash className="h-4 w-4" />
      </Button>
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
                  onDelete();
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
