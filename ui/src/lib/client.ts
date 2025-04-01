import { paths } from "@/schema";
import createClient from "openapi-fetch";

export const client = createClient<paths>({
  baseUrl: "/",
});
