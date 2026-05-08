import { useNavigate, useRouterState } from "@tanstack/react-router";
import { ChangeEvent } from "react";

export const useSortParams = () => {
  const navigate = useNavigate();
  const search = useRouterState({ select: (s) => s.location.searchStr });
  const searchParams = new URLSearchParams(search);
  const objectParams = Object.fromEntries(searchParams);

  const setSearchParams = (params: Record<string, string>) => {
    navigate({ search: () => params, replace: true });
  };

  const handleFieldChange =
    (field: string) => (event: ChangeEvent<HTMLInputElement>) => {
      setSearchParams({
        ...objectParams,
        page: "0",
        [field]: String(event.target.value),
      });
    };

  const handlePageChange = (_event: unknown, newPage: number) => {
    setSearchParams({ ...objectParams, page: String(newPage) });
  };

  const handleTableFieldChange = (fields: string[], values: string[]) => {
    const obj = fields.reduce<Record<string, string>>(
      (accumulator, element, index) => {
        return { ...accumulator, [element]: values[index] ?? "" };
      },
      {}
    );
    setSearchParams({ ...objectParams, ...obj });
  };

  return {
    handleFieldChange,
    handlePageChange,
    handleTableFieldChange,
    searchParams: objectParams,
    setSearchParams,
  };
};
