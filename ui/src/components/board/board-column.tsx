import { CreateProjectTaskDialog } from "@/components/board/create-project-task-dialog";
import { Task, TaskCard } from "@/components/board/task-card";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { ScrollArea, ScrollBar } from "@/components/ui/scroll-area";
import { useDndContext, useDroppable, type UniqueIdentifier } from "@dnd-kit/core";
import { SortableContext } from "@dnd-kit/sortable";
import { cva } from "class-variance-authority";
import { Badge } from "../ui/badge";

export interface Column {
  id: UniqueIdentifier;
  title: string;
  color?: string | null;
}

interface BoardColumnProps {
  column: Column;
  taskIds: string[];
  taskMap: Record<string, Task>;
  isOverlay?: boolean;
  projectId: string;
}

export const BoardColumn = ({
  column,
  taskIds,
  taskMap,
  isOverlay,
  projectId,
}: BoardColumnProps) => {
  const { setNodeRef, isOver } = useDroppable({
    id: column.id,
    data: { type: "Column", column },
  });

  const variants = cva(
    "h-full w-[calc(100vw-2rem)] sm:w-[300px] bg-primary-foreground flex flex-col flex-shrink-0 snap-center mt-4 overflow-y-auto transition-colors",
    {
      variants: {
        state: {
          default: "border-2 border-transparent",
          over: "border-2 border-primary/40 bg-primary/5",
          overlay: "ring-2 ring-primary",
        },
      },
    },
  );

  return (
    <Card
      ref={setNodeRef}
      className={variants({
        state: isOverlay ? "overlay" : isOver ? "over" : "default",
      })}
    >
      <CardHeader className="p-4 font-semibold border-b-2 flex flex-row items-center justify-between">
        <div className="flex items-center gap-2">
          {column.color && (
            <span
              className="inline-block w-3 h-3 rounded-full flex-shrink-0"
              style={{ backgroundColor: column.color }}
            />
          )}
          <h1>{column.title}</h1>
        </div>
        <Badge variant="outline">{taskIds.length}</Badge>
      </CardHeader>
      <ScrollArea>
        <CardContent className="flex flex-grow flex-col gap-2 p-2">
          <SortableContext items={taskIds}>
            {taskIds.length === 0 ? (
              <div className="flex flex-grow items-center justify-center py-8 text-sm text-muted-foreground">
                No tasks here.
              </div>
            ) : (
              taskIds.map((id) => {
                const card = taskMap[id];
                return card ? <TaskCard key={id} task={card} /> : null;
              })
            )}
          </SortableContext>
          <CreateProjectTaskDialog
            projectId={projectId}
            workflowStatusId={column.id.toString()}
          />
        </CardContent>
      </ScrollArea>
    </Card>
  );
};

export const BoardContainer = ({ children }: { children: React.ReactNode }) => {
  const dndContext = useDndContext();

  const variations = cva(
    "px-4 md:px-0 flex lg:justify-center pb-4 overscroll-x-contain",
    {
      variants: {
        dragging: {
          default: "snap-x snap-mandatory",
          active: "snap-none",
        },
      },
    },
  );

  return (
    <div className="relative">
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
      {/* Fade indicator — hints that more columns exist to the right */}
      <div className="pointer-events-none absolute inset-y-0 right-0 w-8 bg-gradient-to-l from-background to-transparent sm:hidden" />
    </div>
  );
};
