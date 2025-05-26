import { landingLinks, LinkDto } from "@/components/links";
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
  // const { pathname } = useLocation();
  const isAdmin = user?.roles?.includes("superuser");
  // const isAdminPath = pathname.startsWith(RouteMap.ADMIN);
  const admin = isAdmin
    ? [{ to: RouteMap.ADMIN, title: "Admin", current: () => false }]
    : [];
  const isUser = user
    ? [{ to: RouteMap.DASHBOARD, title: "Dashboard", current: () => false }]
    : [];
  // const dashboard = !isAdminPath
  const links = [...isUser, ...admin] as LinkDto[];
  useEffect(() => {
    if (user && pathname === RouteMap.HOME) {
      navigate(RouteMap.DASHBOARD);
    }
  }, [user, pathname, navigate]);
  return (
    <>
      <div className="relative flex min-h-screen flex-col">
        <div className="px-4 md:px-6 lg:px-8 py-2 items-center sticky top-0 z-50 w-full bg-background shadow-sm border-b">
          <div className="max-w-[1400px] mx-auto">
            <NexusAILandingHeader leftLinks={landingLinks} rightLinks={links} />
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
