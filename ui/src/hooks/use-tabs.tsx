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
