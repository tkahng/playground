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
import { createUserPermissions, getUserPermissions2 } from "@/lib/queries";
import { UserDetailWithRoles } from "@/schema.types";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";

const formSchema = z.object({
  // userId: z.string().uuid(),
  // roleIds: z.string().uuid().array().min(1),
  permissions: z
    .object({
      value: z.string().uuid(),
      label: z.string(),
    })
    .array()
    .min(1),
});

export function UserPermissionDialog({
  userDetail,
}: {
  userDetail: UserDetailWithRoles;
}) {
  const { user } = useAuthProvider();
  const [isDialogOpen, setDialogOpen] = useState(false);
  const queryClient = useQueryClient();
  const userId = userDetail?.id;
  const { data, isLoading, error } = useQuery({
    queryKey: ["user-permissions-reverse", userId],
    queryFn: async () => {
      if (!user?.tokens.access_token || !userId) {
        throw new Error("Missing access token or role ID");
      }
      const { data } = await getUserPermissions2(
        user.tokens.access_token,
        userId
      );
      return data;
    },
  });
  const mutation = useMutation({
    mutationFn: async (values: z.infer<typeof formSchema>) => {
      if (!user?.tokens.access_token || !userId) {
        throw new Error("Missing access token or role ID");
      }
      await createUserPermissions(user.tokens.access_token, userId, {
        permission_ids: values.permissions.map((perms) => perms.value),
      });
      setDialogOpen(false);
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: ["userInfo", userId],
      });
      await queryClient.invalidateQueries({
        queryKey: ["user-permissions-reverse", userId],
      });
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
  }, [data, form]);

  if (isLoading) {
    return <div>Loading...</div>;
  }

  if (error) {
    return <div>Error: {error.message}</div>;
  }
  return (
    <Dialog open={isDialogOpen} onOpenChange={setDialogOpen}>
      <DialogTrigger asChild>
        <Button variant="outline">Assign Permissions</Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Assign Permissions</DialogTitle>
          <DialogDescription>
            Select the Permissions you want to assign to this user
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
                          defaultOptions={data?.map((role) => ({
                            label: role.name,
                            value: role.id,
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
                  <Button type="submit">Assign roles</Button>
                </DialogFooter>
              </div>
            </div>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
