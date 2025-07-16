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
import {
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
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
import { PopoverContent } from "@/components/ui/popover";
import { PopoverContentNoPortal } from "@/components/ui/popover-noportal";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useDialog } from "@/hooks/use-dialog";
import { useTeam } from "@/hooks/use-team";
import { getTeamTeamMembers, updateTask } from "@/lib/api";
import { cn } from "@/lib/utils";
import { Task, TeamMember } from "@/schema.types";
import { zodResolver } from "@hookform/resolvers/zod";
import { Popover, PopoverTrigger } from "@radix-ui/react-popover";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { format } from "date-fns";
import { CalendarIcon, Check, ChevronsUpDown } from "lucide-react";
import { useEffect, useState } from "react";
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
  const { teamMember } = useTeam();
  const endDateDialog = useDialog();
  const assigneeDialog = useDialog();
  // const [isDialogOpen, setDialogOpen] = useState(false);
  const queryClient = useQueryClient();
  const [members, setMembers] = useState<TeamMember[]>([]);
  // const [search, setSearch] = useState<string>("");
  // const navigate = useNavigate();
  const { data, isFetched } = useQuery({
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
  useEffect(() => {
    setMembers(data?.data?.length ? data.data : []);
  }, [data]);
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
        <DialogTitle>Edit Task Details</DialogTitle>
        <DialogDescription>{task.name}</DialogDescription>
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
                        <SelectItem value="in_progress">In Progress</SelectItem>
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
                    <Popover
                      open={endDateDialog.props.open}
                      onOpenChange={endDateDialog.props.onOpenChange}
                    >
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
                    <Popover
                      open={assigneeDialog.props.open}
                      onOpenChange={assigneeDialog.props.onOpenChange}
                    >
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
                              ? members.find(
                                  (language) => language.id === field.value
                                )?.user?.email
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
                            // onValueChange={(value) => {
                            //   setSearch(value);
                            // }}
                          />
                          <CommandList>
                            <CommandEmpty>No assignee found.</CommandEmpty>
                            <CommandGroup>
                              {isFetched &&
                                members.map((language) => (
                                  <CommandItem
                                    value={language.id}
                                    key={language.id}
                                    onSelect={() => {
                                      form.setValue("assignee_id", language.id);
                                    }}
                                  >
                                    {language.user?.email}
                                    <Check
                                      className={cn(
                                        "ml-auto",
                                        language.id === field.value
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
                <Button type="submit">Update Task</Button>
              </DialogFooter>
            </div>
          </div>
        </form>
      </Form>
    </>
  );
}
