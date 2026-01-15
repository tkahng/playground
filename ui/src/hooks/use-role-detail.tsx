import { RoleDetailContext } from "@/context/role-detail-context";
import { useContext } from "react";

export const useRoleDetail = () => useContext(RoleDetailContext);
