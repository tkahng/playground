import { paths } from "@/schema";
import createClient from "openapi-fetch";

export const client = createClient<paths>({
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
