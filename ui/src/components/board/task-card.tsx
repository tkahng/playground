import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { ConfirmDialog, useDialog } from "@/hooks/use-dialog";
import { EditProjectTaskDialog } from "@/pages/teams/projects/tasks/edit-project-task-dialog";
import { Task as DbTask } from "@/schema.types";
import type { UniqueIdentifier } from "@dnd-kit/core";
import { useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { cva } from "class-variance-authority";
import { GripVertical } from "lucide-react";
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
  card: Task;
};
export function TaskCard({ task, isOverlay }: TaskCardProps) {
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
      card: task,
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
        <EditProjectTaskDialog
          dialog={editDialog.props}
          trigger={editDialog.trigger}
          task={task.task}
        />
      </ConfirmDialog>
    </Card>
  );
}
