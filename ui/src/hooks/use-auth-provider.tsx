import { AuthContext } from "@/context/auth-context";
import { useContext } from "react";

export const useAuthProvider = () => useContext(AuthContext);
