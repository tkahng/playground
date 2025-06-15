import { LinkDto, teamLinks } from "@/components/links";
import { MainNav } from "@/components/main-nav";
import { NexusAIMinimalFooter } from "@/components/nexus-minimal-footer";
import { RouteMap } from "@/components/route-map";
import { TeamHeader } from "@/components/team-header";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useTeam } from "@/hooks/use-team";
import { Outlet } from "react-router";

export default function TeamDashboardLayout() {
  const { user } = useAuthProvider();
  const { team, error, isLoading } = useTeam();
  // const { pathname } = useLocation();
  const isAdmin = user?.roles?.includes("superuser");
  // const isAdminPath = pathname.startsWith(RouteMap.ADMIN);
  const admin: LinkDto[] = isAdmin
    ? [
        {
          to: RouteMap.ADMIN,
          title: "Admin",
          current: () => false,
        },
      ]
    : [];
  // const dashboard = !isAdminPath
  const links = [
    { to: RouteMap.DASHBOARD, title: "Dashboard", current: () => true },
    ...admin,
  ] as LinkDto[];
  // if (!isAdminPath) {
  //   links.push({ to: RouteMap.DASHBOARD, title: "Dashboard" });
  // }
  if (error) {
    return <div>Error loading team: {error.message}</div>;
  }
  if (isLoading) {
    return <div>Loading team...</div>;
  }
  if (!team) {
    return <div>No team found.</div>;
  }
  if (!user) {
    return <div>No user found.</div>;
  }
  return (
    <div className="min-h-screen flex flex-col">
      <div className="px-4 md:px-6 lg:px-8 py-2 items-center sticky top-0 z-50 w-full bg-background shadow-sm border-b">
        <TeamHeader rightLinks={links} />
        <MainNav links={teamLinks(team.slug)} />
      </div>
      <main className="flex-1">
        <Outlet />
      </main>
      <NexusAIMinimalFooter />
    </div>
  );
}
