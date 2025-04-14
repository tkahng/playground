import { useAuthProvider } from "@/hooks/use-auth-provider";
import AccountDropdown from "./account-dropdown";
import { NavLink } from "./link/nav-link";
import { RouteMap } from "./route-map";

export default function AuthenticatedButton() {
  const { user } = useAuthProvider();
  const isAdmin = user?.roles?.includes("superuser");
  return (
    <>
      <NavLink title="Dashboard" to={RouteMap.DASHBOARD_HOME} />
      {isAdmin && <NavLink title="Admin" to={RouteMap.ADMIN_DASHBOARD_HOME} />}
      <AccountDropdown />
    </>
  );
}
