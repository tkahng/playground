import { LinkDto } from "@/components/links";
import { MainNav } from "@/components/main-nav";
import { NexusAILandingHeader } from "@/components/nexus-landing-header";
import { NexusAIMinimalFooter } from "@/components/nexus-minimal-footer";
import { Outlet } from "react-router";

export default function DashboardLayout({
  headerLinks,
}: {
  headerLinks?: LinkDto[];
}) {
  return (
    <div className="min-h-screen flex flex-col">
      <div className="px-4 md:px-6 lg:px-8 py-2 items-center sticky top-0 z-50 w-full bg-background shadow-sm border-b">
        <NexusAILandingHeader />
        {headerLinks && headerLinks.length > 0 && (
          <MainNav links={headerLinks} />
        )}
      </div>
      <main className="flex-1">
        <Outlet />
      </main>
      <NexusAIMinimalFooter />
    </div>
  );
}
