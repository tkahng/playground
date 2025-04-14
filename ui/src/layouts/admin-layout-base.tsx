import { NexusAILandingHeader } from "@/components/nexus-landing-header";
import { NexusAIMinimalFooter } from "@/components/nexus-minimal-footer";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useEffect } from "react";
import { Navigate, useLocation, useOutlet } from "react-router";
import { toast } from "sonner";
// const links: LinkProps[] = [
//   {
//     title: "Home",
//     to: RouteMap.ADMIN_DASHBOARD_HOME,
//     icon: <Home />,
//   },
//   {
//     title: "Users",
//     to: RouteMap.ADMIN_DASHBOARD_USERS,
//     icon: <User />,
//   },
//   {
//     title: "Roles",
//     to: RouteMap.ADMIN_DASHBOARD_ROLES,
//     icon: <Key />,
//   },
// ];

export default function AdminLayoutBase() {
  const location = useLocation();
  const { checkAuth, user } = useAuthProvider();
  const outlet = useOutlet();
  useEffect(() => {
    checkAuth();
  }, [location]);

  if (!user || !user.roles?.includes("superuser")) {
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
      <div className="relative flex min-h-screen flex-col justify-center">
        <NexusAILandingHeader full />
        <main className="flex flex-grow">
          {/* <DashboardSidebar links={links} /> */}
          {outlet}
        </main>
        <NexusAIMinimalFooter />
      </div>
    </>
  );
}
