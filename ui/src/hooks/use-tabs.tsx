import { useNavigate, useRouterState } from "@tanstack/react-router";

export function useTabs<T extends string = string>(
  defaultValue: T,
  allowed: readonly T[]
) {
  const navigate = useNavigate();
  const search = useRouterState({ select: (s) => s.location.searchStr });
  const searchParams = new URLSearchParams(search);
  const value = searchParams.get("tab");

  const tab = allowed.includes(value as T) ? (value as T) : defaultValue;

  const onClick = (next: T) => {
    const params = new URLSearchParams(search);
    params.set("tab", next);
    // @ts-expect-error – search schema not declared per-route; runtime is correct
    navigate({ search: () => Object.fromEntries(params.entries()), replace: true });
  };

  return { tab, onClick };
}
