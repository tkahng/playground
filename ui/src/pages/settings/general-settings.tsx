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
import { GetError } from "@/lib/get-erro";
import { getMe, resetPassword, updateMe } from "@/lib/queries";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";

const formSchema = z.object({
  name: z.string().min(1).nullable().optional(),
  image: z.string().url().nullable().optional(),
});

const resetPasswordSchema = z.object({
  currentPassword: z.string().min(1, "Current password is required"),
  newPassword: z.string().min(1, "New password is required"),
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
  const credentialsAccount = data?.accounts?.find(
    (account) => account.provider === "credentials"
  );
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
      toast.error(`Failed to update Profile: ${error.message}`);
    },
  });
  const resetPasswordMutation = useMutation({
    mutationFn: async (formData: z.infer<typeof resetPasswordSchema>) => {
      if (!user) {
        throw new Error("User not found");
      }
      await resetPassword(
        user.tokens.access_token,
        formData.currentPassword,
        formData.newPassword
      );
      toast.success("Password reset successfully");
    },
    onError: (error: any) => {
      const err = GetError(error);
      if (err) {
        if (err.errors?.length) {
          toast.error(`${err.errors[0].message || err.errors[0].value}`);
        } else if (err.title) toast.error(`${err.detail || err.title}`);
      } else {
        toast.error(`Failed to reset password: ${error.message}`);
      }
      // toast.error(`Failed to reset password: ${error.message}`);
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: ["auth/me"],
      });
      resetPasswordForm.reset();
    },
  });
  const resetPasswordForm = useForm<z.infer<typeof resetPasswordSchema>>({
    resolver: zodResolver(resetPasswordSchema),
    defaultValues: {
      currentPassword: undefined,
      newPassword: undefined,
    },
  });

  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      name: data?.name,
      image: data?.image,
    },
  });
  const onResetPasswordSubmut = (
    values: z.infer<typeof resetPasswordSchema>
  ) => {
    resetPasswordMutation.mutate(values);
  };
  const onSubmit = (values: z.infer<typeof formSchema>) => {
    mutation.mutate(values);
  };
  if (isLoading) return <p>Loading...</p>;
  if (isError) return <p>Error: {error.message}</p>;
  if (!data) return <p>User not found</p>;
  return (
    <div className="flex">
      <DashboardSidebar links={settingsSidebarLinks} />
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
                    <Input
                      {...field}
                      value={field.value ?? undefined}
                      placeholder="Task Name"
                      defaultValue={data.name || ""} // Ensure default value is set
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
                      defaultValue={data.image || ""} // Ensure default value is set
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
        <Separator />
        {credentialsAccount && (
          <Form {...resetPasswordForm}>
            <h1>Reset Password</h1>
            <form
              onSubmit={resetPasswordForm.handleSubmit(onResetPasswordSubmut)}
              className="space-y-8"
            >
              <FormField
                control={resetPasswordForm.control}
                name="currentPassword"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Current Password</FormLabel>
                    <FormControl>
                      <Input
                        {...field}
                        value={field.value ?? undefined}
                        placeholder="Current Password"
                        type="password"
                        defaultValue={undefined} // Ensure default value is set
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={resetPasswordForm.control}
                name="newPassword"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>New Password</FormLabel>
                    <FormControl>
                      <Input
                        {...field}
                        value={field.value ?? undefined}
                        placeholder="New Password"
                        type="password"
                        defaultValue={data.image || ""} // Ensure default value is set
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
    </div>
  );
}
