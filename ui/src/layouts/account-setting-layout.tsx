import BackLink from "@/components/back-link";
import { DashboardSidebar } from "@/components/dashboard-sidebar";
import {
  authenticatedSubHeaderLinks,
  LinkDto,
} from "@/components/landing-links";
import { MainNav } from "@/components/main-nav";
import { NexusAILandingHeader } from "@/components/nexus-landing-header";
import { NexusAIMinimalFooter } from "@/components/nexus-minimal-footer";
import { RouteMap } from "@/components/route-map";
import { Outlet } from "react-router";
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
  return (
    <>
      <div className="relative flex min-h-screen flex-col justify-center">
        <NexusAILandingHeader full leftLinks={headerLinks} />
        <div className="flex items-center justify-between px-6 py-4 lg:px-8 lg:py-4 border-b">
          <MainNav links={authenticatedSubHeaderLinks ?? []} />
        </div>
        <main className="flex flex-grow">
          <DashboardSidebar
            links={links}
            backLink={
              <BackLink to={RouteMap.DASHBOARD_HOME} name="Dashboard" />
            }
          />
          <Outlet />
        </main>
        <NexusAIMinimalFooter />
      </div>
    </>
  );
}
