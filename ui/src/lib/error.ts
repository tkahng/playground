import { components } from "@/schema";
import { ErrorModel } from "@/schema.types";

export class ApiError extends Error {
  constructor(detail: string);
  constructor(
    detail?: string,
    title?: string,
    status?: number,
    errors?: components["schemas"]["ErrorDetail"][] | null,
    type?: string
  );
  constructor(
    readonly detail?: string,
    readonly title?: string,
    readonly status?: number,
    readonly errors?: components["schemas"]["ErrorDetail"][] | null,
    readonly type: string = "ApiError"
  ) {
    super(detail);
  }

  static fromErrorModel(error: ErrorModel): ApiError {
    return new ApiError(
      error.detail,
      error.title,
      error.status,
      error.errors,
      error.type
    );
  }
}
export const GetError = <T extends Error>(error: T): ApiError => {
  if (error instanceof ApiError) {
    return error;
  }
  if (isErrorModel(error)) {
    return ApiError.fromErrorModel(error);
  }
  return new ApiError(error.message);
};

export const isErrorModel = (error: unknown): error is ErrorModel => {
  return (
    typeof error === "object" &&
    error !== null &&
    "$schema" in error &&
    typeof error.$schema === "string" &&
    error.$schema.includes("ErrorModel")
  );
};
