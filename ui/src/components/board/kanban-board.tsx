import { useCallback, useEffect, useId, useMemo, useState } from "react";
import { createPortal } from "react-dom";

import { useAuthProvider } from "@/hooks/use-auth-provider";
import { updateTaskPositionStatus } from "@/lib/api";
import {
  Active,
  DataRef,
  DndContext,
  type DragEndEvent,
  type DragOverEvent,
  DragOverlay,
  type DragStartEvent,
  KeyboardSensor,
  MouseSensor,
  Over,
  TouchSensor,
  useSensor,
  useSensors,
} from "@dnd-kit/core";
import { SortableContext, arrayMove } from "@dnd-kit/sortable";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import {
  BoardColumn,
  BoardContainer,
  Column,
  ColumnDragData,
} from "./board-column";
import { coordinateGetter } from "./keyboard-preset";
import { CardDragData, Task, TaskCard } from "./task-card";

type NestedColumn = Column & {
  children?: NestedColumn[];
};

const defaultCols = [
  {
    id: "todo" as const,
    title: "Todo",
  },
  {
    id: "in_progress" as const,
    title: "In progress",
  },
  {
    id: "done" as const,
    title: "Done",
  },
] satisfies Column[];

export type ColumnId = (typeof defaultCols)[number]["id"];

