import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useEffect } from "react";
import { useLocation } from "react-router";

const AuthVerify = () => {
  const location = useLocation();
  const { checkAuth } = useAuthProvider();
  useEffect(() => {
    checkAuth();
  }, [location]);
  return null;
};

export default AuthVerify;
