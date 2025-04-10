import { UserDetailContext } from "@/context/user-detail-context";
import { useContext } from "react";

export const useUserDetail = () => useContext(UserDetailContext);