export function KanbanBoard(props: { cards: Task[]; projectId: string }) {
  const [columns, setColumns] = useState<Column[]>(defaultCols);
  const [cards, setCards] = useState<Task[]>(props.cards);
  const [activeColumn, setActiveColumn] = useState<Column | null>(null);
  const [activeCard, setActiveCard] = useState<Task | null>(null);
  const dndContextId = useId();

  useEffect(() => {
    setCards(props.cards);
  }, [props.cards]);

  const queryClient = useQueryClient();
  const { user } = useAuthProvider();
  const mutation = useMutation({
    mutationFn: async ({
      taskId,
      status,
      position,
    }: {
      taskId: string;
      status: "todo" | "in_progress" | "done";
      position: number;
    }) => {
      if (!user?.tokens.access_token) return;
      await updateTaskPositionStatus(user?.tokens.access_token, taskId, {
        status: status,
        position: position,
      });
      return;
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: ["project-tasks", props.projectId],
      });
      toast.success("Task updated");
    },
    onError: (error) => {
      toast.error("Failed to update task", {
        description: error.message,
      });
    },
  });
  const sensors = useSensors(
    useSensor(MouseSensor),
    useSensor(TouchSensor),
    useSensor(KeyboardSensor, {
      coordinateGetter: coordinateGetter,
    })
  );

  const hasDraggableData = <T extends Active | Over>(
    entry: T | null | undefined
  ): entry is T & {
    data: DataRef<CardDragData | ColumnDragData>;
  } => {
    if (!entry) {
      return false;
    }

    const data = entry.data.current;

    if (data?.type === "Column" || data?.type === "Task") {
      return true;
    }

    return false;
  };

  // Helper function to flatten nested columns
  const flattenColumns = useCallback((cols: NestedColumn[]): Column[] => {
    return cols.flatMap((col) =>
      col.children
        ? [{ id: col.id, title: col.title }, ...flattenColumns(col.children)]
        : [col]
    );
  }, []);

  const flatColumns = useMemo(
    () => flattenColumns(columns),
    [columns, flattenColumns]
  );
  const columnsId = useMemo(
    () => flatColumns.map((col) => col.id),
    [flatColumns]
  );

  // recursively render nested columns
  const renderNestedColumns = (cols: NestedColumn[]) => {
    return cols.map((col) => {
      const cardsInColumn = cards.filter((card) => card.columnId === col.id);

      if (col.children && col.children.length > 0) {
        return (
          <div key={col.id} className="flex flex-col">
            {cardsInColumn.length > 0 && (
              <BoardColumn
                column={col}
                cards={cardsInColumn}
                projectId={props.projectId}
              />
            )}
            <div className={cardsInColumn.length > 0 ? "ml-4 mt-2" : ""}>
              {renderNestedColumns(col.children)}
            </div>
          </div>
        );
      } else {
        return (
          <BoardColumn
            key={col.id}
            column={col}
            cards={cardsInColumn}
            projectId={props.projectId}
          />
        );
      }
    });
  };

  const onDragStart = (event: DragStartEvent) => {
    if (!hasDraggableData(event.active)) return;
    const data = event.active.data.current;
    if (data?.type === "Column") {
      setActiveColumn(data.column);
      return;
    }

    if (data?.type === "Task") {
      setActiveCard(data.card);
      return;
    }
  };

  const onDragEnd = async (event: DragEndEvent) => {
    setActiveColumn(null);
    setActiveCard(null);

    const { active, over } = event;
    if (!over) return;

    const activeId = active.id;
    const overId = over.id;

    if (!hasDraggableData(active)) return;

    const activeData = active.data.current;

    if (activeId === overId) return;

    const isActiveAColumn = activeData?.type === "Column";
    if (isActiveAColumn) {
      setColumns((columns) => {
        const activeColumnIndex = columns.findIndex(
          (col) => col.id === activeId
        );
        const overColumnIndex = columns.findIndex((col) => col.id === overId);
        return arrayMove(columns, activeColumnIndex, overColumnIndex);
      });
    } else if (activeData?.type === "Task") {
      const newColumnId = hasDraggableData(over)
        ? over.data.current?.type === "Column"
          ? (over.id as ColumnId)
          : over.data.current?.card.columnId
        : (over.id as ColumnId);

      const oldColumnId = activeData.card.columnId;

      if (oldColumnId !== newColumnId) {
        setCards((cars) => {
          return cars.map((car) =>
            car.id === activeId && newColumnId
              ? { ...car, columnId: newColumnId }
              : car
          );
        });
      }
    }
  };

  const onDragOver = (event: DragOverEvent) => {
    const { active, over } = event;
    if (!over) return;

    const activeId = active.id;
    const overId = over.id;

    if (activeId === overId) return;

    if (!hasDraggableData(active) || !hasDraggableData(over)) return;

    const activeData = active.data.current;
    const overData = over.data.current;

    const isActiveACar = activeData?.type === "Task";
    const isOverACar = overData?.type === "Task";

    if (!isActiveACar) return;

    if (isActiveACar && isOverACar) {
      setCards((cars) => {
        const activeIndex = cars.findIndex((car) => car.id === activeId);
        const overIndex = cars.findIndex((car) => car.id === overId);
        const activeCar = cars[activeIndex];
        const overCar = cars[overIndex];
        if (activeCar && overCar && activeCar.columnId !== overCar.columnId) {
          activeCar.columnId = overCar.columnId;
          mutation.mutate({
            taskId: activeCar.id.toString(),
            status: activeCar.columnId,
            position: overData.sortable.index,
          });
          return arrayMove(cars, activeIndex, overIndex - 1);
        }
        mutation.mutate({
          taskId: activeCar.id.toString(),
          status: activeCar.columnId,
          position: overData.sortable.index,
        });
        return arrayMove(cars, activeIndex, overIndex);
      });
    }

    const isOverAColumn = overData?.type === "Column";

    if (isActiveACar && isOverAColumn) {
      setCards((cars) => {
        const activeIndex = cars.findIndex((car) => car.id === activeId);
        const activeCar = cars[activeIndex];
        if (activeCar) {
          activeCar.columnId = overId as ColumnId;
          mutation.mutate({
            taskId: activeCar.id.toString(),
            status: activeCar.columnId,
            position: overData.sortable.index,
          });
          return arrayMove(cars, activeIndex, activeIndex);
        }
        return cars;
      });
    }
  };

  return (
    <DndContext
      id={dndContextId}
      sensors={sensors}
      onDragStart={onDragStart}
      onDragEnd={onDragEnd}
      onDragOver={onDragOver}
    >
      <BoardContainer>
        <SortableContext items={columnsId}>
          {renderNestedColumns(columns)}
        </SortableContext>
      </BoardContainer>

      {typeof window !== "undefined" &&
        createPortal(
          <DragOverlay>
            {activeColumn && (
              <BoardColumn
                projectId={props.projectId}
                column={activeColumn}
                cards={cards.filter((car) => car.columnId === activeColumn.id)}
                isOverlay
              />
            )}
            {activeCard && <TaskCard task={activeCard} isOverlay />}
          </DragOverlay>,
          document.body
        )}
    </DndContext>
  );
}
