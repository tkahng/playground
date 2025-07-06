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
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Separator } from "@/components/ui/separator";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { UserDetailContext } from "@/context/user-detail-context";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { ConfirmDialog, useDialog } from "@/hooks/use-dialog";
import { useTabs } from "@/hooks/use-tabs";
import {
  adminResetUserPassword,
  getUserInfo,
  removeUserPermission,
  removeUserRole,
  updateUser,
} from "@/lib/queries";
import { UserPermissionDialog } from "@/pages/admin/users/user-permissions-dialog";
import { UserRolesDialog } from "@/pages/admin/users/user-roles-dialog";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { CheckCircle, ChevronLeft, Trash } from "lucide-react";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { Link, useParams } from "react-router";
import { toast } from "sonner";
import { z } from "zod";

const formSchema = z.object({
  name: z.string().optional(),
  image: z.string().url().optional(),
  email: z.string().email("Invalid email address"),
  email_verified_at: z.boolean().optional(),
});

const updatePasswordSchema = z.object({
  password: z.string().min(1, "New password is required"),
});

export default function UserEdit() {
  // const navigate = useNavigate();
  const { tab, onClick } = useTabs("profile");
  const queryClient = useQueryClient();
  const { user } = useAuthProvider();
  const { userId } = useParams<{ userId: string }>();
  const {
    data,
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
  const userUpdateMutation = useMutation({
    mutationFn: async (values: z.infer<typeof formSchema>) => {
      if (!user?.tokens.access_token || !userId) {
        throw new Error("Missing access token or user ID");
      }
      return updateUser(user.tokens.access_token, userId, {
        name: values.name,
        email: values.email,
        email_verified_at: values.email_verified_at
          ? new Date().toISOString()
          : undefined,
      });
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: ["userInfo", userId],
      });
      toast.success("User updated successfully");
    },
    onError: (error) => {
      console.error(error);
      toast.error("Failed to update user");
    },
  });
  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      name: data?.name || "",
      email: data?.email || "",
      email_verified_at: !!data?.email_verified_at,
    },
  });

  function onSubmit(values: z.infer<typeof formSchema>) {
    userUpdateMutation.mutate(values);
  }

  const credentialsAccount = data?.accounts?.find(
    (account) => account.provider === "credentials"
  );

  const resetPasswordMutation = useMutation({
    mutationFn: async (values: z.infer<typeof updatePasswordSchema>) => {
      if (!credentialsAccount) {
        throw new Error("No credentials account found");
      }
      if (!user?.tokens.access_token || !userId) {
        throw new Error("Missing access token or user ID");
      }
      return adminResetUserPassword(
        user.tokens.access_token,
        userId,
        values.password
      );
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: ["userInfo", userId],
      });
      toast.success("Password updated successfully");
    },
    onError: (error) => {
      console.error(error);
      toast.error("Failed to update password");
    },
  });

  const resetPasswordForm = useForm<z.infer<typeof updatePasswordSchema>>({
    resolver: zodResolver(updatePasswordSchema),
    defaultValues: {
      password: "",
    },
  });
  function onResetPasswordSubmut(values: z.infer<typeof updatePasswordSchema>) {
    if (!credentialsAccount) {
      toast.error("No credentials account found");
      return;
    }
    resetPasswordMutation.mutate({
      password: values.password,
    });
  }
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

  useEffect(() => {
    if (data) {
      form.reset({
        name: data.name || undefined,
        email: data.email,
        email_verified_at: !!data.email_verified_at,
      });
    }
  }, [data, form]);

  if (loading) {
    return <div>Loading...</div>;
  }
  if (error) {
    return <div>Error: {error.message}</div>;
  }
  if (!data) {
    return <div>User not found</div>;
  }
  return (
    <UserDetailContext.Provider value={data}>
      <div className="space-y-6">
        <Link
          to={RouteMap.ADMIN_USERS}
          className="flex items-center gap-2 text-sm text-muted-foreground"
        >
          <ChevronLeft className="h-4 w-4" />
          Back to Users
        </Link>
        <h1 className="text-2xl font-bold">{data.email}</h1>
        <Tabs value={tab} onValueChange={onClick} className="h-full space-y-6">
          <TabsList>
            <TabsTrigger value="profile">Account</TabsTrigger>
            <TabsTrigger value="roles">roles</TabsTrigger>
            <TabsTrigger value="permissions">permissions</TabsTrigger>
          </TabsList>

          <TabsContent value="profile">
            {/* <div className="flex-1 space-y-6 p-12 w-full"> */}
            <div className="space-y-6">
              {/* <div>
                  <h3 className="text-lg font-medium">Edit User</h3>
                  <p className="text-sm text-muted-foreground">Edit users.</p>
                </div> */}
              <Separator />
              <Form {...form}>
                <form
                  onSubmit={form.handleSubmit(onSubmit)}
                  className="space-y-6"
                >
                  <div className="flex flex-row w-full justify-between">
                    <div className="flex items-center space-x-4">
                      <h1>Email </h1>
                      <p className="text-sm text-muted-foreground">
                        {data.email}
                      </p>
                    </div>

                    {data.email_verified_at ? (
                      <div>
                        <CheckCircle className="h-8 w-8 text-green-600 dark:text-green-300" />
                      </div>
                    ) : (
                      <Button>Send Verification Email</Button>
                    )}
                  </div>
                  <FormField
                    control={form.control}
                    name="email_verified_at"
                    render={({ field }) => (
                      <FormItem className="flex items-center space-x-2">
                        <FormControl>
                          <input
                            type="checkbox"
                            checked={field.value}
                            onChange={(e) => field.onChange(e.target.checked)}
                          />
                        </FormControl>
                        <FormLabel>Email Verified</FormLabel>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <FormField
                    control={form.control}
                    name="name"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Name</FormLabel>
                        <FormControl>
                          <Input {...field} placeholder="Name" />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <FormField
                    control={form.control}
                    name="image"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Image</FormLabel>
                        <FormControl>
                          <Input {...field} placeholder="Image" />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <Button type="submit" disabled={!form.formState.isDirty}>
                    Save
                  </Button>
                </form>
              </Form>
              <Separator />
              {credentialsAccount && (
                <Form {...resetPasswordForm}>
                  <h1>Reset Password</h1>
                  <form
                    onSubmit={resetPasswordForm.handleSubmit(
                      onResetPasswordSubmut
                    )}
                    className="space-y-6"
                  >
                    <FormField
                      control={resetPasswordForm.control}
                      name="password"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>New Password</FormLabel>
                          <FormControl>
                            <Input
                              {...field}
                              placeholder="Current Password"
                              type="password"
                            />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />

                    <Button
                      type="submit"
                      disabled={!resetPasswordForm.formState.isDirty}
                    >
                      Save
                    </Button>
                  </form>
                </Form>
              )}
              <Separator />
              <div className="space-y-2">
                <h3 className="text-lg font-medium">Danger Zone</h3>
                <p className="text-sm text-destructive">
                  This section is for actions that cannot be undone.
                </p>
                <Button
                  variant="destructive"
                  onClick={() => {
                    toast.error("This feature is not implemented yet.");
                  }}
                >
                  Delete Account
                </Button>
              </div>
            </div>
          </TabsContent>
          <TabsContent value="roles">
            <div className="space-y-4 flex flex-row space-x-16">
              <p className="flex-1">Add Roles to this user.</p>
              <UserRolesDialog userDetail={data} />
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
              data={data.roles || []}
            />
          </TabsContent>
          <TabsContent value="permissions">
            <div className="space-y-4 flex flex-row space-x-16">
              <p className="flex-1">Add Permissions to this user.</p>
              <UserPermissionDialog userDetail={data} />
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
              data={data.permissions || []}
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
