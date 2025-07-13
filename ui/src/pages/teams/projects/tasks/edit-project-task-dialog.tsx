import { Button } from "@/components/ui/button";
import {
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
import { createTask } from "@/lib/queries";
import { Task } from "@/schema.types";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";

const formSchema = z.object({
  name: z.string().min(1),
  // name: string;
  description: z.string().min(0).optional(),
  // description?: string;
  status: z.enum(["todo", "in_progress", "done"]),
  // status: "todo" | "in_progress" | "done";
  assignee_id: z.string().nullable(),
  //  assignee_id: string | null;
  created_by_member_id: z.string().nullable(),
  // created_by_member_id: string | null;
  end_at: z.string().nullable(),
  // end_at: string | null;
  parent_id: z.string().nullable(),
  // parent_id: string | null;
  position: z.number().optional(),
  // position?: number;
  project_id: z.string(),
  // project_id: string;
  rank: z.number().optional(),
  // rank?: number;
  reporter_id: z.string().nullable(),
  // reporter_id: string | null;
  start_at: z.string().nullable(),
  // start_at: string | null;
  team_id: z.string(),
});

export function EditProjectTaskDialog({
  task,
  onFinish,
}: {
  task: Task;
  onFinish: () => void;
}) {
  const { user } = useAuthProvider();
  // const [isDialogOpen, setDialogOpen] = useState(false);
  const queryClient = useQueryClient();
  // const navigate = useNavigate();

  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      name: task.name,
      description: task.description || "",
      status: task.status,
      assignee_id: task.assignee_id,
      created_by_member_id: task.created_by_member_id,
      end_at: task.end_at,
      parent_id: task.parent_id,
      project_id: task.project_id,
      rank: task.rank,
      reporter_id: task.reporter_id,
      start_at: task.start_at,
      team_id: task.team_id,
    },
  });

  const mutation = useMutation({
    mutationFn: async (values: z.infer<typeof formSchema>) => {
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token");
      }
      await createTask(user.tokens.access_token, task.id, values);
    },
    onSuccess: async () => {
      onFinish();
      await queryClient.invalidateQueries({
        queryKey: ["project-with-tasks", task.project_id],
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
    <>
      <DialogHeader>
        <DialogTitle>Add Task to Project</DialogTitle>
        <DialogDescription>
          Select the Task you want to add to this project
        </DialogDescription>
      </DialogHeader>
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
          <div className="grid gap-4 py-4">
            <div className="w-full px-10">
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
                        <SelectItem value="in_progress">In Progress</SelectItem>
                        <SelectItem value="done">Done</SelectItem>
                      </SelectContent>
                    </Select>
                  </FormItem>
                )}
              />
              <DialogFooter>
                <Button type="submit">Create Project Task</Button>
              </DialogFooter>
            </div>
          </div>
        </form>
      </Form>
    </>
  );
}
