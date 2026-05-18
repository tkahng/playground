import { describe, expect, it } from "vitest";
import { buildItems, findContainer } from "../kanban-board.utils";

const COL_A = "11111111-1111-1111-1111-111111111111";
const COL_B = "22222222-2222-2222-2222-222222222222";
const COL_C = "33333333-3333-3333-3333-333333333333";

// ---------------------------------------------------------------------------
// buildItems
// ---------------------------------------------------------------------------

describe("buildItems", () => {
  it("groups task IDs by workflowStatusId", () => {
    const cards = [
      { id: "t1", workflowStatusId: COL_A },
      { id: "t2", workflowStatusId: COL_A },
      { id: "t3", workflowStatusId: COL_B },
    ];
    const result = buildItems(cards, [COL_A, COL_B, COL_C]);
    expect(result[COL_A]).toEqual(["t1", "t2"]);
    expect(result[COL_B]).toEqual(["t3"]);
    expect(result[COL_C]).toEqual([]); // pre-initialised even though empty
  });

  it("pre-initialises empty columns so they are valid drag targets", () => {
    const result = buildItems([], [COL_A, COL_B]);
    expect(result[COL_A]).toEqual([]);
    expect(result[COL_B]).toEqual([]);
    expect(Object.keys(result)).toHaveLength(2);
  });

  it("skips tasks with empty workflowStatusId", () => {
    const cards = [
      { id: "t1", workflowStatusId: COL_A },
      { id: "t2", workflowStatusId: "" },
    ];
    const result = buildItems(cards, [COL_A]);
    expect(result[COL_A]).toEqual(["t1"]);
  });

  it("preserves insertion order within each column", () => {
    const cards = [
      { id: "t3", workflowStatusId: COL_A },
      { id: "t1", workflowStatusId: COL_A },
      { id: "t2", workflowStatusId: COL_A },
    ];
    expect(buildItems(cards, [COL_A])[COL_A]).toEqual(["t3", "t1", "t2"]);
  });

  it("returns only column keys when no tasks provided", () => {
    expect(buildItems([], [COL_A, COL_C])).toEqual({
      [COL_A]: [],
      [COL_C]: [],
    });
  });

  it("handles numeric ids by coercing to string", () => {
    const cards = [{ id: 42, workflowStatusId: COL_A }];
    expect(buildItems(cards, [COL_A])[COL_A]).toEqual(["42"]);
  });
});

// ---------------------------------------------------------------------------
// findContainer
// ---------------------------------------------------------------------------

describe("findContainer", () => {
  const items = {
    [COL_A]: ["t1", "t2"],
    [COL_B]: ["t3"],
    [COL_C]: [],
  };

  it("returns the column id when given a column id directly", () => {
    expect(findContainer(items, COL_A)).toBe(COL_A);
  });

  it("finds the container for a task id", () => {
    expect(findContainer(items, "t1")).toBe(COL_A);
    expect(findContainer(items, "t3")).toBe(COL_B);
  });

  it("returns undefined for an unknown id", () => {
    expect(findContainer(items, "ghost")).toBeUndefined();
  });

  it("returns the column id for an empty column (column is still a key)", () => {
    expect(findContainer(items, COL_C)).toBe(COL_C);
  });
});
