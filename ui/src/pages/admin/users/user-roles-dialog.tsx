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
import { createUserRoles, rolesPaginate } from "@/lib/api";
import { UserDetailWithRoles } from "@/schema.types";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";

const formSchema = z.object({
  // userId: z.string().uuid(),
  // roleIds: z.string().uuid().array().min(1),
  roles: z
    .object({
      value: z.string().uuid(),
      label: z.string(),
    })
    .array()
    .min(1),
});

export function UserRolesDialog({
  userDetail,
}: {
  userDetail: UserDetailWithRoles;
}) {
  const { user } = useAuthProvider();
  const [isDialogOpen, setDialogOpen] = useState(false);
  const queryClient = useQueryClient();
  // const [value, setValue] = useState<Option[]>([]);
  const userId = userDetail?.id;
  const { data, isLoading, error } = useQuery({
    queryKey: ["user-roles-reverse", userId],
    queryFn: async () => {
      if (!user?.tokens.access_token || !userId) {
        throw new Error("Missing access token or role ID");
      }
      const { data } = await rolesPaginate(user.tokens.access_token, {
        user_id: userId,
        reverse: "user",
        page: 0,
        per_page: 50,
      });
      return data;
    },
  });
  const mutation = useMutation({
    mutationFn: async (values: z.infer<typeof formSchema>) => {
      if (!user?.tokens.access_token || !userId) {
        throw new Error("Missing access token or role ID");
      }
      await createUserRoles(user.tokens.access_token, userId, {
        role_ids: values.roles.map((role) => role.value),
      });
      setDialogOpen(false);
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: ["userInfo", userId],
      });
      await queryClient.invalidateQueries({
        queryKey: ["user-roles-reverse", userId],
      });
    },
  });
  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      roles: [],
    },
  });

  const onSubmit = (values: z.infer<typeof formSchema>) => {
    mutation.mutate(values);
  };
  useEffect(() => {
    if (data) {
      form.reset({ roles: [] });
    }
  }, [data, form]);

  if (isLoading) {
    return <div>Loading...</div>;
  }

  if (error) {
    return <div>Error: {error.message}</div>;
  }

  if (!data?.length) {
    return <div>User not found</div>;
  }
  return (
    <Dialog open={isDialogOpen} onOpenChange={setDialogOpen}>
      <DialogTrigger asChild>
        <Button variant="outline">Assign Roles</Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Assign Roles</DialogTitle>
          <DialogDescription>
            Select the roles you want to assign to this user
          </DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)}>
            <div className="grid">
              <div className="space-y-4">
                <FormField
                  control={form.control}
                  name="roles"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Roles</FormLabel>
                      <FormControl>
                        <MultipleSelector
                          {...field}
                          defaultOptions={data.map((role) => ({
                            label: role.name,
                            value: role.id,
                          }))}
                          placeholder="Select roles you like..."
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
