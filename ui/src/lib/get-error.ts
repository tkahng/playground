import { ErrorModel } from "@/schema.types";

export const GetError = <T extends Error>(
  error: T | ErrorModel
): ErrorModel => {
  if (isErrorModel(error)) {
    return error;
  }
  return {
    $schema: "ErrorModel",
    title: error.name,
    status: 500,
    detail: error.message,
    errors: [],
    type: "ErrorModel",
  };
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
