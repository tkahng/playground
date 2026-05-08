import { useNavigate, useRouterState } from "@tanstack/react-router";

export function useQueryParams(name: string) {
  const navigate = useNavigate();
  const search = useRouterState({ select: (s) => s.location.searchStr });
  const searchParams = new URLSearchParams(search);
  const value = searchParams.get(name);

  const onClick = (next: string | null) => {
    const params = new URLSearchParams(search);
    if (next) {
      params.set(name, next);
    } else {
      params.delete(name);
    }
    // @ts-expect-error – search schema not declared per-route; runtime is correct
    navigate({ search: () => Object.fromEntries(params.entries()), replace: true });
  };

  return { param: value, onClick };
}
