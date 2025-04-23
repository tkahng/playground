import { DashboardSidebar } from "@/components/dashboard-sidebar";
import { LinkDto, RouteLinks } from "@/components/links";
import { MainNav } from "@/components/main-nav";
import { NexusAILandingHeader } from "@/components/nexus-landing-header";
import { NexusAIMinimalFooter } from "@/components/nexus-minimal-footer";
import { Outlet } from "react-router";

export default function AdminDashboardLayout({
  links,
  headerLinks,
}: {
  links: LinkDto[];
  headerLinks: LinkDto[];
}) {
  return (
    <>
      <div className="relative flex min-h-screen flex-col justify-center">
        <NexusAILandingHeader leftLinks={headerLinks} />
        <div className="flex items-center sticky top-0 z-50 w-full">
          {headerLinks && headerLinks.length > 0 && (
            <MainNav links={headerLinks} />
          )}
        </div>
        <main className="flex flex-grow">
          <DashboardSidebar
            links={links}
            backLink={RouteLinks.DASHBOARD_HOME}
          />
          {/* <div className="mx-auto w-full max-w-[1200px] py-12 px-4 @lg:px-6 @xl:px-12 @2xl:px-20 @3xl:px-24"> */}
          <Outlet />
          {/* </div> */}
        </main>
        <NexusAIMinimalFooter />
      </div>
    </>
  );
}
