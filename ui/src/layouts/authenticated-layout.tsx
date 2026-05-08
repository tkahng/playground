import { useAuthProvider } from "@/hooks/use-auth-provider";
import { Navigate, Outlet, useRouterState } from "@tanstack/react-router";
import { useEffect, useRef } from "react";

export default function AuthenticatedLayout() {
  const location = useRouterState({ select: (s) => s.location });
  const { user, checkAuth } = useAuthProvider();
  const isMounted = useRef(false);
  useEffect(() => {
    if (!isMounted.current) {
      isMounted.current = true;
      checkAuth().catch(() => {});
    }
  }, [checkAuth, location, user]);

  if (!user) {
    const redirectTo = encodeURIComponent(
      location.pathname + (location.searchStr || "")
    );
    return <Navigate to="/signin" search={{ redirect_to: redirectTo }} />;
  }

  return <Outlet />;
}
