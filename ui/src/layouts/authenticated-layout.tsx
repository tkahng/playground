import { DashboardSidebar } from "@/components/dashboard-sidebar";
import { landingLinks, LinkProps } from "@/components/landing-links";
import { NexusAIFooter } from "@/components/nexus-footer";
import { NexusAILandingHeader } from "@/components/nexus-landing-header";
import { RouteMap } from "@/components/route-map";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { Home } from "lucide-react";
import { useEffect } from "react";
import { Navigate, useLocation, useOutlet } from "react-router";

const links: LinkProps[] = [
  {
    title: "Home",
    to: RouteMap.DASHBOARD_HOME,
    icon: <Home />,
  },
];
export default function AuthenticatedLayout() {
  const location = useLocation();
  const { checkAuth, user } = useAuthProvider();
  const outlet = useOutlet();
  useEffect(() => {
    checkAuth();
  }, [location]);

  if (!user) {
    return <Navigate to="/signin" />;
  }
  return (
    <>
      <div className="relative flex min-h-screen flex-col justify-center">
        <NexusAILandingHeader leftLinks={landingLinks} />
        <main className="flex flex-grow">
          <DashboardSidebar links={links} />
          {outlet}
        </main>
        <NexusAIFooter />
      </div>
    </>
  );
}
