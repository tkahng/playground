import { PlusIcon } from "lucide-react";
import type { FormEvent, KeyboardEvent } from "react";
import { useEffect, useRef, useState } from "react";
import { flushSync } from "react-dom";

import { Input } from "@/components/ui/input";
import {
  KanbanBoard,
  KanbanBoardCard,
  KanbanBoardCardButtonGroup,
  KanbanBoardCardDescription,
  KanbanBoardCardTitle,
  KanbanBoardCircleColor,
  KanbanBoardColumn,
  KanbanBoardColumnButton,
  KanbanBoardColumnFooter,
  KanbanBoardColumnHeader,
  KanbanBoardColumnList,
  KanbanBoardColumnListItem,
  KanbanBoardColumnSkeleton,
  KanbanBoardColumnTitle,
  KanbanBoardDropDirection,
  KanbanColorCircle,
  useDndEvents,
} from "@/components/ui/kanban";

import { ConfirmDialog, useDialog } from "@/hooks/use-dialog";
import { useJsLoaded } from "@/hooks/use-js-loaded";
import { groupItems } from "@/lib/array";
import { useUpdateTaskPosition } from "@/lib/mutation";
import { Task, TaskStatus, TeamMember } from "@/schema.types";
import { CreateProjectTaskDialog2 } from "./tasks/create-project-task-dialog copy";
import { EditProjectTaskDialog } from "./tasks/edit-project-task-dialog";
import { TaskEditDropDown } from "./tasks/task-edit-dropdown";

// Types
type CardType = {
  id: string;
  name: string;
  assignee_id: string | null;
  /** Format: date-time */
  created_at: string;
  created_by_member?: TeamMember;
  created_by_member_id: string | null;
  description: string | null;
  /** Format: date-time */
  end_at: string | null;
  parent_id: string | null;
  project_id: string;
  /** Format: double */
  rank: number;
  reporter_id: string | null;
  /** Format: date-time */
  start_at: string | null;
  /** @enum {string} */
  status: TaskStatus;
  team_id: string;
  /** Format: date-time */
  updated_at: string;
};

type Column = {
  id: string;
  title: string;
  status: TaskStatus;
  projectId: string;
  color: KanbanBoardCircleColor;
  items: CardType[];
};

