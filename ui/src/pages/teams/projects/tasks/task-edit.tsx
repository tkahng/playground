import { Button } from "@/components/ui/button";
import { Calendar } from "@/components/ui/calendar";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command";
import { DialogFooter } from "@/components/ui/dialog";
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
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { PopoverContentNoPortal } from "@/components/ui/popover-noportal";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { TaskContext } from "@/context/task-context";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useDialog } from "@/hooks/use-dialog";
import { useTeam } from "@/hooks/use-team";
import { getTeamTeamMembers, updateTask } from "@/lib/api";
import { GetError } from "@/lib/get-error";
import { useTaskQuery } from "@/lib/queries";
import { cn } from "@/lib/utils";
import { zodResolver } from "@hookform/resolvers/zod";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { format } from "date-fns";
import { CalendarIcon, Check, ChevronLeft, ChevronsUpDown } from "lucide-react";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { Link, useParams } from "react-router";
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

export default function TaskEdit() {
  const { taskId } = useParams<{
    projectId: string;
    taskId: string;
    teamId: string;
  }>();
  const {
    data: task,
    isLoading: isTaskLoading,
    error: taskError,
  } = useTaskQuery(taskId);
  const { user } = useAuthProvider();
  const { teamMember, team } = useTeam();
  const endAtDialog = useDialog();
  // const assigneeDialog = useDialog();
  const queryClient = useQueryClient();

  // const [search, setSearch] = useState<string>("");
  const {
    data: members,
    isLoading: isMembersLoading,
    error: membersError,
  } = useQuery({
    queryKey: ["team-members", teamMember?.team_id],
    queryFn: async () => {
      return await getTeamTeamMembers(
        user!.tokens.access_token,
        teamMember!.team_id,
        0,
        50
      );
    },
    enabled: !!teamMember?.team_id && !!user?.tokens.access_token,
  });

  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      name: task?.name,
      description: task?.description || "",
      status: task?.status || "todo",
      assignee_id: task?.assignee_id,
      end_at: task?.end_at,
      parent_id: task?.parent_id,
      rank: task?.rank,
      reporter_id: task?.reporter_id,
      start_at: task?.start_at,
    },
  });

  const mutation = useMutation({
    mutationFn: async (values: z.infer<typeof formSchema>) => {
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token");
      }
      if (!task) {
        throw new Error("Task ID is required");
      }
      await updateTask(user.tokens.access_token, task!.id, {
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
        queryKey: ["task", task?.id],
      });
      toast.success("Task updated successfully");
    },
    onError: (error) => {
      toast.error(`Failed to create task: ${error.message}`);
    },
  });
  useEffect(() => {
    if (task) {
      form.reset({
        assignee_id: task.assignee_id,
        description: task.description || "",
        end_at: task.end_at,
        name: task.name,
        parent_id: task.parent_id,
        rank: task.rank,
        reporter_id: task.reporter_id,
        start_at: task.start_at,
        status: task.status,
      });
    }
  }, [task, form]);
  const onSubmit = (values: z.infer<typeof formSchema>) => {
    mutation.mutate(values);
  };
  if (isTaskLoading || isMembersLoading) {
    return <div>Loading...</div>;
  }
  if (taskError || membersError) {
    const err = GetError(taskError || membersError);
    return <div>Error: {err?.detail}</div>;
  }
  return (
    <TaskContext.Provider value={task!}>
      <div className="space-y-6">
        <Link
          to={`/teams/${team?.slug}/projects/${task?.project_id}`}
          className="flex items-center gap-2 text-sm text-muted-foreground"
        >
          <ChevronLeft className="h-4 w-4" />
          Back to Project
        </Link>
        <h1 className="text-2xl font-bold">Edit Task</h1>
        <h1 className="text-2xl font-bold">{task?.name}</h1>
        <p className="text-sm text-muted-foreground">
          {task?.description || "No description"}
        </p>
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

                <FormField
                  control={form.control}
                  name="end_at"
                  render={({ field }) => (
                    <FormItem className="flex flex-col">
                      <FormLabel>End Date</FormLabel>
                      <Popover {...endAtDialog.props}>
                        <PopoverTrigger asChild>
                          <FormControl>
                            <Button
                              variant={"outline"}
                              className={cn(
                                "w-[240px] pl-3 text-left font-normal",
                                !field.value && "text-muted-foreground"
                              )}
                            >
                              {field.value ? (
                                format(field.value, "PPP")
                              ) : (
                                <span>Pick a date</span>
                              )}
                              <CalendarIcon className="ml-auto h-4 w-4 opacity-50" />
                            </Button>
                          </FormControl>
                        </PopoverTrigger>
                        <PopoverContentNoPortal
                          className="w-auto p-0"
                          align="start"
                        >
                          <Calendar
                            mode="single"
                            selected={
                              field.value ? new Date(field.value) : undefined
                            }
                            onSelect={(selected) => {
                              field.onChange(selected?.toISOString());
                            }}
                            captionLayout="dropdown"
                          />
                        </PopoverContentNoPortal>
                      </Popover>
                      <FormDescription>
                        Set the due date for this task.
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name="assignee_id"
                  render={({ field }) => (
                    <FormItem className="flex flex-col">
                      <FormLabel>Assignee</FormLabel>
                      <Popover>
                        <PopoverTrigger asChild>
                          <FormControl>
                            <Button
                              variant="outline"
                              role="combobox"
                              className={cn(
                                "w-[200px] justify-between",
                                !field.value && "text-muted-foreground"
                              )}
                            >
                              {field.value
                                ? members?.data?.find((member) => {
                                    return member.id === field.value;
                                  })?.user?.email
                                : "Select assignee"}
                              <ChevronsUpDown className="opacity-50" />
                            </Button>
                          </FormControl>
                        </PopoverTrigger>
                        <PopoverContent className="w-[200px] p-0">
                          <Command>
                            <CommandInput
                              placeholder="Search assignee..."
                              className="h-9"
                            />
                            <CommandList>
                              <CommandEmpty>No assignee found.</CommandEmpty>
                              <CommandGroup>
                                <CommandItem
                                  value={"null"}
                                  key={"null"}
                                  onSelect={() => {
                                    form.setValue(field.name, null, {
                                      shouldDirty: true,
                                    });
                                  }}
                                >
                                  None
                                  <Check
                                    className={cn(
                                      "ml-auto",
                                      !field.value ? "opacity-100" : "opacity-0"
                                    )}
                                  />
                                </CommandItem>
                                {members?.data?.map((member) => (
                                  <CommandItem
                                    value={member.user?.email}
                                    key={member.id}
                                    onSelect={() => {
                                      form.setValue(field.name, member.id, {
                                        shouldDirty: true,
                                      });
                                    }}
                                  >
                                    {member.user?.email}
                                    <Check
                                      className={cn(
                                        "ml-auto",
                                        member.id === field.value
                                          ? "opacity-100"
                                          : "opacity-0"
                                      )}
                                    />
                                  </CommandItem>
                                ))}
                              </CommandGroup>
                            </CommandList>
                          </Command>
                        </PopoverContent>
                      </Popover>
                      <FormDescription>
                        This is the language that will be used in the dashboard.
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <DialogFooter>
                  <Button type="submit" disabled={!form.formState.isDirty}>
                    Update Task
                  </Button>
                </DialogFooter>
              </div>
            </div>
          </form>
        </Form>
      </div>
    </TaskContext.Provider>
  );
}
