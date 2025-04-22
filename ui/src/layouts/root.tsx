import { landingLinks } from "@/components/landing-links";
import { NexusAIFooter } from "@/components/nexus-footer";
import { NexusAILandingHeader } from "@/components/nexus-landing-header";
import { RouteMap } from "@/components/route-map";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useEffect } from "react";
import { Outlet, useLocation, useNavigate } from "react-router";

export default function RootLayout() {
  const { user } = useAuthProvider();
  const { pathname } = useLocation();
  const navigate = useNavigate();
  useEffect(() => {
    if (user && pathname === RouteMap.HOME) {
      navigate(RouteMap.DASHBOARD_HOME);
    }
  }, [user, pathname]);
  return (
    <>
      <div className=" flex min-h-screen flex-col">
        <div className="items-center sticky top-0 z-50 w-full bg-background shadow-sm border-b">
          <NexusAILandingHeader leftLinks={landingLinks} />
        </div>
        <main className="flex-grow items-center justify-center">
          <Outlet />
        </main>
        <NexusAIFooter />
      </div>
    </>
  );
}
