import { components } from "@/schema";
import { ErrorModel } from "@/schema.types";

export class ApiError extends Error {
  constructor(detail: string);
  constructor(
    detail?: string,
    title?: string,
    status?: number,
    errors?: components["schemas"]["ErrorDetail"][] | null
  );
  constructor(
    readonly detail?: string,
    readonly title?: string,
    readonly status?: number,
    readonly errors?: components["schemas"]["ErrorDetail"][] | null
  ) {
    super(detail);
  }

  static fromErrorModel(error: ErrorModel): ApiError {
    return new ApiError(error.detail, error.title, error.status, error.errors);
  }
}
