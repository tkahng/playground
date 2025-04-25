import { LinkDto } from "@/components/links";
import { RouteMap } from "@/components/route-map";
import { UserNav } from "@/components/user-nav";
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
  // const { pathname } = useLocation();
  // const isAdmin = user?.roles?.includes("superuser");
  return (
    <>
      {/* {![RouteMap.DASHBOARD, RouteMap.PROTECTED].includes(pathname) && (
        <NavbarLink title="Dashboard" to={RouteMap.DASHBOARD} />
      )} */}
      <UserNav links={links} />
    </>
  );
}
