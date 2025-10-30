import { useSearchParams } from "react-router";

export const useTabs = (defaultValue: string) => {
  const [searchParams, setSearchParams] = useSearchParams();
  const tab = searchParams.get("tab") ?? defaultValue;

  const onClick = (value: string) => {
    setSearchParams(
      (prev) => {
        prev.set("tab", value);
        return prev;
      },
      {
        preventScrollReset: true,
      }
    );
  };
  return { tab, onClick };
};

export function useTabs2<T extends string = string>(
  defaultValue: T,
  allowed: readonly T[]
) {
  const [searchParams, setSearchParams] = useSearchParams();
  const value = searchParams.get("tab");

  // validate query param
  const tab = allowed.includes(value as T) ? (value as T) : defaultValue;

  const onClick = (next: T) => {
    setSearchParams(
      (prev) => {
        prev.set("tab", next);
        return prev;
      },
      {
        preventScrollReset: true,
      }
    );
  };

  return { tab, onClick };
}
