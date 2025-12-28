import { useCallback, useMemo } from "react";
import { useSearchParams } from "react-router-dom";

const ITEM_PARAM = "item";

export function useItemDialog() {
  const [searchParams, setSearchParams] = useSearchParams();

  const selectedItemId = searchParams.get(ITEM_PARAM);

  const openItem = useCallback(
    (id: string) => {
      const next = new URLSearchParams(searchParams);
      next.set(ITEM_PARAM, id);
      setSearchParams(next, { replace: false }); // push into history
    },
    [searchParams, setSearchParams]
  );

  const closeDialog = useCallback(() => {
    const next = new URLSearchParams(searchParams);
    next.delete(ITEM_PARAM);
    setSearchParams(next, { replace: false });
  }, [searchParams, setSearchParams]);

  const isOpen = useMemo(() => selectedItemId !== null, [selectedItemId]);

  return { isOpen, selectedItemId, openItem, closeDialog };
}
