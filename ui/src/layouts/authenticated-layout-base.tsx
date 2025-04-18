import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useEffect } from "react";
import { Navigate, useLocation, useOutlet } from "react-router";

export default function AuthenticatedLayoutBase() {
  const location = useLocation();
  const { checkAuth, user } = useAuthProvider();
  const outlet = useOutlet();
  useEffect(() => {
    checkAuth();
  }, [location]);

  if (!user) {
    return <Navigate to="/signin" />;
  }
  return (
    <>
      {outlet}
    </>
  );
}
