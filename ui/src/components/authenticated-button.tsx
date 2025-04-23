import { LinkDto } from "@/components/links";
import { RouteMap } from "@/components/route-map";
import { UserNav } from "@/components/user-nav";
import { useLocation } from "react-router";
import { NavbarLink } from "./link/nav-link";
const links: LinkDto[] = [
  {
    to: RouteMap.DASHBOARD,
    title: "Dashboard",
  },
  {
    to: RouteMap.SETTINGS,
    title: "Settings",
  },
];

export default function AuthenticatedButton() {
  // const { user } = useAuthProvider();
  const { pathname } = useLocation();
  console.log(pathname);
  // const isAdmin = user?.roles?.includes("superuser");
  return (
    <>
      {!pathname.startsWith(RouteMap.DASHBOARD) && (
        <NavbarLink title="Dashboard" to={RouteMap.DASHBOARD} />
      )}
      <UserNav links={links} />
    </>
  );
}
