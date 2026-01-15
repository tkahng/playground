import { RoleWithPermissions } from "@/schema.types";
import { createContext } from "react";

export const RoleDetailContext = createContext<RoleWithPermissions | null>(
  null
);
