import { paths } from "@/schema";
import createFetchClient from "openapi-fetch";
import createClient from "openapi-react-query";

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

export const $api = createClient(client);
