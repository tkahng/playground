import { landingLinks } from "@/components/links";
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
      navigate(RouteMap.DASHBOARD);
    }
  }, [user, pathname]);
  return (
    <>
      <div className="relative flex min-h-screen flex-col">
        <div className="px-4 md:px-6 lg:px-8 py-2 items-center sticky top-0 z-50 w-full bg-background shadow-sm border-b">
          <div className="max-w-[1400px] mx-auto">
            <NexusAILandingHeader leftLinks={landingLinks} />
          </div>
        </div>
        <main className="flex-grow items-center justify-center">
          <Outlet />
        </main>
        <NexusAIFooter />
      </div>
    </>
  );
}
