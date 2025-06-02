import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useEffect, useRef } from "react";
import {
  createSearchParams,
  Navigate,
  Outlet,
  useLocation,
} from "react-router";

export default function AuthenticatedLayoutBase() {
  const location = useLocation();
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
  }, [location, checkAuth, user]);

  if (!user) {
    return (
      <Navigate
        to={{
          pathname: "/signin",
          search: createSearchParams({
            redirect_to: location.pathname + location.search,
          }).toString(),
        }}
      />
    );
  }

  return <Outlet context={{ user }} />;
}
