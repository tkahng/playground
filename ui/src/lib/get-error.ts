import { ErrorModel } from "@/schema.types";

export const GetError = <T>(error: T | ErrorModel): ErrorModel | null => {
  if (typeof error === "object" && error !== null) {
    if (
      "$schema" in error &&
      typeof error.$schema === "string" &&
      error.$schema.includes("ErrorModel")
    ) {
      return error as ErrorModel;
    }
    if (error instanceof Error) {
      return {
        type: error.name,
        detail: error.message,
      };
    }
  }
  return null;
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
