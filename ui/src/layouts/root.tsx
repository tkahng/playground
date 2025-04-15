import { landingLinks } from "@/components/landing-links";
import { NexusAIFooter } from "@/components/nexus-footer";
import { NexusAILandingHeader } from "@/components/nexus-landing-header";
import { Outlet } from "react-router";

export default function RootLayout() {
  return (
    <>
      <div className=" flex min-h-screen flex-col">
        <NexusAILandingHeader leftLinks={landingLinks} />
        <main className="flex-grow items-center justify-center">
          <Outlet />
        </main>
        <NexusAIFooter />
      </div>
    </>
  );
}
