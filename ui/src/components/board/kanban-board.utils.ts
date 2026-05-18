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

/**
 * Build a container→taskIds map from a flat task array.
 * All known column IDs are pre-initialised (even if empty) so that empty
 * columns are valid droppable targets in the drag system.
 */
export function buildItems(
  cards: { id: string | number; workflowStatusId: string }[],
  columnIds: string[] = [],
): Record<string, string[]> {
  const result: Record<string, string[]> = {};
  // Pre-initialise every column so empty columns exist in the record.
  for (const id of columnIds) {
    result[id] = [];
  }
  for (const card of cards) {
    const col = card.workflowStatusId;
    if (!col) continue;
    if (!result[col]) result[col] = [];
    result[col].push(card.id.toString());
  }
  return result;
}
