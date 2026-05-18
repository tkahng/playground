import { describe, expect, it } from "vitest";
import type { Task } from "../task-card";
import {
  applyCardOverCard,
  applyCardOverColumn,
} from "../kanban-board.utils";

// Use UUID-style column IDs matching the production contract (workflow status IDs).
const COL_A = "11111111-1111-1111-1111-111111111111";
const COL_B = "22222222-2222-2222-2222-222222222222";
const COL_C = "33333333-3333-3333-3333-333333333333";

function makeTask(id: string, columnId: string, rank = 0): Task {
  return {
    id,
    name: `Task ${id}`,
    columnId,
    content: null,
    rank,
    task: {} as Task["task"],
  };
}

describe("applyCardOverCard", () => {
  it("reorders within the same column without changing columnId", () => {
    const cards = [
      makeTask("a", COL_A),
      makeTask("b", COL_A),
      makeTask("c", COL_A),
    ];

    const result = applyCardOverCard(cards, "a", "c");

    expect(result.map((c) => c.id)).toEqual(["b", "c", "a"]);
    expect(result.every((c) => c.columnId === COL_A)).toBe(true);
  });

  it("moves card to the target column when columns differ", () => {
    const cards = [
      makeTask("a", COL_A),
      makeTask("b", COL_B),
      makeTask("c", COL_B),
    ];

    const result = applyCardOverCard(cards, "a", "b");

    const moved = result.find((c) => c.id === "a")!;
    expect(moved.columnId).toBe(COL_B);
  });

  it("does not mutate the original cards array or its elements", () => {
    const cards = [makeTask("a", COL_A), makeTask("b", COL_B)];
    const originalColumnId = cards[0]?.columnId;

    applyCardOverCard(cards, "a", "b");

    expect(cards[0]?.columnId).toBe(originalColumnId);
    expect(cards).toHaveLength(2);
  });

  it("returns the original array when activeId is not found", () => {
    const cards = [makeTask("a", COL_A), makeTask("b", COL_A)];

    const result = applyCardOverCard(cards, "nonexistent", "b");

    expect(result).toBe(cards);
  });

  it("handles a single card (no-op)", () => {
    const cards = [makeTask("a", COL_A)];

    const result = applyCardOverCard(cards, "a", "a");

    expect(result.map((c) => c.id)).toEqual(["a"]);
  });

  it("preserves all other card properties when changing column", () => {
    const cards = [makeTask("a", COL_A), makeTask("b", COL_C)];

    const result = applyCardOverCard(cards, "a", "b");
    const moved = result.find((c) => c.id === "a")!;

    expect(moved.name).toBe("Task a");
    expect(moved.content).toBeNull();
  });

  it("cross-column move: positions moved card before the over card", () => {
    const cards = [
      makeTask("x", COL_A),
      makeTask("a", COL_B),
      makeTask("b", COL_B),
      makeTask("c", COL_B),
    ];

    const result = applyCardOverCard(cards, "x", "b");

    const ids = result.map((c) => c.id);
    expect(ids.indexOf("x")).toBeLessThan(ids.indexOf("b"));
  });
});

describe("applyCardOverColumn", () => {
  it("changes the columnId of the active card", () => {
    const cards = [makeTask("a", COL_A), makeTask("b", COL_A)];

    const result = applyCardOverColumn(cards, "a", COL_C);

    const moved = result.find((c) => c.id === "a")!;
    expect(moved.columnId).toBe(COL_C);
  });

  it("does not change other cards' columns", () => {
    const cards = [makeTask("a", COL_A), makeTask("b", COL_A)];

    const result = applyCardOverColumn(cards, "a", COL_C);

    const other = result.find((c) => c.id === "b")!;
    expect(other.columnId).toBe(COL_A);
  });

  it("does not mutate the original cards array or its elements", () => {
    const cards = [makeTask("a", COL_A), makeTask("b", COL_B)];
    const snapshot = cards[0]?.columnId;

    applyCardOverColumn(cards, "a", COL_C);

    expect(cards[0]?.columnId).toBe(snapshot);
    expect(cards).toHaveLength(2);
  });

  it("returns the original array when activeId is not found", () => {
    const cards = [makeTask("a", COL_A)];

    const result = applyCardOverColumn(cards, "ghost", COL_C);

    expect(result).toBe(cards);
  });

  it("preserves all other card properties", () => {
    const cards = [makeTask("a", COL_A)];

    const result = applyCardOverColumn(cards, "a", COL_B);
    const moved = result[0];

    expect(moved?.name).toBe("Task a");
    expect(moved?.id).toBe("a");
  });

  it("keeps the card at the same index (no reordering)", () => {
    const cards = [
      makeTask("x", COL_A),
      makeTask("a", COL_A),
      makeTask("y", COL_A),
    ];

    const result = applyCardOverColumn(cards, "a", COL_C);

    expect(result.map((c) => c.id)).toEqual(["x", "a", "y"]);
  });
});
