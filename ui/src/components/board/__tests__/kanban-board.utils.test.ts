import { describe, expect, it } from "vitest";
import type { Task } from "../task-card";
import { applyCardOverCard, applyCardOverColumn } from "../kanban-board.utils";

function makeTask(id: string, columnId: "todo" | "in_progress" | "done", rank = 0): Task {
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
      makeTask("a", "todo"),
      makeTask("b", "todo"),
      makeTask("c", "todo"),
    ];

    const result = applyCardOverCard(cards, "a", "c");

    expect(result.map((c) => c.id)).toEqual(["b", "c", "a"]);
    expect(result.every((c) => c.columnId === "todo")).toBe(true);
  });

  it("moves card to the target column when columns differ", () => {
    const cards = [
      makeTask("a", "todo"),
      makeTask("b", "in_progress"),
      makeTask("c", "in_progress"),
    ];

    const result = applyCardOverCard(cards, "a", "b");

    const moved = result.find((c) => c.id === "a")!;
    expect(moved.columnId).toBe("in_progress");
  });

  it("does not mutate the original cards array or its elements", () => {
    const cards = [makeTask("a", "todo"), makeTask("b", "in_progress")];
    const originalColumnId = cards[0].columnId;

    applyCardOverCard(cards, "a", "b");

    expect(cards[0].columnId).toBe(originalColumnId);
    expect(cards).toHaveLength(2);
  });

  it("returns the original array when activeId is not found", () => {
    const cards = [makeTask("a", "todo"), makeTask("b", "todo")];

    const result = applyCardOverCard(cards, "nonexistent", "b");

    expect(result).toBe(cards);
  });

  it("handles a single card (no-op)", () => {
    const cards = [makeTask("a", "todo")];

    const result = applyCardOverCard(cards, "a", "a");

    expect(result.map((c) => c.id)).toEqual(["a"]);
  });

  it("preserves all other card properties when changing column", () => {
    const cards = [
      makeTask("a", "todo"),
      makeTask("b", "done"),
    ];

    const result = applyCardOverCard(cards, "a", "b");
    const moved = result.find((c) => c.id === "a")!;

    expect(moved.name).toBe("Task a");
    expect(moved.content).toBeNull();
  });

  it("cross-column move: positions moved card before the over card", () => {
    const cards = [
      makeTask("x", "todo"),
      makeTask("a", "in_progress"),
      makeTask("b", "in_progress"),
      makeTask("c", "in_progress"),
    ];

    const result = applyCardOverCard(cards, "x", "b");

    const ids = result.map((c) => c.id);
    expect(ids.indexOf("x")).toBeLessThan(ids.indexOf("b"));
  });
});

describe("applyCardOverColumn", () => {
  it("changes the columnId of the active card", () => {
    const cards = [makeTask("a", "todo"), makeTask("b", "todo")];

    const result = applyCardOverColumn(cards, "a", "done");

    const moved = result.find((c) => c.id === "a")!;
    expect(moved.columnId).toBe("done");
  });

  it("does not change other cards' columns", () => {
    const cards = [makeTask("a", "todo"), makeTask("b", "todo")];

    const result = applyCardOverColumn(cards, "a", "done");

    const other = result.find((c) => c.id === "b")!;
    expect(other.columnId).toBe("todo");
  });

  it("does not mutate the original cards array or its elements", () => {
    const cards = [makeTask("a", "todo"), makeTask("b", "in_progress")];
    const snapshot = cards[0].columnId;

    applyCardOverColumn(cards, "a", "done");

    expect(cards[0].columnId).toBe(snapshot);
    expect(cards).toHaveLength(2);
  });

  it("returns the original array when activeId is not found", () => {
    const cards = [makeTask("a", "todo")];

    const result = applyCardOverColumn(cards, "ghost", "done");

    expect(result).toBe(cards);
  });

  it("preserves all other card properties", () => {
    const cards = [makeTask("a", "todo")];

    const result = applyCardOverColumn(cards, "a", "in_progress");
    const moved = result[0];

    expect(moved.name).toBe("Task a");
    expect(moved.id).toBe("a");
  });

  it("keeps the card at the same index (no reordering)", () => {
    const cards = [
      makeTask("x", "todo"),
      makeTask("a", "todo"),
      makeTask("y", "todo"),
    ];

    const result = applyCardOverColumn(cards, "a", "done");

    expect(result.map((c) => c.id)).toEqual(["x", "a", "y"]);
  });
});
