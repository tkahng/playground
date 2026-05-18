// Utilities for multi-container kanban drag-and-drop.

export type ColumnId = string;

/** O(1) lookup: which container holds `id`? */
export function findContainer(
  items: Record<string, string[]>,
  id: string,
): string | undefined {
  if (id in items) return id; // id is itself a container
  return Object.keys(items).find((key) => items[key]?.includes(id));
}

/** Build a container→taskIds map from a flat task array. */
export function buildItems(
  cards: { id: string | number; workflowStatusId: string }[],
): Record<string, string[]> {
  const result: Record<string, string[]> = {};
  for (const card of cards) {
    const col = card.workflowStatusId;
    if (!col) continue;
    if (!result[col]) result[col] = [];
    result[col].push(card.id.toString());
  }
  return result;
}
