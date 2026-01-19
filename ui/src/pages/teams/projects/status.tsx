import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import { TaskStatus } from "@/schema.types";

const projectStatusConfig: Record<
  TaskStatus,
  { label: string; className: string }
> = {
  todo: {
    label: "To Do",
    className: "bg-muted text-muted-foreground hover:bg-muted",
  },
  in_progress: {
    label: "In Progress",
    className:
      "bg-blue-100 text-blue-700 hover:bg-blue-100 dark:bg-blue-950 dark:text-blue-300",
  },
  done: {
    label: "Done",
    className:
      "bg-green-100 text-green-700 hover:bg-green-100 dark:bg-green-950 dark:text-green-300",
  },
};

const taskStatusConfig: Record<
  TaskStatus,
  { label: string; className: string }
> = {
  todo: {
    label: "To Do",
    className: "bg-muted text-muted-foreground hover:bg-muted",
  },
  in_progress: {
    label: "In Progress",
    className:
      "bg-blue-100 text-blue-700 hover:bg-blue-100 dark:bg-blue-950 dark:text-blue-300",
  },

  done: {
    label: "Done",
    className:
      "bg-green-100 text-green-700 hover:bg-green-100 dark:bg-green-950 dark:text-green-300",
  },
};

export function ProjectStatusBadge({ status }: { status: TaskStatus }) {
  const config = projectStatusConfig[status];
  return (
    <Badge variant="secondary" className={cn("font-medium", config.className)}>
      {config.label}
    </Badge>
  );
}

export function TaskStatusBadge({ status }: { status: TaskStatus }) {
  const config = taskStatusConfig[status];
  return (
    <Badge variant="secondary" className={cn("font-medium", config.className)}>
      {config.label}
    </Badge>
  );
}
