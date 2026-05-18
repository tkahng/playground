import { useEffect, useId, useMemo, useState } from "react";
import { createPortal } from "react-dom";

import { useAuthProvider } from "@/hooks/use-auth-provider";
import { updateTaskPositionStatus } from "@/lib/task-queries";
import { WorkflowStatus } from "@/schema.types";
import {
  closestCorners,
  DndContext,
  type DragEndEvent,
  type DragOverEvent,
  type DragStartEvent,
  DragOverlay,
  PointerSensor,
  useSensor,
  useSensors,
} from "@dnd-kit/core";
import { arrayMove, SortableContext } from "@dnd-kit/sortable";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { BoardColumn, BoardContainer, Column } from "./board-column";
import { buildItems, findContainer } from "./kanban-board.utils";
import { Task, TaskCard } from "./task-card";

export type { ColumnId } from "./kanban-board.utils";

type Items = Record<string, string[]>;

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

  const [items, setItems] = useState<Items>(() => buildItems(props.cards));
  const [activeId, setActiveId] = useState<string | null>(null);
  const [snapshot, setSnapshot] = useState<Items | null>(null);
  const dndContextId = useId();

  // Sync from server data. Stable because selectTasks is defined outside the
  // parent component — only fires when the server data actually changes.
  useEffect(() => {
    setItems(buildItems(props.cards));
  }, [props.cards]);

  // O(1) task lookup for rendering.
  const taskMap = useMemo(
    () =>
      Object.fromEntries(props.cards.map((c) => [c.id.toString(), c])),
    [props.cards],
  );

  const activeTask = activeId ? taskMap[activeId] : null;

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
      snapshot: Items;
    }) => {
      if (!user?.tokens.access_token) {
        throw new Error("Not authenticated");
      }
      await updateTaskPositionStatus(user.tokens.access_token, taskId, {
        workflow_status_id: workflowStatusId,
        position,
      });
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: [{ key: "project-tasks", project_id: props.projectId }],
      });
    },
    onError: (error, variables) => {
      setItems(variables.snapshot);
      toast.error("Failed to move task", { description: error.message });
    },
  });

  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: { distance: 5 },
    }),
  );

  const onDragStart = ({ active: dragActive }: DragStartEvent) => {
    const id = dragActive.id.toString();
    setActiveId(id);
    // Deep copy for rollback — only item arrays need cloning.
    setSnapshot(
      Object.fromEntries(
        Object.entries(items).map(([k, v]) => [k, [...v]]),
      ),
    );
  };

  // onDragOver handles cross-container moves only; same-container reordering
  // is finalized in onDragEnd.
  const onDragOver = ({ active, over }: DragOverEvent) => {
    const overId = over?.id?.toString();
    if (!overId) return;

    const activeId = active.id.toString();
    const activeContainer = findContainer(items, activeId);
    const overContainer = findContainer(items, overId);

    if (!activeContainer || !overContainer || activeContainer === overContainer)
      return;

    setItems((prev) => {
      const fromItems = prev[activeContainer] ?? [];
      const toItems = prev[overContainer] ?? [];

      // Insert before the item it's hovering over (or end of container).
      const overIndex = toItems.indexOf(overId);
      const insertAt = overIndex === -1 ? toItems.length : overIndex;

      return {
        ...prev,
        [activeContainer]: fromItems.filter((id) => id !== activeId),
        [overContainer]: [
          ...toItems.slice(0, insertAt),
          activeId,
          ...toItems.slice(insertAt),
        ],
      };
    });
  };

  const onDragEnd = ({ over }: DragEndEvent) => {
    const prevActiveId = activeId;
    const prevSnapshot = snapshot;
    setActiveId(null);
    setSnapshot(null);

    if (!prevActiveId) return;

    if (!over) {
      if (prevSnapshot) setItems(prevSnapshot);
      return;
    }

    const overId = over.id.toString();
    const activeContainer = findContainer(items, prevActiveId);
    const overContainer = findContainer(items, overId);

    if (!activeContainer || !overContainer) {
      if (prevSnapshot) setItems(prevSnapshot);
      return;
    }

    let finalPosition = (items[activeContainer] ?? []).indexOf(prevActiveId);

    if (activeContainer === overContainer) {
      // Same container: apply the final drop position.
      const containerItems = items[activeContainer] ?? [];
      const activeIndex = containerItems.indexOf(prevActiveId);
      const overIndex = containerItems.indexOf(overId);

      if (activeIndex !== -1 && overIndex !== -1 && activeIndex !== overIndex) {
        const reordered = arrayMove(containerItems, activeIndex, overIndex);
        finalPosition = overIndex;
        setItems((prev) => ({ ...prev, [activeContainer]: reordered }));
      }
    }

    mutation.mutate({
      taskId: prevActiveId,
      workflowStatusId: activeContainer,
      position: Math.max(0, finalPosition),
      snapshot: prevSnapshot ?? {},
    });
  };

  // All task IDs across all columns — used for the overlay SortableContext.
  const allTaskIds = useMemo(
    () => Object.values(items).flat(),
    [items],
  );

  return (
    <DndContext
      id={dndContextId}
      sensors={sensors}
      collisionDetection={closestCorners}
      onDragStart={onDragStart}
      onDragEnd={onDragEnd}
      onDragOver={onDragOver}
    >
      <BoardContainer>
        {columns.map((col) => (
          <BoardColumn
            key={col.id}
            column={col}
            taskIds={items[col.id.toString()] ?? []}
            taskMap={taskMap}
            projectId={props.projectId}
          />
        ))}
      </BoardContainer>

      {typeof window !== "undefined" &&
        createPortal(
          <DragOverlay>
            {activeTask && (
              <SortableContext items={allTaskIds}>
                <TaskCard task={activeTask} isOverlay />
              </SortableContext>
            )}
          </DragOverlay>,
          document.body,
        )}
    </DndContext>
  );
}
