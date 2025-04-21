import { KanbanBoard } from "@/components/board/KanbanBoard";
import { RouteMap } from "@/components/route-map";
import { Button } from "@/components/ui/button";
import {
  Form,
  FormControl,
  FormDescription,
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
import { taskList, taskProjectGet, taskProjectUpdate } from "@/lib/queries";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { ChevronLeft } from "lucide-react";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { Link, useParams } from "react-router";
import { toast } from "sonner";
import { z } from "zod";
import { CreateProjectTaskDialog } from "./create-project-task-dialog";

const formSchema = z.object({
  name: z.string().min(2, {
    message: "name must be at least 2 characters.",
  }),
  description: z
    .string()
    .min(2, { message: "description must be at least 2 characters." })
    .optional(),
  status: z.enum(["todo", "in_progress", "done"]),
  order: z.number(),
});

export default function ProjectEdit() {
  // const navigate = useNavigate();
  // const { tab, onClick } = useTabs("tasks");
  const queryClient = useQueryClient();
  const { user } = useAuthProvider();
  const { projectId } = useParams<{ projectId: string }>();
  const {
    data: project,
    isLoading: loading,
    error,
  } = useQuery({
    select: (data) => {
      return {
        ...data,
        tasks: data.tasks?.map((task) => ({
          name: task.name,
          order: task.order,
          columnId: task.status as "todo" | "done" | "in_progress",
          content: task.description,
          id: task.id,
        })),
      };
    },
    queryKey: ["project-with-tasks", projectId],
    queryFn: async () => {
      if (!user?.tokens.access_token || !projectId) {
        throw new Error("Missing access token or project ID");
      }
      const project = await taskProjectGet(user.tokens.access_token, projectId);
      const tasks = await taskList(user.tokens.access_token, {
        project_id: projectId,
        sort_by: "order",
        sort_order: "asc",
        per_page: 50,
      });
      return {
        ...project,
        tasks: tasks.data,
      };
    },
  });

  const mutation = useMutation({
    mutationFn: async (values: z.infer<typeof formSchema>) => {
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token");
      }
      await taskProjectUpdate(user.tokens.access_token, projectId!, {
        name: values.name,
        description: values.description,
        status: values.status,
        order: values.order,
      });
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: ["project-with-tasks"],
      });
      await queryClient.refetchQueries();
      // form.reset();
      toast.success("Project updated!");
    },
    onError: (err: any) => {
      toast.error(`Failed to update project: ${err.message}`);
    },
  });

  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      name: project?.name || "",
      description: project?.description || "",
      status: project?.status as "todo" | "in_progress" | "done" | undefined,
      order: project?.order || 0,
    },
  });
  function onSubmit(values: z.infer<typeof formSchema>) {
    mutation.mutate(values);
  }
  useEffect(() => {
    if (project) {
      form.reset({
        name: project.name,
        description: project.description || "",
        status: project.status as "todo" | "in_progress" | "done",
        order: project.order || 0,
      });
    }
  }, [project]);

  if (loading) return <p>Loading...</p>;
  if (error) return <p>Error: {error.message}</p>;
  if (!project) return <p>Project not found</p>;

  return (
    // <div className="h-full px-4 py-6 lg:px-8 space-y-6">
    <div className="grow">
      <div className="flex flex-row gap-4">
        <div className="px-4 md:px-6 gap-4 flex-1">
          <Link
            to={RouteMap.TASK_PROJECTS}
            className="flex items-center gap-2 text-sm text-muted-foreground"
          >
            <ChevronLeft className="h-4 w-4" />
            Back to projects
          </Link>
          <h1 className="text-2xl font-bold">{project.name}</h1>
          {/* <Tabs
            value={tab}
            onValueChange={onClick}
            className="h-full space-y-6"
          >
            <TabsList>
              <TabsTrigger value="tasks">Tasks</TabsTrigger>
            </TabsList>
            <TabsContent value="tasks" className="flex flex-col gap-4 mx-4 "> */}
          <div className="space-y-4 flex flex-row space-x-16">
            <p className="flex-1">
              Add Tasks to this Project. Users who have this Project will
              receive all Tasks below that match the API of their login request.
            </p>
            <CreateProjectTaskDialog
              projectId={projectId!}
              status={project.status as "todo" | "in_progress" | "done"}
            />
          </div>
          <div className="flex flex-col grow">
            <KanbanBoard cars={project.tasks || []} projectId={projectId!} />
          </div>
          {/* </TabsContent>
          </Tabs> */}
        </div>
        <div className="gap-4 flex-none px-4 md:px-6">
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
              <FormField
                control={form.control}
                name="name"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Name</FormLabel>
                    <FormControl>
                      <Input {...field} />
                    </FormControl>
                    <FormDescription>
                      This is your public display name.
                    </FormDescription>
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
                      {...field}
                      onValueChange={field.onChange}
                      defaultValue={field.value}
                    >
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue placeholder="Select a status" />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        <SelectItem value="todo">Todo</SelectItem>
                        <SelectItem value="in_progress">In Progress</SelectItem>
                        <SelectItem value="done">Done</SelectItem>
                      </SelectContent>
                    </Select>
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
                      <Input placeholder="shadcn" {...field} />
                    </FormControl>
                    <FormDescription>
                      This is your public display name.
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="order"
                render={({ field }) => (
                  <FormItem>
                    {/* <FormLabel>Order</FormLabel> */}
                    <FormControl>
                      <Input {...field} hidden />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <Button type="submit" disabled={!form.formState.isDirty}>
                Submit
              </Button>
            </form>
          </Form>
        </div>
      </div>
    </div>
  );
}
