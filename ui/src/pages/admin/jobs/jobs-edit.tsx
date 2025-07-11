import { RouteMap } from "@/components/route-map";
import { Button } from "@/components/ui/button";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { adminJobQueries } from "@/lib/queries";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { ChevronLeft } from "lucide-react";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { Link, useNavigate, useParams } from "react-router";
import { toast } from "sonner";
import { z } from "zod";

const formSchema = z.object({
  status: z.enum(["pending", "processing", "done", "failed"]),
  //   attempts: number;
  // kind: string;
  kind: z.string(),
  // last_error: string | null;
  last_error: z.string().nullable(),
  // max_attempts: number;
  max_attempts: z.number(),
  payload: z.string(),
  // run_after: string;
  run_after: z.string(),
  // unique_key: string | null;
  unique_key: z.string().nullable(),
  // created_at: string;
  created_at: z.string(),
  // updated_at: string;
  updated_at: z.string(),
  attempts: z.number(),
});
export default function JobsEdit() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { user } = useAuthProvider();
  const { jobId } = useParams<{ jobId: string }>();
  const {
    data: job,
    isLoading: loading,
    error,
  } = useQuery({
    queryKey: ["get-job", jobId],
    queryFn: async () => {
      if (!user?.tokens.access_token || !jobId) {
        throw new Error("Missing access token or Permission ID");
      }

      return adminJobQueries.getJob(user.tokens.access_token, jobId);
    },
  });
  const mutation = useMutation({
    mutationFn: (values: z.infer<typeof formSchema>) => {
      if (!user?.tokens.access_token || !jobId) {
        throw new Error("Missing access token or Permission ID");
      }
      if (!job) {
        throw new Error("Permission not found");
      }
      return adminJobQueries.updateJob(user.tokens.access_token, jobId, {
        // status: values.status,
        // attempts: permission.attempts,
        // kind: permission.kind,
        // last_error: permission.last_error,
        // max_attempts: permission.max_attempts,
        // payload: permission.payload,
        // run_after: permission.run_after,
        // unique_key: permission.unique_key,
        ...values,
      });
    },
    onSuccess: async () => {
      const updatedRole = await queryClient.fetchQuery({
        queryKey: ["get-job", jobId],
        queryFn: () =>
          adminJobQueries.getJob(user!.tokens.access_token, jobId!),
      });
      form.reset({ status: updatedRole.status });
      toast.success("Permission updated!");
    },
    onError: (err) => {
      toast.error(`Failed to update Permission: ${err.message}`);
    },
  });
  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      status: job?.status || "pending",
    },
  });
  function onSubmit(values: z.infer<typeof formSchema>) {
    mutation.mutate(values);
  }
  useEffect(() => {
    if (job) {
      form.reset(job);
    }
  }, [job, form.reset, form]);
  if (!user) {
    navigate(RouteMap.SIGNIN);
  }
  if (loading) return <p>Loading...</p>;
  if (error) return <p>Error: {error.message}</p>;
  if (!job) return <p>Role not found</p>;

  return (
    <div className="space-y-6">
      <Link
        to={RouteMap.ADMIN_JOBS}
        className="flex items-center gap-2 text-sm text-muted-foreground"
      >
        <ChevronLeft className="h-4 w-4" />
        Back to jobs
      </Link>
      <h1 className="text-2xl font-bold">{job.id}</h1>
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
          <FormField
            control={form.control}
            name="status"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Status</FormLabel>
                <Select
                  onValueChange={field.onChange}
                  defaultValue={field.value}
                >
                  <FormControl {...field}>
                    <SelectTrigger>
                      <SelectValue placeholder="Select Task Status" />
                    </SelectTrigger>
                  </FormControl>
                  <SelectContent>
                    <SelectItem value="todo">Todo</SelectItem>
                    <SelectItem value="in_progress">In Progress</SelectItem>
                    <SelectItem value="done">Done</SelectItem>
                  </SelectContent>
                </Select>
              </FormItem>
            )}
          />
          <FormField
            control={form.control}
            name="kind"
            disabled
            render={({ field }) => (
              <FormItem>
                <FormLabel>Kind</FormLabel>
                <FormControl>
                  <Input {...field} />
                </FormControl>
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
