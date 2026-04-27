import type { UniqueIdentifier } from "@dnd-kit/core";
import { arrayMove } from "@dnd-kit/sortable";

export const defaultCols = [
  { id: "todo" as const, title: "Todo" },
  { id: "in_progress" as const, title: "In progress" },
  { id: "done" as const, title: "Done" },
] satisfies { id: string; title: string }[];

export type ColumnId = (typeof defaultCols)[number]["id"];

type CardLike = { id: UniqueIdentifier; columnId: ColumnId };

/**
 * Computes next card array when an active card is dragged over another card.
 * Pure — never mutates the input array or its elements.
 */
export function applyCardOverCard<T extends CardLike>(
  cards: T[],
  activeId: UniqueIdentifier,
  overId: UniqueIdentifier,
): T[] {
  const activeIndex = cards.findIndex((c) => c.id === activeId);
  const overIndex = cards.findIndex((c) => c.id === overId);
  const active = cards[activeIndex];
  const over = cards[overIndex];
  if (!active) return cards;
  if (over && active.columnId !== over.columnId) {
    const updated = cards.map((c) =>
      c.id === activeId ? { ...c, columnId: over.columnId } : c,
    ) as T[];
    return arrayMove(updated, activeIndex, overIndex - 1);
  }
  return arrayMove(cards, activeIndex, overIndex);
}

/**
 * Computes next card array when an active card is dragged over a column drop zone.
 * Pure — never mutates the input array or its elements.
 */
export function applyCardOverColumn<T extends CardLike>(
  cards: T[],
  activeId: UniqueIdentifier,
  columnId: ColumnId,
): T[] {
  const activeIndex = cards.findIndex((c) => c.id === activeId);
  if (activeIndex === -1) return cards;
  return arrayMove(
    cards.map((c) => (c.id === activeId ? { ...c, columnId } : c)) as T[],
    activeIndex,
    activeIndex,
  );
}
