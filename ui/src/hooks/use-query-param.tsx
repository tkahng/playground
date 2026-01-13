import { useSearchParams } from "react-router";

export function useQueryParams(name: string) {
  const [searchParams, setSearchParams] = useSearchParams();
  const value = searchParams.get(name);

  // validate query param
  const param = value;

  const onClick = (next: string | null) => {
    setSearchParams(
      (prev) => {
        if (next) {
          prev.set(name, next);
        } else {
          prev.delete(name);
        }
        return prev;
      },
      {
        preventScrollReset: true,
      },
    );
  };

  return { param, onClick };
}
