import { NavLink } from "@/components/link/nav-link";
import { RouteMap } from "@/components/route-map";

export default function NonAuthenticatedButton() {
  return (
    <>
      <NavLink title="Sign In" to={RouteMap.SIGNIN} />
      <NavLink title="Sign Up" to={RouteMap.SIGNUP} />
    </>
  );
}
