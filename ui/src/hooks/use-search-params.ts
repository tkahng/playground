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
    if (typeof updater === "function") {
      const current = new URLSearchParams(rawSearch);
      const next = updater(current);
      navigate({
        search: () => Object.fromEntries(next.entries()),
        replace,
      });
    } else {
      navigate({ search: () => updater, replace });
    }
  };

  return [searchParams, setSearchParams];
}
