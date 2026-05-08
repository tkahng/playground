import {
  RouterContextProvider,
  createMemoryHistory,
  createRootRoute,
  createRouter,
  Outlet,
} from "@tanstack/react-router";
import { act, renderHook, waitFor } from "@testing-library/react";
import React from "react";
import { describe, expect, it } from "vitest";
import { useQueryParams } from "@/hooks/use-query-param";
import { useSearchParams } from "@/hooks/use-search-params";
import { useSortParams } from "@/hooks/use-sort-params";
import { useTabs } from "@/hooks/use-tabs";

function createWrapper(initialPath = "/") {
  const rootRoute = createRootRoute({ component: Outlet });
  const router = createRouter({
    routeTree: rootRoute.addChildren([]),
    history: createMemoryHistory({ initialEntries: [initialPath] }),
  });

  function Wrapper({ children }: { children: React.ReactNode }) {
    return (
      <RouterContextProvider router={router}>{children}</RouterContextProvider>
    );
  }

  return { Wrapper, router };
}

// ---------------------------------------------------------------------------
// useQueryParams
// ---------------------------------------------------------------------------

describe("useQueryParams", () => {
  it("returns null when param is absent", () => {
    const { Wrapper } = createWrapper("/");
    const { result } = renderHook(() => useQueryParams("filter"), {
      wrapper: Wrapper,
    });
    expect(result.current.param).toBeNull();
  });

  it("reads existing param from URL", () => {
    const { Wrapper } = createWrapper("/?filter=active");
    const { result } = renderHook(() => useQueryParams("filter"), {
      wrapper: Wrapper,
    });
    expect(result.current.param).toBe("active");
  });

  it("sets a param via onClick", async () => {
    const { Wrapper } = createWrapper("/");
    const { result } = renderHook(() => useQueryParams("color"), {
      wrapper: Wrapper,
    });

    act(() => {
      result.current.onClick("blue");
    });

    await waitFor(() => {
      expect(result.current.param).toBe("blue");
    });
  });

  it("removes a param when onClick is called with null", async () => {
    const { Wrapper } = createWrapper("/?color=blue");
    const { result } = renderHook(() => useQueryParams("color"), {
      wrapper: Wrapper,
    });

    act(() => {
      result.current.onClick(null);
    });

    await waitFor(() => {
      expect(result.current.param).toBeNull();
    });
  });

  it("preserves unrelated params when setting a new value", async () => {
    const { Wrapper, router } = createWrapper("/?sort=asc&page=2");
    const { result } = renderHook(() => useQueryParams("filter"), {
      wrapper: Wrapper,
    });

    act(() => {
      result.current.onClick("active");
    });

    await waitFor(() => {
      const search = new URLSearchParams(router.state.location.search);
      expect(search.get("filter")).toBe("active");
      expect(search.get("sort")).toBe("asc");
      expect(search.get("page")).toBe("2");
    });
  });

  it("preserves unrelated params when removing a value", async () => {
    const { Wrapper, router } = createWrapper("/?filter=old&sort=asc");
    const { result } = renderHook(() => useQueryParams("filter"), {
      wrapper: Wrapper,
    });

    act(() => {
      result.current.onClick(null);
    });

    await waitFor(() => {
      const search = new URLSearchParams(router.state.location.search);
      expect(search.get("filter")).toBeNull();
      expect(search.get("sort")).toBe("asc");
    });
  });
});

// ---------------------------------------------------------------------------
// useTabs
// ---------------------------------------------------------------------------

describe("useTabs", () => {
  const ALLOWED = ["overview", "settings", "billing"] as const;

  it("returns defaultValue when no tab param in URL", () => {
    const { Wrapper } = createWrapper("/");
    const { result } = renderHook(() => useTabs("overview", ALLOWED), {
      wrapper: Wrapper,
    });
    expect(result.current.tab).toBe("overview");
  });

  it("returns tab from URL when the value is in the allowed list", () => {
    const { Wrapper } = createWrapper("/?tab=settings");
    const { result } = renderHook(() => useTabs("overview", ALLOWED), {
      wrapper: Wrapper,
    });
    expect(result.current.tab).toBe("settings");
  });

  it("falls back to defaultValue when tab value is not allowed", () => {
    const { Wrapper } = createWrapper("/?tab=hackedvalue");
    const { result } = renderHook(() => useTabs("overview", ALLOWED), {
      wrapper: Wrapper,
    });
    expect(result.current.tab).toBe("overview");
  });

  it("updates the URL tab param via onClick", async () => {
    const { Wrapper } = createWrapper("/");
    const { result } = renderHook(() => useTabs("overview", ALLOWED), {
      wrapper: Wrapper,
    });

    act(() => {
      result.current.onClick("billing");
    });

    await waitFor(() => {
      expect(result.current.tab).toBe("billing");
    });
  });

  it("preserves other params when switching tabs", async () => {
    const { Wrapper, router } = createWrapper("/?sort=desc&tab=overview");
    const { result } = renderHook(() => useTabs("overview", ALLOWED), {
      wrapper: Wrapper,
    });

    act(() => {
      result.current.onClick("settings");
    });

    await waitFor(() => {
      const search = new URLSearchParams(router.state.location.search);
      expect(search.get("tab")).toBe("settings");
      expect(search.get("sort")).toBe("desc");
    });
  });
});

