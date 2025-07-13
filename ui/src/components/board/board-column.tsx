import { Task, TaskCard } from "@/components/board/task-card";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { ScrollArea, ScrollBar } from "@/components/ui/scroll-area";
import { CreateProjectTaskDialog } from "@/pages/teams/projects/tasks/create-project-task-dialog";
import { useDndContext, type UniqueIdentifier } from "@dnd-kit/core";
import { SortableContext, useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { cva } from "class-variance-authority";
import { useMemo } from "react";
import { Badge } from "../ui/badge";

export interface Column {
  id: UniqueIdentifier;
  title: string;
}

export type ColumnType = "Column";

export type ColumnDragData = {
  type: ColumnType;
  column: Column;
};

interface BoardColumnProps {
  column: Column;
  cards: Task[];
  isOverlay?: boolean;
  projectId: string;
}

export const BoardColumn = ({
  column,
  cards: cards,
  isOverlay,
  projectId,
}: BoardColumnProps) => {
  const cardIds = useMemo(() => {
    return cards.map((card) => card.id);
  }, [cards]);

  const { setNodeRef, transform, transition, isDragging } = useSortable({
    id: column.id,
    data: {
      type: "Column",
      column,
    } satisfies ColumnDragData,
    attributes: {
      roleDescription: `Column: ${column.title}`,
    },
  });

  const style = {
    transition,
    transform: CSS.Translate.toString(transform),
  };

  const variants = cva(
    "h-full w-[300px] bg-primary-foreground flex flex-col flex-shrink-0 snap-center mt-4 overflow-y-auto",
    {
      variants: {
        dragging: {
          default: "border-2 border-transparent",
          over: "ring-2 opacity-30",
          overlay: "ring-2 ring-primary",
        },
      },
    }
  );

  return (
    <Card
      ref={setNodeRef}
      style={style}
      className={variants({
        dragging: isOverlay ? "overlay" : isDragging ? "over" : undefined,
      })}
    >
      <CardHeader className="p-4 font-semibold border-b-2 flex flex-row items-center justify-between">
        <h1>{column.title}</h1>
        <Badge variant="outline">{cards.length}</Badge>
      </CardHeader>
      <ScrollArea>
        <CardContent className="flex flex-grow flex-col gap-2 p-2">
          <SortableContext items={cardIds}>
            {cards.length === 0 ? (
              <div className="flex flex-grow items-center justify-center">
                <p className="">No tasks here.</p>
              </div>
            ) : (
              cards.map((card) => <TaskCard key={card.id} task={card} />)
            )}
          </SortableContext>
          <CreateProjectTaskDialog
            projectId={projectId}
            status={column.id as "todo" | "done" | "in_progress"}
          />
        </CardContent>
      </ScrollArea>
    </Card>
  );
};

export const BoardContainer = ({ children }: { children: React.ReactNode }) => {
  const dndContext = useDndContext();

  const variations = cva("px-2 md:px-0 flex lg:justify-center pb-4", {
    variants: {
      dragging: {
        default: "snap-x snap-mandatory",
        active: "snap-none",
      },
    },
  });

  return (
    <ScrollArea
      className={variations({
        dragging: dndContext.active ? "active" : "default",
      })}
    >
      <div className="flex gap-4 items-start flex-row justify-center">
        {children}
      </div>
      <ScrollBar orientation="horizontal" />
    </ScrollArea>
  );
};
