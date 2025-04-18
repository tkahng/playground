import { RouteMap } from "@/components/route-map";
import { Button } from "@/components/ui/button";
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useTabs } from "@/hooks/use-tabs";
import { getRoleWithPermission, updateRole } from "@/lib/queries";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { ChevronLeft } from "lucide-react";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { Link, useNavigate, useParams } from "react-router";
import { toast } from "sonner";
import { z } from "zod";
import { CreateRolePermissionDialog } from "./create-role-permission-dialog";

const formSchema = z.object({
  name: z.string().min(2, {
    message: "name must be at least 2 characters.",
  }),
  description: z
    .string()
    .min(2, { message: "description must be at least 2 characters." })
    .optional(),
});
export default function RoleEdit() {
  const navigate = useNavigate();
  const { tab, onClick } = useTabs("general");
  const queryClient = useQueryClient();
  const { user } = useAuthProvider();
  const { roleId } = useParams<{ roleId: string }>();
  const {
    data: role,
    isLoading: loading,
    error,
  } = useQuery({
    queryKey: ["role-with-permission", roleId],
    queryFn: async () => {
      if (!user?.tokens.access_token || !roleId) {
        throw new Error("Missing access token or role ID");
      }
      return getRoleWithPermission(user.tokens.access_token, roleId);
    },
  });
  const mutation = useMutation({
    mutationFn: (values: z.infer<typeof formSchema>) =>
      updateRole(user!.tokens.access_token, roleId!, values),
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: ["role-with-permission", roleId],
      });
      const updatedRole = await queryClient.fetchQuery({
        queryKey: ["role-with-permission", roleId],
        queryFn: () =>
          getRoleWithPermission(user!.tokens.access_token, roleId!),
      });
      form.reset(updatedRole);
      toast.success("Role updated!");
    },
    onError: (err: any) => {
      toast.error(`Failed to update role: ${err.message}`);
    },
  });
  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      name: role?.name || "",
      description: role?.description || "",
    },
  });
  function onSubmit(values: z.infer<typeof formSchema>) {
    mutation.mutate(values);
  }
  useEffect(() => {
    if (role) {
      form.reset(role);
    }
  }, [role, form.reset]);
  if (!user) {
    navigate(RouteMap.SIGNIN);
  }
  if (loading) return <p>Loading...</p>;
  if (error) return <p>Error: {error.message}</p>;
  if (!role) return <p>Role not found</p>;

  return (
    // <div className="h-full px-4 py-6 lg:px-8 space-y-6">
    <div className="mx-auto w-full max-w-[1200px] py-12 px-4 @lg:px-6 @xl:px-12 @2xl:px-20 @3xl:px-24">
      <Link
        to={RouteMap.ADMIN_DASHBOARD_ROLES}
        className="flex items-center gap-2 text-sm text-muted-foreground"
      >
        <ChevronLeft className="h-4 w-4" />
        Back to roles
      </Link>
      <h1 className="text-2xl font-bold">{role.name}</h1>
      <Tabs value={tab} onValueChange={onClick} className="h-full space-y-6">
        <TabsList>
          <TabsTrigger value="general">General</TabsTrigger>
          <TabsTrigger value="permissions">Permissions</TabsTrigger>
        </TabsList>
        <TabsContent value="general">
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
              <FormField
                control={form.control}
                name="name"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Name</FormLabel>
                    <FormControl>
                      <Input {...field} />
                    </FormControl>
                    <FormDescription>
                      This is your public display name.
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="description"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Description</FormLabel>
                    <FormControl>
                      <Input placeholder="shadcn" {...field} />
                    </FormControl>
                    <FormDescription>
                      This is your public display name.
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <Button type="submit" disabled={!form.formState.isDirty}>
                Submit
              </Button>
            </form>
          </Form>
        </TabsContent>
        <TabsContent value="permissions">
          <div className="space-y-4 flex flex-row space-x-16">
            <p className="flex-1">
              Add Permissions to this Role. Users who have this Role will
              receive all Permissions below that match the API of their login
              request.
            </p>
            <CreateRolePermissionDialog roleId={roleId!} />
          </div>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Permission</TableHead>
                <TableHead>Description</TableHead>
                <TableHead>Status</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {role.permissions?.map((permission) => (
                <TableRow key={permission.id}>
                  <TableCell>{permission.name}</TableCell>
                  <TableCell>{permission.description}</TableCell>
                  {/* <TableCell>{permission.status}</TableCell> */}
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TabsContent>
      </Tabs>
    </div>
  );
}
