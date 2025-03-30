import { NexusAILandingHeader } from "@/components/header";
import { landingLinks } from "@/components/landing-links";
import { NexusAIFooter } from "@/components/nexus-footer";
import { Outlet } from "react-router";

export default function RootLayout() {
  return (
    <>
      <div className="relative flex min-h-screen flex-col justify-center">
        <NexusAILandingHeader leftLinks={landingLinks} />
        <main className="relative flex flex-col items-center justify-center">
          <Outlet />
        </main>
        <NexusAIFooter />
      </div>
    </>
  );
}
