import { useEffect, useId, useMemo, useState } from "react";
import { createPortal } from "react-dom";

import { useAuthProvider } from "@/hooks/use-auth-provider";
import { updateTaskPositionStatus } from "@/lib/task-queries";
import { WorkflowStatus } from "@/schema.types";
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
import { SortableContext } from "@dnd-kit/sortable";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import {
  BoardColumn,
  BoardContainer,
  Column,
  ColumnDragData,
} from "./board-column";
import {
  applyCardOverCard,
  type ColumnId,
} from "./kanban-board.utils";
import { coordinateGetter } from "./keyboard-preset";
import { CardDragData, Task, TaskCard } from "./task-card";

export type { ColumnId } from "./kanban-board.utils";

export function KanbanBoard(props: {
  cards: Task[];
  projectId: string;
  workflowStatuses: WorkflowStatus[];
}) {
  const columns = useMemo<Column[]>(
    () =>
      props.workflowStatuses.map((s) => ({
        id: s.id,
        title: s.name,
        color: s.color,
      })),
    [props.workflowStatuses],
  );

  // Track per-card column overrides applied by drag. Overrides are cleared when
  // the server data updates (after a successful mutation + refetch).
  const [overrides, setOverrides] = useState<Record<string, ColumnId>>({});
  const [activeColumn, setActiveColumn] = useState<Column | null>(null);
  const [activeCard, setActiveCard] = useState<Task | null>(null);
  const dndContextId = useId();

  // Merge server cards with local drag overrides.
  const cards = useMemo<Task[]>(
    () =>
      props.cards.map((card) =>
        overrides[card.id as string]
          ? { ...card, columnId: overrides[card.id as string]! }
          : card,
      ),
    [props.cards, overrides],
  );

  // Clear overrides that are now consistent with server data (after refetch).
  useEffect(() => {
    setOverrides((prev) => {
      const stale = Object.keys(prev).filter((id) => {
        const card = props.cards.find((c) => c.id === id);
        return card && card.columnId === prev[id];
      });
      if (stale.length === 0) return prev;
      const next = { ...prev };
      stale.forEach((id) => delete next[id]);
      return next;
    });
  }, [props.cards]);

  const queryClient = useQueryClient();
  const { user } = useAuthProvider();
  const mutation = useMutation({
    mutationFn: async ({
      taskId,
      workflowStatusId,
      position,
    }: {
      taskId: string;
      workflowStatusId: string;
      position: number;
      previousCards: Record<string, ColumnId>;
    }) => {
      if (!user?.tokens.access_token) {
        throw new Error("Not authenticated");
      }
      await updateTaskPositionStatus(user.tokens.access_token, taskId, {
        workflow_status_id: workflowStatusId,
        position: position,
      });
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: [{ key: "project-tasks", project_id: props.projectId }],
      });
    },
    onError: (error, variables) => {
      // Roll back to pre-drag override state
      setOverrides(variables.previousCards);
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
    }),
  );

  const hasDraggableData = <T extends Active | Over>(
    entry: T | null | undefined,
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

  const columnsId = useMemo(() => columns.map((col) => col.id), [columns]);

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

  const onDragEnd = (event: DragEndEvent) => {
    setActiveColumn(null);
    setActiveCard(null);

    const { active } = event;

    if (!hasDraggableData(active)) return;
    const activeData = active.data.current;
    if (activeData?.type !== "Task") return;

    const activeId = active.id.toString();

    // The target column is whatever onDragOver set in overrides.
    // We use this instead of event.over.id because after onDragOver moves the
    // card to its new column, dnd-kit may report the dragged card itself as the
    // over target (activeId === over.id), which would cause the early-return guard
    // to fire and prevent the mutation.
    const newColumnId = overrides[activeId];
    if (!newColumnId) return; // card was never dragged over a valid target

    const previousOverrides = { ...overrides };

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const position: number = (event.over?.data?.current as any)?.sortable?.index ?? 0;

    mutation.mutate({
      taskId: activeId,
      workflowStatusId: newColumnId,
      position,
      previousCards: previousOverrides,
    });
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

    if (activeData?.type !== "Task") return;

    if (overData?.type === "Task") {
      const result = applyCardOverCard(cards, activeId, overId);
      const moved = result.find((c) => c.id === activeId);
      if (moved) {
        setOverrides((prev) => ({
          ...prev,
          [activeId.toString()]: moved.columnId,
        }));
      }
    } else if (overData?.type === "Column") {
      setOverrides((prev) => ({
        ...prev,
        [activeId.toString()]: overId as ColumnId,
      }));
    }

    // Also initialise override on first drag so onDragEnd can always read it,
    // even if the card never leaves its original column.
    const currentCard = cards.find((c) => c.id === activeId);
    if (currentCard && !overrides[activeId.toString()]) {
      setOverrides((prev) => ({
        ...prev,
        [activeId.toString()]: currentCard.columnId,
      }));
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
          {columns.map((col) => (
            <BoardColumn
              key={col.id}
              column={col}
              cards={cards.filter((card) => card.columnId === col.id)}
              projectId={props.projectId}
            />
          ))}
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
          document.body,
        )}
    </DndContext>
  );
}
