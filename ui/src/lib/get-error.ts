import { ErrorModel } from "@/schema.types";

export const GetError = <T>(error: T | ErrorModel) => {
  if (typeof error === "object" && error !== null && "$schema" in error) {
    if (
      typeof error.$schema === "string" &&
      error.$schema.includes("ErrorModel")
    ) {
      return error as ErrorModel;
    }
  }
  return null;
};
