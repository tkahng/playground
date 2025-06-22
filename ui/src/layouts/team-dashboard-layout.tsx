import { LinkDto, teamLinks } from "@/components/links";
import { MainNav } from "@/components/main-nav";
import { NexusAIMinimalFooter } from "@/components/nexus-minimal-footer";
import { RouteMap } from "@/components/route-map";
import { TeamHeader } from "@/components/team-header";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useTeamContext } from "@/hooks/use-team-context";
import { getTeamBySlug } from "@/lib/queries";
import { useQuery } from "@tanstack/react-query";
import { useEffect, useRef } from "react";
import { Outlet, useParams } from "react-router";

export default function TeamDashboardLayout() {
  // const { user } = useAuthProvider();
  // const { team, error, isLoading } = useTeam();
  const { user } = useAuthProvider();
  const { teamSlug } = useParams<{ teamSlug: string }>();
  const { setTeam, team } = useTeamContext();
  const { isLoading, error, refetch } = useQuery({
    queryKey: ["team-by-slug-layout"],
    queryFn: async () => {
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token");
      }
      if (!teamSlug) {
        throw new Error("Team slug is required");
      }
      const response = await getTeamBySlug(user.tokens.access_token, teamSlug);
      // if (!team) {
      setTeam(response.team);
      // }
      return response;
    },
    enabled: !!user?.tokens.access_token && !!teamSlug,
  });
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
  const isMounted = useRef(false);
  useEffect(() => {
    if (!isMounted.current) {
      isMounted.current = true;
      // if (teamSlug && user?.tokens.access_token) {
      refetch().then(() => {
        isMounted.current = false;
      });
      // }
    }
  }, [refetch, teamSlug]);
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
