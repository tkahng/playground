import { UserDetailWithRoles } from "@/schema.types";
import { createContext } from "react";

export const UserDetailContext = createContext<UserDetailWithRoles | null>(
  null
);