export function MyKanbanBoard({
  projectId,
  tasks,
}: {
  projectId: string;
  tasks: Task[];
}) {
  const [columns, setColumns] = useState<Column[]>([
    {
      id: "todo",
      title: "Todo",
      status: "todo",
      projectId: projectId,
      color: "blue",
      items: [],
    },
    {
      id: "in_progress",
      title: "In progress",
      status: "in_progress",
      projectId: projectId,
      color: "red",
      items: [],
    },
    {
      id: "done",
      title: "Done",
      status: "done",
      projectId: projectId,
      color: "green",
      items: [],
    },
  ]);
  useEffect(() => {
    const groupedItems = groupItems(tasks || [], (task) => task.status);
    const g = columns.map((c) => ({
      ...c,
      items: groupedItems[c.status] || [],
    }));
    setColumns(g);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [tasks]);
  const mutation = useUpdateTaskPosition();
  // Scroll to the right when a new column is added.
  const scrollContainerReference = useRef<HTMLDivElement>(null);

  function scrollRight() {
    if (scrollContainerReference.current) {
      scrollContainerReference.current.scrollLeft =
        scrollContainerReference.current.scrollWidth;
    }
  }

  /*
  Column logic
  */

  function handleDeleteColumn(columnId: string) {
    console.log(columnId);
    // flushSync(() => {
    //   setColumns((previousColumns) =>
    //     previousColumns.filter((column) => column.id !== columnId)
    //   );
    // });

    scrollRight();
  }

  function handleUpdateColumnTitle(columnId: string, title: string) {
    console.log(columnId, title);
    // setColumns((previousColumns) =>
    //   previousColumns.map((column) =>
    //     column.id === columnId ? { ...column, title } : column
    //   )
    // );
  }

  /*
  Card logic
  */

  function handleAddCard(columnId: string, cardContent: string) {
    console.log(columnId, cardContent);
    // setColumns((previousColumns) =>
    //   previousColumns.map((column) =>
    //     column.id === columnId
    //       ? {
    //           ...column,
    //           items: [
    //             ...column.items,
    //             { id: new Date().getTime().toString(), name: cardContent },
    //           ],
    //         }
    //       : column
    //   )
    // );
  }

  function handleDeleteCard(cardId: string) {
    setColumns((previousColumns) =>
      previousColumns.map((column) =>
        column.items.some((card) => card.id === cardId)
          ? { ...column, items: column.items.filter(({ id }) => id !== cardId) }
          : column
      )
    );
  }

  function handleMoveCardToColumn(
    columnId: string,
    index: number,
    card: CardType
  ) {
    mutation.mutate({
      projectId,
      taskId: card.id,
      status: columnId as TaskStatus,
      position: index,
    });
  }

  function handleUpdateCardTitle(cardId: string, cardTitle: string) {
    console.log(cardId, cardTitle);
    // setColumns((previousColumns) =>
    //   previousColumns.map((column) =>
    //     column.items.some((card) => card.id === cardId)
    //       ? {
    //           ...column,
    //           items: column.items.map((card) =>
    //             card.id === cardId ? { ...card, name: cardTitle } : card
    //           ),
    //         }
    //       : column
    //   )
    // );
  }

  /*
  Moving cards with the keyboard.
  */
  const [activeCardId, setActiveCardId] = useState<string>("");
  const originalCardPositionReference = useRef<{
    columnId: string;
    cardIndex: number;
  } | null>(null);
  const { onDragStart, onDragEnd, onDragCancel, onDragOver } = useDndEvents();

  // This helper returns the appropriate overId after a card is placed.
  // If there's another card below, return that card's id, otherwise return the column's id.
  function getOverId(column: Column, cardIndex: number): string {
    if (cardIndex < column.items.length - 1) {
      return column.items[cardIndex + 1].id;
    }

    return column.id;
  }

  // Find column and index for a given card.
  function findCardPosition(cardId: string): {
    columnIndex: number;
    cardIndex: number;
  } {
    for (const [columnIndex, column] of columns.entries()) {
      const cardIndex = column.items.findIndex((c) => c.id === cardId);

      if (cardIndex !== -1) {
        return { columnIndex, cardIndex };
      }
    }

    return { columnIndex: -1, cardIndex: -1 };
  }

  function moveActiveCard(
    cardId: string,
    direction: "ArrowLeft" | "ArrowRight" | "ArrowUp" | "ArrowDown"
  ) {
    const { columnIndex, cardIndex } = findCardPosition(cardId);
    if (columnIndex === -1 || cardIndex === -1) return;

    const card = columns[columnIndex].items[cardIndex];

    let newColumnIndex = columnIndex;
    let newCardIndex = cardIndex;

    switch (direction) {
      case "ArrowUp": {
        newCardIndex = Math.max(cardIndex - 1, 0);

        break;
      }
      case "ArrowDown": {
        newCardIndex = Math.min(
          cardIndex + 1,
          columns[columnIndex].items.length - 1
        );

        break;
      }
      case "ArrowLeft": {
        newColumnIndex = Math.max(columnIndex - 1, 0);
        // Keep same cardIndex if possible, or if out of range, insert at end
        newCardIndex = Math.min(
          newCardIndex,
          columns[newColumnIndex].items.length
        );

        break;
      }
      case "ArrowRight": {
        newColumnIndex = Math.min(columnIndex + 1, columns.length - 1);
        newCardIndex = Math.min(
          newCardIndex,
          columns[newColumnIndex].items.length
        );

        break;
      }
    }

    // Perform state update in flushSync to ensure immediate state update.
    flushSync(() => {
      handleMoveCardToColumn(columns[newColumnIndex].id, newCardIndex, card);
    });

    // Find the card's new position and announce it.
    const { columnIndex: updatedColumnIndex, cardIndex: updatedCardIndex } =
      findCardPosition(cardId);
    const overId = getOverId(columns[updatedColumnIndex], updatedCardIndex);

    onDragOver(cardId, overId);
  }

  function handleCardKeyDown(
    event: KeyboardEvent<HTMLButtonElement>,
    cardId: string
  ) {
    const { key } = event;

    if (activeCardId === "" && key === " ") {
      // Pick up the card.
      event.preventDefault();
      setActiveCardId(cardId);
      onDragStart(cardId);

      const { columnIndex, cardIndex } = findCardPosition(cardId);
      originalCardPositionReference.current =
        columnIndex !== -1 && cardIndex !== -1
          ? { columnId: columns[columnIndex].id, cardIndex }
          : null;
    } else if (activeCardId === cardId) {
      // Card is already active.
      if (key === " " || key === "Enter") {
        event.preventDefault();
        // Drop the card.
        flushSync(() => {
          setActiveCardId("");
        });

        const { columnIndex, cardIndex } = findCardPosition(cardId);
        if (columnIndex !== -1 && cardIndex !== -1) {
          const overId = getOverId(columns[columnIndex], cardIndex);
          onDragEnd(cardId, overId);
        } else {
          // If we somehow can't find the card, just call onDragEnd with cardId.
          onDragEnd(cardId);
        }

        originalCardPositionReference.current = null;
      } else if (key === "Escape") {
        event.preventDefault();

        // Cancel the drag.
        if (originalCardPositionReference.current) {
          const { columnId, cardIndex } = originalCardPositionReference.current;
          const {
            columnIndex: currentColumnIndex,
            cardIndex: currentCardIndex,
          } = findCardPosition(cardId);

          // Revert card only if it moved.
          if (
            currentColumnIndex !== -1 &&
            (columnId !== columns[currentColumnIndex].id ||
              cardIndex !== currentCardIndex)
          ) {
            const card = columns[currentColumnIndex].items[currentCardIndex];
            flushSync(() => {
              handleMoveCardToColumn(columnId, cardIndex, card);
            });
          }
        }

        onDragCancel(cardId);
        originalCardPositionReference.current = null;

        setActiveCardId("");
      } else if (
        key === "ArrowLeft" ||
        key === "ArrowRight" ||
        key === "ArrowUp" ||
        key === "ArrowDown"
      ) {
        event.preventDefault();
        moveActiveCard(cardId, key);
        // onDragOver is called inside moveActiveCard after placement.
      }
    }
  }

  function handleCardBlur() {
    setActiveCardId("");
  }

  const jsLoaded = useJsLoaded();

  return (
    <KanbanBoard ref={scrollContainerReference} className="flex-grow">
      {columns.map((column) =>
        jsLoaded ? (
          <MyKanbanBoardColumn
            activeCardId={activeCardId}
            column={column}
            key={column.id}
            onAddCard={handleAddCard}
            onCardBlur={handleCardBlur}
            onCardKeyDown={handleCardKeyDown}
            onDeleteCard={handleDeleteCard}
            onDeleteColumn={handleDeleteColumn}
            onMoveCardToColumn={handleMoveCardToColumn}
            onUpdateCardTitle={handleUpdateCardTitle}
            onUpdateColumnTitle={handleUpdateColumnTitle}
          />
        ) : (
          <KanbanBoardColumnSkeleton key={column.id} />
        )
      )}

      {/* Add a new column */}
    </KanbanBoard>
  );
}

function MyKanbanBoardColumn({
  activeCardId,
  column,
  onAddCard,
  onCardBlur,
  onCardKeyDown,
  onDeleteCard,
  onMoveCardToColumn,
  onUpdateCardTitle,
  onUpdateColumnTitle,
}: {
  activeCardId: string;
  column: Column;
  onAddCard: (columnId: string, cardContent: string) => void;
  onCardBlur: () => void;
  onCardKeyDown: (
    event: KeyboardEvent<HTMLButtonElement>,
    cardId: string
  ) => void;
  onDeleteCard: (cardId: string) => void;
  onDeleteColumn: (columnId: string) => void;
  onMoveCardToColumn: (columnId: string, index: number, card: CardType) => void;
  onUpdateCardTitle: (cardId: string, cardTitle: string) => void;
  onUpdateColumnTitle: (columnId: string, columnTitle: string) => void;
}) {
  const [isEditingTitle, setIsEditingTitle] = useState(false);
  const listReference = useRef<HTMLUListElement>(null);
  const moreOptionsButtonReference = useRef<HTMLButtonElement>(null);
  const { onDragCancel, onDragEnd } = useDndEvents();

  function scrollList() {
    if (listReference.current) {
      listReference.current.scrollTop = listReference.current.scrollHeight;
    }
  }

  function closeDropdownMenu() {
    flushSync(() => {
      setIsEditingTitle(false);
    });

    moreOptionsButtonReference.current?.focus();
  }

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const formData = new FormData(event.currentTarget);
    const columnTitle = formData.get("columnTitle") as string;
    onUpdateColumnTitle(column.id, columnTitle);
    closeDropdownMenu();
  }

  function handleDropOverColumn(dataTransferData: string) {
    const card = JSON.parse(dataTransferData) as CardType;
    onMoveCardToColumn(column.id, 0, card);
  }

  function handleDropOverListItem(cardId: string) {
    return (
      dataTransferData: string,
      dropDirection: KanbanBoardDropDirection
    ) => {
      const card = JSON.parse(dataTransferData) as CardType;
      const cardIndex = column.items.findIndex(({ id }) => id === cardId);
      const currentCardIndex = column.items.findIndex(
        ({ id }) => id === card.id
      );

      const baseIndex = dropDirection === "top" ? cardIndex : cardIndex + 1;
      const targetIndex =
        currentCardIndex !== -1 && currentCardIndex < baseIndex
          ? baseIndex - 1
          : baseIndex;

      // Safety check to ensure targetIndex is within bounds
      const safeTargetIndex = Math.max(
        0,
        Math.min(targetIndex, column.items.length)
      );
      const overCard = column.items[safeTargetIndex];

      if (card.id === overCard?.id) {
        onDragCancel(card.id);
      } else {
        onMoveCardToColumn(column.id, safeTargetIndex, card);
        onDragEnd(card.id, overCard?.id || column.id);
      }
    };
  }

  return (
    <KanbanBoardColumn
      columnId={column.id}
      key={column.id}
      onDropOverColumn={handleDropOverColumn}
      className="grow"
    >
      <KanbanBoardColumnHeader>
        {isEditingTitle ? (
          <form
            className="w-full"
            onSubmit={handleSubmit}
            onBlur={(event) => {
              if (!event.currentTarget.contains(event.relatedTarget)) {
                closeDropdownMenu();
              }
            }}
          >
            <Input
              aria-label="Column title"
              autoFocus
              defaultValue={column.title}
              name="columnTitle"
              onKeyDown={(event) => {
                if (event.key === "Escape") {
                  closeDropdownMenu();
                }
              }}
              required
            />
          </form>
        ) : (
          <>
            <KanbanBoardColumnTitle columnId={column.id}>
              <KanbanColorCircle color={column.color} />
              {column.title}
            </KanbanBoardColumnTitle>
          </>
        )}
      </KanbanBoardColumnHeader>

      <KanbanBoardColumnList ref={listReference}>
        {column.items.map((card) => (
          <KanbanBoardColumnListItem
            cardId={card.id}
            key={card.id}
            onDropOverListItem={handleDropOverListItem(card.id)}
          >
            <MyKanbanBoardCard
              card={card}
              isActive={activeCardId === card.id}
              onCardBlur={onCardBlur}
              onCardKeyDown={onCardKeyDown}
              onDeleteCard={onDeleteCard}
              onUpdateCardTitle={onUpdateCardTitle}
            />
          </KanbanBoardColumnListItem>
        ))}
      </KanbanBoardColumnList>

      <MyNewKanbanBoardCard
        column={column}
        onAddCard={onAddCard}
        scrollList={scrollList}
      />
    </KanbanBoardColumn>
  );
}

