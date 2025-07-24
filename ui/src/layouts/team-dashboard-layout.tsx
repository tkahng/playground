import { LinkDto, teamLinks } from "@/components/links";
import { MainNav } from "@/components/main-nav";
import { PlaygroundMinimalFooter } from "@/components/playground-minimal-footer";
import { RouteMap } from "@/components/route-map";
import { TeamHeader } from "@/components/team-header";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useTeam } from "@/hooks/use-team";
import { getTeamBySlug } from "@/lib/api";
import { GetError } from "@/lib/get-error";
import { useQuery } from "@tanstack/react-query";
import { useEffect, useRef } from "react";
import { Navigate, Outlet, useLocation, useParams } from "react-router";
import { toast } from "sonner";

export default function TeamDashboardLayout() {
  const { user } = useAuthProvider();
  const { teamSlug } = useParams<{ teamSlug: string }>();
  const { setTeam, team, teamMember } = useTeam();
  const { pathname } = useLocation();
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

      setTeam({ ...response.team, member: response.member });
      return response;
    },
    enabled: false,
  });
  const isMounted = useRef(false);
  useEffect(() => {
    if (!isMounted.current) {
      isMounted.current = true;
      refetch().then(() => {});
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [teamSlug]);
  const isAdmin = user?.roles?.includes("superuser");
  const admin: LinkDto[] = isAdmin
    ? [
        {
          to: RouteMap.ADMIN,
          title: "Admin",
          current: () => false,
        },
      ]
    : [];
  const links = [
    { to: RouteMap.DASHBOARD, title: "Dashboard", current: () => false },
    ...admin,
  ] as LinkDto[];

  const isNotUserTeam = teamMember?.user_id !== user?.user.id;

  if (error) {
    const err = GetError(error);
    toast.error("Error loading team: " + err?.detail, {
      description: "Please try again",
      action: {
        label: "Undo",
        onClick: () => console.log("Undo"),
      },
    });
    return (
      <Navigate
        to={{
          pathname: "/teams",
        }}
      />
    );
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
  if (isNotUserTeam) {
    return (
      <Navigate
        to={{
          pathname: "/teams",
        }}
      />
    );
  }
  if (pathname === "/teams/settings/billing") {
    return (
      <Navigate
        to={{
          pathname: `/teams/${team.slug}/settings/billing`,
        }}
      />
    );
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
      <PlaygroundMinimalFooter />
    </div>
  );
}
