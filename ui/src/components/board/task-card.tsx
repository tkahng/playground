import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { ConfirmDialog, useDialog } from "@/hooks/use-dialog";
import { taskQueries } from "@/lib/queries";
import { cn } from "@/lib/utils";
import { Task as DbTask } from "@/schema.types";
import type { UniqueIdentifier } from "@dnd-kit/core";
import { useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@radix-ui/react-popover";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@radix-ui/react-select";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { cva } from "class-variance-authority";
import { format } from "date-fns";
import { CalendarIcon, GripVertical } from "lucide-react";
import { useForm } from "react-hook-form";
import { Form } from "react-router";
import { toast } from "sonner";
import { z } from "zod";
import { Calendar } from "../ui/calendar";
import {
  DialogClose,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "../ui/dialog";
import {
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "../ui/form";
import { Input } from "../ui/input";
import { ColumnId } from "./kanban-board";

export type Task = {
  id: UniqueIdentifier;
  name: string;
  columnId: ColumnId;
  content: string | null;
  rank: number;
  task: DbTask;
};

type TaskCardProps = {
  task: Task;
  isOverlay?: boolean;
};

export type CardType = "Task";

export type CardDragData = {
  type: CardType;
  car: Task;
};
const formSchema = z.object({
  name: z.string().min(1),
  // name: string;
  description: z.string().min(0).nullable(),
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
export function TaskCard({ task, isOverlay }: TaskCardProps) {
  const { user } = useAuthProvider();
  const {
    setNodeRef,
    attributes,
    listeners,
    transform,
    transition,
    isDragging,
  } = useSortable({
    id: task.id,
    data: {
      type: "Task",
      car: task,
    } satisfies CardDragData,
    attributes: {
      roleDescription: "Task",
    },
  });
  const editDialog = useDialog();
  const style = {
    transition,
    transform: CSS.Translate.toString(transform),
  };

  const variants = cva("", {
    variants: {
      dragging: {
        over: "ring-2 opacity-30",
        overlay: "ring-2 ring-primary",
      },
    },
  });
  // const handlePointerDown = () => {
  //   dragStartTime.current = Date.now();
  // };

  // const handlePointerUp = () => {
  //   const now = Date.now();
  //   const delta = now - (dragStartTime.current ?? now);

  //   if (delta < 200) {
  //     // Assume it's a click, not a drag
  //     setOpen(true);
  //   }
  // };
  const queryClient = useQueryClient();
  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      name: task.task.name,
      description: task.task.description || "",
      status: task.task.status,
      assignee_id: task.task.assignee_id,
      created_by_member_id: task.task.created_by_member_id,
      end_at: task.task.end_at,
      parent_id: task.task.parent_id,
      project_id: task.task.project_id,
      rank: task.task.rank,
      reporter_id: task.task.reporter_id,
      start_at: task.task.start_at,
      team_id: task.task.team_id,
    },
  });
  const mutation = useMutation({
    mutationFn: async (values: z.infer<typeof formSchema>) => {
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token");
      }
      await taskQueries.updateTask(user.tokens.access_token, task.task.id, {
        ...values,
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
        queryKey: ["project-with-tasks", task.task.project_id],
      });
      toast.success("Task updated successfully");
    },
    onError: (error) => {
      toast.error(`Failed to create task: ${error.message}`);
    },
  });
  const onSubmit = (values: z.infer<typeof formSchema>) => {
    mutation.mutate(values);
  };
  return (
    <Card
      ref={setNodeRef}
      style={style}
      className={variants({
        dragging: isOverlay ? "overlay" : isDragging ? "over" : undefined,
      })}
      onDoubleClick={editDialog.trigger}
    >
      <CardContent className="p-4 flex items-center align-middle text-left whitespace-pre-wrap">
        <Button
          variant="ghost"
          {...attributes}
          {...listeners}
          className="p-1 -ml-2 h-auto cursor-grab"
        >
          <span className="sr-only">Move car</span>
          <GripVertical />
        </Button>
        <div className="text-sm">
          {task.name}
          <br />
          {task.content}
          <br />
          {/* {task.order} */}
        </div>
      </CardContent>
      <ConfirmDialog dialogProps={editDialog.props}>
        <>
          <DialogHeader>
            <DialogTitle>Edit Task</DialogTitle>
          </DialogHeader>
          <DialogDescription>Edit Task</DialogDescription>
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
                    defaultValue={task.task.description || ""}
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Description</FormLabel>
                        <FormControl>
                          <Input
                            {...field}
                            value={field.value || ""}
                            placeholder="Task Description"
                          />
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
                        <FormLabel>Date of birth</FormLabel>
                        <Popover>
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
                          <PopoverContent className="w-auto p-0" align="start">
                            <Calendar
                              mode="single"
                              selected={
                                field.value ? new Date(field.value) : undefined
                              }
                              onSelect={field.onChange}
                              disabled={(date) =>
                                date > new Date() ||
                                date < new Date("1900-01-01")
                              }
                              captionLayout="dropdown"
                            />
                          </PopoverContent>
                        </Popover>
                        <FormDescription>
                          Your date of birth is used to calculate your age.
                        </FormDescription>
                        <FormMessage />
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
          <DialogFooter>
            <DialogClose asChild>
              <Button
                variant="outline"
                onClick={() => {
                  console.log("cancel");
                  // editDialog.props.onOpenChange(false);
                }}
              >
                Cancel
              </Button>
            </DialogClose>
            <DialogClose asChild>
              <Button
                variant="destructive"
                onClick={() => {
                  console.log("delete");
                  // editDialog.props.onOpenChange(false);
                  // onDelete(permissionId);
                }}
              >
                Delete
              </Button>
            </DialogClose>
          </DialogFooter>
        </>
      </ConfirmDialog>
    </Card>
  );
}
