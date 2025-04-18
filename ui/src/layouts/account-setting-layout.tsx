import { DashboardSidebar } from "@/components/dashboard-sidebar";
import { LinkDto } from "@/components/landing-links";
import { NexusAILandingHeader } from "@/components/nexus-landing-header";
import { NexusAIMinimalFooter } from "@/components/nexus-minimal-footer";
import { RouteMap } from "@/components/route-map";
import { useOutlet } from "react-router";
const links: LinkDto[] = [
  {
    title: "Account",
    to: RouteMap.ACCOUNT_SETTINGS,
  },
  {
    title: "Billing",
    to: RouteMap.BILLING_SETTINGS,
  },
];

const headerLinks: LinkDto[] = [
  { title: "Dashboard", to: RouteMap.DASHBOARD_HOME },
];
export default function AccountSettingsLayout() {
  const outlet = useOutlet();

  return (
    <>
      <div className="relative flex min-h-screen flex-col justify-center">
        <NexusAILandingHeader full leftLinks={headerLinks} />
        <main className="flex flex-grow">
          <DashboardSidebar links={links} />
          {outlet}
        </main>
        <NexusAIMinimalFooter />
      </div>
    </>
  );
}
