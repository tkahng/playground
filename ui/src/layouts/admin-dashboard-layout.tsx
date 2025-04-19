import { DashboardSidebar } from "@/components/dashboard-sidebar";
import { LinkDto } from "@/components/landing-links";
import { MainNav } from "@/components/main-nav";
import { NexusAILandingHeader } from "@/components/nexus-landing-header";
import { NexusAIMinimalFooter } from "@/components/nexus-minimal-footer";
import { RouteMap } from "@/components/route-map";
import { ChevronLeft } from "lucide-react";
import { Link, Outlet } from "react-router";

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
        <NexusAILandingHeader full leftLinks={headerLinks} />
        <MainNav links={links} />
        <main className="flex flex-grow">
          <DashboardSidebar
            links={links}
            backLink={
              <>
                <Link
                  to={RouteMap.DASHBOARD_HOME}
                  className="flex items-center gap-2 text-sm text-muted-foreground"
                >
                  <ChevronLeft className="h-4 w-4" />
                  Back to Dashboard
                </Link>
              </>
            }
          />
          <div className="mx-auto w-full max-w-[1200px] py-12 px-4 @lg:px-6 @xl:px-12 @2xl:px-20 @3xl:px-24">
            <Outlet />
          </div>
        </main>
        <NexusAIMinimalFooter />
      </div>
    </>
  );
}
