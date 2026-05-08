import { landingLinks } from "@/components/links";
import { PlaygroundLandingHeader } from "@/components/playground-landing-header";
import { PlaygroundMinimalFooter } from "@/components/playground-minimal-footer";
import { Outlet } from "@tanstack/react-router";

export default function RootLayout() {
  return (
    <>
      <div className="relative flex min-h-screen flex-col">
        <div className="px-4 md:px-6 lg:px-8 py-2 items-center sticky top-0 z-50 w-full bg-background shadow-sm border-b">
          <div className="max-w-[1400px] mx-auto">
            <PlaygroundLandingHeader leftLinks={landingLinks} />
          </div>
        </div>
        <main className="flex-grow items-center justify-center">
          <Outlet />
        </main>
        <PlaygroundMinimalFooter />
      </div>
    </>
  );
}
