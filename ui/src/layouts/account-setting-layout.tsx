import { DashboardSidebar } from "@/components/dashboard-sidebar";
import { LinkProps } from "@/components/landing-links";
import { RouteMap } from "@/components/route-map";
import { useOutlet } from "react-router";
const links: LinkProps[] = [
  {
    title: "Account",
    to: RouteMap.ACCOUNT_SETTINGS,
  },
  {
    title: "Billing",
    to: RouteMap.BILLING_SETTINGS,
  },
];
export default function AccountSettingsLayout() {
  const outlet = useOutlet();

  return (
    <div className="flex-row flex grow">
      <DashboardSidebar links={links} />
      {outlet}
    </div>
  );
}
