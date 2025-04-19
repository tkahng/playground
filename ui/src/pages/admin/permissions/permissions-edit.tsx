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
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { getPermission, updatePermission } from "@/lib/queries";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { ChevronLeft } from "lucide-react";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { Link, useNavigate, useParams } from "react-router";
import { toast } from "sonner";
import { z } from "zod";

const formSchema = z.object({
  name: z.string().min(2, {
    message: "name must be at least 2 characters.",
  }),
  description: z
    .string()
    .min(2, { message: "description must be at least 2 characters." })
    .optional(),
});
export default function PermissionEdit() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { user } = useAuthProvider();
  const { permissionId } = useParams<{ permissionId: string }>();
  const {
    data: permission,
    isLoading: loading,
    error,
  } = useQuery({
    queryKey: ["permission", permissionId],
    queryFn: async () => {
      if (!user?.tokens.access_token || !permissionId) {
        throw new Error("Missing access token or Permission ID");
      }
      return getPermission(user.tokens.access_token, permissionId);
    },
  });
  const mutation = useMutation({
    mutationFn: (values: z.infer<typeof formSchema>) =>
      updatePermission(user!.tokens.access_token, permissionId!, values),
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: ["permission", permissionId],
      });
      const updatedRole = await queryClient.fetchQuery({
        queryKey: ["permission", permissionId],
        queryFn: () => getPermission(user!.tokens.access_token, permissionId!),
      });
      form.reset(updatedRole);
      toast.success("Permission updated!");
    },
    onError: (err: any) => {
      toast.error(`Failed to update Permission: ${err.message}`);
    },
  });
  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      name: permission?.name || "",
      description: permission?.description || "",
    },
  });
  function onSubmit(values: z.infer<typeof formSchema>) {
    mutation.mutate(values);
  }
  useEffect(() => {
    if (permission) {
      form.reset(permission);
    }
  }, [permission, form.reset]);
  if (!user) {
    navigate(RouteMap.SIGNIN);
  }
  if (loading) return <p>Loading...</p>;
  if (error) return <p>Error: {error.message}</p>;
  if (!permission) return <p>Role not found</p>;

  return (
    <div className="space-y-6">
      <Link
        to={RouteMap.ADMIN_DASHBOARD_PERMISSIONS}
        className="flex items-center gap-2 text-sm text-muted-foreground"
      >
        <ChevronLeft className="h-4 w-4" />
        Back to permissions
      </Link>
      <h1 className="text-2xl font-bold">{permission.name}</h1>
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
    </div>
  );
}
