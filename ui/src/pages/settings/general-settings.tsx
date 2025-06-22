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
import { GetError } from "@/lib/get-error";
import {
  deleteUser,
  getMe,
  requestVerification,
  resetPassword,
  updateMe,
} from "@/lib/queries";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { CheckCircle } from "lucide-react";
import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
// import { useNavigate } from "react-router-dom";

const formSchema = z.object({
  name: z.string().min(1).optional(),
  image: z.string().url().optional(),
});

const resetPasswordSchema = z.object({
  currentPassword: z.string().min(1, "Current password is required"),
  newPassword: z.string().min(1, "New password is required"),
});

export default function AccountSettingsPage() {
  const [, setIsPending] = useState(false);
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
    onError: (error) => {
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
      setIsPending(true);
      await queryClient.invalidateQueries({
        queryKey: ["auth/me"],
      });
      resetPasswordForm.reset();
    },
  });
  const requestVerificationEmailMutation = useMutation({
    mutationFn: async () => {
      if (!user) {
        throw new Error("User not found");
      }
      await requestVerification(user.tokens.access_token);
    },
    onSuccess: async () => {
      setIsPending(false);
      await queryClient.invalidateQueries({
        queryKey: ["auth/me"],
      });
      toast.success("Verification email sent successfully");
    },
    onError: (error) => {
      setIsPending(false);
      const err = GetError(error);
      if (err) {
        if (err.errors?.length) {
          toast.error(`${err.errors[0].message || err.errors[0].value}`);
        } else if (err.title) toast.error(`${err.detail || err.title}`);
      } else {
        toast.error(`Failed to send verification email: ${error.message}`);
      }
    },
  });

  const requestVerificationEmail = () => {
    setIsPending(true);
    requestVerificationEmailMutation.mutate();
  };

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
      name: data?.name ?? undefined,
      image: data?.image ?? undefined,
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
  useEffect(() => {
    if (data) {
      form.reset({
        name: data.name || "",
        image: data.image || "",
      });
    }
  }, [data, form]);
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
        <div className="flex flex-row w-full justify-between">
          <div className="flex items-center space-x-4">
            <h1>Email </h1>
            <p className="text-sm text-muted-foreground">{data.email}</p>
          </div>

          {data.email_verified_at ? (
            <div>
              <CheckCircle className="h-8 w-8 text-green-600 dark:text-green-300" />
            </div>
          ) : (
            <Button onClick={requestVerificationEmail}>
              Send Verification Email
            </Button>
          )}
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
            <FormField
              control={form.control}
              name="image"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Image</FormLabel>
                  <FormControl>
                    <Input {...field} placeholder="Image" type="url" />
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
                        placeholder="Current Password"
                        type="password"
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
                        placeholder="New Password"
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
