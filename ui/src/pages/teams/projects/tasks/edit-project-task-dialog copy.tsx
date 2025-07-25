import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
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
import { useTeam } from "@/hooks/use-team";
import { updateTask } from "@/lib/api";
import { Task } from "@/schema.types";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { Link } from "react-router";
import { toast } from "sonner";
import { z } from "zod";

const formSchema = z.object({
  name: z.string().min(1),
  description: z.string().min(0).optional(),
  status: z.enum(["todo", "in_progress", "done"]),
  assignee_id: z.string().nullable(),
  end_at: z.string().nullable(),
  parent_id: z.string().nullable(),
  position: z.number().optional(),
  rank: z.number().optional(),
  reporter_id: z.string().nullable(),
  start_at: z.string().nullable(),
});

export function EditTaskDialog2({
  task,
  isDialogOpen,
  setDialogOpen,
}: {
  task: Task;
  isDialogOpen: boolean;
  setDialogOpen: React.Dispatch<React.SetStateAction<boolean>>;
}) {
  const { user } = useAuthProvider();
  const { team } = useTeam();

  const queryClient = useQueryClient();

  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      name: task.name,
      description: task.description || "",
      status: task.status,
      assignee_id: task.assignee_id,

      end_at: task.end_at,
      parent_id: task.parent_id,

      rank: task.rank,
      reporter_id: task.reporter_id,
      start_at: task.start_at,
    },
  });

  const mutation = useMutation({
    mutationFn: async (values: z.infer<typeof formSchema>) => {
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token");
      }
      await updateTask(user.tokens.access_token, task.id, {
        name: values.name,
        status: values.status,
        description: values.description || null,
        assignee_id: values.assignee_id || null,
        end_at: values.end_at || null,
        parent_id: values.parent_id || null,
        reporter_id: values.reporter_id || null,
        start_at: values.start_at || null,
      });
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: ["project-tasks", task.project_id],
      });
      toast.success("Task created successfully");
    },
    onError: (error) => {
      toast.error(`Failed to create task: ${error.message}`);
    },
  });
  const onSubmit = (values: z.infer<typeof formSchema>) => {
    mutation.mutate(values);
  };
  return (
    <Dialog open={isDialogOpen} onOpenChange={setDialogOpen}>
      {/* <DialogTrigger asChild>
        <Button variant="outline">Create Team</Button>
      </DialogTrigger> */}
      <DialogContent
        className="sm:max-w-[425px]"
        onInteractOutside={() => {
          setDialogOpen(false);
        }}
      >
        <DialogHeader>
          <DialogTitle>Edit Task Details</DialogTitle>
          <DialogDescription>{task.name}</DialogDescription>
          <Link
            to={`/teams/${team?.slug}/projects/${task.project_id}/tasks/${task.id}`}
          >
            View Task
          </Link>
        </DialogHeader>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
            <div className="grid gap-4 py-4">
              <div className="w-full px-10 space-y-4">
                <FormField
                  control={form.control}
                  name="name"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Name</FormLabel>
                      <FormControl>
                        <Input {...field} placeholder="Task Name" />
                      </FormControl>
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
                        <Input {...field} placeholder="Task Description" />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
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
                          <SelectItem value="in_progress">
                            In Progress
                          </SelectItem>
                          <SelectItem value="done">Done</SelectItem>
                        </SelectContent>
                      </Select>
                    </FormItem>
                  )}
                />
                <DialogFooter>
                  <Button type="submit">Update Task</Button>
                </DialogFooter>
              </div>
            </div>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
