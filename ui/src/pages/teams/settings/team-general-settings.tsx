import { DashboardSidebar } from "@/components/dashboard-sidebar";
import { teamSettingLinks } from "@/components/links";
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
import { useTeamContext } from "@/hooks/use-team-context";
import { GetError } from "@/lib/get-error";
import { deleteUser, updateTeam } from "@/lib/queries";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";

const formSchema = z.object({
  name: z.string().min(1).optional(),
});

export default function TeamSettingsPage() {
  const { user } = useAuthProvider();
  const { team: data } = useTeamContext();

  const queryClient = useQueryClient();
  const mutation = useMutation({
    mutationFn: async (formData: z.infer<typeof formSchema>) => {
      if (!user) {
        throw new Error("User not found");
      }
      if (!data) {
        throw new Error("Team not found");
      }
      if (!formData.name) {
        throw new Error("Name is required");
      }
      await updateTeam(user.tokens.access_token, data.id, {
        name: formData.name,
        slug: data.slug,
      });
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: ["auth/me"],
      });
      toast.success("Profile updated successfully");
    },
    onError: (error) => {
      toast.error(`Failed to update Profile: ${error.message}`);
    },
  });
  const deleteUserMutation = useMutation({
    mutationFn: async () => {
      if (!user) {
        throw new Error("User not found");
      }
      await deleteUser(user.tokens.access_token);
    },
    onError: (error) => {
      const err = GetError(error);
      if (err) {
        if (err.errors?.length) {
          toast.error(`${err.errors[0].message || err.errors[0].value}`);
        } else if (err.title) toast.error(`${err.detail || err.title}`);
      } else {
        toast.error(`Failed to reset password: ${error.message}`);
      }
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: ["auth/me"],
      });
      toast.success("Account deleted successfully");
    },
  });

  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      name: data?.name ?? undefined,
    },
  });

  const onSubmit = (values: z.infer<typeof formSchema>) => {
    mutation.mutate(values);
  };
  useEffect(() => {
    if (data) {
      form.reset({
        name: data.name || "",
      });
    }
  }, [data, form]);
  // if (isLoading) return <p>Loading...</p>;
  // if (error) return <p>Error: {error.message}</p>;
  if (!data) return <p>User not found</p>;
  return (
    <div className="flex">
      <DashboardSidebar links={teamSettingLinks(data.slug)} />
      <div className="flex-1 space-y-6 p-12 w-full">
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
                    <Input {...field} placeholder="Name" />
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

        <Separator />
        <div className="space-y-2">
          <h3 className="text-lg font-medium">Danger Zone</h3>
          <p className="text-sm text-destructive">
            This section is for actions that cannot be undone.
          </p>
          <Button
            variant="destructive"
            onClick={() => {
              deleteUserMutation.mutate();
            }}
          >
            Delete Account
          </Button>
        </div>
      </div>
    </div>
  );
}