function MyKanbanBoardCard({
  card,
  isActive,
  onCardBlur,
  onCardKeyDown,
  onDeleteCard,
}: {
  card: CardType;
  isActive: boolean;
  onCardBlur: () => void;
  onCardKeyDown: (
    event: KeyboardEvent<HTMLButtonElement>,
    cardId: string
  ) => void;
  onDeleteCard: (cardId: string) => void;
  onUpdateCardTitle: (cardId: string, cardTitle: string) => void;
}) {
  const [isEditingTitle, setIsEditingTitle] = useState(false);
  const kanbanBoardCardReference = useRef<HTMLButtonElement>(null);
  // This ref tracks the previous `isActive` state. It is used to refocus the
  // card after it was discarded with the keyboard.
  const previousIsActiveReference = useRef(isActive);
  // This ref tracks if the card was cancelled via Escape.
  const wasCancelledReference = useRef(false);
  const editDialog = useDialog();
  useEffect(() => {
    // Maintain focus after the card is picked up and moved.
    if (isActive && !isEditingTitle) {
      kanbanBoardCardReference.current?.focus();
    }

    // Refocus the card after it was discarded with the keyboard.
    if (
      !isActive &&
      previousIsActiveReference.current &&
      wasCancelledReference.current
    ) {
      kanbanBoardCardReference.current?.focus();
      wasCancelledReference.current = false;
    }

    previousIsActiveReference.current = isActive;
  }, [isActive, isEditingTitle]);
  function handleEditCardClick() {
    if (!editDialog.props.open) {
      editDialog.props.onOpenChange(true);
      setIsEditingTitle(true);
    }
  }
  return (
    <KanbanBoardCard
      data={card}
      isActive={isActive}
      onBlur={onCardBlur}
      onClick={() => {
        handleEditCardClick();
      }}
      onKeyDown={(event) => {
        if (event.key === " ") {
          // Prevent the button "click" action on space because that should
          // be used to pick up and move the card using the keyboard.
          event.preventDefault();
        }

        if (event.key === "Escape") {
          // Mark that this card was cancelled.
          wasCancelledReference.current = true;
        }

        onCardKeyDown(event, card.id);
      }}
      ref={kanbanBoardCardReference}
    >
      <KanbanBoardCardTitle>{card.name}</KanbanBoardCardTitle>
      <KanbanBoardCardDescription>
        {card.description}
      </KanbanBoardCardDescription>
      <KanbanBoardCardButtonGroup disabled={isActive}>
        <TaskEditDropDown task={card} onDelete={() => onDeleteCard(card.id)} />
        {/* <KanbanBoardCardButton
          className="text-destructive"
          onClick={() => onDeleteCard(card.id)}
          tooltip="Delete card"
        >
          <Ellipsis />

          <span className="sr-only">Delete card</span>
        </KanbanBoardCardButton> */}
      </KanbanBoardCardButtonGroup>
      <ConfirmDialog dialogProps={editDialog.props}>
        <EditProjectTaskDialog task={card} props={editDialog.props} />
      </ConfirmDialog>
    </KanbanBoardCard>
  );
}

