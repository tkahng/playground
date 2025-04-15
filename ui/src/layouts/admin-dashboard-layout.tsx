import { DashboardSidebar } from "@/components/dashboard-sidebar";
import { LinkProps } from "@/components/landing-links";
import { NexusAILandingHeader } from "@/components/nexus-landing-header";
import { NexusAIMinimalFooter } from "@/components/nexus-minimal-footer";
import { RouteMap } from "@/components/route-map";
import { Home, Key, User } from "lucide-react";
import { useOutlet } from "react-router";
const links: LinkProps[] = [
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

export default function AdminDashboardLayout() {
  const outlet = useOutlet();

  return (
    <>
      <div className="relative flex min-h-screen flex-col justify-center">
        <NexusAILandingHeader full />
        <main className="flex flex-grow">
          <DashboardSidebar links={links} />
          {outlet}
        </main>
        <NexusAIMinimalFooter />
      </div>
    </>
  );
}
