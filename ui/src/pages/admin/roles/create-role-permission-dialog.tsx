import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import MultipleSelector from "@/components/ui/multiple-selector";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { createRolePermission, permissionsPaginate } from "@/lib/queries";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";

const formSchema = z.object({
  permissions: z
    .object({
      value: z.string().uuid(),
      label: z.string(),
    })
    .array()
    .min(1),
});

export function CreateRolePermissionDialog({ roleId }: { roleId: string }) {
  const { user, checkAuth } = useAuthProvider();
  const [isDialogOpen, setDialogOpen] = useState(false);
  const queryClient = useQueryClient();
  const { data, isLoading, error } = useQuery({
    queryKey: ["role-permissions-reverse", roleId],
    queryFn: async () => {
      await checkAuth(); // Ensure user is authenticated
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token or role ID");
      }
      const data = await permissionsPaginate(user.tokens.access_token, {
        page: 0,
        per_page: 50,
        role_id: roleId,
        role_reverse: true,
      });
      if (!data.data) {
        throw new Error("No data available");
      }
      return data.data;
    },
  });
  const mutation = useMutation({
    mutationFn: async (values: z.infer<typeof formSchema>) => {
      if (!user?.tokens.access_token || !roleId) {
        throw new Error("Missing access token or role ID");
      }
      await createRolePermission(user.tokens.access_token, roleId, {
        permission_ids: values.permissions.map((perms) => perms.value),
      });
      setDialogOpen(false);
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: ["role-with-permission", roleId],
      });
      toast.success("Permissions assigned successfully");
      setDialogOpen(false);
    },
  });
  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      permissions: [],
    },
  });

  const onSubmit = (values: z.infer<typeof formSchema>) => {
    mutation.mutate(values);
  };
  useEffect(() => {
    if (data) {
      form.reset({ permissions: [] });
    }
  }, [data, form.reset]);

  if (isLoading) {
    return <div>Loading...</div>;
  }

  if (error) {
    return <div>Error: {error.message}</div>;
  }
  if (!data) {
    return <div>No data available</div>;
  }
  return (
    <Dialog open={isDialogOpen} onOpenChange={setDialogOpen}>
      <DialogTrigger asChild>
        <Button variant="outline">Add Permissions to Role</Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Assign Permissions</DialogTitle>
          <DialogDescription>
            Select the Permissions you want to assign to this role
          </DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)}>
            <div className="grid">
              <div className="space-y-4">
                <FormField
                  control={form.control}
                  name="permissions"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Permissions</FormLabel>
                      <FormControl>
                        <MultipleSelector
                          {...field}
                          defaultOptions={data.map((permission) => ({
                            label: permission.name,
                            value: permission.id,
                          }))}
                          placeholder="Select permissions you like..."
                          emptyIndicator={
                            <p className="text-center text-lg leading-10 text-gray-600 dark:text-gray-400">
                              no results found.
                            </p>
                          }
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <DialogFooter>
                  <Button type="submit">Assign Permissions</Button>
                </DialogFooter>
              </div>
            </div>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
