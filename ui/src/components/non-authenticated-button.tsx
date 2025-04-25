import { NavbarLink } from "@/components/link/nav-link";
import { RouteMap } from "@/components/route-map";

export default function NonAuthenticatedButton() {
  return (
    <>
      <NavbarLink title="Sign In" to={RouteMap.SIGNIN} />
      <NavbarLink title="Sign Up" to={RouteMap.SIGNUP} />
    </>
  );
}
