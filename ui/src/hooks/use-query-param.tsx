import { useNavigate, useRouterState } from "@tanstack/react-router";

export function useQueryParams(name: string) {
  const navigate = useNavigate();
  const search = useRouterState({ select: (s) => s.location.searchStr });
  const searchParams = new URLSearchParams(search);
  const value = searchParams.get(name);

  const onClick = (next: string | null) => {
    navigate({
      search: (prev) => {
        const newParams = { ...(prev as Record<string, string>) };
        if (next) {
          newParams[name] = next;
        } else {
          delete newParams[name];
        }
        return newParams;
      },
      replace: true,
    });
  };

  return { param: value, onClick };
}
