import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useEffect, useRef } from "react";
import {
  createSearchParams,
  Navigate,
  Outlet,
  useLocation,
} from "react-router";

export default function AuthenticatedLayoutOutlet() {
  const location = useLocation();
  const { pathname } = location;
  const { user, checkAuth } = useAuthProvider();
  // const { team, teamMember } = useTeam();
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
    if (pathname.startsWith("/team-invitation")) {
      return (
        <Navigate
          to={{
            pathname: "/signup",
            search: createSearchParams({
              redirect_to: location.pathname + location.search,
            }).toString(),
          }}
        />
      );
    }
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
  // if (user) {
  //   // if (pathname === "/") {
  //   //   if (team && teamMember?.user_id === user.user.id) {
  //   //     return <Navigate to={`/teams/${team.slug}/dashboard`} />;
  //   //   } else {
  //   //     return <Navigate to="/teams" />;
  //   //   }
  //   // }
  //   // if (pathname === "/dashboard") {
  //   //   if (team && teamMember?.user_id === user.user.id) {
  //   //     return <Navigate to={`/teams/${team.slug}/dashboard`} />;
  //   //   }
  //   // }
  // }
  return <Outlet context={{ user }} />;
}