// ---------------------------------------------------------------------------
// useSortParams
// ---------------------------------------------------------------------------

describe("useSortParams", () => {
  it("returns empty searchParams object when URL has no params", () => {
    const { Wrapper } = createWrapper("/");
    const { result } = renderHook(() => useSortParams(), { wrapper: Wrapper });
    expect(result.current.searchParams).toEqual({});
  });

  it("reflects existing URL params in searchParams", () => {
    const { Wrapper } = createWrapper("/?page=2&sort=name");
    const { result } = renderHook(() => useSortParams(), { wrapper: Wrapper });
    expect(result.current.searchParams).toEqual({ page: "2", sort: "name" });
  });

  it("handlePageChange updates page, keeps other params", async () => {
    const { Wrapper, router } = createWrapper("/?page=0&sort=name");
    const { result } = renderHook(() => useSortParams(), { wrapper: Wrapper });

    act(() => {
      result.current.handlePageChange(null, 3);
    });

    await waitFor(() => {
      const search = new URLSearchParams(router.state.location.search);
      expect(search.get("page")).toBe("3");
      expect(search.get("sort")).toBe("name");
    });
  });

  it("handleFieldChange updates field value and resets page to 0", async () => {
    const { Wrapper, router } = createWrapper("/?page=5&q=hello");
    const { result } = renderHook(() => useSortParams(), { wrapper: Wrapper });

    act(() => {
      result.current.handleFieldChange("q")({
        target: { value: "world" },
      } as React.ChangeEvent<HTMLInputElement>);
    });

    await waitFor(() => {
      const search = new URLSearchParams(router.state.location.search);
      expect(search.get("q")).toBe("world");
      expect(search.get("page")).toBe("0");
    });
  });

  it("handleTableFieldChange updates multiple fields simultaneously", async () => {
    const { Wrapper, router } = createWrapper("/?page=1");
    const { result } = renderHook(() => useSortParams(), { wrapper: Wrapper });

    act(() => {
      result.current.handleTableFieldChange(
        ["sort_by", "sort_order"],
        ["name", "asc"]
      );
    });

    await waitFor(() => {
      const search = new URLSearchParams(router.state.location.search);
      expect(search.get("sort_by")).toBe("name");
      expect(search.get("sort_order")).toBe("asc");
      expect(search.get("page")).toBe("1");
    });
  });

  it("setSearchParams replaces the full search string", async () => {
    const { Wrapper, router } = createWrapper("/?old=value");
    const { result } = renderHook(() => useSortParams(), { wrapper: Wrapper });

    act(() => {
      result.current.setSearchParams({ new: "param" });
    });

    await waitFor(() => {
      const search = new URLSearchParams(router.state.location.search);
      expect(search.get("new")).toBe("param");
      expect(search.get("old")).toBeNull();
    });
  });
});

// ---------------------------------------------------------------------------
// useSearchParams
// ---------------------------------------------------------------------------

describe("useSearchParams", () => {
  it("returns URLSearchParams reflecting the current URL", () => {
    const { Wrapper } = createWrapper("/?foo=bar&baz=qux");
    const { result } = renderHook(() => useSearchParams(), {
      wrapper: Wrapper,
    });
    const [params] = result.current;
    expect(params.get("foo")).toBe("bar");
    expect(params.get("baz")).toBe("qux");
  });

  it("returns empty URLSearchParams when URL has no query string", () => {
    const { Wrapper } = createWrapper("/");
    const { result } = renderHook(() => useSearchParams(), {
      wrapper: Wrapper,
    });
    const [params] = result.current;
    expect(params.toString()).toBe("");
  });

  it("setSearchParams with an object replaces the entire search", async () => {
    const { Wrapper, router } = createWrapper("/?old=val");
    const { result } = renderHook(() => useSearchParams(), {
      wrapper: Wrapper,
    });
    const [, setParams] = result.current;

    act(() => {
      setParams({ newKey: "newVal" });
    });

    await waitFor(() => {
      const search = new URLSearchParams(router.state.location.search);
      expect(search.get("newKey")).toBe("newVal");
      expect(search.get("old")).toBeNull();
    });
  });

  it("setSearchParams with a function receives current params and applies update", async () => {
    const { Wrapper, router } = createWrapper("/?count=3");
    const { result } = renderHook(() => useSearchParams(), {
      wrapper: Wrapper,
    });
    const [, setParams] = result.current;

    act(() => {
      setParams((prev) => {
        prev.set("count", String(Number(prev.get("count")) + 1));
        return prev;
      });
    });

    await waitFor(() => {
      const search = new URLSearchParams(router.state.location.search);
      expect(search.get("count")).toBe("4");
    });
  });

  it("setSearchParams with a function preserves unrelated params", async () => {
    const { Wrapper, router } = createWrapper("/?page=2&filter=active");
    const { result } = renderHook(() => useSearchParams(), {
      wrapper: Wrapper,
    });
    const [, setParams] = result.current;

    act(() => {
      setParams((prev) => {
        prev.set("filter", "inactive");
        return prev;
      });
    });

    await waitFor(() => {
      const search = new URLSearchParams(router.state.location.search);
      expect(search.get("filter")).toBe("inactive");
      expect(search.get("page")).toBe("2");
    });
  });
});
