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
    navigate({
      search: (prev) => {
        const newParams = { ...(prev as Record<string, string>) };
        newParams["tab"] = next;
        return newParams;
      },
      replace: true,
    });
  };

  return { tab, onClick };
}
