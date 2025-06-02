import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useEffect, useRef } from "react";
import {
  createSearchParams,
  Navigate,
  Outlet,
  useLocation,
  useParams,
} from "react-router";

export default function TeamLayoutBase() {
  const location = useLocation();
  const { teamSlug } = useParams<{ teamSlug: string }>();
  const { user } = useAuthProvider();
  const isMounted = useRef(false);
  useEffect(() => {
    if (!isMounted.current) {
      isMounted.current = true;
      console.log("Checking auth for team layout");
      console.log("Team slug:", teamSlug);
    }
  }, [location, teamSlug, user]);

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
