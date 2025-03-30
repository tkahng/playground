import { useEffect, useState } from "react";
import { useAuthProvider } from "./use-auth-provider";

export const useCheckAuth = () => {
  const auth = useAuthProvider();
  const [loading, setLoading] = useState(true);
  const [isAuthenticated, setIsAuthenticated] = useState(false);

  useEffect(() => {
    const verifyAuth = async () => {
      //Simulate async request
      try {
        await auth.checkAuth();
        setIsAuthenticated(true);
        setLoading(false);
      } catch (error) {
        setIsAuthenticated(false);
        setLoading(false);
      }
    };

    verifyAuth();
  }, []);

  return { loading, isAuthenticated };
};
