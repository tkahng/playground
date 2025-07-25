import { useAuthProvider } from "@/hooks/use-auth-provider";
import {
  createSearchParams,
  Navigate,
  Outlet,
  useLocation,
} from "react-router";
import { toast } from "sonner";

export default function AdminLayoutBase() {
  const location = useLocation();
  const { user } = useAuthProvider();

  if (!user || !user.permissions?.includes("superuser")) {
    toast.error("Unauthorized", {
      description: "You are not an admin",
      action: {
        label: "Close",
        onClick: () => console.log("Close"),
      },
    });
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
