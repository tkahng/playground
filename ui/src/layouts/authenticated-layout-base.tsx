import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useEffect } from "react";
import { Navigate, Outlet, useLocation } from "react-router";

export default function AuthenticatedLayoutBase() {
  const location = useLocation();
  const { checkAuth, user } = useAuthProvider();
  useEffect(() => {
    checkAuth();
  }, [location]);

  if (!user) {
    return <Navigate to="/signin" />;
  }
  return (
    <>
      <Outlet />
    </>
  );
}
