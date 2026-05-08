import { useNavigate, useRouterState } from "@tanstack/react-router";

type SetSearchParams = (
  updater:
    | Record<string, string>
    | ((prev: URLSearchParams) => URLSearchParams),
  options?: { replace?: boolean }
) => void;

export function useSearchParams(): [URLSearchParams, SetSearchParams] {
  const navigate = useNavigate();
  const rawSearch = useRouterState({ select: (s) => s.location.searchStr });
  const searchParams = new URLSearchParams(rawSearch);

  const setSearchParams: SetSearchParams = (updater, options) => {
    const replace = options?.replace ?? true;
    const next =
      typeof updater === "function"
        ? updater(new URLSearchParams(rawSearch))
        : new URLSearchParams(updater);
    // @ts-expect-error – search schema not declared per-route; runtime is correct
    navigate({ search: () => Object.fromEntries(next.entries()), replace });
  };

  return [searchParams, setSearchParams];
}
