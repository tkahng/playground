import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useTeam } from "@/hooks/use-team";
import {
  createSearchParams,
  Navigate,
  Outlet,
  useLocation,
} from "react-router";

export default function TeamsLayoutBase() {
  const location = useLocation();
  const { user } = useAuthProvider();
  const { team } = useTeam();
  const isNotUsersTeam = team?.member?.user_id !== user?.user.id;
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

  if (!team) {
    return (
      <Navigate
        to={{
          pathname: "/team-select",
          search: createSearchParams({
            redirect_to: location.pathname + location.search,
          }).toString(),
        }}
      />
    );
  }
  if (isNotUsersTeam) {
    return (
      <Navigate
        to={{
          pathname: "/team-select",
          search: createSearchParams({
            redirect_to: location.pathname + location.search,
          }).toString(),
        }}
      />
    );
  }
  return <Outlet context={{ user, team }} />;
}
