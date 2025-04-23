import { DashboardSidebar } from "@/components/dashboard-sidebar";
import { LinkDto } from "@/components/landing-links";
import { MainNav } from "@/components/main-nav";
import { NexusAILandingHeader } from "@/components/nexus-landing-header";
import { NexusAIMinimalFooter } from "@/components/nexus-minimal-footer";
import { Outlet } from "react-router";

export default function DashboardLayout({
  sidebarLinks: links,
  headerLinks,
  sidebarBackLink: backLink,
}: {
  sidebarLinks?: LinkDto[];
  headerLinks?: LinkDto[];
  sidebarBackLink?: LinkDto;
}) {
  return (
    <>
      <div className="relative flex min-h-screen flex-col">
        <div className="px-4 md:px-6 lg:px-8 py-2 items-center sticky top-0 z-50 w-full bg-background shadow-sm border-b">
          <NexusAILandingHeader />
          {headerLinks && headerLinks.length > 0 && (
            <MainNav links={headerLinks} />
          )}
        </div>
        <main className="flex flex-1">
          {links && links.length > 0 && (
            <DashboardSidebar links={links} backLink={backLink} />
          )}
          <Outlet />
        </main>
        <NexusAIMinimalFooter />
      </div>
    </>
  );
}
