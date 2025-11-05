import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useEffect, useRef } from "react";
import { Navigate, Outlet, useLocation } from "react-router";

export default function AuthenticatedLayout() {
  const location = useLocation();
  const { user, checkAuth } = useAuthProvider();
  const isMounted = useRef(false);
  useEffect(() => {
    if (!isMounted.current) {
      isMounted.current = true;
      checkAuth()
        .then(() => {
          // isMounted.current = false;
        })
        .catch(() => {
          // isMounted.current = false;
        });
    }
  }, [checkAuth, location, user]);

  if (!user) {
    return (
      <Navigate
        to={{
          pathname: "/signin",
          search:
            "redirect_to=" +
            encodeURIComponent(location.pathname + location.search),
        }}
      />
    );
  }

  return <Outlet context={{ user }} />;
}
