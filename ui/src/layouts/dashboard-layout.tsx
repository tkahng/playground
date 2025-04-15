import { DashboardSidebar } from "@/components/dashboard-sidebar";
import { LinkProps } from "@/components/landing-links";
import { RouteMap } from "@/components/route-map";
import { Home } from "lucide-react";
import { useOutlet } from "react-router";
const links: LinkProps[] = [
  {
    title: "Home",
    to: RouteMap.DASHBOARD_HOME,
    icon: <Home />,
  },
  {
    title: "Basic Route",
    to: RouteMap.PROTECTED_BASIC,
  },
  {
    title: "Pro Route",
    to: RouteMap.PROTECTED_PRO,
  },
  {
    title: "Advanced Route",
    to: RouteMap.PROTECTED_ADVANCED,
  },
];

export default function DashboardLayout() {
  const outlet = useOutlet();

  return (
    <>
      <DashboardSidebar links={links} />
      {outlet}
    </>
  );
}
