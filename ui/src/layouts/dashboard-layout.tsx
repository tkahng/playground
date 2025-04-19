import { DashboardSidebar } from "@/components/dashboard-sidebar";
import { LinkDto } from "@/components/landing-links";
import { MainNav } from "@/components/main-nav";
import { NexusAILandingHeader } from "@/components/nexus-landing-header";
import { NexusAIMinimalFooter } from "@/components/nexus-minimal-footer";
import { JSX } from "react";
import { Outlet } from "react-router";

export default function DashboardLayout({
  links,
  headerLinks,
  backLink,
}: {
  links: LinkDto[];
  headerLinks?: LinkDto[];
  backLink?: JSX.Element;
}) {
  return (
    <>
      <div className="relative flex min-h-screen flex-col justify-center">
        <NexusAILandingHeader full />
        <div className="flex items-center justify-between px-6 py-4 lg:px-8 lg:py-4 border-b">
          <MainNav links={headerLinks ?? []} />
        </div>
        <main className="flex flex-grow">
          <DashboardSidebar links={links} backLink={backLink} />
          <div className="mx-auto w-full max-w-[1200px] py-12 px-4 @lg:px-6 @xl:px-12 @2xl:px-20 @3xl:px-24">
            <Outlet />
          </div>
        </main>
        <NexusAIMinimalFooter />
      </div>
    </>
  );
}
