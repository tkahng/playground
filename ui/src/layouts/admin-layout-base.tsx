import { AdminDashboardSidebar } from "@/components/admin-dashboard-sidebar";
import { NexusAILandingHeader } from "@/components/header";
import { adminLinks } from "@/components/landing-links";
import { NexusAIFooter } from "@/components/nexus-footer";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useEffect } from "react";
import { Navigate, useLocation, useOutlet } from "react-router";

export default function AdminLayoutBase() {
  const location = useLocation();
  const { checkAuth, user } = useAuthProvider();
  const outlet = useOutlet();
  useEffect(() => {
    checkAuth();
  }, [location]);

  if (!user || !user.roles?.includes("superuser")) {
    return <Navigate to="/signin" />;
  }
  return (
    <>
      <div className="relative flex min-h-screen flex-col justify-center">
        <NexusAILandingHeader leftLinks={adminLinks} />
        <main className="flex flex-grow">
          <AdminDashboardSidebar />
          {outlet}
        </main>
        <NexusAIFooter />
      </div>
    </>
  );
}
