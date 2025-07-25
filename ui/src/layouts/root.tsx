import { landingLinks, LinkDto } from "@/components/links";
import { PlaygroundLandingHeader } from "@/components/playground-landing-header";
import { PlaygroundMinimalFooter } from "@/components/playground-minimal-footer";
import { RouteMap } from "@/components/route-map";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { Outlet } from "react-router";

export default function RootLayout() {
  const { user } = useAuthProvider();
  // const { team, teamMember } = useTeam();
  // const { pathname } = useLocation();
  const isAdmin = user?.roles?.includes("superuser");
  // const isAdminPath = pathname.startsWith(RouteMap.ADMIN);
  const admin = isAdmin
    ? [{ to: RouteMap.ADMIN, title: "Admin", current: () => false }]
    : [];
  const isUser = user
    ? [
        {
          to: RouteMap.DASHBOARD,
          title: "Dashboard",
          current: () => false,
        },
      ]
    : [];
  const links = [...isUser, ...admin] as LinkDto[];
  // if (user) {
  //   // if (pathname === "/") {
  //   //   return <Navigate to="/account" />;
  //   //   // if (team && teamMember?.user_id === user.user.id) {
  //   //   //   return <Navigate to={`/teams/${team.slug}/dashboard`} />;
  //   //   // } else {
  //   //   //   return <Navigate to="/teams" />;
  //   //   // }
  //   // }
  //   // if (pathname === "/dashboard") {
  //   //   if (team && teamMember?.user_id === user.user.id) {
  //   //     return <Navigate to={`/teams/${team.slug}/dashboard`} />;
  //   //   }
  //   // }
  // }
  return (
    <>
      <div className="relative flex min-h-screen flex-col">
        <div className="px-4 md:px-6 lg:px-8 py-2 items-center sticky top-0 z-50 w-full bg-background shadow-sm border-b">
          <div className="max-w-[1400px] mx-auto">
            <PlaygroundLandingHeader
              leftLinks={landingLinks}
              rightLinks={links}
            />
          </div>
        </div>
        <main className="flex-grow items-center justify-center">
          <Outlet />
        </main>
        <PlaygroundMinimalFooter />
      </div>
    </>
  );
}
