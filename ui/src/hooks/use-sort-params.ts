import { ChangeEvent } from "react";
import { useSearchParams } from "react-router";

export const useSortParams = () => {
  const [searchParams, setSearchParams] = useSearchParams();

  const objectParams = Object.fromEntries(searchParams);

  const handleFieldChange =
    (field: string) => (event: ChangeEvent<HTMLInputElement>) => {
      const params = Object.fromEntries([...searchParams]);
      setSearchParams({
        ...params,
        page: "0",
        [field]: String(event.target.value),
      });
    };

  const handlePageChange = (_event: unknown, newPage: number) => {
    const params = Object.fromEntries([...searchParams]);
    setSearchParams({ ...params, page: String(newPage) });
  };

  const handleTableFieldChange = (fields: string[], values: string[]) => {
    const params = Object.fromEntries([...searchParams]);
    const obj = fields.reduce((accumulator, element, index) => {
      return { ...accumulator, [element]: values[index] };
    }, {});
    setSearchParams({ ...params, ...obj });
  };

  return {
    handleFieldChange,
    handlePageChange,
    handleTableFieldChange,
    searchParams: objectParams,
    setSearchParams,
    // setPaginationParams,
  };
};
