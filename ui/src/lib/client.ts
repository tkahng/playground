import { paths } from "@/schema";
import createFetchClient from "openapi-fetch";

export const client = createFetchClient<paths>({
  baseUrl: "/",
  querySerializer: {
    array: {
      style: "form", // "form" (default) | "spaceDelimited" | "pipeDelimited"
      explode: false,
    },
    object: {
      style: "deepObject", // "form" | "deepObject" (default)
      explode: true,
    },
  },
});
