import { DashboardSidebar } from "@/components/dashboard-sidebar";
import { settingsSidebarLinks } from "@/components/links";
import { Button } from "@/components/ui/button";
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
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { getMe, updateMe } from "@/lib/queries";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";

const formSchema = z.object({
  name: z.string().min(1).nullable().optional(),
  image: z.string().url().nullable().optional(),
});

export default function AccountSettingsPage() {
  const { user } = useAuthProvider();
  const { data, isLoading, isError, error } = useQuery({
    queryKey: ["auth/me"],
    queryFn: async () => {
      if (!user) {
        throw new Error("User not found");
      }
      return getMe(user.tokens.access_token);
    },
  });
  const queryClient = useQueryClient();
  const mutation = useMutation({
    mutationFn: async (formData: z.infer<typeof formSchema>) => {
      if (!user) {
        throw new Error("User not found");
      }
      await updateMe(user.tokens.access_token, {
        name: formData.name ?? null,
        image: formData.image ?? null,
      });
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: ["auth/me"],
      });
      toast.success("Profile updated successfully");
    },
    onError: (error: any) => {
      toast.error(`Failed to update project: ${error.message}`);
    },
  });
  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      name: data?.name,
      image: data?.image,
    },
  });
  const onSubmit = (values: z.infer<typeof formSchema>) => {
    mutation.mutate(values);
  };
  if (isLoading) return <p>Loading...</p>;
  if (isError) return <p>Error: {error.message}</p>;
  if (!data) return <p>User not found</p>;
  return (
    <div className="flex">
      <DashboardSidebar links={settingsSidebarLinks} />
      <div className="space-y-6 p-12">
        <div>
          <h3 className="text-lg font-medium">Profile</h3>
          <p className="text-sm text-muted-foreground">
            This is how others will see you on the site.
          </p>
        </div>
        <Separator />
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Name</FormLabel>
                  <FormControl>
                    <Input
                      {...field}
                      value={field.value ?? undefined}
                      placeholder="Task Name"
                    />
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
                    <Input
                      {...field}
                      value={field.value ?? undefined}
                      placeholder="Image"
                    />
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
      </div>
    </div>
  );
}
