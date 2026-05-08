import { useAuthProvider } from "@/hooks/use-auth-provider";
import { createLedgerWallet } from "@/lib/api";
import { Navigate, Outlet, useRouterState } from "@tanstack/react-router";
import { useEffect, useRef } from "react";

export default function AuthenticatedLayoutOutlet() {
  const location = useRouterState({ select: (s) => s.location });
  const { pathname } = location;
  const { user, checkAuth, getOrRefreshToken } = useAuthProvider();
  const isMounted = useRef(false);
  useEffect(() => {
    if (!isMounted.current) {
      isMounted.current = true;
      getOrRefreshToken()
        .then((u) => {
          createLedgerWallet(u.tokens.access_token).catch(() => {});
        })
        .catch(() => {});
    }
  }, [checkAuth, getOrRefreshToken, location, user]);

  if (!user) {
    const redirectTo = pathname + (location.searchStr || "");
    if (pathname.startsWith("/team-invitation")) {
      return <Navigate to="/signup" search={{ redirect_to: redirectTo }} />;
    }
    return <Navigate to="/signin" search={{ redirect_to: redirectTo }} />;
  }

  return <Outlet />;
}
