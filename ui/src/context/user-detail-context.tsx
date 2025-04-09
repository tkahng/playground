import { UserDetail } from "@/schema.types";
import { createContext } from "react";

export const UserDetailContext = createContext<UserDetail | null>(null);
