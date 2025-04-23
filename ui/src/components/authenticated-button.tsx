import { LinkDto } from "@/components/landing-links";
import { RouteMap } from "@/components/route-map";
import { UserNav } from "@/components/user-nav";
import { useLocation } from "react-router";
import { NavLink } from "./link/nav-link";
const links: LinkDto[] = [
  {
    to: RouteMap.DASHBOARD_HOME,
    title: "Dashboard",
  },
  {
    to: RouteMap.ACCOUNT_SETTINGS,
    title: "Settings",
  },
];

export default function AuthenticatedButton() {
  // const { user } = useAuthProvider();
  const { pathname } = useLocation();
  // const isAdmin = user?.roles?.includes("superuser");
  return (
    <>
      {!pathname.startsWith(RouteMap.DASHBOARD_HOME) && (
        <NavLink title="Dashboard" to={RouteMap.DASHBOARD_HOME} />
      )}
      <UserNav links={links} />
    </>
  );
}
