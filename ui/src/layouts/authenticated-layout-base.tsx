import { useAuthProvider } from "@/hooks/use-auth-provider";
import {
  createSearchParams,
  Navigate,
  Outlet,
  useLocation,
} from "react-router";

export default function AuthenticatedLayoutBase() {
  const location = useLocation();
  const { user } = useAuthProvider();

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