function MyNewKanbanBoardCard({
  column,
  scrollList,
}: {
  column: Column;
  onAddCard: (columnId: string, cardContent: string) => void;
  scrollList: () => void;
}) {
  const editDialog = useDialog();
  // const [cardContent, setCardContent] = useState("");
  const newCardButtonReference = useRef<HTMLButtonElement>(null);
  // const submitButtonReference = useRef<HTMLButtonElement>(null);
  // const [showNewCardForm, setShowNewCardForm] = useState(false);

  function handleAddCardClick() {
    if (!editDialog.props.open) {
      editDialog.props.onOpenChange(true);
      scrollList();
    }
  }
  return (
    <KanbanBoardColumnFooter>
      <KanbanBoardColumnButton
        onClick={handleAddCardClick}
        ref={newCardButtonReference}
      >
        <PlusIcon />

        <span aria-hidden>New card</span>

        <span className="sr-only">Add new card to {column.title}</span>
      </KanbanBoardColumnButton>
      <ConfirmDialog dialogProps={editDialog.props}>
        <CreateProjectTaskDialog2
          projectId={column.projectId}
          status={column.status}
          onFinish={() => {
            editDialog.props.onOpenChange(false);
          }}
        />
      </ConfirmDialog>
    </KanbanBoardColumnFooter>
  );
}
