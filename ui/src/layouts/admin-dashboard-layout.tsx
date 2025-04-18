import { DashboardSidebar } from "@/components/dashboard-sidebar";
import { LinkDto } from "@/components/landing-links";
import { NexusAILandingHeader } from "@/components/nexus-landing-header";
import { NexusAIMinimalFooter } from "@/components/nexus-minimal-footer";
import { RouteMap } from "@/components/route-map";
import { Home, Key, User } from "lucide-react";
import { useOutlet } from "react-router";
const links: LinkDto[] = [
  {
    title: "Home",
    to: RouteMap.ADMIN_DASHBOARD_HOME,
    icon: <Home />,
  },
  {
    title: "Users",
    to: RouteMap.ADMIN_DASHBOARD_USERS,
    icon: <User />,
  },
  {
    title: "Roles",
    to: RouteMap.ADMIN_DASHBOARD_ROLES,
    icon: <Key />,
  },
];
const headerLinks: LinkDto[] = [
  { title: "Dashboard", to: RouteMap.DASHBOARD_HOME },
];
export default function AdminDashboardLayout() {
  const outlet = useOutlet();

  return (
    <>
      <div className="relative flex min-h-screen flex-col justify-center">
        <NexusAILandingHeader full leftLinks={headerLinks} />
        <main className="flex flex-grow">
          <DashboardSidebar links={links} />
          <div className="mx-auto w-full max-w-[1200px] py-12 px-4 @lg:px-6 @xl:px-12 @2xl:px-20 @3xl:px-24">
            {outlet}
          </div>
        </main>
        <NexusAIMinimalFooter />
      </div>
    </>
  );
}
