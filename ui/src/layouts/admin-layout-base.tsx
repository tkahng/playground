import { useAuthProvider } from "@/hooks/use-auth-provider";
import { Navigate, Outlet, useRouterState } from "@tanstack/react-router";
import { toast } from "sonner";

export default function AdminLayoutBase() {
  const location = useRouterState({ select: (s) => s.location });
  const { user } = useAuthProvider();

  if (!user || !user.permissions?.includes("superuser")) {
    toast.error("Unauthorized", {
      description: "You are not an admin",
      action: {
        label: "Close",
        onClick: () => console.log("Close"),
      },
    });
    const redirectTo = location.pathname + (location.searchStr || "");
    return (
      <Navigate to="/signin" search={{ redirect_to: redirectTo }} />
    );
  }
  return <Outlet />;
}
