import { LinkDto } from "@/components/links";
import { MainNav } from "@/components/main-nav";
import { PlaygroundLandingHeader } from "@/components/nexus-landing-header";
import { PlaygroundMinimalFooter } from "@/components/nexus-minimal-footer";
import { RouteMap } from "@/components/route-map";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { Outlet } from "react-router";

export default function AdminLayout({
  headerLinks,
}: {
  leftLinks?: LinkDto[];
  rightLinks?: LinkDto[];
  headerLinks?: LinkDto[];
}) {
  const { user } = useAuthProvider();
  // const { pathname } = useLocation();
  const isAdmin = user?.roles?.includes("superuser");
  // const isAdminPath = pathname.startsWith(RouteMap.ADMIN);
  const admin: LinkDto[] = isAdmin
    ? [
        {
          to: RouteMap.ADMIN,
          title: "Admin",
          current: () => true,
        },
      ]
    : [];
  // const dashboard = !isAdminPath
  const links = [
    { to: RouteMap.DASHBOARD, title: "Dashboard", current: () => false },
    ...admin,
  ] as LinkDto[];
  // if (!isAdminPath) {
  //   links.push({ to: RouteMap.DASHBOARD, title: "Dashboard" });
  // }
  return (
    <div className="min-h-screen flex flex-col">
      <div className="px-4 md:px-6 lg:px-8 py-2 items-center sticky top-0 z-50 w-full bg-background shadow-sm border-b">
        <PlaygroundLandingHeader rightLinks={links} />
        {headerLinks && headerLinks.length > 0 && (
          <MainNav links={headerLinks} />
        )}
      </div>
      <main className="flex-1">
        <Outlet />
      </main>
      <PlaygroundMinimalFooter />
    </div>
  );
}
