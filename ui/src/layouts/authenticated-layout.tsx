import { DashboardSidebar } from "@/components/dashboard-sidebar";
import { NexusAILandingHeader } from "@/components/header";
import { landingLinks } from "@/components/landing-links";
import { NexusAIFooter } from "@/components/nexus-footer";
import { Outlet } from "react-router";

export default function AuthenticatedLayout() {
  return (
    <>
      <div className=" flex min-h-screen flex-col">
        <NexusAILandingHeader leftLinks={landingLinks} />
        <main className="flex-row flex flex-grow">
          <DashboardSidebar />
          <Outlet />
        </main>
        <NexusAIFooter />
      </div>
    </>
  );
}
