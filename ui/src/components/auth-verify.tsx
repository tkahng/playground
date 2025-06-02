import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useEffect, useRef } from "react";
import { useLocation, useNavigate } from "react-router";

const AuthVerify = () => {
  const location = useLocation();
  const navigate = useNavigate();
  const { user, checkAuth } = useAuthProvider();
  const isMounted = useRef(false);
  useEffect(() => {
    if (!isMounted.current) {
      isMounted.current = true;
      checkAuth()
        .then(() => {
          isMounted.current = false;
        })
        .catch(() => {
          isMounted.current = false;
        });
    }
  }, [location, checkAuth]);
  if (!user) {
    navigate("/signin");
  }
  return null;
};

export default AuthVerify;
