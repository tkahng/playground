import { LinkProps } from "./landing-links";
import { RouteMap } from "./route-map";
import { UserNav } from "./user-nav";
const links: LinkProps[] = [
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
  // const isAdmin = user?.roles?.includes("superuser");
  return (
    <>
      {/* <NavLink title="Dashboard" to={RouteMap.DASHBOARD_HOME} /> */}
      {/* {isAdmin && <NavLink title="Admin" to={RouteMap.ADMIN_DASHBOARD_HOME} />} */}
      {/* <AccountDropdown /> */}
      <UserNav links={links} />
      {/* <NavUser
        user={{
          avatar: user?.user.image || "",
          name: user?.user.name || "",
          email: user?.user.email || "",
        }}
      /> */}
    </>
  );
}
