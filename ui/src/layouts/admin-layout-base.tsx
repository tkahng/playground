import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useEffect } from "react";
import { Navigate, Outlet, useLocation } from "react-router";
import { toast } from "sonner";

export default function AdminLayoutBase() {
  const location = useLocation();
  const { checkAuth, user } = useAuthProvider();
  useEffect(() => {
    checkAuth();
  }, [location]);

  if (!user || !user.permissions?.includes("superuser")) {
    toast.error("Unauthorized", {
      description: "You are not an admin",
      action: {
        label: "Close",
        onClick: () => console.log("Close"),
      },
    });
    return <Navigate to="/signin" />;
  }
  return (
    <>
      {/* <div className="relative flex min-h-screen flex-col justify-center">
        <NexusAILandingHeader full />
        <main className="flex flex-grow"> */}
      {/* <DashboardSidebar links={links} /> */}
      <Outlet />
      {/* </main>
        <NexusAIMinimalFooter />
      </div> */}
    </>
  );
}
